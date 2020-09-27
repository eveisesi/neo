<template>
    <b-container v-else>
        <b-row>
            <b-col md="6">
                <ComponentLoading v-if='$apollo.queries.information.loading' />
                <Error v-else-if='$apollo.queries.information.error' />
                <b-table-simple v-else>
                    <b-tbody>
                        <b-tr>
                            <b-td
                                rowspan="4"
                                width="130"
                            >
                                <b-img
                                    :src="image+'types/'+information.id+'/render?size=128'"
                                    rounded
                                    fluid
                                    height="128"
                                    width="128"
                                />
                            </b-td>
                            <b-td>Ship</b-td>
                            <b-td>
                                {{information.name}}
                                <span v-if="information.group">
                                    (
                                    <nuxt-link :to="'shipGroups/'+information.group.id">{{information.group.name}}</nuxt-link>)
                                </span>
                            </b-td>
                        </b-tr>
                        <b-tr>
                            <b-td v-if="information.group"></b-td>
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
                <KillmailHighlight :mv="mv" />
            </b-col>
        </b-row>
        <b-row></b-row>
        <b-row>
            <b-col sm="12">
                <div class="float-right mt-2">
                    <b-pagination
                        v-model="compPage"
                        total-rows="500"
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
                    v-else
                    :killmails="killmails"
                />
                <hr style="background-color: white" />
                <b-pagination
                    v-model="compPage"
                    total-rows="500"
                    per-page="50"
                    @change="handlePagination"
                    hide-ellipsis
                    align="center"
                ></b-pagination>
            </b-col>
        </b-row>
    </b-container>
</template>


<script>
import numeral from "numeral";

import {
    KILLMAILS,
    TYPE_INFORMATION,
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
            title: head(this.information.name, "Ship"),
        };
    },
    key: (to) => to.fullPath,
    apollo: {
        killmails: {
            query: KILLMAILS,
            variables() {
                const page = this.$router.currentRoute.query.page
                    ? this.$router.currentRoute.query.page
                    : 1;
                return {
                    entity: "ship",
                    id: this.$router.currentRoute.params.id,
                    page: page,
                };
            },
            result(result, key) {
                this.killmails = result.data.killmails;
            },
            result(result, key) {
                this.error = JSON.stringify(result.message);
            },
        },
        information: {
            query: TYPE_INFORMATION,
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
        mv: {
            query: MOST_VALUABLE,
            variables() {
                return {
                    category: "kill",
                    type: "ship",
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