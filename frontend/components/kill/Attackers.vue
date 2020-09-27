<template>
    <div>
        <b-table-simple table-variant="active">
            <b-tbody>
                <b-tr>
                    <b-td class="text-center">Final Blow</b-td>
                    <b-td class="text-center">Most Damage</b-td>
                </b-tr>
                <b-tr class="text-center">
                    <b-td class="text-center">
                        <img
                            :src="EVEONLINE_IMAGE+'characters/'+finalBlow.character.id+'/portrait?size=64'"
                            class="rounded"
                            v-if="finalBlow.character"
                        />
                        <img
                            :src="EVEONLINE_IMAGE+'types/'+finalBlow.ship.id+'/icon?size=64'"
                            class="rounded"
                            v-else-if="finalBlow.ship"
                        />
                    </b-td>
                    <b-td class="text-center">
                        <img
                            :src="EVEONLINE_IMAGE+'characters/'+mostDamage.character.id+'/portrait?size=64'"
                            v-if="mostDamage.character"
                            class="rounded"
                        />
                        <img
                            :src="EVEONLINE_IMAGE+'types/'+mostDamage.ship.id+'/icon?size=64'"
                            v-else-if="mostDamage.ship"
                            class="rounded"
                        />
                    </b-td>
                </b-tr>
                <b-tr>
                    <b-td class="text-center">
                        <div v-if="finalBlow.character">
                            <nuxt-link :to="'/characters/'+finalBlow.character.id">
                                {{finalBlow.character.name}}
                            </nuxt-link>
                            <span v-if="finalBlow.corporation">[<nuxt-link :to="'/corporations/'+finalBlow.corporation.id">{{finalBlow.corporation.ticker}}</nuxt-link>]</span>
                        </div>
                        <span v-else>{{finalBlow.ship.name}}</span>
                    </b-td>
                    <b-td class="text-center">
                        <div v-if="mostDamage.character">
                            <nuxt-link
                                v-if="mostDamage.character"
                                :to="'/characters/'+mostDamage.character.id"
                            >
                                {{finalBlow.character.name}}
                            </nuxt-link>
                            <span v-if="mostDamage.corporation">[<nuxt-link :to="'/corporations/'+mostDamage.corporation.id">{{mostDamage.corporation.ticker}}</nuxt-link>]</span>
                        </div>
                        <span v-else>{{mostDamage.ship.name}}</span>
                    </b-td>
                </b-tr>
            </b-tbody>
        </b-table-simple>
        <b-table-simple table-variant="active">
            <b-tbody>
                <b-tr>
                    <b-td colspan="3">{{killmail.attackers.length}} Involved</b-td>
                    <b-td>
                        <span class="float-right">Damage</span>
                    </b-td>
                </b-tr>
                <b-tr
                    v-for="attacker in killmail.attackers"
                    :key="attacker.id"
                >
                    <b-td
                        style="width: 66px"
                        class="p-0"
                    >
                        <img
                            :src="EVEONLINE_IMAGE+'characters/'+attacker.character.id+'/portrait?size=64'"
                            v-if="attacker.character"
                            class="float-left"
                            style="height:64px; width:64px"
                        />
                        <img
                            :src="EVEONLINE_IMAGE+'types/'+attacker.ship.id+'/icon?size=64'"
                            v-else
                            class="float-left"
                            style="height:64px; width:64px"
                        />
                    </b-td>
                    <b-td
                        style="width: 34px"
                        class="p-0"
                    >
                        <img
                            :src="EVEONLINE_IMAGE+'types/'+attacker.ship.id+'/icon?size=32'"
                            v-if="attacker.ship"
                            style="height:32px; width:32px"
                        />
                        <br />
                        <img
                            :src="EVEONLINE_IMAGE+'types/'+attacker.weapon.id+'/icon?size=32'"
                            v-if="attacker.weapon"
                            style="height:32px; width:32px"
                        />
                    </b-td>
                    <b-td>
                        <nuxt-link
                            v-if="attacker.character"
                            :to="'/characters/'+attacker.character.id"
                        >{{attacker.character.name}}</nuxt-link>
                        <nuxt-link
                            v-else-if="attacker.corporation"
                            :to="'/corporations/'+attacker.corporation.id"
                        >{{attacker.corporation.name}}</nuxt-link>
                        <span v-else>{{attacker.ship.name}}</span>

                        <nuxt-link
                            v-if="attacker.corporation"
                            :to="'/corporations/'+ attacker.corporation.id"
                        >
                            <br />
                            {{attacker.corporation.name}}
                        </nuxt-link>
                        <nuxt-link
                            v-if="attacker.alliance"
                            :to="'/alliances/'+ attacker.alliance.id"
                        >
                            <br />
                            {{attacker.alliance.name}}
                        </nuxt-link>
                    </b-td>
                    <b-td>
                        <span class="float-right">{{humanize((attacker.damageDone/killmail.victim.damageTaken)*100)}}%</span>
                        <br>
                        <span class="float-right">{{humanize((attacker.damageDone))}}</span>
                    </b-td>
                </b-tr>
            </b-tbody>
        </b-table-simple>
    </div>
</template>

<script>
import numeral from "numeral";

import { EVEONLINE_IMAGE } from "../../util/const/urls";

export default {
    name: "Attackers",
    data() {
        return {
            EVEONLINE_IMAGE: EVEONLINE_IMAGE,
            finalBlow: {},
            mostDamage: {},
        };
    },
    props: ["killmail"],
    methods: {
        humanize(total) {
            return numeral(total).format("0,0");
        },
        getAttackerCharacterPortraitURL(attacker) {
            if (attacker.character != null) {
                return `${this.EVEONLINE_IMAGE}characters/${attacker.character.id}/portrait?size=64`;
            }
            return `${this.EVEONLINE_IMAGE}types/${attacker.ship.id}/icon?size=64`;
        },
    },
    created() {
        this.mostDamage = this.killmail.attackers.reduce((prev, current) => {
            return prev.damageDone > current.damageDone ? prev : current;
        });
        this.finalBlow = this.killmail.attackers.find(
            (attacker) => attacker.finalBlow
        );
    },
};
</script>