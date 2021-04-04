<template>
  <div class="background-image">
    <!-- <b-navbar v-if="loggedInMember !== null" class="mb-2" sticky type="dark" variant="dark">
      <b-navbar-brand href="#">
        <img src="../../assets/mAIRobotics Logo White Transparent-h30.png" alt />
      </b-navbar-brand>
      <b-navbar-nav class="ml-auto">
        <b-nav-item-dropdown v-if="showUser" right> -->
          <!--  Using 'button-content' slot -->
          <!-- <template #button-content>
            <em>{{loggedInMemberName}}</em>
          </template>
          <b-dropdown-item @click="onSignOut">Sign Out</b-dropdown-item>
        </b-nav-item-dropdown>
      </b-navbar-nav>
    </b-navbar> -->
    <router-view></router-view>
  </div>
</template>
<script>
// import config from "../../config";
export default {
  name: "MobileContainer",

  data() {
    return {
      loggedInMember: null,
    };
  },

  computed: {
    loggedInMemberName() {
      if (this.loggedInMember === null) {
        return "";
      }
      return this.loggedInMember.first_name;
    },
    showUser() {
      return this.loggedInMember != null
    }
  },
  mounted() {
    this.loggedInMember = this.$store.state.loggedInMember;
  },
  watch: {
    /* eslint-disable */
    $route(to, from) {
      var _this = this;
      console.log("router watcher invoked!");
      this.loggedInMember = this.$store.state.loggedInMember;

      // this.getInstitutionFromDb(
      //   this.loggedInMember.institution_id,
      //   (institution) => {
      //     _this.$store.commit("setInstitution", institution);
      //   }
      // );
    },
    /* eslint-enable */
  },
  methods: {
    onSignOut() {
      console.log(`onSignOut`);
      this.$store.commit("resetLoggedInFamily");
      this.$store.commit("resetLoggedInMember");
      this.$store.commit("resetInstitution");
      this.$router.push("/mobile/login");
    },
  },
};
</script>

<style>
.background-image{
 background-image: url("../../assets/background.png");
 background-size: 100vw;
 background-repeat: no-repeat;
 background-color: rgb(250, 250, 250);
}
</style>
