<template>
  <b-container class="h-100 ml-2 mr-2" fluid>
    <b-row class="ml-2 mt-4" style="margin-bottom: 30px;" align-h="between">
      <b-col xl="3" lg="4">
        <h2>Tags Dashboard</h2>
      </b-col>
      <b-col xl="2" md="4">
        <b-button variant="primary" @click="onOpenTagImportModal()">Import Tags</b-button>
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
              <b-button class="mr-2 mb-2" variant="info" :disabled="disableSingleAction" @click="onEditSelected">Edit</b-button>
              <b-button class="mr-2 mb-2" variant="danger" :disabled="disableManyAction" @click="onDeleteSelected">Delete</b-button>
              <b-button class="mr-2 mb-2" variant="outline-secondary" :disabled="disableManyAction" @click="onClearSelect">Clear Select</b-button>
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
        <b-modal ref="edit-modal" hide-footer title="Edit Tag">
          <div class="d-block text-left">
            <h3>Update Tag</h3>
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
                <b-button variant="primary" @click="onFormSubmit()">Update</b-button>
              </b-row>
            </b-form>
          </div>
        </b-modal>
      </b-container>
      <b-modal ref="import-modal" hide-footer title="Import Tags">
        <h3>Select a CSV File</h3>
        <b-form-file v-model="tagCSVFile" plain></b-form-file>
        <b-row class="mr-1" align-h="end">
          <b-button variant="primary" @click="onImportTags">Import</b-button>
        </b-row>
      </b-modal>
    </b-card>
  </b-container>
</template>

<script>
import config from "../../../config";
const queryString = require("query-string");
export default {
  name: "TagDashboard",
  data() {
    var modalItem = this.getNewTagModalItem();
    return {
      instId: "",
      loggedInToken: "",
      currentPage: 1,
      perPage: 20,
      fields: ["tag_number", "name", "group", "phone", "email"],
      items: [],
      selectedItems: [],
      nameFilter: "",
      filterOn: ["tag_number"],
      modalItem: modalItem,
      tagCSVFile: null,
    };
  },
  mounted() {
    // console.log(`mode: ${this.mode}`)
    // console.log(`isTagMode: ${this.isTagMode}`)
    var activeUser = this.$store.state.activeUser;
    this.loggedInToken = this.$store.state.loggedInToken
    if (activeUser != null) {
      console.log(
        `Member Dashboard mounted!, instId - ${activeUser.institution_id}`
      );
      this.instId = activeUser.institution_id;
    }
    this.getTagsFromDb();
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
  },
  methods: {
    mapTagToItem(tag) {
      return {
        self: tag,
        tag_number: tag.tag_string,
        name: tag.first_name + " " + tag.last_name,
        group: tag.group,
        phone: tag.phone_num,
        email: tag.email,
      };
    },
    onRowSelected(items) {
      // handle "undefined" error when clicking the same item multiple times
      if (items[0] === undefined) return;
      this.selectedItems = items;
    },
    onClearSelect() {
      this.selectedItems = [];
      this.$refs.dashboardTable.clearSelected();
    },
    onEditSelected() {
      var item = this.selectedItems[0];
      if (item != null) {
        this.modalItem = item.self;
        console.log(
          `onEditMember - modalItem: ${JSON.stringify(this.modalItem)}`
        );
        this.$refs["edit-modal"].show();
      }
    },
    onDeleteSelected() {
      var IDListToDelete = this.selectedItems.map((item) => {
        return item.self._id;
      })
      IDListToDelete.forEach((idToDelete) => {
        this.deleteTagByIdInDb(idToDelete);
      })
      setTimeout(() => {
        this.getTagsFromDb();
      }, 500)

    },
    onFormSubmit() {
      this.updateTagToDb();
    },
    onOpenTagImportModal(){
      this.$refs["import-modal"].show();
    },
    onImportTags() {
      this.$refs["import-modal"].hide();
      this.importTagsToDb();
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
    getTagsFromDb() {
      var _this = this;
      const queryParams = { instID: this.instId };
      const queryArgs = queryString.stringify(queryParams);

      var query = config.API_LOCATION + "tags?" + queryArgs;

      const http = new XMLHttpRequest();
      console.log(`Members: getMembers query -  ${query}`);
      http.open("GET", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // console.log(this.responseText);
          //// if data received with no error, add to table
          var tags = JSON.parse(this.responseText).data;
          // console.log(JSON.stringify(_this.families));
          if (tags) {
            _this.items = tags.map((tag) => {
              return _this.mapTagToItem(tag);
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
    updateTagToDb() {
      var tagToSubmit = this.modalItem;
      var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "tag/" + this.modalItem._id;
      console.log(`Tags: update Tags query -  ${query}`);
      http.open("PUT", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("update succeed!");
          _this.getTagsFromDb();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        http.send(JSON.stringify(tagToSubmit));
      } catch (e) {
        alert(e);
      }
    },
    deleteTagByIdInDb(tagID) {
      var _this = this;
      const http = new XMLHttpRequest();
      const query = config.API_LOCATION + "tag/" + tagID;
      // console.log(`Tags: update Tags query -  ${query}`);
      http.open("DELETE", query, true);
      http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
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
    importTagsToDb() {
      var _this = this;
      const queryParams = { instID: this.instId };
      const queryArgs = queryString.stringify(queryParams);
      const query = config.API_LOCATION + "import/tags?" + queryArgs;
      console.log(`Tags: import Tags query -  ${query}`);
      var formData = new FormData();
      formData.append('file', this.tagCSVFile)

      const http = new XMLHttpRequest();
      http.open("POST", query, true);
      // http.setRequestHeader("Content-Type", "application/json;charset=UTF-8");
      http.setRequestHeader("Authorization", `Bearer ${this.loggedInToken}`);
      http.onreadystatechange = function () {
        if (this.readyState === 4 && this.status === 200) {
          // var response = JSON.parse(this.responseText);
          console.log("import succeed!");
          _this.getTagsFromDb();
          _this.$bvToast.toast(this.responseText, {
            title: "DataBase Message",
            autoHideDelay: 2000,
          });
        } else if (this.readyState === 4) {
          alert(this.responseText);
        }
      };
      try {
        http.send(formData);
      } catch (e) {
        alert(e);
      }
    },
    getNewTagModalItem() {
      return {
        tag_string: "",
        first_name: "",
        last_name: "",
        group: "",
        email: "",
      };
    },
  },
};
</script>