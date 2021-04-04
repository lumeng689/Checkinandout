<template>
  <div id="wrap">
    <div class="activate-head">
      <div class="activate-logo">
        <img
          class="header-image"
          src="../../assets/mAIRobotics_Logo_300px.png"
        />
      </div>
      <div class="welcome">
        <p class="welcome-text">Sign Up</p>
      </div>
    </div>

    <b-form class="activate-form">
      <div class="row" id="first-name-row">
        <b-input-group>
          <input
            id="first-name-input"
            type="text"
            class="form-control form-rounded form-control-lg"
            v-model="firstName"
            placeholder="First Name"
            required
          />
        </b-input-group>
      </div>
      <div class="row" id="last-name-row">
        <b-input-group>
          <input
            id="last-name-input"
            v-model="lastName"
            placeholder="Last Name"
            type="text"
            class="form-control form-rounded form-control-lg"
            required
          />
        </b-input-group>
        <p class="notes"></p>
      </div>
      <div class="row">
        <b-input-group>
          <input
            id="phone-number-input"
            v-model="phoneNum"
            placeholder="Phone Number"
            type="text"
            class="form-control form-rounded form-control-lg"
            required
            @keyup="focusOut"
            @blur="focusOut"
          />
        </b-input-group>
        <p class="notes"></p>
      </div>

      <div class="row" id="registration-button-div">
        <b-button
          class="registration-button"
          pill
          type="submit"
          block
          variant="dark"
          size="lg"
        >
          Sign Up
        </b-button>
      </div>
      <div class="link-to-pages">
          Already have an account?
        <b-link @click="$router.push('/mobile/login')"
          > Log In</b-link
        >
      </div>
    </b-form>
  </div>
</template>

<script>
export default {
  name: "MobileRegistration",
  data() {
    return {
      firstName: "",
      lastName: "",
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
  },
};
</script>

<style>
#first-name-input,
#last-name-input {
  border-radius: 50px;
  background-color: rgba(182, 182, 182, 0.658);
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
