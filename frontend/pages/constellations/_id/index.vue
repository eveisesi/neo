<template>
    <b-container>
        <b-row>
            <b-col md="6">
                <ComponentLoading v-if="$apollo.queries.information.loading" />
                <Error v-else-if="$apollo.queries.information.error" />
                <b-table-simple v-else>
                    <b-tbody>
                        <tr>
                            <td
                                rowspan="4"
                                width="130"
                            >
                                <b-img
                                    :src="image+'types/9/render?size=128'"
                                    rounded
                                    fluid
                                    height="128"
                                    width="128"
                                />
                            </td>
                            <b-td>Constellation</b-td>
                            <b-td>{{information.name}}</b-td>
                        </tr>

                        <b-tr v-if="information.region">
                            <b-td>Region</b-td>
                            <b-td>
                                <router-link :to="'/regions/' + information.region.id">{{information.region.name}}</router-link>
                            </b-td>
                        </b-tr>
                    </b-tbody>
                </b-table-simple>
            </b-col>
        </b-row>
        <b-row>
            <b-col md="12">
                <h4 class="text-center">Most Valuable Kills - Last 7 Days</h4>
                <hr style="background-color: white" />
                <ComponentLoading v-if="$apollo.queries.mv.loading" />
                <Error v-else-if="$apollo.queries.mv.error" />
                <KillmailHighlight
                    :mv="mv"
                    v-else
                />
            </b-col>
        </b-row>
        <b-row>
            <b-col sm="12">

                <div>
                    <div class="float-right mt-2">
                        <b-pagination
                            v-model="compPage"
                            total-rows="1000"
                            per-page="50"
                            @change="handlePagination"
                            hide-ellipsis
                        ></b-pagination>
                    </div>
                    <h3>Recent Activity</h3>
                    <hr style="background-color: white" />
                    <ComponentLoading v-if="$apollo.queries.killmails.loading" />
                    <Error v-else-if="$apollo.queries.killmails.error" />
                    <KillTable
                        :killmails="killmails"
                        :target="information.id"
                        v-else
                    />
                    <hr style="background-color: white" />
                    <b-pagination
                        v-model="compPage"
                        total-rows="1000"
                        per-page="50"
                        @change="handlePagination"
                        hide-ellipsis
                        align="center"
                    ></b-pagination>
                </div>
            </b-col>
        </b-row>
    </b-container>
</template>

<script>
import {
    KILLMAILS,
    CONSTELLATION_INFORMATION,
    MOST_VALUABLE,
} from "../../../util/queries";
import { EVEONLINE_IMAGE } from "../../../util/const/urls";
import head from "../../../util/head";

export default {
    watchQuery: ["page"],
    validate({ query }) {
        return query.page != undefined && !query.page > 20;
    },
    data() {
        return {
            information: {},
            killmails: [],
            mv: [],
            image: EVEONLINE_IMAGE,
        };
    },
    validate({ query }) {
        return query.page == undefined || query.page <= 20;
    },
    head() {
        return {
            title: head(this.information.name, "Constellation"),
        };
    },
    key: (to) => to.fullPath,
    apollo: {
        information: {
            query: CONSTELLATION_INFORMATION,
            variables() {
                return {
                    id: this.$router.currentRoute.params.id,
                };
            },
            result(result, key) {
                this.information = result.data.information;
            },
            error(result, key) {
                this.error = JSON.stringify(result.message);
            },
        },
        killmails: {
            query: KILLMAILS,
            variables() {
                const page = this.$router.currentRoute.query.page
                    ? this.$router.currentRoute.query.page
                    : 1;
                return {
                    entity: "constellation",
                    id: this.$router.currentRoute.params.id,
                    page: page,
                };
            },
            result(result, key) {
                this.killmails = result.data.killmails;
            },
            error(result, key) {
                this.error = JSON.stringify(result.message);
            },
        },
        mv: {
            query: MOST_VALUABLE,
            variables() {
                return {
                    category: "kill",
                    type: "constellation",
                    id: this.$router.currentRoute.params.id,
                    age: 7,
                    limit: 6,
                };
            },
            result(result, key) {
                this.mv = result.data.mv;
            },
            error(result, key) {
                this.error = JSON.stringify(result.message);
            },
        },
    },
    computed: {
        compPage: {
            get: function () {
                return this.$router.currentRoute.query &&
                    this.$router.currentRoute.query.page
                    ? this.$router.currentRoute.query.page
                    : 1;
            },
            set: function (newValue) {
                this.page = newValue;
            },
        },
    },
    methods: {
        handlePagination(page) {
            this.$router.push({
                path: this.$router.currentRoute.path,
                params: this.$router.currentRoute.params,
                query: { page: page },
            });
        },
    },
};
</script>

<style>
</style>