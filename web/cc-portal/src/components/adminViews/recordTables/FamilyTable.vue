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
          <b-button class="mr-2 mb-2" variant="info" :disabled="disableSingleAction" @click="onViewFamily">View Family</b-button>
          <b-button v-if="false" class="mr-2 mb-2" variant="primary" :disabled="disableManyAction || disableCheckout" @click="onCheckOutEvents">CheckOut</b-button>
          <b-button class="mr-2 mb-2" variant="danger" :disabled="disableManyAction" @click="onDeleteEvents">Delete</b-button>
          <b-button class="mr-2 mb-2" variant="outline-secondary" :disabled="disableManyAction" @click="onClearSelect">Clear Select</b-button>
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
import config from "../../../config";
const queryString = require("query-string");
const moment = require("moment");
const DATE_FORMAT_DISP = "MM/DD/YYYY";
// const DATE_FORMAT = "YYYY-MM-DD";
const TIME_FORMAT = "hh:mm A";
export default {
  name: "family-table",
  props: { ccRecords: Array },
  data() {
    return {
      institution: null,
      loggedInToken: "",
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
    };
  },
  computed: {
    fields() {
      if (this.institution != null && this.institution.workflow_type === "cc") {
        return [
          "date",
          "student_name",
          "group",
          "guardian_name",
          "temperature",
          "phone_number",
          "drop_off",
          "scheduled_pickup",
          "actual_pickup",
        ];
      }
      return [
        "time",
        "student_name",
        "group",
        "guardian_name",
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
    this.loggedInToken = this.$store.state.loggedInToken;
  },
  methods: {
    mapCCRecordToItem(ccRecord) {
      if (ccRecord.gw === null) {
        return {
          _id: ccRecord._id,
          self: ccRecord,
          date: "",
          time: "",
          has_expired: "",
          student_name: "",
          group: "",
          guardian_name: "",
          temperature: "",
          phone_number: "",
          drop_off: "",
          scheduled_pickup: "",
          actual_pickup: "",
        };
      }
      var checkInTime = moment(ccRecord.gw.check_in_event.time);
      var scheduleTime = moment(ccRecord.check_out_scheduled_at);
      var checkOutTime = moment(ccRecord.gw.check_out_event.time);
      var checkInDisplay = checkInTime.isBefore("1970-01-01")
        ? ""
        : checkInTime.format(TIME_FORMAT);
      var scheduleDisplay = scheduleTime.isBefore("1970-01-01")
        ? ""
        : scheduleTime.format(TIME_FORMAT);
      var checkOutDisplay = checkOutTime.isBefore("1970-01-01")
        ? ""
        : checkOutTime.format(TIME_FORMAT);
      var row = {
        _id: ccRecord._id,
        self: ccRecord,
        date: checkInTime.format(DATE_FORMAT_DISP),
        time: checkInTime.format(DATE_FORMAT_DISP) + checkInDisplay,
        has_expired: ccRecord.has_expired,
        student_name: ccRecord.gw.ward_info.name,
        group: ccRecord.gw.ward_info.group,
        guardian_name: ccRecord.gw.check_in_event.guardian_info.name,
        temperature: ccRecord.temperature.toFixed(1),
        phone_number: ccRecord.gw.check_in_event.guardian_info.phone_num,
        drop_off: checkInDisplay,
        scheduled_pickup: scheduleDisplay,
        actual_pickup: checkOutDisplay,
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

    onViewFamily() {
      var _this = this;
      var wardID = this.selectedItems[0].self.gw.ward_info.id
      this.getFamilyWithMemberByWardID(wardID, (familyWithMembers) => {
        _this.$store.commit("setLoadedFamilyWithMembers", familyWithMembers)
        _this.$router.push("/portal/families/info");
      })
    },

    onCheckOutEventes() {},

    onDeleteEvents() {},

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
    getFamilyWithMemberByWardID(wardID, callback) {
      
      const queryParams = {wardID: wardID };
      const queryArgs = queryString.stringify(queryParams)

      const http = new XMLHttpRequest()
      const query = config.API_LOCATION + "family-with-members?" + queryArgs;
      console.log(`getFamilyWithMemberByAWardID - query: ${query}`)
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
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
</script>