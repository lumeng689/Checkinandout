<template>
    <b-nav vertical>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/cc-records'">Records</b-nav-item>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/members/dashboard/0'">Employees</b-nav-item>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/admins'">Admins</b-nav-item>
      </b-nav>
</template>


<script>
export default {
    name: "member-corporate-sidebar",
}
</script>