<template>
  <div>
    <b-card :title="title" align="left">
      <template #header>
        <b-row align-h="end">
          <b-button @click="onCreateEntity()">Add {{title}}</b-button>
        </b-row>
      </template>
      <b-table bordered :items="items" :fields="fields" stacked="md">
        <template #cell(actions)="row">
          <b-link v-if="itemType === 'guardian'" class="mr-2" style="color:green" @click="onCheckRegCode(row.item)">
            <b>Reg.Code</b>
          </b-link>
          <b-link class="mr-2" style="color:blue" @click="onEditEntity(row.item)">
            <b>Edit</b>
          </b-link>
          <b-link class="mr-1" style="color:red" @click="onDeleteEntity(row.item)">
            <b>Delete</b>
          </b-link>
        </template>
      </b-table>
    </b-card>
    <b-modal ref="reg-code-modal" hide-footer title="Check Registration Code">
      <div class="d-block test-left ml-4 mb-2">
        Registration Code for
        <b>{{modalItem.first_name + " " + modalItem.last_name}}</b> is
        <b>{{(currentRegCode) ? currentRegCode.reg_code : ""}}</b>
      </div>
      <b-form>
        <b-input-group class="mb-2" prepend="Phone #">
          <b-form-input id="regcode-phone-input" v-model="regCodePhoneNum" placeholder="xxx-xxx-xxxx"></b-form-input>
        </b-input-group>
      </b-form>
      <b-row class="mr-1" align-h="end">
        <b-button variant="primary" @click="onSendRegCode">Send RegCode</b-button>
      </b-row>
    </b-modal>
    <b-modal v-if="itemType === 'guardian'" ref="entity-modal-guardian" hide-footer title="Edit Guardian">
      <div class="d-block text-left">
        <h3>{{modalAction}} Guardian</h3>
      </div>
      <b-form>
        <b-form-row class="mb-2 mr-2">
          <b-col sm="6">
            <b-input-group prepend="First Name">
              <b-form-input id="guardian-input-first-name'" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.first_name"></b-form-input>
            </b-input-group>
          </b-col>
          <b-col sm="6">
            <b-input-group prepend="Last Name">
              <b-form-input id="guardian-input-last-name" placeholder="???" v-model="modalItem.last_name"></b-form-input>
            </b-input-group>
          </b-col>
        </b-form-row>
        <b-input-group class="mb-2" prepend="Email">
          <b-form-input id="guardian-input-email" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.email"></b-form-input>
        </b-input-group>
        <b-form-row class="mb-2 mr-2">
          <b-col sm="6">
            <b-input-group prepend="Phone #">
              <b-form-input id="guardian-input-phone-number" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.phone_num"></b-form-input>
            </b-input-group>
          </b-col>
          <b-col sm="6">
            <b-input-group prepend="Relation">
              <b-form-input id="guardian-input-relation" placeholder="???" v-model="modalItem.family_info.relation"></b-form-input>
            </b-input-group>
          </b-col>
        </b-form-row>
        <b-row class="mr-2" align-h="end">
          <b-button variant="primary" @click="onEntityFormSubmit">{{modalAction}}</b-button>
        </b-row>
      </b-form>
    </b-modal>
    <b-modal v-if="itemType === 'vehicle'" ref="entity-modal-vehicle" hide-footer title="Edit Vehicle">
      <div class="d-block text-left">
        <h3>{{modalAction}} Vehicle</h3>
      </div>
      <b-form>
        <b-form-row class="mb-2 mr-2">
          <b-col sm="6">
            <b-input-group prepend="Make">
              <b-form-input id="vehicle-input-make'" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.make"></b-form-input>
            </b-input-group>
          </b-col>
          <b-col sm="6">
            <b-input-group prepend="Model">
              <b-form-input id="vehicle-input-model" placeholder="???" v-model="modalItem.model"></b-form-input>
            </b-input-group>
          </b-col>
        </b-form-row>
        <b-form-row class="mb-2 mr-2">
          <b-col sm="6">
            <b-input-group prepend="Color">
              <b-form-input id="vehicle-input-color'" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.color"></b-form-input>
            </b-input-group>
          </b-col>
          <b-col sm="6">
            <b-input-group prepend="Plate #">
              <b-form-input id="vehicle-input-plate" placeholder="???" v-model="modalItem.plate_number"></b-form-input>
            </b-input-group>
          </b-col>
        </b-form-row>
        <b-row class="mr-2" align-h="end">
          <b-button variant="primary" @click="onEntityFormSubmit">{{modalAction}}</b-button>
        </b-row>
      </b-form>
    </b-modal>
    <b-modal v-if="itemType === 'ward'" ref="entity-modal-ward" hide-footer title="Edit Ward">
      <div class="d-block text-left">
        <h3>{{modalAction}} Ward</h3>
      </div>
      <b-form>
        <b-form-row class="mb-2 mr-2">
          <b-col sm="6">
            <b-input-group prepend="First Name">
              <b-form-input id="ward-input-first-name'" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.first_name"></b-form-input>
            </b-input-group>
          </b-col>
          <b-col sm="6">
            <b-input-group prepend="Last Name">
              <b-form-input id="ward-input-last-name" placeholder="???" v-model="modalItem.last_name"></b-form-input>
            </b-input-group>
          </b-col>
        </b-form-row>
        <b-form-row>
          <b-col sm="4">
            <b-input-group prepend="Group">
              <b-form-input id="ward-input-group" placeholder="???" v-model="modalItem.group"></b-form-input>
            </b-input-group>
          </b-col>
        </b-form-row>
        <b-row class="mr-2" align-h="end">
          <b-button variant="primary" @click="onEntityFormSubmit">{{modalAction}}</b-button>
        </b-row>
      </b-form>
    </b-modal>
  </div>
</template>

<script>
import config from "../../../config";
const queryString = require("query-string");
export default {
  name: "FamilyInfoCard",
  props: { itemType: String, entities: Array },
  data() {
    var modalItem = this.getNewModalItem();
    return {
      modalItem: modalItem,
      modalAction: "Update",
      currentRegCode: null,
      regCodePhoneNum: "",
    };
  },

  computed: {
    fields() {
      if (this.itemType === "guardian") {
        return [
          "guardian_name",
          "phone",
          "email",
          "relation",
          "reg_code_sent",
          "status",
          "actions",
        ];
      }
      if (this.itemType === "vehicle") {
        return ["make", "model", "color", "plate_number", "actions"];
      }
      if (this.itemType === "ward") {
        return ["name", "group", "actions"];
      }
      return [];
    },

    items() {
      if (this.itemType === "guardian") {
        return this.entities.map((entity) => {
          return this.mapMemberToItem(entity);
        });
      }
      if (this.itemType === "vehicle") {
        return this.entities.map((entity) => {
          return this.mapVehicleToItem(entity);
        });
      }
      if (this.itemType === "ward") {
        return this.entities.map((entity) => {
          return this.mapWardToItem(entity);
        });
      }
      return [];
    },
    entityModalName() {
      if (this.itemType === "guardian") {
        return "entity-modal-guardian";
      }
      if (this.itemType === "vehicle") {
        return "entity-modal-vehicle";
      }
      if (this.itemType === "ward") {
        return "entity-modal-ward";
      }
      return [];
    },
    title() {
      return this.itemType.charAt(0).toUpperCase() + this.itemType.slice(1)
    },
  },
  created() {},
  watch: {
    entities: function () {
      if (this.items.length === 0) {
        this.modalItem = this.getNewModalItem();
      }
    },
  },
  methods: {
    mapMemberToItem(member) {
      var reg_code_sent = member.status === 0 ? false : true;
      var regCodeSentDisplay = reg_code_sent ? "YES" : "NO";
      var statusDisplay = member.status === 2 ? "activated" : "assigned";
      return {
        self: member,
        guardian_name: member.first_name + " " + member.last_name,
        phone: member.phone_num,
        email: member.email,
        relation: member.relation,
        device_id: member.device_id,
        reg_code_sent: regCodeSentDisplay,
        status: statusDisplay,
      };
    },
    mapVehicleToItem(vehicle) {
      return {
        self: vehicle,
        make: vehicle.make,
        model: vehicle.model,
        color: vehicle.color,
        plate_number: vehicle.plate_num,
      };
    },
    mapWardToItem(ward) {
      return {
        self: ward,
        name: ward.first_name + " " + ward.last_name,
        group: ward.group,
      };
    },
    onCheckRegCode(item) {
      // console.log(`gItem - ${JSON.stringify(gItem)}`);
      this.getRegCodeByMemberIDFromDB(item.self._id);
      this.modalItem = item.self;
      this.regCodePhoneNum = item.self.phone_num;
      this.$refs["reg-code-modal"].show();
    },
    onSendRegCode() {
      this.sendRegCodeWithSMS();
      this.$refs["reg-code-modal"].hide();
    },
    onCreateEntity() {
      this.modalAction = "Create";
      this.modalItem = this.getNewModalItem();
      this.$refs[this.entityModalName].show();
    },
    onEditEntity(item) {
      this.modalAction = "Update";
      this.modalItem = item.self;
      this.$refs[this.entityModalName].show();
    },
    onEntityFormSubmit() {
      this.$refs[this.entityModalName].hide();
      if (this.modalAction === "Update") {
        this.$emit("need-entity-update", this.modalItem);
      } else if (this.modalAction === "Create") {
        this.$emit("need-entity-create", this.modalItem);
      }
    },
    onDeleteEntity(item) {
      this.$emit("need-entity-delete", item.self._id)
    },
    // RegCode requests
    getRegCodeByMemberIDFromDB(memberID) {
      var _this = this;
      const http = new XMLHttpRequest();
      const queryParams = { memberID: memberID };
      const queryArgs = queryString.stringify(queryParams);
      const query = config.API_LOCATION + "reg-code?" + queryArgs;
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("get regcode succeed!");
          _this.currentRegCode = JSON.parse(this.responseText).data;
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
    sendRegCodeWithSMS() {
      var _this = this;
      const http = new XMLHttpRequest();
      const requestBody = {
        id: this.currentRegCode._id,
        phone_num: this.regCodePhoneNum,
        first_name: this.modalItem.first_name,
      };
      const query = config.API_LOCATION + "reg-code/sms";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          var response = JSON.parse(this.responseText);
          _this.getFamilyByIdFromDb();
          console.log("send regcode with SMS succeed!");
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
        http.send(JSON.stringify(requestBody));
      } catch (e) {
        alert(e);
      }
    },
    getNewModalItem() {
      if (this.itemType === "guardian") {
        return this.getNewModalMemberItem();
      }
      if (this.itemType === "vehicle") {
        return this.getNewModalVehicleItem();
      }
      if (this.itemType === "ward") {
        return this.getNewModalWardItem();
      }
    },
    getNewModalMemberItem() {
      return {
        phone_num: "",
        email: "",
        first_name: "",
        last_name: "",
        family_info: {
          relation: "",
        },
      };
    },
    getNewModalVehicleItem() {
      return {
        make: "",
        model: "",
        color: "",
        plate_num: "",
      };
    },
    getNewModalWardItem() {
      return {
        first_name: "",
        last_name: "",
        group: "",
      };
    },
  },
};
</script>