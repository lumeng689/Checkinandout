<template>
  <b-container>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;" align-h="between">
      <b-col xl="3" lg="4">
        <h2>Redirect to Check-InNOut Portal</h2>
      </b-col>
      <b-col xl="3" lg="4">
        <b-button @click="$router.push('./add-admin')" variant="info">Create Admin</b-button>
      </b-col>
    </b-row>
    <b-row class="ml-4 mr-4">
      <b-col lg="4">
        <b-form-group id="input-group-1" label="Select Institution" label-align="left" label-for="institution-select">
          <b-form-select class="mb-4" id="institution-select" v-model="instSelected" :options="instOptions"></b-form-select>
        </b-form-group>
        <b-form-group id="input-group-2" label="Select Admin" label-align="left" label-for="admin-select">
          <b-form-select class="mb-4" id="admin-select" v-model="adminSelected" :options="adminOptions"></b-form-select>
        </b-form-group>
        <b-button variant="success" :disabled="adminSelected === null" @click="onRedirectToCCPortal">To CC-Portal</b-button>
      </b-col>
    </b-row>
  </b-container>
</template>
<script>
import config from "../../config";
const queryString = require("query-string");
export default {
  name: "RedirectCCPortal",
  data() {
    return {
      instSelected: null,
      adminSelected: null,
      instOptions: [],
      adminOptions: [],
    };
  },
  watch: {
    instSelected: function () {
      if (this.instSelected === null) {
        this.adminOptions = [];
      }
      this.getAdminsFromDb();
    },
  },

  mounted() {
    this.getInstitutionsFromDb();
  },

  methods: {
    mapInstToOptions(inst) {
      return {
        value: inst,
        text: inst.name,
      };
    },
    mapAdminToOptions(admin) {
      return {
        value: admin,
        text: admin.fras_username,
      };
    },

    onRedirectToCCPortal() {
      // set config
      // config.FRAS_USERNAME = this.adminSelected.fras_username;
      var _this = this
      this.logInAdminByDb(
        (response) => {
          var admin = response.data
          var token = response.token
          _this.$store.commit("setInstitution", _this.instSelected);
          _this.$store.commit("setActiveUser", admin);
            _this.$store.commit("setLoggedInToken", token)
            _this.$router.push("/portal/cc-records")
        }
      )
      // redirect
    },

    getInstitutionsFromDb() {
      var _this = this;
      const query = config.API_LOCATION + "institutions";
      const http = new XMLHttpRequest();
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${config.SUPER_ADMIN_TOKEN}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          var response = JSON.parse(this.responseText);
          var institutions = response.data;
          if (institutions) {
            _this.instOptions = institutions.map((inst) => {
              return _this.mapInstToOptions(inst);
            });
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
    getAdminsFromDb() {
      var _this = this;
      const queryParams = { instID: this.instSelected._id };
      const queryArgs = queryString.stringify(queryParams);
      const query = config.API_LOCATION + "admins?" + queryArgs;
      const http = new XMLHttpRequest();
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${config.SUPER_ADMIN_TOKEN}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          var response = JSON.parse(this.responseText);
          var admins = response.data;
          if (admins) {
            _this.adminOptions = admins.map((admin) => {
              return _this.mapAdminToOptions(admin);
            });
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
          fras_username: this.adminSelected.fras_username,
          password: this.adminSelected.password,
        }));
      } catch (e) {
        alert(e);
      }
    },
  },
};
</script>
