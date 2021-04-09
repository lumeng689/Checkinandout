<template>
    <div id="wrap">
      <div class="activate-head">
        <div class="activate-logo">
          <img class="header-image" src="../../assets/mAIRobotics_Logo_300px.png">
        </div>
        <div class="welcome">
          <p class="welcome-text">First Time User</p>
        </div>
      </div>
   
      <b-form class="activate-form" @submit="onMemberActivate">
          <div class="row">
            <b-input-group>
              <input
                id="phone-number-input"
                type="text" 
                class="form-control form-rounded form-control-lg"
                v-model="phoneNum"
                placeholder="Phone Number"
                required
                @keyup="focusOut"
                @blur="focusOut"
              >
            </b-input-group>
              <p class="notes">
                * Phone number is for authentication purpose only
              </p>
          
          </div>
          <div class="row">
            <b-input-group>
              <input
                id="reg-code-input"
                v-model="regCode"
                placeholder="Registration Code"
                type="text" 
                class="form-control form-rounded form-control-lg"
                required
              >
            </b-input-group>
              <p class="notes">
                * Please check your message for registration code
              </p>
          
          </div>
          <div class="row" id="activate-button-div">
            <b-button class="activate-button" pill type="submit" block variant="dark" size="lg"
              > Activate
              </b-button>
          </div>
          <div class="link-to-pages">
            Not a first time user?
        <b-link @click="$router.push('/mobile/login')"
          > <span class="link-text">Log In</span></b-link
        >
      </div>
        
      </b-form>
    </div>
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
          var responseData = response.data
          var token = response.token
          console.log(`onMemberLogin response - ${responseData}`);
          var member = responseData.member;
          var family = responseData.family;
          if (!member) return;
          _this.$store.commit("setLoggedInMember", member);
          _this.$store.commit("setLoggedInToken", token)
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
      http.onreadystatechange = function() {
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
      http.onreadystatechange = function() {
        if (this.readyState === 4 && this.status === 200) {
        var response = JSON.parse(this.responseText)
          if (response.data && successCallback != null) {
            successCallback(response);
            return;
          }
        } else if (this.readyState === 4) {
          var responseFailed = JSON.parse(this.responseText)
          if (responseFailed.message != undefined) {
            if (
              responseFailed.message.includes("Not Activated") &&
              needRegCallback != null
            ) {
              needRegCallback();
              return;
            }
            alert(responseFailed.message);
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
      http.onreadystatechange = function() {
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

<style>

#wrap .activate-head {
  height: 30vh;
  margin-left: auto;
  margin-right: auto;
  width: 100%;
  padding: 5vh;
}

.activate-logo{
  padding-top: 6vh;
  height: 15vh;
}

#wrap .activate-form {
  position: absolute;
  height: 60vh;
  width: 100%;
  align-self: center;
  padding-left: 4rem;
  padding-right: 4rem;
}

.notes{
  font-size: 0.8rem;
  padding-top: 2%;
}

#reg-code-input{
  border-radius: 50px;
  background-color: rgb(231, 231, 231);
  border: 0;
}

#activate-button-div { 
  padding-top: 16vh;
  text-transform: uppercase;
}

.activate-button{
  color:rgb(59, 231, 223);
  font: Proxima Nova;
}

</style>
