import Vue from 'vue'
import Vuex from 'vuex'
import VuexPersistence from 'vuex-persist'
import BootstrapVue from "bootstrap-vue"
import VueRouter from 'vue-router'
import App from './App.vue'
import "bootstrap/dist/css/bootstrap.min.css"
import "bootstrap-vue/dist/bootstrap-vue.css"
import routes from './router'

Vue.config.productionTip = false
Vue.use(BootstrapVue)
Vue.use(VueRouter)
Vue.use(Vuex)

const router = new VueRouter({
  routes,
  mode: "history",
})

const vuexSession = new VuexPersistence({
  storage: window.sessionStorage
})

const store = new Vuex.Store({
  state: {
    activeUser: null,
    institution: null,
    loggedInFamily: null,
    loggedInMember: null,
    loadedFamilyWithMembers: null,
    wardIndexForCheckIn: -1,
  },
  mutations: {
    setActiveUser (state, u) {
      console.log(`setActiveUser - ${JSON.stringify(u)}`)
      state.activeUser = u
    },
    setInstitution (state, i) {
      console.log(`setInstitution - ${JSON.stringify(i)}`)
      state.institution = i
    },
    setLoadedFamilyWithMembers (state, f) {
      console.log(`setLoadedFamily - ${JSON.stringify(f)}`)
      state.loadedFamilyWithMembers = f
    },
    setLoggedInFamily (state, f) {
      console.log(`setLoggedInFamily - ${JSON.stringify(f)}`)
      state.loggedInFamily = f
    },
    setLoggedInMember(state, m) {
      console.log(`setLoggedInMember - ${JSON.stringify(m)}`)
      state.loggedInMember = m
    },
    setWardIndexForCheckIn (state, index) {
      console.log(`setWardIndexForCheckIn - ${JSON.stringify(index)}`)
      state.wardIndexForCheckIn = index
    },
    resetInstitution(state) {
      console.log('institution reset to null')
      state.institution = null
    },
    resetLoadedFamilyWithMembers (state) {
      console.log('loadedFamily reset to null')
      state.loadedFamilyWithMembers = null
    },
    resetLoggedInFamily (state) {
      console.log('loggedInFamily reset to null')
      state.loggedInFamily = null
    },
    resetLoggedInMember (state) {
      console.log('loggedInMember reset to null')
      state.loggedInMember = null
    },
    resetWardIndexForCheckIn (state) {
      console.log('wardIndexForCheckIn reset to -1')
      state.wardIndexForCheckIn = -1
    }
  },
  plugins: [vuexSession.plugin]
})

new Vue({
  router,
  store,
  render: h => h(App),
}).$mount('#app')
