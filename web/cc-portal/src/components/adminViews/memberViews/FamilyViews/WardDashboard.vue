<template>
  <b-container class="h-100 ml-2 mr-2" fluid>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;">
      <b-col xl="3" lg="4">
        <h2>Student Dashboard</h2>
      </b-col>
    </b-row>
    <b-card class="h-100">
      <b-container class="h-100" fluid>
        <b-row algin-h="between">
          <b-col xl="2" lg="4" class="my-1">
            <b-form-group label="Patient Name" label-for="ward-name-filter-input" label-cols-sm="5" label-align-sm="right" label-size="sm">
              <b-input-group>
                <b-form-input id="ward-name-filter-input" v-model="wardNameFilter" type="search" placeholder="Type to Search"></b-form-input>
              </b-input-group>
            </b-form-group>
          </b-col>
          <b-col xl="2" lg="4" class="my-1">
            <b-form-group label="Group" label-for="group-filter-input" label-cols-sm="3" label-align-sm="right" label-size="sm">
              <b-input-group>
                <b-form-input id="group-filter-input" v-model="wardGroupFilter" type="search" placeholder="Type to Search"></b-form-input>
              </b-input-group>
            </b-form-group>
          </b-col>
          <b-col xl="2" lg="4" class="my-1">
            <b-form-group label="Guardian Name" label-for="guardian-name-filter-input" label-cols-sm="5" label-align-sm="right" label-size="sm">
              <b-input-group>
                <b-form-input id="guardian-name-filter-input" v-model="guardianNameFilter" type="search" placeholder="Type to Search"></b-form-input>
              </b-input-group>
            </b-form-group>
          </b-col>
          <b-col xl="6" lg="6" class="my-1">
            <b-row align-h="end">
              <b-button class="mr-2 mb-2" variant="success" @click="onRefresh">Refresh</b-button>
              <b-button class="mr-2 mb-2" variant="info" :disabled="disableSingleAction" @click="onViewFamily">View Family</b-button>
              <b-button v-if="false" class="mr-2 mb-2" variant="primary" :disabled="disableManyAction || disableCheckin" @click="onCheckInEvents">CheckIn</b-button>
              <b-button class="mr-2 mb-2" variant="danger" :disabled="disableManyAction" @click="onDeleteEvents">Delete</b-button>
              <b-button class="mr-2 mb-2" variant="outline-secondary" :disabled="disableManyAction" @click="onClearSelect">Clear Select</b-button>
            </b-row>
          </b-col>
        </b-row>
        <b-table
          bordered
          sticky-header="80%"
          selectable
          id="ward-table"
          ref="wardTable"
          responsive="sm"
          :items="items"
          :fields="fields"
          :current-page="currentPage"
          :per-page="perPage"
          :filter="filter"
          :filter-included-fields="filterOn"
          :filter-function="filterFunction"
          @row-selected="onRowSelected"
          @filtered="onFiltered"
        ></b-table>
        <b-pagination v-model="currentPage" :total-rows="numRows" :per-page="perPage" aria-controls="ward-table"></b-pagination>
      </b-container>
    </b-card>
  </b-container>
</template>
<script>
import config from "../../../../config";
const queryString = require("query-string");
export default {
  name: "Wards",
  data() {
    return {
      instId: "",
      loggedInToken: "",
      currentPage: 1,
      perPage: 20,
      wards: [],
      fields: ["name", "group", "guardian_name", "guardian_phone"],
      items: [],
      selectedItems: [],
      wardGroupFilter: "",
      wardNameFilter: "",
      guardianNameFilter: "",
      filterOn: ["group", "name", "guardian_name"],
    };
  },

  mounted() {
    // console.log("Wards Dashboard mounted")

    var activeUser = this.$store.state.activeUser;
    this.loggedInToken = this.$store.state.loggedInToken;
    if (activeUser != null) {
      console.log(
        `Wards Dashboard mounted!, instId - ${activeUser.institution_id}`
      );
      this.instId = activeUser.institution_id;
    }
    this.getWardsFromDb();
  },
  computed: {
    disableSingleAction() {
      return this.selectedItems.length != 1;
    },
    disableManyAction() {
      return this.selectedItems.length == 0;
    },
    disableCheckin() {
      return true;
    },
    numRows() {
      return this.items.length;
    },
    filter() {
      return (
        this.wardGroupFilter +
        " " +
        this.wardNameFilter +
        " " +
        this.guardianNameFilter
      );
    },
  },
  methods: {
    mapWardToItem(ward) {
      return {
        _id: ward._id,
        name: ward.name,
        group: ward.group,
        guardian_name: ward.guardian_name,
        guardian_phone: ward.phone_num,
      };
    },
    onRowSelected(items) {
      // handle "undefined" error when clicking the same item multiple times
      if (items[0] === undefined) return;
      this.selectedItems = items;
    },
    onRefresh() {
      this.getWardsFromDb();
    },
    onViewFamily() {
      this.getFamilyByWardIdFromDb(this.selectedItems[0], (family) => {
        if (family) {
          this.$store.commit("setLoadedFamily", family);
          this.$router.push("../families/info");
        }
      });
    },
    onCheckInEvents() {},
    onDeleteEvents() {
      var IDListToDelete = this.selectedItems.map((item) => {
        return item.self._id;
      })
      IDListToDelete.forEach((idToDelete) => {
        this.deleteWardFromDb(idToDelete);
      })
      setTimeout(() => {
        this.getWardsFromDb();
      }, 500)
      
    },
    onClearSelect() {
      (this.selectedItems = []), this.$refs.wardTable.clearSelected();
    },
    onFiltered(filteredItems) {
      this.totalRows = filteredItems.length;
      this.currentPage = 1;
      this.onClearSelect();
    },
    filterFunction(row) {
      // console.log(`row: ${JSON.stringify(row)}`)
      var wardGroupPred = true;
      var wardNamePred = true;
      var guardianNamePred = true;
      if (this.wardGroupFilter.length > 0) {
        wardGroupPred = row.group.toLowerCase().includes(this.wardGroupFilter);
      }
      if (this.wardNameFilter.length > 0) {
        wardNamePred = row.name.toLowerCase().includes(this.wardNameFilter);
      }
      if (this.guardianNameFilter.length > 0) {
        guardianNamePred = row.guardian_name
          .toLowerCase()
          .includes(this.guardianNameFilter);
      }
      console.log(
        `expiredPred: ${wardGroupPred}, tempPred: ${wardNamePred}, groupPred: ${guardianNamePred}, `
      );
      return wardGroupPred && wardNamePred && guardianNamePred;
    },
    getWardsFromDb() {
      // console.log("CUSTOMER_TABLE: customers: " + JSON.stringify(this.customers));
      var _this = this;
      //// getting Families from database
      const queryParams = { instID: this.instId };
      const queryArgs = queryString.stringify(queryParams);
      const http = new XMLHttpRequest();
      var query = config.API_LOCATION + "families?" + queryArgs;
      console.log(`Wards: get Wards query -  ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText);
          //// if data received with no error, add to table
          var families = JSON.parse(this.responseText).data;
          // console.log(JSON.stringify(_this.families));
          if (families) {
            _this.wards = _this.getWardsFromFamilies(families);
            _this.items = _this.wards.map((ward) => {
              return _this.mapWardToItem(ward);
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
    getFamilyByWardIdFromDb(item, callback) {
      // var _this = this;
      const queryParams = { wardID: item._id };
      const queryArgs = queryString.stringify(queryParams);
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "family?" + queryArgs;
      console.log(`CC-Records: getCCRecord query -  ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText);
          //// if data received with no error, add to table
          var family = JSON.parse(this.responseText).data;
          if (family && callback != null) {
            callback(family);
          }
          // console.log("Record_Table: items: " + JSON.stringify(this.items));
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
    deleteWardFromDb(item) {
      // Send a Delete Request per each of selected Items
      var _this = this;
      const http = new XMLHttpRequest();
      var query = config.API_LOCATION + "ward/" + item._id;
      console.log(`CC-Records: delete CCRecord query -  ${query}`);
      http.open("DELETE", query, true);

      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText);
          var response = JSON.parse(this.responseText);

          if (response.message != undefined) {
            _this.$bvToast.toast(response.message, {
              title: "DataBase Message",
              autoHideDelay: 2000,
            });
            // console.log("Record_Table: items: " + JSON.stringify(this.items));
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
    getWardsFromFamilies(families) {
      var wardsWithContactInfo = [];
      for (const family of families) {
        var wsWithContactInfo = this.getWardWithContactInfo(family);
        wardsWithContactInfo.push(...wsWithContactInfo);
      }
      return wardsWithContactInfo;
    },
    getWardWithContactInfo(family) {
      var contact = family.contact_member_info
      if (contact === null) return null;
      console.log(`getWardWithContactInfo - contactInfo: ${contact}`);
      return family.wards.map((ward) => {
        return {
          _id: ward._id,
          name: ward.first_name + " " + ward.last_name,
          group: ward.group,
          guardian_name: contact.name,
          phone_num: contact.phone_num,
        };
      });
    },
    
  },
};
</script>