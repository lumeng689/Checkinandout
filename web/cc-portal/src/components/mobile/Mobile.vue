<template>
  <div>
    <family-cc-mobile v-if="useFamilyCCMobile"></family-cc-mobile>
    <member-cc-mobile v-if="useMemberCCMobile"></member-cc-mobile>
    <member-check-in-mobile v-if="useMemberCheckInMobile"></member-check-in-mobile>
  </div>
</template>
<script>
// import config from "../config";
import common from "../../common";
import FamilyCCMobile from "./mobileViews/FamilyCCMobile";
import MemberCCMobile from "./mobileViews/MemberCCMobile";
import MemberCheckInMobile from "./mobileViews/MemberCheckInMobile";
// import MemberCheckInMobile from "./mobileViews/MemberCheckInMobile_old";
export default {
  name: "Mobile",
  components: {
    "family-cc-mobile": FamilyCCMobile,
    "member-cc-mobile": MemberCCMobile,
    "member-check-in-mobile": MemberCheckInMobile,
  },
  data() {
    return {
      institution: null,
      loggedInMember: null,
    };
  },
  computed: {
    useFamilyCCMobile() {
      if (this.institution === null) return false;
      return (
        this.institution.member_type === common.MEMBER_TYPE_GUARDIAN &&
        this.institution.workflow_type === common.WORKFLOW_TYPE_CC
      );
    },
    useMemberCCMobile() {
      if (this.institution === null) return false;
      return (
        this.institution.member_type === common.MEMBER_TYPE_STANDARD &&
        this.institution.workflow_type === common.WORKFLOW_TYPE_CC
      );
    },
    useMemberCheckInMobile() {
      if (this.institution === null) return false;
      return (
        this.institution.member_type === common.MEMBER_TYPE_STANDARD &&
        this.institution.workflow_type === common.WORKFLOW_TYPE_CHECK_IN
      );
    },
  },
  mounted() {
    var institution = this.$store.state.institution;
    this.loggedInMember = this.$store.state.loggedInMember;

    // Check if User is actually Logged In
    if (this.loggedInMember === null) {
      alert("Something went wrong, try loggin in again!");
      this.$router.push("/mobile/login");
      return;
    }

    if (institution === null) {
      alert("Something went wrong, try loggin in again!");
      this.$router.push("/mobile/login");
      return;
    }
    this.institution = institution;

    console.log(
      `Mobile mounted! - institution: ${JSON.stringify(this.institution)}`
    );
  },
  
};
</script>