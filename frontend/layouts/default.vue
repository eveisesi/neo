<template>
    <div>
        <Navbar />
        <Nuxt />
    </div>
</template>

<script>
import Navbar from "../components/Navbar";
import "bootswatch/dist/cyborg/bootstrap.min.css";

export default {
    components: {
        Navbar,
    },
};
</script>

<style>
</style>
