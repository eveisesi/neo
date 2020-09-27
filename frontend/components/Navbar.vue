<template>
    <b-navbar
        toggleable="lg"
        type="dark"
        variant="dark"
    >
        <b-container>
            <b-navbar-brand href="/">New Eden Obituary</b-navbar-brand>
            <Search />
        </b-container>
    </b-navbar>
</template>

<script>
import Search from "./Search";
export default {
    name: "Navbar",
    components: {
        Search,
    },
};
</script>

<style>
</style>