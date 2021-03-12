<template>
  <b-container class="h-100" fluid>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;">
      <b-col xl="3" lg="4">
        <h2>Admin Dashboard</h2>
      </b-col>
    </b-row>
    <b-card class="h-100">
      <b-container class="h-100" fluid>
        <b-table bordered sticky-header="90%" id="admin-table" ref="adminTable" responsive="sm" :items="items" :fields="fields" :current-page="currentPage" :per-page="perPage"></b-table>
        <b-pagination v-model="currentPage" :total-rows="numRows" :per-page="perPage" aria-controls="ward-table"></b-pagination>
      </b-container>
    </b-card>
  </b-container>
</template>
<script>
import config from "../config";
const queryString = require("query-string");
const moment = require("moment");
const DATETIME_FORMAT = "MMM DD YYYY hh:mm A";
export default {
  name: "Admins",
  data() {
    return {
      instId: "",
      currentPage: 1,
      perPage: 20,
      admins: [],
      fields: ["FRAS_username", "last_login"],
      items: [],
    };
  },
  mounted() {
    // console.log("Wards Dashboard mounted")
    var activeUser = this.$store.state.activeUser;
    if (activeUser != null) {
      console.log(
        `Admins Dashboard mounted!, instId - ${activeUser.institution_id}`
      );
      this.instId = activeUser.institution_id;
    }
    this.getAdminsFromDb();
  },
  computed: {
    numRows() {
      return this.items.length;
    },
  },
  methods: {
    mapAdminToItem(admin) {
      var logInTime = moment(admin.last_login_at);
      var logInDisplay = logInTime.isBefore("1970-01-01")
        ? ""
        : logInTime.format(DATETIME_FORMAT);
      return {
        _id: admin._id,
        FRAS_username: admin.fras_username,
        last_login: logInDisplay,
      };
    },
    getAdminsFromDb() {
      // console.log("CUSTOMER_TABLE: customers: " + JSON.stringify(this.customers));
      var _this = this;
      //// getting Families from database
      const queryParams = { instID: this.instId };
      const queryArgs = queryString.stringify(queryParams);
      const http = new XMLHttpRequest();

      var query = config.API_LOCATION + "admins?" + queryArgs;
      console.log(`Admins: get Admin query -  ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText);
          //// if data received with no error, add to table
          _this.admins = JSON.parse(this.responseText).data;
          // console.log(JSON.stringify(_this.families));
          if (_this.admins) {
            _this.items = _this.admins.map((admin) => {
              return _this.mapAdminToItem(admin);
            });
            // _this.filteredItems = _this.items;
          }
          // console.log("Event_Table: items: " + JSON.stringify(this.items));
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