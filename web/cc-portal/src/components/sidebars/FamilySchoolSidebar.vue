<template>
    <b-nav vertical>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/cc-records'">CC Records</b-nav-item>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/wards'">Students</b-nav-item>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/families/dashboard'">Families</b-nav-item>
        <b-nav-item class="cc-sidebar-item" active :to="'/portal/admins'">Admins</b-nav-item>
      </b-nav>
</template>


<script>
export default {
    name: "family-school-sidebar",
}
</script>