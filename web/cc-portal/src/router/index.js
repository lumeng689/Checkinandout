import CCPortal from "../components//adminViews/CCPortal"
import CCPortalLogin from "../components/adminViews/CCPortalLogin"
import CCRecords from "../components/adminViews/CCRecordDashboard";
import TagDashboard from "../components/adminViews/memberViews/TagDashboard"
import MemberDashboard from "../components/adminViews/memberViews/MemberDashboard"
import FamilyDashboard from "../components/adminViews/memberViews/FamilyViews/FamilyDashboard";
import WardDashboard from "../components/adminViews/memberViews/FamilyViews/WardDashboard";
import AddFamily from "../components/adminViews/memberViews/FamilyViews/AddFamily";
import FamilyInfo from "../components/adminViews/memberViews/FamilyViews/FamilyInfo";
import Admins from "../components/adminViews/Admins";
import AdminProfile from "../components/adminViews/AdminProfile"
import AdminSettings from "../components/adminViews/AdminSettings"
import Mobile from "../components/mobile/Mobile"
import MobileRegistration from "../components/mobile/MobileRegistration"
import MobileLogin from "../components/mobile/MobileLogin"
import MobileActivate from "../components/mobile/MobileActivate"
import MobileContainer from "../components/mobile/MobileContainer"
import AddInstitution from "../components/superAdminViews/AddInstitution"
import AddAdmin from "../components/superAdminViews/AddAdmin"
import RedirectCCPortal from "../components/superAdminViews/RedirectCCPortal"
const routes = [
  {
    path: "/portal",
    name: "CCPortal",
    component: CCPortal,
    children: [
      {
        path: "/portal/cc-records",
        component: CCRecords,
      },
      {
        path: "/portal/members/dashboard",
        component: MemberDashboard,
        props: true,
      },
      {
        path: "/portal/tags/dashboard",
        component: TagDashboard,
        props: true,
      },
      {
        path: "/portal/families/dashboard",
        component: FamilyDashboard,
      },
      {
        path: "/portal/wards/",
        component: WardDashboard,
      },
      {
        path: `/portal/families/add-family`,
        component: AddFamily,
      },
      {
        path: `/portal/families/info`,
        component: FamilyInfo,
      },
      {
        path: "/portal/admins",
        component: Admins,
      },
      {
        path: "/portal/profile",
        component: AdminProfile,
      },
      {
        path: "/portal/settings",
        component: AdminSettings,
      },
    ]
  },
  {
    path: "/mobile",
    name: "Mobile",
    component: MobileContainer,
    children: [
      {
        path: "/mobile/home",
        component: Mobile,
      },
      {
        path:"/mobile/activate",
        component: MobileActivate,
      },
      {
        path:"/mobile/login",
        component: MobileLogin,
      },
      {
        path:"/mobile/registration/:instID",
        component: MobileRegistration,
        props: true,
      }
      
    ]
  },
  {
    path: "/user-login",
    name: "CCPortalLogin",
    component: CCPortalLogin,
  },
  {
    path: "/super-admin/add-institution",
    name: "AddInstitution",
    component: AddInstitution,
  },
  {
    path: "/super-admin/add-admin",
    name: "AddAdmin",
    component: AddAdmin,
  },
  {
    path: "/super-admin/redirect",
    name: "RedirectCCPortal",
    component: RedirectCCPortal,
  }
];

export default routes;
