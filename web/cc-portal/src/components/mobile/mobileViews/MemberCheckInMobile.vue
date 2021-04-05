<template>
  <div id="wrap">
    <div class="activate-head">
      <div class="activate-logo">
        <img
          class="header-image"
          src="../../../assets/mAIRobotics_Logo_300px.png"
        />
      </div>
      <div class="person-welcome">
        <p class="person-welcome-text">Hello, {{ loggedInGuardianName }}</p>
      </div>
    </div>
    <div class="profile">
      <div class="person-photo">
        <b-avatar size="7em"></b-avatar>
      </div>
      <div class="info">
        <p class="info-text">Welcome Back</p>
        <p class="info-date">{{ date }}</p>
      </div>
    </div>
    <div class="button-div">
      <div class="checkin-button-div">
        <b-button
          class="checkin-button"
          pill
          block
          variant="outline-dark"
          size="lg"
          @click="onCheckIn"
        >
          Check In
        </b-button>
      </div>
      <div class="signout-button-div">
        <b-button
          class="signout-button"
          pill
          block
          variant="dark"
          size="lg"
          @click="onSignOut"
        >
          Sign Out
        </b-button>
      </div>
    </div>
    <b-modal
      static
      hide-footer
      ref="cc-modal"
      :title="modalTitle + ' QRCode'"
      @hidden="onModalHidden"
    >
      <b-container>
        <div id="qrcode" style="width:250; height:250; margin-top:15px;"></div>
      </b-container>
    </b-modal>
  </div>
</template>

<script>
import config from "../../../config";
import QRCode from "../../../utils/qrcode";
const queryString = require("query-string");
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
  name: "member-check-in-mobile",
  data() {
    return {
      institution: null,
      timer: null,
      timeObj: null,
      date: "",
      time: "",
      // loggedInFamily: null,
      // loggedInMemberID: "",
      loggedInMember: null,
      ccRecord: null,
      qrcode: null,
      modalTitle: "",
      // selectedSchedule: 0,
      // scheduleOptions: [
      //   { value: 0, text: "Now" },
      //   { value: 5, text: "5 minutes" },
      //   { value: 10, text: "10 minutes" },
      //   { value: 20, text: "20 minutes" },
      //   { value: 30, text: "30 minutes" },
      //   { value: 60, text: "1 hour" },
      // ],
      // schedulingWardIndex: -1,
      // CCMode: "single",
      // surveyState: INIT_STATE,
      // ccIndex: -1,
    };
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
    // this.loggedInFamily = this.$store.state.loggedInFamily;
    // this.loggedInMemberID = this.$store.state.loggedInMemberID;
    this.institution = this.$store.state.institution;
    this.loggedInMember = this.$store.state.loggedInMember;
    console.log(
      // `Mobile mounted - loggedInMemberID: ${this.loggedInMemberID}`
      `Mobile mounted - loggedInMemberID: ${JSON.stringify(
        this.loggedInMember
      )}`
    );
    if (this.loggedInMember === null) {
      // do something when family info is not available
      this.$router.push("/mobile/login");
      return;
    }

    // Initialize CCRecord and QRCode
    this.qrcode = new QRCode(document.getElementById("qrcode"), {
      width: 250,
      height: 250,
    });
    // Open Check-In QRCode if survey succeed
    var surveyPassed = this.$route.query.proceed;
    var _this = this;
    this.syncCCRecordWithDb(() => {
      console.log(`ccRecord received - ${_this.ccRecord}`);
      if (surveyPassed === "true") {
        _this.showCCModal();
      }
    });
  },
  beforeDestroy() {
    clearInterval(this.timer);
  },
  computed: {
    wards() {
      if (this.loggedInFamily === null) {
        return [];
      }
      return this.loggedInFamily.wards;
    },

    memberStage() {
      if (this.ccRecord === null) {
        return "";
      }
      if (this.ccRecord.status === 0) return CHECK_IN_STAGE;
      if (this.ccRecord.status === 1) return SCHEDULE_STAGE;
      if (this.ccRecord.status === 2) return CHECK_OUT_STAGE;

      return "";
    },
    // allCheckIn() {
    //   return this.wardStages.every((s) => {
    //     return s === CHECK_IN_STAGE;
    //   });
    // },
    // allScheduleCheckOut() {
    //   return this.wardStages.every((s) => {
    //     return s === SCHEDULE_STAGE;
    //   });
    // },
    // allCheckOut() {
    //   return this.wardStages.every((s) => {
    //     return s === CHECK_OUT_STAGE;
    //   });
    // },
    loggedInGuardianName() {
      if (!this.loggedInMember) {
        return "";
      }
      return this.loggedInMember.first_name;
    },
  },

  methods: {
    onSync() {
      this.syncCCRecordWithDb(() => {
        console.log("On Sync - ccRecord received!");
      });
    },
    onCheckIn() {
      if (!this.institution.require_survey) {
        this.showCCModal();
        return;
      }
      const queryParams = {
        instID: this.institution._id,
        memberID: this.loggedInMember._id,
      };
      const queryArgs = queryString.stringify(queryParams);
      window.location.href = "/surveys/check-in-survey.html?" + queryArgs;
    },
    onSignOut() {
      console.log(`onSignOut`);
      this.$store.commit("resetLoggedInFamily");
      this.$store.commit("resetLoggedInMember");
      this.$store.commit("resetInstitution");
      this.$router.push("/mobile/login");
    },
    onModalHidden() {
      this.syncCCRecordWithDb();
    },
    showCCModal() {
      this.syncCCRecordWithDb(() => {
        console.log("ccRecords received!");
      });
      this.modalTitle = this.memberStage;
      this.$refs["cc-modal"].show();
      var qrcodeString = this.getQRCodeString();
      console.log(`qrcodeString - ${qrcodeString}`);
      this.qrcode.makeCode(qrcodeString);
    },
    syncCCRecordWithDb(callback) {
      console.log(
        `syncCCRecordWithDb - ID to check: ${this.loggedInMember._id}`
      );
      var _this = this;
      const http = new XMLHttpRequest();
      const requestBody = {
        institution_id: this.loggedInMember.institution_id,
        member_id: this.loggedInMember._id,
      };
      const query = config.API_LOCATION + "cc-record/sync";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function() {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          var prevCCRecord = _this.ccRecord;
          _this.ccRecord = JSON.parse(this.responseText).data;
          if (_this.ccRecord && callback != null) {
            callback(prevCCRecord, _this.ccRecord);
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
    // sendCCScheduleToDb(callback) {
    //   // var _this = this;
    //   const http = new XMLHttpRequest();
    //   var requestBody = null;
    //   if (this.schedulingWardIndex == -1) {
    //     requestBody = {
    //       ward_ids: this.wardIds,
    //       timestamp: moment().add(this.selectedSchedule, "minutes").unix(),
    //     };
    //   } else {
    //     requestBody = {
    //       ward_ids: [this.wardIds[this.schedulingWardIndex]],
    //       timestamp: moment().add(this.selectedSchedule, "minutes").unix(),
    //     };
    //   }

    //   const query = config.API_LOCATION + "cc-record/schedule";
    //   http.open("POST", query, true);
    //   http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
    //   http.onreadystatechange = function () {
    //     if (this.readyState === 4 && this.status === 200) {
    //       // var response = JSON.parse(this.responseText);
    //       if (callback != null) {
    //         callback();
    //       }
    //     } else if (this.readyState === 4) {
    //       alert(this.responseText);
    //     }
    //   };
    //   try {
    //     http.send(JSON.stringify(requestBody));
    //   } catch (e) {
    //     alert(e);
    //   }
    // },

    getQRCodeString() {
      var stage;
      if (this.memberStage == CHECK_IN_STAGE) stage = "checkin";
      if (this.memberStage == CHECK_OUT_STAGE) stage = "checkout";

      return [this.loggedInMember._id, stage].join("|");
    },
  },
};
</script>

<style>
.person-welcome {
  text-align: center;
  font: 1.5em sans-serif;
}

#wrap .profile {
  position: absolute;
  height: 38vh;
  width: 100%;
  align-self: center;
  padding-top: 3vh;
  padding-left: 4rem;
  padding-right: 4rem;
}

.person-photo {
  height: 18vh;
}

.info {
  height: 17vh;
}

.info-text {
  text-align: center;
  font-size: 1.5em;
  margin: 0;
  padding: 0;
}

.info-date {
  text-align: center;
  font: 1em sans-serif;
}

#wrap .button-div {
  padding-top: 38vh;
  height: 27vh;
  width: 100%;
  align-self: center;
  padding-left: 4rem;
  padding-right: 4rem;
}

.checkin-button-div {
  padding-bottom: 5%;
}

.signout-button {
  color: rgb(59, 231, 223);
}
</style>
