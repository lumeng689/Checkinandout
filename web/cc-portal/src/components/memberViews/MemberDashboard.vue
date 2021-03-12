<template>
  <b-container class="h-100 ml-2 mr-2" fluid>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;" align-h="between">
      <b-col xl="3" lg="4">
        <h2>Members Dashboard</h2>
      </b-col>
      <b-col xl="2" md="4">
        <b-button variant="primary" @click="onCreateMember()">Add Member</b-button>
      </b-col>
    </b-row>
    <b-card class="h-100 ml-4 mr-4">
      <b-container class="h-100" fluid>
        <b-row>
          <b-col xl="3" lg="4" class="my-1">
            <b-form-group label="Contact Name" label-for="contact-name-filter-input" label-cols-sm="4" label-align-sm="right" label-size="sm">
              <b-input-group>
                <b-form-input id="contact-name-filter-input" v-model="nameFilter" type="search" placeholder="Type to Search"></b-form-input>
              </b-input-group>
            </b-form-group>
          </b-col>
          <b-col xl="7" lg="12" class="my-1" align-self="end">
            <b-row align-h="end">
              <b-button class="mr-2 mb-2" variant="info" :disabled="disableSingleAction" @click="onCheckRegCode">RegCode</b-button>
              <b-button class="mr-2 mb-2" variant="info" :disabled="disableSingleAction" @click="onEditSelected">Edit</b-button>
              <b-button class="mr-2 mb-2" variant="danger" :disabled="disableManyAction" @click="onDeleteSelected">Delete</b-button>
              <b-button class="mr-2 mb-2" variant="outline-secondary" :disabled="disableManyAction" @click="onClearSelected">Clear Select</b-button>
            </b-row>
          </b-col>
        </b-row>
        <b-table
          bordered
          sticky-header="80%"
          selectable
          id="dashboard-table"
          ref="dashboardTable"
          responsive="sm"
          :items="items"
          :fields="fields"
          :current-page="currentPage"
          :per-page="perPage"
          :filter="filter"
          :filter-function="filterFunction"
          :filter-included-fields="filterOn"
          @row-selected="onRowSelected"
          @filtered="onFiltered"
        ></b-table>
        <b-pagination v-model="currentPage" :total-rows="numRows" :per-page="perPage" aria-controls="dashboard-table"></b-pagination>
        <b-modal ref="edit-modal" hide-footer title="Edit Member">
          <div class="d-block text-left">
            <h3>{{modalAction}} Member</h3>
            <b-form>
              <b-form-row class="mb-2 mr-2">
                <b-col sm="6">
                  <b-input-group prepend="First Name">
                    <b-form-input id="tag-input-first-name'" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.first_name"></b-form-input>
                  </b-input-group>
                </b-col>
                <b-col sm="6">
                  <b-input-group prepend="Last Name">
                    <b-form-input id="tag-input-last-name'" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.last_name"></b-form-input>
                  </b-input-group>
                </b-col>
              </b-form-row>
              <b-input-group class="mb-2" prepend="Email">
                <b-form-input id="tag-input-email" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.email"></b-form-input>
              </b-input-group>
              <b-form-row class="mb-2 mr-2">
                <b-col sm="6">
                  <b-input-group prepend="Phone #">
                    <b-form-input id="tag-input-phone-number" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.phone_num"></b-form-input>
                  </b-input-group>
                  <b-input-group prepend="Group">
                    <b-form-input id="tag-input-group" class="mb-2 mr-sm-2 mb-sm-0" placeholder="???" v-model="modalItem.group"></b-form-input>
                  </b-input-group>
                </b-col>
              </b-form-row>
              <b-row class="mr-2" align-h="end">
                <b-button variant="primary" @click="onFormSubmit">{{modalAction}}</b-button>
              </b-row>
            </b-form>
          </div>
        </b-modal>
        <b-modal ref="reg-code-modal" hide-footer title="Check Registration Code">
          <div class="d-block test-left ml-4 mb-2">
            Registration Code for
            <b>{{modalItem.first_name + " " + modalItem.last_name}}</b> is
            <b>{{displayRegCode}}</b>
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
      </b-container>
    </b-card>
  </b-container>
</template>

<script>
import config from "../../config";
const queryString = require("query-string");
export default {
  name: "MemberDashboard",
  data() {
    var modalItem = this.getNewMemberModalItem();
    return {
      instId: "",
      // Table fields
      currentPage: 1,
      perPage: 20,
      fields: ["name", "group", "phone", "email", "status"],
      items: [],
      selectedItems: [],
      nameFilter: "",
      filterOn: ["name"],
      // Member Edit Modal Fields
      modalItem: modalItem,
      modalAction: "Update",
      // RegCode fields
      currentRegCode: null,
      currentMember: null,
      regCodePhoneNum: "",
    };
  },
  mounted() {
    // console.log(`mode: ${this.mode}`)
    // console.log(`isTagMode: ${this.isTagMode}`)
    var activeUser = this.$store.state.activeUser;
    if (activeUser != null) {
      console.log(
        `Member Dashboard mounted!, instId - ${activeUser.institution_id}`
      );
      this.instId = activeUser.institution_id;
    }
    this.getMembersFromDb();
  },
  computed: {
    disableSingleAction() {
      return this.selectedItems.length != 1;
    },
    disableManyAction() {
      return this.selectedItems.length == 0;
    },
    numRows() {
      return this.items.length;
    },
    filter() {
      return this.nameFilter;
    },
    displayRegCode() {
      if (this.currentRegCode === null) return "";
      return this.currentRegCode.reg_code;
    }
  },
  methods: {
    mapMemberToItem(member) {
      var statusDisplay = member.status === 2 ? "activated" : "assigned";
      return {
        self: member,
        name: member.first_name + " " + member.last_name,
        group: member.group,
        phone: member.phone_num,
        email: member.email,
        status: statusDisplay,
      };
    },
    onRowSelected(items) {
      // handle "undefined" error when clicking the same item multiple times
      if (items[0] === undefined) return;
      this.selectedItems = items;
    },
    onClearSelected() {
      this.selectedItems = [];
      this.$refs.dashboardTable.clearSelected();
    },
    onEditSelected() {
      var item = this.selectedItems[0];
      if (item != null) {
        this.modalAction = "Update";
        this.modalItem = item.self;
        console.log(
          `onEditMember - modalItem: ${JSON.stringify(this.modalItem)}`
        );
        this.$refs["edit-modal"].show();
      }
    },
    onCreateMember() {
      this.modalAction = "Create";
      this.modalItem = this.getNewMemberModalItem();
      this.$refs["edit-modal"].show();
    },
    onDeleteSelected() {
      var IDListToDelete = this.selectedItems.map((item) => {
        return item.self._id;
      })
      IDListToDelete.forEach((idToDelete) => {
        this.deleteMemberInDb(idToDelete);
      })
      setTimeout(() => {
        this.getMembersFromDb();
      }, 500)
      // this.deleteMemberInDb(this.selectedItems[0].self._id)
    },
    onFormSubmit() {
      if (this.modalAction === "Update") {
        this.updateMemberToDb();
      } else if (this.modalAction === "Create") {
        this.createMemberInDb();
      }
    },
    onCheckRegCode() {
      this.currentMember = this.selectedItems[0]
      console.log(`currentMember - ${JSON.stringify(this.currentMember)}`)
      this.getRegCodeByMemberIDFromDb(this.currentMember.self._id);
      this.regCodePhoneNum = this.currentMember.self.phone_num;
      this.$refs["reg-code-modal"].show();
      this.onClearSelected()
    },
    onSendRegCode() {
      this.sendRegCodeWithSMS();
      this.$refs["reg-code-modal"].hide();
    },
    onFiltered(filteredItems) {
      this.totalRows = filteredItems.length;
      this.currentPage = 1;
      this.onClearSelect();
    },
    filterFunction(row) {
      var namePred = true;
      if (this.nameFilter) {
        namePred = row.name.toLowerCase().includes(this.nameFilter);
      }
      return namePred;
    },
    getMembersFromDb() {
      var _this = this;
      const queryParams = { instID: this.instId };
      const queryArgs = queryString.stringify(queryParams);

      var query = config.API_LOCATION + "members?" + queryArgs;

      const http = new XMLHttpRequest();
      console.log(`Members: getMembers query -  ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText);
          //// if data received with no error, add to table
          var members = JSON.parse(this.responseText).data;
          console.log(JSON.stringify(members));
          if (members) {
            _this.items = members.map((member) => {
              return _this.mapMemberToItem(member);
            });
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
    updateMemberToDb() {
      var memberToSubmit = this.modalItem;
      var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "member/" + this.modalItem._id;
      console.log(`Tags: update Tags query -  ${query}`);
      http.open("PUT", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("update succeed!");
          _this.getMembersFromDb();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        http.send(JSON.stringify(memberToSubmit));
      } catch (e) {
        alert(e);
      }
    },
    createMemberInDb() {
      var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "member";
      console.log(`Tags: update Tags query -  ${query}`);
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 201) {
          // var response = JSON.parse(this.responseText);
          console.log("create succeed!");
          _this.getMembersFromDb();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        http.send(
          JSON.stringify({
            institution_id: this.instId,
            phone_num: this.modalItem.phone_num,
            email: this.modalItem.email,
            first_name: this.modalItem.first_name,
            last_name: this.modalItem.last_name,
          })
        );
      } catch (e) {
        alert(e);
      }
    },
    deleteMemberInDb(memberID) {
      var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "member/" + memberID;
      // console.log(`Tags: update Tags query -  ${query}`);
      http.open("DELETE", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("delete succeed!");
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
    // RegCode requests
    getRegCodeByMemberIDFromDb(memberID) {
      var _this = this;
      const http = new XMLHttpRequest();
      const queryParams = { memberID: memberID };
      const queryArgs = queryString.stringify(queryParams);
      const query = config.API_LOCATION + "reg-code?" + queryArgs;
      console.log(`getRegCode Query - ${memberID}`)
      console.log(`getRegCode Query - ${JSON.stringify(queryParams)}`)
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
        first_name: this.currentMember.first_name,
      };
      const query = config.API_LOCATION + "reg-code/sms";
      http.open("POST", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          var response = JSON.parse(this.responseText);
          _this.getMembersFromDb();
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
    getNewMemberModalItem() {
      return {
        first_name: "",
        last_name: "",
        group: "",
        email: "",
      };
    },
  },
};
</script>