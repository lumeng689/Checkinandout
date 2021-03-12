<template>
  <b-container>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;">
      <b-col xl="3" lg="4">
        <h2>Add Family</h2>
      </b-col>
    </b-row>
    <b-card bg-variant="light">
      <b-form state="true" @submit="onFamilyFormSubmit">
        <b-form-group label="Guardian Info" label-size="lg" label-align="left" label-class="font-weight-bold pt-0" class="mb-0">
          <b-form-group v-for="g in guardians" v-bind:key="g.index" :label="'Guardian '+g.index" label-size="md" label-align="left" label-class="font-weight-bold pt-0" class="mb-0">
            <b-form-row class="mb-2">
              <b-col sm="3">
                <b-input-group prepend="First Name">
                  <b-form-input required :id="'guardian-input-first-name-'+g.index" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="guardians[g.index-1].first_name"></b-form-input>
                </b-input-group>
              </b-col>
              <b-col sm="3">
                <b-input-group prepend="Last Name">
                  <b-form-input required :id="'guardian-input-last-name-'+g.index" placeholder="???" v-model="guardians[g.index-1].last_name"></b-form-input>
                </b-input-group>
              </b-col>
              <b-col sm="3">
                <b-input-group prepend="Phone #">
                  <b-form-input required :id="'guardian-input-phone-num-'+g.index" class="mb-2 mr-sm-2 mb-sm-0" placeholder="xxx-xxx-xxxx" v-model="guardians[g.index-1].phone_num"></b-form-input>
                </b-input-group>
              </b-col>
            </b-form-row>
            <b-form-row class="mb-3">
              <b-col sm="5">
                <b-input-group prepend="Email">
                  <b-form-input required :id="'guardian-input-email-'+g.index" placeholder="xxxxxxx@xxx.com" v-model="guardians[g.index-1].email"></b-form-input>
                </b-input-group>
              </b-col>
              <b-col sm="3">
                <b-input-group prepend="Relation">
                  <b-form-input required :id="'guardian-input-relation-'+g.index" placeholder="Brother" v-model="guardians[g.index-1].relation"></b-form-input>
                </b-input-group>
              </b-col>
            </b-form-row>
          </b-form-group>
          <b-row>
            <b-button class="ml-2 mt-2" variant="outline-info" @click="onAddGuardian">Add Guardian</b-button>
            <b-button class="ml-2 mt-2" variant="outline-danger" v-if="guardians.length>1" @click="onRemoveGuardian">Remove Guardian</b-button>
          </b-row>
        </b-form-group>
        <b-form-group label="Vehicle Info" label-size="lg" label-align="left" label-class="font-weight-bold pt-0" class="mt-4 mb-0">
          <b-form-group v-for="v in vehicles" v-bind:key="v.index" :label="'Vehicle '+v.index" label-size="md" label-align="left" label-class="font-weight-bold pt-0" class="mb-0">
            <b-form-row class="mb-2">
              <b-col sm="3">
                <b-input-group prepend="Make">
                  <b-form-input :id="'vehicle-input-make-'+v.index" class="mb-2 mr-sm-2 mb-sm-0" placeholder="Ford" v-model="vehicles[v.index-1].make"></b-form-input>
                </b-input-group>
              </b-col>
              <b-col sm="3">
                <b-input-group prepend="Model">
                  <b-form-input :id="'vehicle-input-model-'+v.index" placeholder="Model T" v-model="vehicles[v.index-1].model"></b-form-input>
                </b-input-group>
              </b-col>
              <b-col sm="3">
                <b-input-group prepend="Color">
                  <b-form-input :id="'vehicle-input-color-'+v.index" placeholder="Blue" v-model="vehicles[v.index-1].color"></b-form-input>
                </b-input-group>
              </b-col>
              <b-col sm="3">
                <b-input-group prepend="Plate #">
                  <b-form-input :id="'vehicle-input-plate-number-'+v.index" placeholder="XXXXXX" v-model="vehicles[v.index-1].plate_num"></b-form-input>
                </b-input-group>
              </b-col>
            </b-form-row>
          </b-form-group>
          <b-row>
            <b-button class="ml-2 mt-2" variant="outline-info" @click="onAddVehicle">Add Vehicle</b-button>
            <b-button class="ml-2 mt-2" variant="outline-danger" v-if="vehicles.length>1" @click="onRemoveVehicle">Remove Vehicle</b-button>
          </b-row>
        </b-form-group>
        <b-form-group label="Children Info" label-size="lg" label-align="left" label-class="font-weight-bold pt-0" class="mt-4 mb-0">
          <b-form-group v-for="w in wards" v-bind:key="w.index" :label="'Child '+w.index" label-size="md" label-align="left" label-class="font-weight-bold pt-0" class="mb-0">
            <b-form-row class="mb-2">
              <b-col sm="3">
                <b-input-group prepend="First Name">
                  <b-form-input :id="'ward-input-first-name-'+w.index" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="wards[w.index-1].first_name"></b-form-input>
                </b-input-group>
              </b-col>
              <b-col sm="3">
                <b-input-group prepend="Last Name">
                  <b-form-input :id="'ward-input-last-name-'+w.index" placeholder="???" v-model="wards[w.index-1].last_name"></b-form-input>
                </b-input-group>
              </b-col>
              <b-col sm="3">
                <b-input-group prepend="Group">
                  <b-form-input :id="'ward-input-group-'+w.index" placeholder="Group?" v-model="wards[w.index-1].group"></b-form-input>
                </b-input-group>
              </b-col>
            </b-form-row>
          </b-form-group>
          <b-row>
            <b-button class="ml-2 mt-2" variant="outline-info" @click="onAddWard">Add Child</b-button>
            <b-button class="ml-2 mt-2" variant="outline-danger" v-if="wards.length>1" @click="onRemoveWard">Remove Child</b-button>
          </b-row>
        </b-form-group>
        <b-row align-h="end" class="mr-2">
          <b-button variant="primary" type="submit">Submit</b-button>
        </b-row>
      </b-form>
    </b-card>
  </b-container>
</template>

<script>
import config from "../../../config";
export default {
  name: "AddFamily",
  data() {
    return {
      guardians: [this.getNewGuardian(1)],
      wards: [this.getNewWard(1)],
      vehicles: [this.getNewVehicle(1)],
    };
  },
  watch: {},
  computed: {
    instId() {
      var activeUser = this.$store.state.activeUser;
      if (activeUser === null || activeUser.institution_id === undefined) {
        return "";
      }
      return activeUser.institution_id;
    },
    // gValidation() {
    //   var gValidation = [];
    //   for (var i = 0; i < this.guardians.length; i++) {
    //     var g = this.guardians[i];
    //     gValidation[i].firstName = g.instId;
    //   }
    // }
  },
  methods: {
    onAddGuardian() {
      var guardianToAdd = this.getNewGuardian(this.guardians.length + 1);
      this.guardians.push(guardianToAdd);
      // console.log(`guardians: ${this.guardians}`);
    },
    onRemoveGuardian() {
      this.guardians.pop();
    },
    onAddVehicle() {
      var vehicleToAdd = this.getNewWard(this.vehicles.length + 1);
      this.vehicles.push(vehicleToAdd);
      // console.log(`wards: ${this.wards}`);
    },
    onRemoveVehicle() {
      this.vehicles.pop();
    },
    onAddWard() {
      var wardToAdd = this.getNewWard(this.wards.length + 1);
      this.wards.push(wardToAdd);
      // console.log(`wards: ${this.wards}`);
    },
    onRemoveWard() {
      this.wards.pop();
    },
    onFamilyFormSubmit(event) {
      event.preventDefault()
      var familyToSubmit = this.getFamilyToSubmit();
      var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "family";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 201) {
          var response = JSON.parse(this.responseText);
          if (response.message != undefined) {
            _this.$bvToast.toast(response.message, {
              title: "DataBase Message",
              autoHideDelay: 2000,
            });
          }
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        http.send(JSON.stringify(familyToSubmit));
      } catch (e) {
        alert(e);
      }
    },

    getNewGuardian(i) {
      return {
        index: i,
        first_name: "",
        last_name: "",
        phone_num: "",
        email: "",
        relation: "",
      };
    },
    getNewWard(i) {
      return {
        index: i,
        first_name: "",
        last_name: "",
        group: "",
      };
    },
    getNewVehicle(i) {
      return {
        index: i,
        make: "",
        model: "",
        color: "",
        plate_num: "",
      };
    },
    getFamilyToSubmit() {
      return {
        institution_id: this.instId,
        members: this.guardians,
        vehicles: this.vehicles,
        wards: this.wards,
      };
    },
  },
};
</script>