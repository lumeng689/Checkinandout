<template>
  <b-container>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;">
      <b-col xl="3" lg="4">
        <h2>Settings</h2>
      </b-col>
    </b-row>
    <b-row class="ml-4 mr-4">
      <b-col md="3" cols="6">
        <b-button variant="info" @click="onExportCCRecords">Export CC-Records</b-button>
      </b-col>
      <b-col v-if="enableMemberExport" md="3" cols="6">
        <b-button  variant="info" @click="onExportMembers">Export {{MemberDisplay}}</b-button>
      </b-col>
      <b-col v-if="enableFamilyExport" md="3" cols="6">
        <b-button  variant="info" @click="onExportFamilies">Export Families</b-button>
      </b-col>
      <b-col v-if="enableFamilyExport" md="3" cols="6">
        <b-button  variant="info" @click="onExportWards">Export Students</b-button>
      </b-col>
      <b-col v-if="enableSurveyExport" md="3" cols="6">
        <b-button variant="info" @click="onExportSurveys">Export Surveys</b-button>
      </b-col>
    </b-row>
  </b-container>
</template>
<script>
import config from "../../config";
const moment = require("moment");
const queryString = require("query-string");
export default {
  name: "AdminProfile",
  data() {
    return {
      institution: null,
      loggedInToken: "",
    };
  },
  computed: {
    MemberDisplay() {
      if (this.institution === null) return "";
      if (this.institution.type === "hospital") return "Patients"
      if (this.institution.type === "corporate") return "Employee"
      return "Member"
    },
    enableMemberExport() {
      if (this.institution === null) return false;
      return this.institution.member_type === "standard";
    },
    enableFamilyExport() {
      if (this.institution === null) return false;
      return this.institution.member_type === "guardian";
    },
    enableSurveyExport() {
      if (this.institution === null) return false;
      return this.institution.require_survey;
    },
  },
  mounted() {
    var institution = this.$store.state.institution;
    this.loggedInToken = this.$store.state.loggedInToken
    if (institution != null) {
      console.log(
        `Admub Settings mounted!, institution - ${JSON.stringify(institution)}`
      );
      this.institution = institution;
    }
  },
  methods: {
    onExportCCRecords() {
      var _this = this;
      this.getCCRecordsExportFromDb("cc-records", (csvContent) => {
        var filename =
          "export-cc-records-" + moment().format("MMM.DD") + ".csv";
        _this.downloadExport(filename, csvContent);
      });
    },
    onExportMembers() {
      var _this = this;
      this.getCCRecordsExportFromDb("members", (csvContent) => {
        var filename = "export-members-" + moment().format("MMM.DD") + ".csv";
        _this.downloadExport(filename, csvContent);
      });
    },
    onExportWards() {
      var _this = this;
      this.getCCRecordsExportFromDb("wards", (csvContent) => {
        var filename = "export-wards-" + moment().format("MMM.DD") + ".csv";
        _this.downloadExport(filename, csvContent);
      });
    },
    onExportFamilies() {
      var _this = this;
      this.getCCRecordsExportFromDb("families", (csvContent) => {
        var filename = "export-families-" + moment().format("MMM.DD") + ".csv";
        _this.downloadExport(filename, csvContent);
      });
    },
    onExportSurveys() {
      var _this = this;
      this.getCCRecordsExportFromDb("surveys", (csvContent) => {
        var filename = "export-surveys-" + moment().format("MMM.DD") + ".csv";
        _this.downloadExport(filename, csvContent);
      });
    },
    // recordType: {"cc-records", "wards", "families", "surveys"}
    getCCRecordsExportFromDb(recordType, callback) {
      // var _this = this
      const queryParams = { 
        instID: this.institution._id,
        hourOffset: - (new Date().getTimezoneOffset() / 60),
      };
      
      const queryArgs = queryString.stringify(queryParams);
      const http = new XMLHttpRequest();
      const query =
        config.API_LOCATION + "export/" + recordType + "?" + queryArgs;
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(`downloaded content: ${this.responseText}`);
          callback(this.responseText);
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
    downloadExport(filename, csvContent) {
      // var blob = new Blob([content])
      var file = new File([csvContent], filename, { type: "text/csv" });
      var a = window.document.createElement("a");
      a.href = window.URL.createObjectURL(file);
      a.download = filename;
      a.click();
    },
  },
};
</script>