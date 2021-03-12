<template>
  <b-container>
    <h4>Activate User</h4>
    <b-form style="margin-top: 50px;" @submit="onMemberActivate">
      <b-form-group label="Phone #: " label-align="left" label-for="phone-number-input" description="Phone # is for authentication purpose only">
        <b-form-input id="phone-number-input" v-model="phoneNum" placeholder="Enter Phone #" required @keyup="focusOut" @blur="focusOut"></b-form-input>
      </b-form-group>
      <b-form-group label="Reg. Code" label-align="left" label-for="reg-code-input" description="Please check your email for registration code">
        <b-form-input id="reg-code-input" v-model="regCode" placeholder="Enter Reg.Code" required @keyup="focusOut" @blur="focusOut"></b-form-input>
      </b-form-group>
      <b-row class="ml-1 mb-2">
        <b-button type="submit" variant="warning">Activate</b-button>
      </b-row>
    </b-form>
    <b-row class="ml-1">
      <b-button variant="primary" @click="$router.push('/mobile/login')">To Login Page</b-button>
    </b-row>
  </b-container>
</template>

<script>
import config from "../../config";
// const moment = require('moment')
export default {
  name: "MobileActivate",
  data() {
    return {
      phoneValue: "",
      phoneNum: "",
      regCode: "",
      preventNextIteration: false,
    };
  },
  methods: {
    focusOut(event) {
      if (["Arrow", "Backspace", "Shift"].includes(event.key)) {
        this.preventNextIteration = true;
        return;
      }
      if (this.preventNextIteration) {
        this.preventNextIteration = false;
        return;
      }
      if (this.phoneNum.length > 0) {
        this.phoneValue = this.phoneNum
          .replace(/-/g, "")
          .match(/(\d{1,10})/g)[0];
      }
      // Format display value based on calculated currencyValue
      this.phoneNum = this.phoneValue.replace(
        /(\d{1,3})(\d{1,3})(\d{1,4})/g,
        "$1-$2-$3"
      );
    },
    onMemberActivate(event) {
      var _this = this;
      event.preventDefault();
      this.activateMemberInDb(
        () => {
          _this.loginMember();
        },
        () => {
          alert("Activation Failed:( - Try Another Reg. Code");
        }
      );
    },
    loginMember() {
      var _this = this;
      this.getMemberByPhoneNumFromDb(
        // On Success, set LoggedInMember and Login
        (response) => {
          // Determine if Family Mode should be used
          console.log(`onMemberLogin response - ${response}`)
          var member = response.member
          var family = response.family
          if (!member) return;
          _this.$store.commit("setLoggedInMember", member);
          if (family) _this.$store.commit("setLoggedInFamily", family);

          //   _this.$store.commit("resetLoggedInFamily")
          //   _this.$store.commit("resetInstitution")
          _this.getInstitutionFromDb(member.institution_id, (inst) => {
            _this.$store.commit("setInstitution", inst);
            _this.$router.push("/mobile/home");
          });
        },
        () => {
          _this.$router.push("/mobile/activate");
        }
      );
    },
    activateMemberInDb(successCallback, activateFailedCallback) {
      // var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "member/activate";
      console.log(`activateGuaridanInDb - query: ${query}`);
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          if (successCallback != null) {
            successCallback();
          }
        } else if (this.readyState === 4) {
          var response = JSON.parse(this.responseText);
          if (response.message != undefined) {
            if (
              response.message.includes("Try Another") &&
              activateFailedCallback != null
            ) {
              activateFailedCallback();
              return;
            }
            alert(response.message);
            return;
          }
          alert(this.responseText);
        }
      };
      try {
        http.send(
          JSON.stringify({
            phone_num: this.phoneNum,
            reg_code: this.regCode,
          })
        );
      } catch (e) {
        alert(e);
      }
    },
    getMemberByPhoneNumFromDb(successCallback, needRegCallback) {
      // var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "member/login";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          var responseData = JSON.parse(this.responseText).data;
          if (responseData && successCallback != null) {
            successCallback(responseData);
            return;
          }
        } else if (this.readyState === 4) {
          var response = JSON.parse(this.responseText);
          if (response.message != undefined) {
            if (
              response.message.includes("Not Activated") &&
              needRegCallback != null
            ) {
              needRegCallback();
              return;
            }
            alert(response.message);
            return;
          }
          alert(this.responseText);
        }
      };
      try {
        http.send(
          JSON.stringify({
            phone_num: this.phoneNum,
          })
        );
      } catch (e) {
        alert(e);
      }
    },
    getInstitutionFromDb(instID, callback) {
      //// get active user in database
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "institution/" + instID;
      console.log(`MobileLogin - getInst query: ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText)
          if (this.responseText.length == 0) {
            return;
          }
          var responseData = JSON.parse(this.responseText).data;
          if (responseData && callback != null) {
            callback(responseData);
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