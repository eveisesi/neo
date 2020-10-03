<template>
    <b-container
        class="mt-2"
        v-else
    >
        <b-row class="mb-2">
            <b-col lg="12">
                <h4 class="text-center">Most Valuable Kills - Last 7 days</h4>
                <hr style="background-color: white;" />
                <ComponentLoading v-if="$apollo.queries.mv.loading" />
                <Error v-else-if="$apollo.queries.mv.error" />
                <KillmailHighlight
                    v-else
                    :mv="mv"
                />
            </b-col>
        </b-row>
        <b-row>
            <b-col lg="12">
                <h3>Most Recent Killmails</h3>
                <hr style="background-color: white" />
                <ComponentLoading v-if="$apollo.queries.killmails.loading" />
                <Error v-else-if="$apollo.queries.killmails.error" />
                <KillTable
                    v-else
                    :killmails="killmails"
                />
            </b-col>
        </b-row>
    </b-container>
</template>

<script>
import {
    RECENT_KILLMAILS,
    MOST_VALUABLE,
    KILLMAIL_FEED,
} from "../util/queries";
import { AbbreviateNumber } from "../util/abbreviate";
import head from "../util/head";

let received = 0;

export default {
    head: {
        title: head("Welcome!"),
    },
    apollo: {
        killmails: {
            query: RECENT_KILLMAILS,
            // subscribeToMore: [
            //     {
            //         document: KILLMAIL_FEED,
            //         updateQuery: (previous, { subscriptionData }) => {
            //             const newKill = subscriptionData.data.feed;
            //             received++;
            //             console.log(previous, received);
            //             previous.killmails.unshift(newKill);
            //             previous.killmails.pop();
            //             return previous;
            //         },
            //     },
            // ],
        },
        mv: {
            query: MOST_VALUABLE,
        },
    },
    data() {
        return {
            killmails: [],
            mv: [],
            feed: [],
        };
    },
    methods: {
        AbbreviateNumber(total) {
            return AbbreviateNumber(total);
        },
    },
};
</script>

<style>
</style>
