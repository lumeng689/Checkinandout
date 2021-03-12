<template>
  <div>
    <b-row>
      <b-col xl="2" lg="4" class="my-1">
        <b-form-group label="Group" label-for="group-filter-input" label-cols-sm="5" label-align-sm="right" label-size="sm">
          <b-input-group>
            <b-form-input id="group-filter-input" v-model="groupFilter" type="search" placeholder="Type to Search"></b-form-input>
          </b-input-group>
        </b-form-group>
      </b-col>
      <b-col xl="3" lg="4" class="my-1">
        <b-form-group label="Temp Threshold" label-for="temp-threshold-filter-input" label-cols-sm="5" label-align-sm="right" label-size="sm" class="mb-0">
          <b-input-group>
            <b-input-group-prepend>
              <b-form-checkbox v-model="enableTempThrdFilter"></b-form-checkbox>
            </b-input-group-prepend>
            <b-form-input type="number" step="0.1" id="temp-threshold-filter-input" v-model="tempThrdFilter" placeholder="99.2"></b-form-input>
          </b-input-group>
        </b-form-group>
      </b-col>
      <b-col xl="2" lg="4" class="my-1">
        <b-form-group label="Include Expired" label-for="temp-threshold-filter-input" label-cols-sm="8" label-align-sm="right" label-size="sm" class="mb-0">
          <b-input-group>
            <b-form-checkbox id="include-expired-filter-input" v-model="includeExpiredFilter"></b-form-checkbox>
          </b-input-group>
        </b-form-group>
      </b-col>
      <b-col xl="5" lg="12" class="my-1" align-self="end">
        <b-row align-h="end">
          <!-- <b-button class="mr-2 mb-2" variant="success" @click="onEditTag">Edit Tag</b-button> -->
          <!-- <b-button class="mr-2 mb-2" variant="info" :disabled="disableSingleAction" @click="onViewFamily">View Family</b-button> -->
          <b-button v-if="false" class="mr-2 mb-2" variant="primary" :disabled="disableManyAction || disableCheckout" @click="onCheckOutEvents">CheckOut</b-button>
          <b-button v-if="false" class="mr-2 mb-2" variant="danger" :disabled="disableManyAction" @click="onDeleteEvents">Delete</b-button>
          <b-button v-if="false" class="mr-2 mb-2" variant="outline-secondary" :disabled="disableManyAction" @click="onClearSelect">Clear Select</b-button>
        </b-row>
      </b-col>
    </b-row>
    <b-table
      bordered
      sticky-header="80%"
      selectable
      id="cc-record-table"
      ref="ccRecordTable"
      responsive="sm"
      :items="items"
      :fields="fields"
      :current-page="currentPage"
      :per-page="perPage"
      :filter="filter"
      :filter-included-fields="filterOn"
      :filter-function="filterFunction"
      @row-selected="onRowSelected"
      @filtered="onFiltered"
    ></b-table>
    <b-pagination v-model="currentPage" :total-rows="numRows" :per-page="perPage" aria-controls="cc-record-table"></b-pagination>
  </div>
</template>
<script>
import config from "../../config";
const queryString = require("query-string");
const moment = require("moment");
const DATE_FORMAT_DISP = "MM/DD/YYYY";
// const DATE_FORMAT = "YYYY-MM-DD";
const TIME_FORMAT = "hh:mm A";
export default {
  name: "tag-table",
  // template: "tag-corporate-table-template",
  props: { ccRecords: Array },
  data() {
    var tModalItem = this.getNewTagModalItem();
    return {
      institution: null,
      currentPage: 1,
      perPage: 20,
      items: [],
      // filteredItems: [],
      selectedItems: [],
      groupFilter: "",
      tempThrdFilter: "99.2",
      enableTempThrdFilter: false,
      includeExpiredFilter: false,
      filterOn: ["group", "temperature"],
      tModalItem: tModalItem,
    };
  },
  computed: {
    fields() {
      // Assumes that only "Corporates" would use Tag mode
      if (this.institution != null && this.institution.workflow_type === "cc") {
        return [
          "date",
          "tag_number",
          "employee",
          "group",
          "temperature",
          "phone_number",
          "check_in",
          "check_out",
        ];
      }

      return [
        "time",
        "tag_number",
        "employee",
        "group",
        "temperature",
        "phone_number",
      ];
    },
    disableSingleAction() {
      return this.selectedItems.length != 1;
    },
    disableManyAction() {
      return this.selectedItems.length == 0;
    },
    disableCheckout() {
      return true;
    },
    numRows() {
      return this.items.length;
    },
    filter() {
      return this.groupFilter + " " + this.tempThrdFilter;
    },
  },
  created() {
    this.institution = this.$store.state.institution;
  },
  methods: {
    mapCCRecordToItem(ccRecord) {
      // map CCRecord to item
      if (ccRecord.mt === null) {
        return {
          _id: ccRecord._id,
          date: "",
          time: "",
          has_expired: "",
          tag_number: "",
          employee: "",
          group: "",
          temperature: "",
          phone_number: "",
          check_in: "",
          check_out: "",
        };
      }
      var checkInTime = moment(ccRecord.mt.check_in_event.time);
      var checkOutTime = moment(ccRecord.mt.check_out_event.time);
      var checkInDisplay = checkInTime.isBefore("1970-01-01")
        ? ""
        : checkInTime.format(TIME_FORMAT);
      var checkOutDisplay = checkOutTime.isBefore("1970-01-01")
        ? ""
        : checkOutTime.format(TIME_FORMAT);
      var row = {
        _id: ccRecord._id,
        date: checkInTime.format(DATE_FORMAT_DISP),
        time: checkInTime.format(DATE_FORMAT_DISP) + " " + checkInDisplay,
        has_expired: ccRecord.has_expired,
        tag_number: ccRecord.mt.info.id,
        employee: ccRecord.mt.info.name,
        group: ccRecord.mt.info.group,
        temperature: ccRecord.temperature.toFixed(1),
        phone_number: ccRecord.mt.info.phone_num,
        check_in: checkInDisplay,
        check_out: checkOutDisplay,
      };
      if (row.temperature > parseFloat(this.tempThrdFilter)) {
        row._rowVariant = 'danger'
      }
      return row;
    },
    onRowSelected(items) {
      // handle "undefined" error when clicking the same item multiple times
      if (items[0] === undefined) return;
      this.selectedItems = items;
    },
    // Button Handlers
    onEditTag() {
      var _this = this;

      this.getTagByTagStringFromDb((tag) => {
        _this.tModalItem = tag;
        // _this.$emit('sync-cc-records')
      });
    },
    onCheckOutEvents() {},
    onDeleteEvents() {
      // TODO
      // for (const item of this.selectedItems) {
      //   this.deleteCCRecordFromDb(item);
      // }
      // this.getCCRecordsFromDb(null);
    },
    onClearSelect() {
      this.selectedItems = [];
      this.$refs.ccRecordTable.clearSelected();
    },
    onFiltered(filteredItems) {
      this.totalRows = filteredItems.length;
      this.currentPage = 1;
      this.onClearSelect();
    },
    filterFunction(row) {
      // console.log(`row: ${JSON.stringify(row)}, filterString: ${filterString}`)
      var expiredPred = false;
      if (!this.includeExpiredFilter) {
        expiredPred = row.has_expired;
      }
      var tempPred = true;
      if (this.enableTempThrdFilter) {
        tempPred = false;
        if (this.tempThrdFilter.length > 0) {
          tempPred = row.temperature > parseFloat(this.tempThrdFilter);
        }
      }
      var groupPred = true;
      if (this.groupFilter.length > 0) {
        groupPred = row.group.toLowerCase().includes(this.groupFilter);
      }
      // console.log(
      //   `expiredPred: ${expiredPred}, tempPred: ${tempPred}, groupPred: ${groupPred}, `
      // );
      return !expiredPred && tempPred && groupPred;
    },
    getTagByTagStringFromDb(callback) {
      // var _this = this
      const item = this.selectedItems[0];
      const queryParams = { tagString: item.mt.info.id };
      const queryArgs = queryString.stringify(queryParams);

      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "tag?" + queryArgs;
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText)
          if (this.responseText.length == 0) {
            return;
          }
          var response = JSON.parse(this.responseText).data;
          if (response && callback != null) {
            callback(response);
          }
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        http.send();
      } catch (e) {
        alert(e);
      }
    },
    getNewTagModalItem() {
      return {
        tag_string: "",
        first_name: "",
        last_name: "",
        group: "",
        email: "",
      };
    },
  },
  watch: {
    ccRecords() {
      this.onClearSelect();
      if (this.ccRecords) {
        this.items = this.ccRecords.map((ccRecord) => {
          return this.mapCCRecordToItem(ccRecord);
        });
      }
    },
  },
};

// module.exports = {
//   TagCorporateTable
// }
</script>