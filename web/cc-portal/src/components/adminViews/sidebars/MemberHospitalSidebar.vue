<template>
    <b-nav vertical>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/cc-records'">Records</b-nav-item>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/members/dashboard'">Patients</b-nav-item>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/admins'">Admins</b-nav-item>
      </b-nav>
</template>


<script>
export default {
    name: "member-hospital-sidebar",
}
</script>