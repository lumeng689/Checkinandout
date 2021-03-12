<template>
<b-container class="h-100 ml-2 mr-2" fluid>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;" align-h="between">
      <b-col xl="3" lg="4">
        <h2>Family Dashboard</h2>
      </b-col>
      <b-col xl="2" md="4">
        <b-button  variant="primary" to="./add-family">Add Family</b-button>
      </b-col>
    </b-row>
    <b-card class="h-100 ml-4 mr-4">
      <b-container class="h-100" fluid>
        <b-row>
          <b-col xl="3" lg="4" class="my-1">
            <b-form-group label="Contact Name" label-for="contact-name-filter-input" label-cols-sm="4" label-align-sm="right" label-size="sm">
              <b-input-group>
                <b-form-input id="contact-name-filter-input" v-model="contactNameFilter" type="search" placeholder="Type to Search"></b-form-input>
              </b-input-group>
            </b-form-group>
          </b-col>
          <b-col xl="2" lg="4" class="my-1">
            <b-form-group label="Reg Code Not Sent Only" label-for="reg-code-filter-input" label-cols-sm="9" label-align-sm="right" label-size="sm">
              <b-form-checkbox id="reg-code-filter-input" v-model="regCodeSentFilter"></b-form-checkbox>
            </b-form-group>
          </b-col>
          <b-col xl="7" lg="12" class="my-1" align-self="end">
            <b-row align-h="end">
              <b-button class="mr-2 mb-2" variant="info" :disabled="disableSingleAction" @click="onViewFamily">View Info</b-button>
              <b-button class="mr-2 mb-2" variant="danger" :disabled="disableManyAction" @click="onDeleteFamilies">Delete</b-button>
              <b-button class="mr-2 mb-2" variant="outline-secondary" :disabled="disableManyAction" @click="onClearSelect">Clear Select</b-button>
            </b-row>
          </b-col>
        </b-row>
        <b-table
          bordered
          sticky-header="80%"
          selectable
          id="family-table"
          ref="familyTable"
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
        <b-pagination v-model="currentPage" :total-rows="numRows" :per-page="perPage" aria-controls="ward-table"></b-pagination>
      </b-container>
    </b-card>
  </b-container>
</template>
<script>
import config from "../../../config";
const queryString = require("query-string");
export default {
  name: "FamilyDashboard",
  data() {
    return {
      instId: "",
      currentPage: 1,
      perPage: 20,
      fields: ["family_contact", "phone", "relation", "reg_code_sent"],
      items: [],
      selectedItems: [],
      regCodeSentFilter: false,
      contactNameFilter: "",
      filterOn: ["reg_code_sent", "family_contact"],
    };
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
      return this.regCodeSentFilter + " " + this.contactNameFilter;
    },
  },
  mounted() {
    var activeUser = this.$store.state.activeUser;
    if (activeUser != null) {
      console.log(
        `Family Dashboard mounted!, instId - ${activeUser.institution_id}`
      );
      this.instId = activeUser.institution_id;
    }
    this.getFamiliesFromDb();
  },
  methods: {
    mapFamilyToItem(family) {
      var contact = family.contact_member_info;
      return {
        self: family,
        family_contact: contact.name,
        phone: contact.phone_num,
        relation: contact.relation,
        reg_code_sent: family.all_reg_code_sent,
      };
    },
    onRowSelected(items) {
      // handle "undefined" error when clicking the same item multiple times
      if (items[0] === undefined) return;
      this.selectedItems = items;
    },
    onClearSelect() {
      this.selectedItems = [];
      this.$refs.familyTable.clearSelected();
    },
    onViewFamily() {
      var _this = this
      var familyID = this.selectedItems[0].self._id
      this.getFamilyWithMemberByIDFromDB(familyID, (familyWithMembers) => {
        _this.$store.commit("setLoadedFamilyWithMembers", familyWithMembers)
        _this.$router.push("/portal/families/info");
      });
    },
    onDeleteFamilies() {
      // // var _this = this
      // this.selectedItems.forEach((item) => {
      //   var familyID = item.self._id
      //   this.deleteFamilyFromDB(familyID) 
      // })
      // setTimeout(() => {
      //   this.getFamiliesFromDb();
      // }, 500)
    },
    onFiltered(filteredItems) {
      this.totalRows = filteredItems.length;
      this.currentPage = 1;
      this.onClearSelect();
    },
    filterFunction(row) {
      // console.log(`row: ${JSON.stringify(row)}, filterString: ${filterString}`)
      var regCodePred = true;
      if (this.regCodeSentFilter) {
        regCodePred = !row.reg_code_sent;
      }
      var contactNamePred = true;
      if (this.contactNameFilter.length > 0) {
        contactNamePred = row.family_contact
          .toLowerCase()
          .includes(this.contactNameFilter);
      }
      // console.log(
      //   `regCodePred: ${regCodePred}, contactNamePred: ${contactNamePred}`
      // );
      return regCodePred && contactNamePred;
    },
    getFamiliesFromDb() {
      // console.log("CUSTOMER_TABLE: customers: " + JSON.stringify(this.customers));
      var _this = this;
      //// getting CC-Events from database
      const queryParams = { instID: this.instId };
      const queryArgs = queryString.stringify(queryParams);

      const http = new XMLHttpRequest();
      var query = config.API_LOCATION + "families?" + queryArgs;
      console.log(`CC-Families: getFamilies query -  ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText);
          //// if data received with no error, add to table
          _this.families = JSON.parse(this.responseText).data;
          // console.log(JSON.stringify(_this.families));
          if (_this.families) {
            _this.items = _this.families.map((family) => {
              return _this.mapFamilyToItem(family);
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
    getFamilyWithMemberByIDFromDB(familyID, callback) {
      // var family = this.selectedItems[0].self;
      const query = config.API_LOCATION + "family-with-members/" + familyID;
      console.log(`getFamilyWithMembersByID - query: ${query}`);
      const http = new XMLHttpRequest();
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText)
          if (this.responseText.length == 0) {
            return;
          }
          var response = JSON.parse(this.responseText).data;
          if (response && callback != null) {
            callback(response);
          }
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
    deleteFamilyFromDB(familyID) {
      var _this = this;
      const query = config.API_LOCATION + "family/" + familyID;
      const http = new XMLHttpRequest();
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
  },
};
</script>