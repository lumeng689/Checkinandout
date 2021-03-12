<template>
  <b-container>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;" align-h="between">
      <b-col xl="3" lg="4">
        <h2>Add Admin</h2>
      </b-col>
      <b-col xl="2" lg="3">
        <b-button class="mb-2" @click="$router.push('./add-institution')" variant="info">Create Institution</b-button>
        <b-button @click="$router.push('./redirect')" variant="warning">To CC-Portal</b-button>
      </b-col>
    </b-row>
    <b-row class="ml-4 mr-4">
      <b-col lg="4">
        <b-form @submit="onSubmit">
          <b-form-group id="input-group-2" label="Affliated Instituttion" label-align="left" label-for="institution-select">
            <b-form-select id="institution-select" v-model="form.institutionId" :options="instOptions" required></b-form-select>
          </b-form-group>
          <b-form-group class="mb-4" id="input-group-1" label="Username" label-align="left" label-for="name-input">
            <b-form-input id="name-input" v-model="form.FrasUsername" placeholder="Enter Fras Username" required></b-form-input>
          </b-form-group>
          <b-form-group class="mb-4" id="input-group-1" label="Password" label-align="left" label-for="password-input">
            <b-form-input id="password-input" type="password" v-model="form.password" placeholder="Enter Password" required></b-form-input>
          </b-form-group>
          <b-form-group class="mb-4" id="input-group-1" label="Confirm Password" label-align="left" label-for="password-input-confirm">
            <b-form-input id="password-input-confirm" type="password" v-model="form.passwordConfirm" placeholder="Confirm Password" required></b-form-input>
          </b-form-group>
          <b-button type="submit" variant="primary">Create</b-button>
        </b-form>
      </b-col>
    </b-row>
  </b-container>
</template>

<script>
import config from "../config";
export default {
  name: "AddAdmin",
  data() {
    return {
      form: {
        FrasUsername: "",
        password: "",
        passwordConfirm: "",
        institutionId: "",
      },
      instOptions: [],
    };
  },
  mounted() {
    this.getInstitutionsFromDb();
  },

  methods: {
    mapInstToOptions(inst) {
      return {
        value: inst._id,
        text: inst.name,
      };
    },

    onSubmit(event) {
      var _this = this;
      event.preventDefault();
      if (this.form.password != this.form.passwordConfirm) {
        alert("Password Confirmation Does NOT Match! Try Again.")
        return;
      }
      this.createAdminInDb((message) => {
        _this.$bvToast.toast(message, {
          title: "DataBase Message",
          autoHideDelay: 2000,
        });
      });
    },
    createAdminInDb(callback) {
      const http = new XMLHttpRequest();
      var requestBody = {
        fras_username: this.form.FrasUsername,
        password: this.form.password,
        institution_id: this.form.institutionId,
      };
      const query = config.API_LOCATION + "admin/register";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 201) {
          var response = JSON.parse(this.responseText);
          if (response.message != undefined && callback != null) {
            callback(response.message);
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
    getInstitutionsFromDb() {
      var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "institutions";
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
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
  },
};
</script>
