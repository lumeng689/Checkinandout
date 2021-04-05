<template>
  <b-container>
    <b-row class="ml-2">
      <h4>Welcome {{(loggedInMember) ? loggedInMember.first_name : ""}}</h4>
    </b-row>
    <b-row class="ml-4 mr-4" align-h="between">
      <b-col lg="3" cols="5">
        <b-row>{{ date }}</b-row>
        <b-row>{{ time }}</b-row>
      </b-col>
      <b-col lg="2" cols="3">
        <b-button variant="success" @click="onSync">Sync</b-button>
      </b-col>
    </b-row>
    <b-row style="margin-top: 80px;">
      <b-col cols="12">
        <b-row class="mb-2" align-h="around">
          <b-button variant="info" :disabled="!allCheckIn" @click="onCheckInAll()">Check In All</b-button>
          <b-button variant="info" :disabled="!allScheduleCheckOut && !allCheckOut" @click="onCheckOutAll()">Check Out All</b-button>
        </b-row>
        <b-row v-for="(w, index) in wards" v-bind:key="w._id">
          <b-card style="width: 100%">
            <template #header>
              <b-row align-h="between">
                <b-col>
                  <h5>
                    <b>{{ w.first_name + " " + w.last_name }}</b>
                  </h5>
                </b-col>
                <b-col>{{ w.group }}</b-col>
              </b-row>
            </template>
            <b-row align-h="around">
              <b-button variant="primary" :disabled="wardStages[index] != 'Check-In'" @click="onCheckInSingle(index)">Check In</b-button>
              <b-button variant="primary" :disabled="wardStages[index] === 'Check-In'" @click="onCheckOutSingle(index)">Check Out</b-button>
            </b-row>
          </b-card>
        </b-row>
      </b-col>
    </b-row>
    <b-modal static hide-footer ref="cc-modal" :title="modalTitle + ' QRCode'" @hidden="onModalHidden">
      <b-container>
        <div id="qrcode" style="width:250; height:250; margin-top:15px;"></div>
      </b-container>
    </b-modal>
    <b-modal hide-footer ref="schedule-modal" title="Schedule Checkout" @hidden="onModalHidden">
      <b-container>
        <b-form @submit="onScheduleCheckOut">
          <b-form-group label="Checkout After: ">
            <b-form-select required v-model="selectedSchedule" :options="scheduleOptions"></b-form-select>
          </b-form-group>
          <b-button type="submit" variant="primary">OK</b-button>
        </b-form>
      </b-container>
    </b-modal>
  </b-container>
</template>

<script>
import config from "../../../config";
const queryString = require("query-string");
import QRCode from "../../../utils/qrcode";
const moment = require("moment");
const DATE_FORMAT = "MMM DD, YYYY";
const TIME_FORMAT = "hh:mm:ss A";
const CHECK_IN_STAGE = "Check-In";
const SCHEDULE_STAGE = "Schedule";
const CHECK_OUT_STAGE = "Check-Out";

// survey-related state machine
// const INIT_STATE = 0;
// const CHECK_IN_ON_STATE = 1;
// const CHECK_IN_DONE_STATE = 2;
export default {
  name: "family-cc-mobile",
  data() {
    return {
      institution: null,
      timer: null,
      date: "",
      time: "",
      loggedInMember: null,
      loggedInFamily: null,
      loggedInToken: "",
      ccRecords: [],
      qrcode: null,
      modalTitle: "",
      selectedSchedule: 0,
      scheduleOptions: [
        { value: 0, text: "Now" },
        { value: 5, text: "5 minutes" },
        { value: 10, text: "10 minutes" },
        { value: 20, text: "20 minutes" },
        { value: 30, text: "30 minutes" },
        { value: 60, text: "1 hour" },
      ],
      schedulingWardIndex: -1,
    };
  },
  computed: {
    wards() {
      if (this.loggedInFamily === null) {
        return [];
      }
      return this.loggedInFamily.wards;
    },
    wardIds() {
      return this.loggedInFamily.wards.map((ward) => {
        return ward._id;
      });
    },
    wardStages() {
      if (this.ccRecords === null) {
        return [];
      }
      var wardStages = this.ccRecords.map((ccr) => {
        if (ccr.status === 0) return CHECK_IN_STAGE;
        if (ccr.status === 1) return SCHEDULE_STAGE;
        if (ccr.status === 2) return CHECK_OUT_STAGE;
      });
      console.log(`wardStages - ${wardStages}`);
      return wardStages;
    },
    allCheckIn() {
      return this.wardStages.every((s) => {
        return s === CHECK_IN_STAGE;
      });
    },
    allScheduleCheckOut() {
      return this.wardStages.every((s) => {
        return s === SCHEDULE_STAGE;
      });
    },
    allCheckOut() {
      return this.wardStages.every((s) => {
        return s === CHECK_OUT_STAGE;
      });
    },
    
  },
  created() {
    var _this = this;
    var _moment = moment;
    this.timer = setInterval(() => {
      _this.timeObj = _moment();
      _this.date = _this.timeObj.format(DATE_FORMAT);
      _this.time = _this.timeObj.format(TIME_FORMAT);
    }, 1000);
  },
  mounted() {
    this.institution = this.$store.state.institution;
    this.loggedInMember = this.$store.state.loggedInMember;
    this.loggedInFamily = this.$store.state.loggedInFamily;
    this.loggedInToken = this.$store.state.loggedInToken;
    if (this.loggedInFamily === null) {
      // do something when family info is not available
      this.$router.push("/mobile/login");
      return;
    }
    console.log(`loggedInFamily - ${JSON.stringify(this.loggedInFamily)}`);
    
    this.qrcode = new QRCode(document.getElementById("qrcode"), {
      width: 250,
      height: 250,
    });

    // Open Check-In QRCode if survey succeed
    var surveyPassed = this.$route.query.proceed;
    var wardIndexForCheckIn = this.$store.state.wardIndexForCheckIn;
    var _this = this;
    this.syncCCRecordWithDb(() => {
      console.log("ccRecords received!");
      if (surveyPassed != "true") return;
      if (wardIndexForCheckIn === -1) {
        _this.showCCModalAll();
      } else {
        _this.showCCModalSingle(wardIndexForCheckIn);
        // Reset WardIndex to -1
        this.$store.commit("resetWardIndexForCheckIn")
      }
    });
  },
  methods: {
    onSync() {
      this.syncCCRecordWithDb(() => {
        console.log("On Sync - ccRecord received!");
      });
    },
    onCheckInSingle(index) {
      // this.surveyState = CHECK_IN_ON_STATE;
      // this.showCCModalSingle(index);
      if (!this.institution.require_survey) {
        this.showCCModalSingle(index);
        return;
      }
      this.$store.commit("setWardIndexForCheckIn", index);
      const queryParams = {
        instID: this.institution._id,
        memberID: this.loggedInMember._id,
      }
      const queryArgs = queryString.stringify(queryParams)
      window.location.href = "/surveys/check-in-survey.html?" + queryArgs 
    },
    onCheckInAll() {
      // this.surveyState = CHECK_IN_ON_STATE;
      // this.showCCModalAll();
      if (!this.institution.require_survey) {
        this.showCCModalAll();
        return;
      }
      const queryParams = {
        instID: this.institution._id,
        memberID: this.loggedInMember._id,
      }
      const queryArgs = queryString.stringify(queryParams)
      window.location.href = "/surveys/check-in-survey.html?" + queryArgs 
    },
    onCheckOutSingle(index) {
      // this.surveyState = INIT_STATE;
      if (this.wardStages[index] === SCHEDULE_STAGE) {
        this.showScheduleModalSingle(index);
        return;
      }
      if (this.wardStages[index] === CHECK_OUT_STAGE) {
        this.showCCModalSingle(index);
      }
      console.log(`CheckOutAll - ${this.allCheckOut}`);
    },
    onCheckOutAll() {
      // this.surveyState = INIT_STATE;
      if (this.allScheduleCheckOut) {
        this.showScheduleModalAll();
        return;
      }
      if (this.allCheckIn || this.allCheckOut) {
        this.showCCModalAll();
      }
    },
    onModalHidden() {
      // var _this = this;
      this.syncCCRecordWithDb(() => {
        // _this.surveyState = INIT_STATE;
        console.log("Modal Hidden - ccRecords received!");
      });
    },
    onScheduleCheckOut() {
      this.sendCCScheduleToDb(() => {
        this.syncCCRecordWithDb(() => {
          console.log("On Schedule Checkout - ccRecords received!");
        });
      });
      this.$refs["schedule-modal"].hide();
    },
    showCCModalSingle(index) {
      
      this.syncCCRecordWithDb(() => {
        console.log("ccRecords received!");
      });
      this.modalTitle = this.wardStages[index];
      this.showCCModal();
      var qrcodeString = this.getQRCodeStringSingle(index);
      console.log(`qrcodeString - ${qrcodeString}`);
      this.qrcode.makeCode(qrcodeString);
    },
    showCCModalAll() {
      this.syncCCRecordWithDb(() => {
        console.log("ccRecords received!");
      });
      this.modalTitle =
        (this.allCheckIn ? CHECK_IN_STAGE : CHECK_OUT_STAGE) + " All";
      this.showCCModal();
      var qrcodeString = this.getQRCodeStringAll();
      console.log(`qrcodeString - ${qrcodeString}`);
      this.qrcode.makeCode(qrcodeString);
    },
    showCCModal() {
      this.$refs["cc-modal"].show();
    },
    showScheduleModalSingle(index) {
      this.schedulingWardIndex = index;
      this.$refs["schedule-modal"].show();
    },
    showScheduleModalAll() {
      this.schedulingWardIndex = -1;
      this.$refs["schedule-modal"].show();
    },
    syncCCRecordWithDb(callback) {
      console.log(`syncCCRecordWithDb - Ids to check: ${this.wardIds}`);
      var _this = this;
      const http = new XMLHttpRequest();
      
      const query = config.API_LOCATION + "cc-record/sync";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          _this.ccRecords = JSON.parse(this.responseText).data;
          if (_this.ccRecords && callback != null) {
            callback(_this.ccRecords);
          }
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        http.send(JSON.stringify({
          institution_id: this.institution._id,
          ward_ids: this.wardIds
        }));
      } catch (e) {
        alert(e);
      }
    },
    sendCCScheduleToDb(callback) {
      // var _this = this;
      const http = new XMLHttpRequest();
      var requestBody = null;
      if (this.schedulingWardIndex == -1) {
        requestBody = {
          ward_ids: this.wardIds,
          timestamp: moment().add(this.selectedSchedule, "minutes").unix(),
        };
      } else {
        requestBody = {
          ward_ids: [this.wardIds[this.schedulingWardIndex]],
          timestamp: moment().add(this.selectedSchedule, "minutes").unix(),
        };
      }

      const query = config.API_LOCATION + "cc-record/schedule";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          if (callback != null) {
            callback();
          }
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        http.send(JSON.stringify(requestBody));
      } catch (e) {
        alert(e);
      }
    },

    getQRCodeStringSingle(index) {
      var stageString =
        this.wardStages[index] === CHECK_IN_STAGE ? "checkin" : "checkout";
      return [
        this.loggedInMember._id,
        this.wardIds[index],
        stageString,
        "single",
      ].join("|");
    },
    getQRCodeStringAll() {
      var stageString = this.allCheckIn ? "checkin" : "checkout";
      return [this.loggedInMember._id, "none", stageString, "all"].join(
        "|"
      );
    },
  },
};
</script>