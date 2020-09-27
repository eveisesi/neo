<template>
    <b-container>
        <b-row>
            <b-col lg="6">
                <ComponentLoading v-if='$apollo.queries.information.loading' />
                <Error v-else-if='$apollo.queries.information.error' />
                <b-table-simple v-else>
                    <b-tbody>
                        <b-tr>
                            <b-td
                                rowspan="3"
                                width="130"
                            >
                                <b-img :src="image +'alliances/'+information.id+'/logo?size=128'" />
                            </b-td>
                            <b-td>
                                Alliance
                            </b-td>
                            <b-td>
                                {{information.name}} [{{information.ticker}}]
                            </b-td>
                        </b-tr>
                        <b-tr>
                            <b-td>
                                Member Count
                            </b-td>
                            <b-td>
                                {{humanize(information.memberCount)}}
                            </b-td>
                        </b-tr>
                    </b-tbody>
                </b-table-simple>
            </b-col>
        </b-row>
        <b-row>
            <b-col sm="12">
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
            <b-col lg="12">
                <div class="float-right mt-2">
                    <b-pagination
                        v-model="compPage"
                        total-rows="1000"
                        per-page=50
                        @change="handlePagination"
                        hide-ellipsis
                    />
                </div>
                <h3>Recent Activity</h3>
                <hr style="background-color: white" />
                <ComponentLoading v-if="$apollo.queries.killmails.loading" />
                <Error v-else-if="$apollo.queries.killmails.error" />
                <KillTable
                    v-else
                    :killmails='killmails'
                    :target="information.id"
                />
                <hr style="background-color: white" />
                <b-pagination
                    v-model="compPage"
                    total-rows="500"
                    per-page="50"
                    @change="handlePagination"
                    hide-ellipsis
                />
            </b-col>
        </b-row>
    </b-container>
</template>

<script>
import numeral from "numeral";

import {
    KILLMAILS,
    ALLIANCE_INFORMATION,
    MOST_VALUABLE,
} from "../../../util/queries";
import { EVEONLINE_IMAGE } from "../../../util/const/urls";
import head from "../../../util/head";

export default {
    watchQuery: ["page"],
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
            title: head(this.information.name, "Alliance"),
        };
    },
    key: (to) => to.fullPath,
    apollo: {
        information: {
            query: ALLIANCE_INFORMATION,
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
                    entity: "alliance",
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
                    type: "alliance",
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
        humanize(total) {
            return numeral(total).format("0,0");
        },
    },
};
</script>