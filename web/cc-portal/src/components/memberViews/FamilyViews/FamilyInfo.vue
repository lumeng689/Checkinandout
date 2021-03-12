<template>
  <b-container>
    <h3>Family Information</h3>
    <family-info-card
      item-type="guardian"
      :entities="members"
      v-on:need-entity-action="updateMemberToDB($event)"
      v-on:need-entity-create="addNewMemberToDB($event)"
      v-on:need-entity-delete="deleteMemberFromDB($event)"
    ></family-info-card>
    <family-info-card
      item-type="vehicle"
      :entities="vehicles"
      v-on:need-entity-update="updateVehicleToDB($event)"
      v-on:need-entity-create="addNewVehicleToDB($event)"
      v-on:need-entity-delete="deleteVehicleFromDB($event)"
    ></family-info-card>
    <family-info-card
      item-type="ward"
      :entities="wards"
      v-on:need-entity-update="updateWardToDB($event)"
      v-on:need-entity-create="addNewWardToDB($event)"
      v-on:need-entity-delete="deleteWardFromDB($event)"
    ></family-info-card>
  </b-container>
</template>
<script>
import config from "../../../config";
const queryString = require("query-string");
import FamilyInfoCard from "./FamilyInfoCard.vue";
export default {
  name: "FamilyInfo",
  components: {
    "family-info-card": FamilyInfoCard,
  },
  data() {
    return {
      loadedFamilyWithMembers: null,
    };
  },
  computed: {
    members() {
      if (this.loadedFamilyWithMembers === null) return [];
      return this.loadedFamilyWithMembers.members;
    },
    vehicles() {
      if (this.loadedFamilyWithMembers === null) return [];
      return this.loadedFamilyWithMembers.vehicles;
    },
    wards() {
      if (this.loadedFamilyWithMembers === null) return [];
      return this.loadedFamilyWithMembers.wards;
    },
  },
  created() {
    this.loadedFamilyWithMembers = this.$store.state.loadedFamilyWithMembers;
  },
  methods: {
    addNewMemberToDB(entity) {
      var _this = this;
      const query = config.API_LOCATION + "member";
      const http = new XMLHttpRequest();
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 201) {
          // var response = JSON.parse(this.responseText);
          console.log("create succeed!");
          _this.getFamilyWithMemberByIDFromDB();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        var memberToSend = JSON.stringify(this.getMemberToSubmit(entity));
        console.log(`memberToSend - ${memberToSend}`);
        http.send(memberToSend);
      } catch (e) {
        alert(e);
      }
    },
    updateMemberToDB(entity) {
      var _this = this;
      const query = config.API_LOCATION + "member/" + entity._id;
      const http = new XMLHttpRequest();
      http.open("PUT", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("update succeed!");
          _this.getFamilyWithMemberByIDFromDB();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        console.log(`memberToSend - ${JSON.stringify(entity)}`);
        http.send(JSON.stringify(entity));
      } catch (e) {
        alert(e);
      }
    },
    deleteMemberFromDB(memberID) {
      var _this = this;
      const query = config.API_LOCATION + "member/" + memberID;
      const http = new XMLHttpRequest();
      http.open("DELETE", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("delete succeed!");
          _this.getFamilyWithMemberByIDFromDB();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
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
    addNewVehicleToDB(entity) {
      var _this = this;
      const queryParams = {familyID: this.loadedFamilyWithMembers._id}
      const queryArgs = queryString.stringify(queryParams)

      const query = config.API_LOCATION + "vehicle/add-new?" + queryArgs;
      const http = new XMLHttpRequest();
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 201) {
          // var response = JSON.parse(this.responseText);
          console.log("create succeed!");
          _this.getFamilyWithMemberByIDFromDB();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        console.log(`vehicleToSend - ${JSON.stringify(entity)}`);
        http.send(JSON.stringify(entity));
      } catch (e) {
        alert(e);
      }
    },
    updateVehicleToDB(entity) {
      var _this = this;
      const query = config.API_LOCATION + "vehicle/" + entity._id;
      const http = new XMLHttpRequest();
      http.open("PUT", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("create succeed!");
          _this.getFamilyWithMemberByIDFromDB();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        console.log(`vehicleToSend - ${JSON.stringify(entity)}`);
        http.send(JSON.stringify(entity));
      } catch (e) {
        alert(e);
      }
    },
    deleteVehicleFromDB(vehicleID) {
      var _this = this;
      const query = config.API_LOCATION + "vehicle/" + vehicleID;
      const http = new XMLHttpRequest();
      http.open("DELETE", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("delete succeed!");
          _this.getFamilyWithMemberByIDFromDB();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
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
    addNewWardToDB(entity) {
      var _this = this;
      const queryParams = {familyID: this.loadedFamilyWithMembers._id}
      const queryArgs = queryString.stringify(queryParams)

      const query = config.API_LOCATION + "ward/add-new?" + queryArgs;
      const http = new XMLHttpRequest();
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 201) {
          // var response = JSON.parse(this.responseText);
          console.log("create succeed!");
          _this.getFamilyWithMemberByIDFromDB();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        console.log(`wardToSend - ${JSON.stringify(entity)}`);
        http.send(JSON.stringify(entity));
      } catch (e) {
        alert(e);
      }
    },
    updateWardToDB(entity) {
      var _this = this;
      const query = config.API_LOCATION + "ward/" + entity._id;
      const http = new XMLHttpRequest();
      http.open("PUT", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("create succeed!");
          _this.getFamilyWithMemberByIDFromDB();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        console.log(`vehicleToSend - ${JSON.stringify(entity)}`);
        http.send(JSON.stringify(entity));
      } catch (e) {
        alert(e);
      }
    },
    deleteWardFromDB(wardID) {
      var _this = this;
      const query = config.API_LOCATION + "ward/" + wardID;
      const http = new XMLHttpRequest();
      http.open("DELETE", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("delete succeed!");
          _this.getFamilyWithMemberByIDFromDB();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
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
    getFamilyWithMemberByIDFromDB() {
      var _this = this;
      const http = new XMLHttpRequest();
      const query =
        config.API_LOCATION +
        "family-with-members/" +
        this.loadedFamilyWithMembers._id;
      console.log(`getFamilyWithMembersByID - query: ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText)
          if (this.responseText.length == 0) {
            return;
          }
          var response = JSON.parse(this.responseText).data;
          _this.loadedFamilyWithMembers = response;
          _this.$store.commit("setLoadedFamilyWithMembers", response);
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
    getMemberToSubmit(entity) {
      return {
        institution_id: this.loadedFamilyWithMembers.institution_id,
        family_info: {
          id: this.loadedFamilyWithMembers._id,
          relation: entity.relation,
        },
        phone_num: entity.phone_num,
        email: entity.email,
        first_name: entity.first_name,
        last_name: entity.last_name,
      };
    },
  },
};
</script>