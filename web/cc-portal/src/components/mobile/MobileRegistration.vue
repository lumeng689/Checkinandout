<template>
  <div id="wrap">
    <div class="activate-head">
      <div class="activate-logo">
        <img class="header-image" src="../../assets/mAIRobotics_Logo_300px.png" />
      </div>
      <div class="welcome">
        <p class="welcome-text">Sign Up</p>
      </div>
    </div>

    <b-form class="activate-form" @submit="onMemberRegister">
      <div class="row" id="first-name-row">
        <b-input-group>
          <input id="first-name-input" type="text" class="form-control form-rounded form-control-lg" v-model="firstName" placeholder="First Name" required />
        </b-input-group>
      </div>
      <div class="row" id="last-name-row">
        <b-input-group>
          <input id="last-name-input" v-model="lastName" placeholder="Last Name" type="text" class="form-control form-rounded form-control-lg" required />
        </b-input-group>
        <p class="notes"></p>
      </div>
      <div class="row">
        <b-input-group>
          <input id="phone-number-input" v-model="phoneNum" placeholder="Phone Number" type="text" class="form-control form-rounded form-control-lg" required @keyup="focusOut" @blur="focusOut" />
        </b-input-group>
        <p class="notes"></p>
      </div>

      <div class="row" id="registration-button-div">
        <b-button class="registration-button" pill type="submit" block variant="dark" size="lg">Sign Up</b-button>
      </div>
      <div class="link-to-pages">
        Already have an account?
        <b-link @click="$router.push('/mobile/login')"><span class="link-text">Log In</span></b-link>
      </div>
    </b-form>
  </div>
</template>

<script>
import config from "../../config";
export default {
  name: "MobileRegistration",
  props: ["instID"],
  data() {
    return {
      firstName: "",
      lastName: "",
      phoneValue: "",
      phoneNum: "",
      preventNextIteration: false,
    };
  },
  created() {
    console.log(`MobileRegistration - created! instID: ${this.instID}`);
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
    onMemberRegister(event) {
      var _this = this;
      event.preventDefault();
      this.registerMemberInDb(() => {
        alert(
          `Registration Succeed! Please Check SMS for RegCode`
        );
        _this.$router.push("/mobile/activate");
      });
    },
    registerMemberInDb(successCallback) {
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "member/register-and-sms";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          var response = JSON.parse(this.responseText);
          console.log(response)
          if (response.data && successCallback != null) {
            successCallback(response);
            return;
          }
        } else if (this.readyState === 4) {
          var responseFailed = JSON.parse(this.responseText);
          if (responseFailed.message != undefined) {
            alert(responseFailed.message);
            return;
          }
          alert(this.responseText);
        }
      };
      try {
        http.send(
          JSON.stringify({
            institution_id: this.instID,
            phone_num: this.phoneNum,
            first_name: this.firstName,
            last_name: this.lastName,
          })
        );
      } catch (e) {
        alert(e);
      }
    },
  },
};
</script>

<style>
#first-name-input,
#last-name-input {
  border-radius: 50px;
  background-color: rgb(231, 231, 231);
  border: 0;
}

#first-name-row,
#last-name-row {
  padding-bottom: 5%;
}

#registration-button-div {
  padding-top: 16vh;
  text-transform: uppercase;
}

.registration-button {
  color: rgb(59, 231, 223);
}
</style>
