<template>
    <b-table-simple
        response
        width="100%"
    >
        <tbody>
            <tr>
                <th style="width: 100px !important">
                    Date Time
                </th>
                <th class="text-center">
                    Ship
                </th>
                <th class="text-center">
                    System
                </th>
                <th class="text-center">
                    Victim
                </th>
                <th class="text-center">
                    Final Blow
                </th>
            </tr>

            <b-tr
                v-for="killmail in killmails"
                :key="killmail.id"
                :class="DetermineRowClass(killmail)"
            >
                <b-td>
                    <nuxt-link :to="'/kill/' + killmail.id">
                        {{fmtDate(killmail.killmailTime)}}
                        <br>
                        {{fmtTime(killmail.killmailTime)}}
                        <br>
                        {{AbbreviateNumber(killmail.totalValue)}}
                    </nuxt-link>
                </b-td>
                <b-td class="text-center">
                    <b-link
                        :href="'/kill/' + killmail.id"
                        :id="'ship-link-detail-'+killmail.id"
                    >
                        <img
                            :src="EVEONLINE_IMAGE+'types/'+(killmail.victim.ship != null ? killmail.victim.ship.id :1 )+'/render?size=64'"
                            class="rounded"
                        />
                    </b-link>
                    <b-tooltip
                        :target="'ship-link-detail-'+killmail.id"
                        triggers="hover"
                    >
                        Detail for {{killmail.id}}
                    </b-tooltip>
                </b-td>
                <b-td>
                    <nuxt-link :to="'/systems/' + killmail.system.id">{{killmail.system.name}}</nuxt-link>
                    (<span :class="killmail.system.security >= 0 ? 'text-success' : 'text-danger'">{{killmail.system.security.toFixed(2)}}</span>)
                    <br>
                    <nuxt-link :to="'/constellations/' + killmail.system.constellation.id">{{killmail.system.constellation.name}}</nuxt-link>
                    <br />
                    <nuxt-link :to="'/regions/' + killmail.system.constellation.region.id">{{killmail.system.constellation.region.name}}</nuxt-link>
                </b-td>
                <!-- Victim Column -->
                <b-td>
                    <img
                        :src="EVEONLINE_IMAGE+ (killmail.victim.alliance != null ? 'alliances/'+killmail.victim.alliance.id: (killmail.victim.corporation != null ? 'corporations/'+killmail.victim.corporation.id: 1)) + '/logo?size=64'"
                        class="float-left mr-2"
                    />
                    <nuxt-link
                        v-if="killmail.victim.character"
                        :to="'/characters/'+ killmail.victim.character.id"
                    >{{killmail.victim.character != null ? killmail.victim.character.name : '' }}</nuxt-link>
                    <br />
                    <nuxt-link
                        v-if="killmail.victim.corporation != null"
                        :to="'/corporations/'+ killmail.victim.corporation.id"
                    >{{killmail.victim.corporation.name}}</nuxt-link>
                    <br />
                    <nuxt-link
                        v-if="killmail.victim.alliance != null"
                        :to="'/alliances/'+ killmail.victim.alliance.id"
                    >{{killmail.victim.alliance.name}}</nuxt-link>

                </b-td>
                <!-- Attacker (Final Blow Column) -->
                <b-td>
                    <img
                        :src="EVEONLINE_IMAGE+'alliances/'+finalBlow(killmail.attackers).alliance.id+'/logo?size=64'"
                        v-if="finalBlow(killmail.attackers).alliance != null"
                        class="float-left mr-2"
                    />
                    <img
                        :src="EVEONLINE_IMAGE+'corporations/'+finalBlow(killmail.attackers).corporation.id+'/logo?size=64'"
                        v-else-if="finalBlow(killmail.attackers).corporation != null"
                        class="float-left mr-2"
                    />
                    <img
                        :src="EVEONLINE_IMAGE+'types/'+finalBlow(killmail.attackers).ship.id+'/icon?size=64'"
                        v-else-if="finalBlow(killmail.attackers).ship != null"
                        class="float-left mr-2"
                    />
                    <nuxt-link
                        v-if="finalBlow(killmail.attackers).character"
                        :to="'/characters/'+ finalBlow(killmail.attackers).character.id"
                    >
                        {{finalBlow(killmail.attackers).character.name}}
                        <br />
                    </nuxt-link>
                    <nuxt-link
                        v-if="finalBlow(killmail.attackers).alliance"
                        :to="'/alliances/'+ finalBlow(killmail.attackers).alliance.id"
                    >{{finalBlow(killmail.attackers).alliance.name}}</nuxt-link>
                    <nuxt-link
                        v-else-if="finalBlow(killmail.attackers).corporations"
                        :to="'/corporations/'+ finalBlow(killmail.attackers).corporation.id"
                    >{{finalBlow(killmail.attackers).corporation.name}}</nuxt-link>
                    <div v-else-if="finalBlow(killmail.attackers).ship != null">{{finalBlow(killmail.attackers).ship.name}}</div>
                </b-td>
            </b-tr>
        </tbody>
    </b-table-simple>
</template>

<script>
import moment from "moment";
import { EVEONLINE_IMAGE } from "../util/const/urls";
import { AbbreviateNumber } from "../util/abbreviate";

export default {
    name: "KillTable",
    props: ["killmails", "scope", "target"],
    computed: {},
    data() {
        return {
            EVEONLINE_IMAGE: EVEONLINE_IMAGE,
        };
    },
    methods: {
        fmtDate: (time) => {
            return moment(time, "YYYY-MM-DDTHH:mm:ssZ").format("YYYY-MM-DD");
        },
        fmtTime: (time) => {
            return moment(time, "YYYY-MM-DDTHH:mm:ssZ").format("HH:mm:ss");
        },
        finalBlow: (attackers) => {
            return attackers.find((attacker) => attacker.finalBlow);
        },
        AbbreviateNumber(total) {
            return AbbreviateNumber(total);
        },
        DetermineRowClass(item) {
            if (!this.scope || !this.target) {
                return "";
            }

            switch (this.scope) {
                case "character":
                    if (
                        item.victim.character &&
                        item.victim.character.id != this.target
                    ) {
                        return "success";
                    } else {
                        return "danger";
                    }
                case "corporation":
                    if (
                        item.victim.corporation &&
                        item.victim.corporation.id != this.target
                    ) {
                        return "success";
                    } else {
                        return "danger";
                    }
                case "alliance":
                    if (
                        item.victim.alliance &&
                        item.victim.alliance.id != this.target
                    ) {
                        return "success";
                    } else {
                        return "danger";
                    }
                default:
                    return "success";
            }
        },
    },
};
</script>

<style scoped>
.success {
    background: #001600;
}
.danger {
    background: #2f0202;
}
</style>