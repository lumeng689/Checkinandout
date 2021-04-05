<template>
  <b-container class="h-100 ml-2 mr-2" fluid>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;">
      <h2>Dashboard</h2>
    </b-row>
    <b-row class="ml-2">
      <b-alert variant="success" :show="dismissCountDown" dismissible @dismissed="dismissCountDown=0" @dismiss-count-down="countDownChanged">
        # of Scan Records Today:
        <b>{{numTodayRecords}}</b>
      </b-alert>
    </b-row>
    <b-card class="h-100">
      <b-container class="h-100" fluid>
        <b-row>
          <b-col xl="8" lg="10">
            <b-form-group class="mb-0" label="Select A Date Range" label-for="sort-by-select" label-cols-sm="3" label-align-sm="right" label-size="sm">
              <b-input-group>
                <b-form-datepicker
                  class="ml-2"
                  id="datepicker-start"
                  placeholder="Choose a Start Date"
                  v-model="startDate"
                  :date-format-options="{ year: 'numeric', month: 'numeric', day: 'numeric' }"
                  locale="en"
                ></b-form-datepicker>
                <p class="ml-2">to</p>
                <b-form-datepicker
                  class="ml-2"
                  id="datepicker-end"
                  placeholder="Choose a End Date"
                  v-model="endDate"
                  :date-format-options="{ year: 'numeric', month: 'numeric', day: 'numeric' }"
                  locale="en"
                ></b-form-datepicker>
                <b-button class="ml-2" variant="primary" :disabled="!hasDateRanges" @click="onApplyDateRangeFilter">Filter</b-button>
              </b-input-group>
            </b-form-group>
          </b-col>
        </b-row>
        <family-table v-if="useFamilyTable" :ccRecords="ccRecords"></family-table>
        <member-table v-if="useMemberTable" :ccRecords="ccRecords" :instType="instType"></member-table>
        <tag-table v-if="useTagTable" :ccRecords="ccRecords"></tag-table>
      </b-container>
    </b-card>
  </b-container>
</template>
<script>
// import TagCorporateTable from "./recordTables/TagCorporateTable";
import config from "../../config";
import FamilyTable from "./recordTables/FamilyTable";
import MemberTable from "./recordTables/MemberTable";
import TagTable from "./recordTables/TagTable";
const queryString = require("query-string");
const moment = require("moment");
// const DATE_FORMAT_DISP = "MM/DD/YYYY";
const DATE_FORMAT = "YYYY-MM-DD";
// const TIME_FORMAT = "hh:mm A";
export default {
  name: "CCRecords",
  components: {
    "tag-table": TagTable,
    "member-table": MemberTable,
    "family-table": FamilyTable,
  },
  data() {
    return {
      institution: null,
      loggedInToken: "",
      ccRecords: [],
      startDate: "",
      endDate: "",
      numTodayRecords: 0,
      dismissSecs: 5,
      dismissCountDown: 0,
    };
  },
  mounted() {
    console.log(`CC-Record mounted!`);
    this.startDate = moment().format(DATE_FORMAT);
    this.endDate = moment().format(DATE_FORMAT);

    var institution = this.$store.state.institution;
    if (institution != null) {
      console.log(
        `CC-Record mounted!, institution - ${JSON.stringify(institution)}`
      );
      this.institution = institution;
    }
    this.loggedInToken = this.$store.state.loggedInToken
    var _this = this;
    this.getCCRecordsFromDb((items) => {
      _this.showAlert();
      _this.numTodayRecords = items.length;
    });
  },
  computed: {
    hasDateRanges() {
      return this.startDate.length > 0 && this.endDate.length > 0;
    },
    useFamilyTable() {
      if (this.institution === null) return false;
      return (
        this.institution.member_type === "guardian"
      );
    },
    useMemberTable() {
      if (this.institution === null) return false;
      return (
        this.institution.member_type === "standard"
      );
    },
    useTagTable() {
      if (this.institution === null) return false;
      return (
        this.institution.member_type === "tag"
      );
    },
    instType() {
      if (this.institution === null) return "corporate"
      return this.institution.type
    },
  },
  methods: {
    countDownChanged(dismissCountDown) {
      this.dismissCountDown = dismissCountDown;
    },
    showAlert() {
      this.dismissCountDown = this.dismissSecs;
    },
    onApplyDateRangeFilter() {
      this.currentPage = 1;
      this.getCCRecordsFromDb(null);
    },
    getCCRecordsFromDb() {
      // console.log("CUSTOMER_TABLE: customers: " + JSON.stringify(this.customers));
      var _this = this;
      //// getting CC-Records from database
      const queryParams = { instID: this.institution._id };
      if (this.startDate.length > 0) {
        queryParams.startDate = moment(this.startDate)
          .hour(0)
          .minute(0)
          .second(0)
          .format();
      }
      if (this.endDate.length > 0) {
        queryParams.endDate = moment(this.endDate)
          .hour(23)
          .minute(59)
          .second(59)
          .format();
      }
      const queryArgs = queryString.stringify(queryParams);

      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "cc-records?" + queryArgs;
      console.log(`CC-Records: getCCRecord query -  ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText);
          //// if data received with no error, add to table
          _this.ccRecords = JSON.parse(this.responseText).data;
          // if (_this.ccRecords) {
          //   _this.items = _this.ccRecords.map((ccRecord) => {
          //     return _this.mapCCRecordToItem(ccRecord);
          //   });
          // }
          // console.log("Record_Table: items: " + JSON.stringify(this.items));
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
    deleteCCRecordFromDb(item) {
      // Send a Delete Request per each of selected Items
      var _this = this;
      const http = new XMLHttpRequest();
      var query = config.API_LOCATION + "cc-record/" + item._id;
      console.log(`CC-Records: delete CCRecord query -  ${query}`);
      http.open("DELETE", query, true);

      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText);
          var response = JSON.parse(this.responseText);

          if (response.message != undefined) {
            _this.$bvToast.toast(response.message, {
              title: "DataBase Message",
              autoHideDelay: 2000,
            });
            // console.log("Record_Table: items: " + JSON.stringify(this.items));
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
};
</script>