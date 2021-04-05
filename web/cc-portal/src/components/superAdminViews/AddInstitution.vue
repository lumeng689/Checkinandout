<template>
  <b-container>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;" align-h="between">
      <b-col xl="3" lg="4">
        <h2>Add Institution</h2>
      </b-col>
      <b-col xl="3" lg="4">
        <b-button @click="$router.push('./add-admin')" variant="info">Create Admin</b-button>
      </b-col>
    </b-row>
    <b-row class="ml-4 mr-4">
      <b-col lg="4">
        <b-form @submit="onSubmit">
          <b-form-group id="input-group-1" label="Name" label-align="left" label-for="name-input">
            <b-form-input id="name-input" v-model="form.name" placeholder="Enter name" required></b-form-input>
          </b-form-group>
          <b-form-group id="input-group-2" label="Address" label-align="left" label-for="address-input">
            <b-form-input id="address-input" v-model="form.address" placeholder="Enter address" required></b-form-input>
          </b-form-group>
          <b-form-group id="input-group-3" label="State" label-align="left" label-for="state-input">
            <b-form-input id="state-input" v-model="form.state" placeholder="Enter state" required></b-form-input>
          </b-form-group>
          <b-form-group id="input-group-4" label="Zip Code" label-align="left" label-for="zip-code-input">
            <b-form-input id="zip-code-input" v-model="form.zipCode" placeholder="Enter Zip Code" required></b-form-input>
          </b-form-group>
          <b-form-group id="input-group-5" label="Institution Type" label-align="left" label-for="inst-type-select">
            <b-form-select id="inst-type-select" v-model="form.instType" :options="instTypeOptions" required></b-form-select>
          </b-form-group>
          <b-form-group id="input-group-6" label="Workflow Type" label-align="left" label-for="workflow-type-select">
            <b-form-select id="workflow-type-select" v-model="form.workflowType" :options="workflowTypeOptions" required></b-form-select>
          </b-form-group>
          <b-form-group id="input-group-7" label="Member Type" label-align="left" label-for="member-type-select">
            <b-form-select id="member-type-select" v-model="form.memberType" :options="memberTypeOptions" required></b-form-select>
          </b-form-group>
          <b-form-group  v-if="isMemberType" id="input-group-8" label="Require Survey When Checking-In" label-align="left" label-for="require-survey-checkbox">
            <b-form-checkbox  id="require-survey-checkbox" v-model="form.requireSurvey" required></b-form-checkbox>
          </b-form-group>
          <b-form-group id="input-group-9" v-if="isTagType" label="Identifier String" label-align="left" label-for="identifier-input">
            <b-form-input id="address-identifier" v-model="form.identifier" placeholder="Enter Identifier" required></b-form-input>
          </b-form-group>
          <b-form-group id="input-group-10" v-if="isTagType" label="Regex Expression For Tag" label-align="left" label-for="regex-input">
            <b-form-input id="regex-input" v-model="form.tagStringRegex" placeholder="Enter Regex"></b-form-input>
          </b-form-group>
          <b-button type="submit" variant="primary">Create</b-button>
        </b-form>
      </b-col>
    </b-row>
  </b-container>
</template>

<script>
import Vue from 'vue'
import config from "../../config";
import common from "../../common";
export default Vue.extend({
  name: "AddInstitution",
  data() {
    return {
      form: {
        name: "",
        address: "",
        state: "",
        zipCode: "",
        instType: "",
        workflowType: "",
        memberType: "",
        identifier: "",
        tagStringRegex: "",
        requireSurvey: false,
      },
      instTypeOptions: [
        { value: common.INST_TYPE_SCHOOL, text: "School" },
        { value: common.INST_TYPE_HOSPITAL, text: "Hospital" },
        { value: common.INST_TYPE_CORPORATE, text: "Corporate" },
      ],
      workflowTypeOptions: [
        { value: common.WORKFLOW_TYPE_CC, text: "Check-In & Check-Out" },
        { value: common.WORKFLOW_TYPE_CHECK_IN, text: "Check-In Only" },
      ],
      memberTypeOptions: [
        {
          value: common.MEMBER_TYPE_GUARDIAN,
          text: "Guardian & Students/Patients",
        },
        {
          value: common.MEMBER_TYPE_STANDARD,
          text: "Patients/Employees/Members",
        },
        { value: common.MEMBER_TYPE_TAG, text: "Printed QRCode/Tags" },
      ],
    };
  },
  computed: {
    isTagType() {
      return this.form.memberType === common.MEMBER_TYPE_TAG
    },
    isMemberType() {
      return (this.form.memberType === common.MEMBER_TYPE_STANDARD) || (this.form.memberType === common.MEMBER_TYPE_GUARDIAN)
    }
  },

  methods: {
    onSubmit(event) {
      var _this = this;
      event.preventDefault();
      this.createInstitutionInDb((message) => {
        _this.$bvToast.toast(message, {
          title: "DataBase Message",
          autoHideDelay: 2000,
        });
      });
    },
    createInstitutionInDb(callback) {
      const http = new XMLHttpRequest();
      var requestBody = {
        name: this.form.name,
        address: this.form.address,
        state: this.form.state,
        zip_code: this.form.zipCode,
        type: this.form.instType,
        workflow_type: this.form.workflowType,
        member_type: this.form.memberType,
        identifier: this.form.identifier,
        require_survey: this.form.requireSurvey,
        custom_tag_string_regex: this.form.tagStringRegex,
      };
      const query = config.API_LOCATION + "institution";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${config.SUPER_ADMIN_TOKEN}`);
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
  },
})
</script>
