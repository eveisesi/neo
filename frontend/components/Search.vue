<template>
    <client-only>
        <vue-bootstrap-typeahead
            class="mt-1"
            style="min-width: 250px"
            v-model="term"
            :data="results"
            :serializer="item => item.name"
            @hit="handleSelection"
            placeholder="Search"
        >
            <template
                slot="suggestion"
                slot-scope="{data, htmlText}"
            >
                <div class="d-flex align-items-center">
                    <img
                        class="rounded-circle"
                        :src="image+data.image+'?size=32'"
                        style="width: 32px; height: 32px;"
                    />
                    <span
                        class="ml-4"
                        v-html="htmlText"
                    />
                </div>
            </template>
        </vue-bootstrap-typeahead>
    </client-only>
</template>

<script>
import { EVEONLINE_IMAGE } from "../util/const/urls";
import _ from "underscore";

const APP_URL = process.env.apiURL;

export default {
    data() {
        return {
            image: EVEONLINE_IMAGE,
            results: [],
            term: "",
        };
    },
    methods: {
        getResults() {
            if (this.term.length <= 1) {
                return;
            }
            console.log(APP_URL);
            this.$axios
                .get(`${APP_URL}/search?term=${this.term}`)
                .then((response) => {
                    this.results = response.data;
                });
            return;
        },
        handleSelection(item) {
            this.term = "";
            this.$router.push(`/${item.type}/${item.id}`);
            return;
        },
    },
    watch: {
        term: _.debounce(function () {
            this.getResults();
        }, 500),
    },
};
</script>

<style>
</style>