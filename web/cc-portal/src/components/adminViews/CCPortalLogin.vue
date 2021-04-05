<template>
  <b-container>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;" align-h="between">
      <b-col xl="3" lg="4">
        <h2>Login to Check-InNOut Portal</h2>
      </b-col>
    </b-row>
    <b-row class="ml-4 mr-4">
      <b-col lg="4">
        <b-form @submit="onSubmit">
        <b-form-group id="input-group-1" label="Username" label-align="left" label-for="username-input">
          <b-form-input class="mb-4" id="username-input" v-model="form.frasUsername" placeholder="Enter Username" required></b-form-input>
        </b-form-group>
        <b-form-group id="input-group-2" label="Password" label-align="left" label-for="password-input">
          <b-form-input class="mb-4" id="password-input" type="password" v-model="form.password" placeholder="Enter Password" required></b-form-input>
        </b-form-group>
        <b-button type="submit" variant="primary">Login</b-button>
        </b-form>
      </b-col>
    </b-row>
  </b-container>
</template>

<script>
import config from "../../config";
export default {
  name: "CCPortalLogin",
  data() {
    return {
      form: {
        frasUsername: "",
        password: "",
      }
    };
  },

  methods: {
    onSubmit(event) {
      
      var _this = this;
      event.preventDefault();
      this.logInAdminByDb(
        (response) => {
          var admin = response.data
          var token = response.token
          _this.getInstitutionByIdFromDb(admin.institution_id, (inst) => {
            _this.$store.commit("setInstitution", inst)
            _this.$store.commit("setActiveUser", admin)
            _this.$store.commit("setLoggedInToken", token)
            _this.$router.push("/portal/cc-records")
          })
        }
      )
    },
    logInAdminByDb(callback) {
      const query = config.API_LOCATION + "admin/login"
      console.log(`logIn query: ${query}`)
      const http = new XMLHttpRequest();
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        var response = JSON.parse(this.responseText)
        if (this.readyState === 4 && this.status === 200) {
          if (this.responseText.length == 0) {
            return;
          }
          if (response.data && callback != null) {
            callback(response);
          }
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        http.send(JSON.stringify({
          fras_username: this.form.frasUsername,
          password: this.form.password,
        }));
      } catch (e) {
        alert(e);
      }
    },
    getInstitutionByIdFromDb(instID, callback) {
      const query = config.API_LOCATION + "institution/" + instID;
      console.log(`getInstitution query: ${query}`)
      const http = new XMLHttpRequest();
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
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
    }
  }
};
</script>