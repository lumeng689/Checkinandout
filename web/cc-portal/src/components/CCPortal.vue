<template>
  <div>
    <b-navbar sticky type="dark" variant="dark">
      <b-navbar-brand href="#">
        <img class="mr-2" src="../assets/mAIRobotics Logo White Transparent-h30.png" alt />
          Check Me! </b-navbar-brand>
      <b-navbar-nav class="ml-auto">
        <b-nav-item-dropdown right>
          <!-- Using 'button-content' slot -->
          <template #button-content>
            <em>{{frasUsername}}</em>
          </template>
          <b-dropdown-item :to="'/portal/profile'">Profile</b-dropdown-item>
          <b-dropdown-item :to="'/portal/settings'">Settings</b-dropdown-item>
          <b-dropdown-item href="/user-login">Sign Out</b-dropdown-item>
        </b-nav-item-dropdown>
      </b-navbar-nav>
    </b-navbar>
    <div class="cc-sidebar">
      <family-school-sidebar v-if="useFamilySchoolSidebar"></family-school-sidebar>
      <member-corporate-sidebar v-if="useMemberCorporateSidebar"></member-corporate-sidebar>
      <member-hospital-sidebar v-if="useMemberHospitalSidebar"></member-hospital-sidebar>
      <tag-corporate-sidebar v-if="useTagCorporateSidebar"></tag-corporate-sidebar>
    </div>
    <div class="cc-content">
      <router-view v-if="activeUserLoaded"></router-view>
    </div>
  </div>
</template>

<script>
import config from "../config";
import FamilySchoolSidebar from "./sidebars/FamilySchoolSidebar.vue";
import MemberCorporateSidebar from "./sidebars/MemberCorporateSidebar.vue";
import MemberHospitalSidebar from "./sidebars/MemberHospitalSidebar.vue";
import TagCorporateSidebar from "./sidebars/TagCorporateSidebar.vue";
const queryString = require("query-string");

export default {
  name: "CCPortal",
  components: {
    "family-school-sidebar": FamilySchoolSidebar,
    "member-corporate-sidebar": MemberCorporateSidebar,
    "member-hospital-sidebar": MemberHospitalSidebar,
    "tag-corporate-sidebar": TagCorporateSidebar,
  },
  data() {
    return {
      frasUsername: config.FRAS_USERNAME,
      activeUser: null,
      institution: null,
      activeUserLoaded: false,
    };
  },

  computed: {
    activeInstitutionId() {
      console.log("CC-Portal: activeInstitutionId");
      if (this.activeUser === null) {
        return "";
      }
      return this.activeUser.institution_id;
    },
    useFamilySchoolSidebar() {
      if (this.institution === null) return false;
      return (
        this.institution.type === "school" &&
        this.institution.member_type === "guardian"
      );
    },
    useMemberCorporateSidebar() {
      if (this.institution === null) return false;
      return (
        this.institution.type === "corporate" &&
        this.institution.member_type === "standard"
      );
    },
    useMemberHospitalSidebar() {
      if (this.institution === null) return false;
      return (
        this.institution.type === "hospital" &&
        this.institution.member_type === "standard"
      );
    },
    useTagCorporateSidebar() {
      if (this.institution === null) return false;
      return (
        this.institution.type === "corporate" &&
        this.institution.member_type === "tag"
      );
    },
  },
  mounted() {
    console.log("CC-Portal: mounted!");

    // If activeUser is already in storage, use it directly, rather than getting from DB
    this.activeUser = this.$store.state.activeUser;
    this.institution = this.$store.state.institution;
    console.log(
      `CCPortal: created! - active user: ${JSON.stringify(this.activeUser)}`
    );
    if (this.activeUser != null && this.institution != null) {
      this.frasUsername = this.activeUser.fras_username;
      this.activeUserLoaded = true;
      return;
    }
    var _this = this;
    this.getActiveUserFromDb(this.frasUsername, (user) => {
      _this.activeUser = user;
      _this.$store.commit("setActiveUser", user);
      _this.getInstitutionFromDb(user.institution_id, (institution) => {
        // console.log(`institution - ${JSON.stringify(institution)}`)
        _this.institution = institution
        _this.$store.commit("setInstitution", institution);
        _this.activeUserLoaded = true;
        _this.institution = institution;
      });
    });
  },

  methods: {
    getActiveUserFromDb(frasUsername, callback) {
      // console.log('CC-Portal: getActiveUserFromDb() called!')
      //// get active user in database
      const http = new XMLHttpRequest();
      const queryParams = { frasUsername: frasUsername };
      const queryArgs = queryString.stringify(queryParams);

      const query = config.API_LOCATION + "admin?" + queryArgs;
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText)
          if (this.responseText.length === 0) {
            return;
          }
          var responseData = JSON.parse(this.responseText).data;
          callback(responseData);
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
    getInstitutionFromDb(instID, callback) {
      //// get active user in database
      const http = new XMLHttpRequest();
      const query =
        config.API_LOCATION + "institution/" + instID;
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText)
          if (this.responseText.length == 0) {
            return;
          }
          var responseData = JSON.parse(this.responseText).data;
          callback(responseData);
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
.cc-sidebar {
  height: 100%;
  width: 200px;
  position: fixed !important;
  background-color: rgb(5, 57, 75);
}
.nav-link {
  color: rgb(162, 195, 235) !important;
}
.cc-content {
  margin-left: 200px;
  height: 90%;
}
html,
body {
  height: 100%;
}
</style>
