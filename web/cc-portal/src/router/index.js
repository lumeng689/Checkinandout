import CCPortal from "../components/CCPortal"
import CCRecords from "../components/CCRecordDashboard";
import TagDashboard from "../components/memberViews/TagDashboard"
import MemberDashboard from "../components/memberViews/MemberDashboard"
import FamilyDashboard from "../components/memberViews/FamilyViews/FamilyDashboard";
import WardDashboard from "../components/memberViews/FamilyViews/WardDashboard";
import AddFamily from "../components/memberViews/FamilyViews/AddFamily";
import FamilyInfo from "../components/memberViews/FamilyViews/FamilyInfo";
import Admins from "../components/Admins";
import AdminProfile from "../components/AdminProfile"
import AdminSettings from "../components/AdminSettings"
import Mobile from "../components/mobile/Mobile"
import MobileRegistration from "../components/mobile/MobileRegistration"
import MobileLogin from "../components/mobile/MobileLogin"
import MobileActivate from "../components/mobile/MobileActivate"
import MobileContainer from "../components/mobile/MobileContainer"
import AddInstitution from "../components/AddInstitution"
import AddAdmin from "../components/AddAdmin"
import RedirectCCPortal from "../components/RedirectCCPortal"
import CCPortalLogin from "../components/CCPortalLogin"
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
        path:"/mobile/registration",
        component: MobileRegistration,
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
