<template>
  <div id="wrap">
    <div class="form-head">
      <div class="logo">
        <img
          class="header-image"
          src="../../assets/mAIRobotics_Logo_300px.png"
        />
      </div>
      <div class="welcome">
        <p class="welcome-text">Welcome Back</p>
      </div>
    </div>

    <b-form class="form" @submit="onMemberLogin">
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
          />
        </b-input-group>
      </div>
      <div class="row" id="sign-in-button-div">
        <b-button
          class="sign-in-button"
          pill
          type="submit"
          block
          variant="dark"
          size="lg"
        >
          Sign In
        </b-button>
      </div>
      <div class="link-to-pages">
        First time user?
        <b-link @click="$router.push('/mobile/activate')"
          > Active Account
        </b-link>
      </div>
    </b-form>
  </div>
</template>

<script>
import config from "../../config";
// const moment = require('moment')
export default {
  name: "MobileLogin",
  data() {
    return {
      phoneValue: "",
      phoneNum: "",
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
    onMemberLogin(event) {
      var _this = this;
      event.preventDefault();
      this.getMemberByPhoneNumFromDb(
        // On Success, set LoggedInMember and Login
        (response) => {
          // Determine if Family Mode should be used
          console.log(`onMemberLogin response - ${response}`);
          var member = response.member;
          var family = response.family;
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
    getMemberByPhoneNumFromDb(successCallback, needRegCallback) {
      // var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "member/login";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function() {
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
    setStatesInStorage() {},
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
#wrap {
  width: 100vw;
  height: 90vh;
}

#wrap .form-head {
  height: 40vh;
  margin-left: auto;
  margin-right: auto;
  width: 100%;
  padding: 5vh;
}

.logo {
  height: 27vh;
}

.header-image {
  width: 25vh;
  height: 4vh;
}

.welcome {
  text-align: center;
  font: 2em sans-serif;
}

.welcome-text {
  color: rgb(59, 231, 223);
  font-weight: bold;
}

#phone-number-input {
  border-radius: 50px;
  background-color: rgba(182, 182, 182, 0.568);
  border: 0;
}

#wrap .form {
  position: absolute;
  height: 50vh;
  width: 100%;
  align-self: center;
  position: absolute;
  padding-left: 4rem;
  padding-right: 4rem;
}

#sign-in-button-div {
  padding-top: 22vh;
  text-transform: uppercase;
}

.sign-in-button {
  color: rgb(59, 231, 223);
}

.link-to-pages {
  text-align: center;
  padding-top: 5%;
  font-size: 1rem;
}
</style>
