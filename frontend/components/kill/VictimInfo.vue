/* eslint function-paren-newline: ["error", "always"] */
<template>
    <b-table-simple
        borderless
        striped
        table-variant="active"
    >
        <b-tbody>
            <b-tr>
                <b-td
                    style="width:66px"
                    class="p-0"
                    v-if="killmail.victim.character"
                >
                    <img
                        :src="EVEONLINE_IMAGE+'characters/'+killmail.victim.character.id+'/portrait?size=64'"
                        style="height:64px; width:64px"
                    />
                </b-td>
                <b-td
                    style="width:66px"
                    class="p-0"
                    v-else
                >
                    <img
                        :src="EVEONLINE_IMAGE+'corporations/'+killmail.victim.corporation.id+'/logo?size=64'"
                        style="height:64px; width:64px"
                    />
                </b-td>
                <b-td
                    style="width:34"
                    class="p-0"
                >
                    <img :src="EVEONLINE_IMAGE+'corporations/'+killmail.victim.corporation.id+'/logo?size=32'" />
                    <br v-if="killmail.victim.alliance != null" />
                    <img
                        v-if="killmail.victim.alliance != null"
                        :src="EVEONLINE_IMAGE+'alliances/'+killmail.victim.alliance.id+'/logo?size=32'"
                    />
                </b-td>
                <b-td>
                    <nuxt-link
                        v-if="killmail.victim.character"
                        :to="'/characters/'+killmail.victim.character.id"
                    >{{killmail.victim.character.name}}</nuxt-link>
                    <br v-if="killmail.victim.character" />
                    <nuxt-link :to="'/corporations/'+killmail.victim.corporation.id">{{killmail.victim.corporation.name}}</nuxt-link>
                    <br />
                    <nuxt-link
                        v-if="killmail.victim.alliance != null"
                        :to="'/alliances/'+killmail.victim.alliance.id"
                    >{{killmail.victim.alliance.name}}</nuxt-link>
                </b-td>
            </b-tr>
            <b-tr>
                <b-td colspan="2">Ship</b-td>
                <b-td>
                    <nuxt-link :to="'/ships/'+killmail.victim.ship.id">{{killmail.victim.ship.name}}</nuxt-link>
                    (<nuxt-link :to="'/shipGroups/'+killmail.victim.ship.group.id">{{killmail.victim.ship.group.name}}</nuxt-link>)
                </b-td>
            </b-tr>
            <b-tr>
                <b-td colspan="2">System</b-td>
                <b-td>
                    <nuxt-link :to="'/systems/'+killmail.system.id">{{killmail.system.name}}</nuxt-link>
                    (<span :class="killmail.system.security >= 0 ? 'text-success' : 'text-danger'">{{killmail.system.security.toFixed(2)}}</span>) / <nuxt-link :to=" 'regions/'+killmail.system.constellation.region.id">{{killmail.system.constellation.region.name}}</nuxt-link>
                </b-td>
            </b-tr>
            <b-tr>
                <b-td colspan="2">Time</b-td>
                <b-td>{{fmtKillTime()}}</b-td>
            </b-tr>
            <b-tr>
                <b-td colspan="2">Damage:</b-td>
                <b-td>{{ killmail.victim.damageTaken.toLocaleString() }}</b-td>
            </b-tr>
            <b-tr>
                <b-td colspan="2">Value Destroyed</b-td>
                <b-td>{{ humanize(killmail.destroyedValue) }} ISK</b-td>
            </b-tr>
            <b-tr>
                <b-td colspan="2">Value Dropped</b-td>
                <b-td>{{ humanize(killmail.droppedValue) }} ISK</b-td>
            </b-tr>
            <b-tr>
                <b-td colspan="2">Fitted Value</b-td>
                <b-td>{{ humanize(killmail.fittedValue) }} ISK</b-td>
            </b-tr>
            <b-tr>
                <b-td colspan="2">Total Value</b-td>
                <b-td>{{ humanize(killmail.totalValue) }} ISK</b-td>
            </b-tr>
        </b-tbody>
    </b-table-simple>
</template>

<script>
import moment from "moment";
import numeral from "numeral";

import { EVEONLINE_IMAGE } from "@/util/const/urls";

export default {
    name: "VictimInfo",
    props: {
        killmail: {
            type: Object,
            required: true,
        },
    },
    data() {
        return {
            EVEONLINE_IMAGE: EVEONLINE_IMAGE,
        };
    },
    methods: {
        humanize(total) {
            return numeral(total).format("0,0.00");
        },
        fmtKillTime() {
            return moment(
                this.killmail.killmailTime,
                "YYYY-MM-DDTHH:mm:ssZ"
            ).format("YYYY-MM-DD HH:mm");
        },
    },
};
</script>
