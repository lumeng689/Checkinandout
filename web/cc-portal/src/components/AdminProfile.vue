<template>
  <b-container class="h-100 ml-2 mr-2" fluid>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;">
      <b-col xl="3" lg="4">
        <h2>Profile</h2>
      </b-col>
    </b-row>
    <b-row class="ml-4 mr-4">
      <b-col lg="4">
        <b-card>
          <b-row class="ml-2 mt-2"></b-row>
          <h5> <b> Institution: </b> {{name}}</h5>
          <h5><b> Address: </b> {{address + ", " + state}}</h5>
        </b-card>
      </b-col>
    </b-row>
  </b-container>
</template>
<script>
import config from "../config";
export default {
  name: "AdminProfile",
  data() {
    return {
      instId: "",
      institution: null,
    };
  },
  computed: {
    name: function () {
      return (this.institution === null) ? "" : this.institution.name 
    },
    address: function () {
      return (this.institution === null) ? "" : this.institution.address 
    },
    state: function () {
      return (this.institution === null) ? "" : this.institution.state
    },
  },
  mounted() {
    // console.log("Wards Dashboard mounted")

    var activeUser = this.$store.state.activeUser;
    if (activeUser != null) {
      console.log(
        `Family Dashboard mounted!, instId - ${activeUser.institution_id}`
      );
      this.instId = activeUser.institution_id;
    }
    this.getInstFromDb();
  },
  methods: {
    getInstFromDb() {
      // console.log("CUSTOMER_TABLE: customers: " + JSON.stringify(this.customers));
      var _this = this;
      const http = new XMLHttpRequest();
      var query = config.API_LOCATION + "institution/" + this.instId;
      console.log(`Profile: get Institution query -  ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          _this.institution = JSON.parse(this.responseText).data;
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