<template>
    <div>
        <b-row
            cols-lg="6"
            cols-sm="3"
            cols="1"
        >
            <b-col
                v-for="killmail in mv"
                :key="killmail.id"
                class="mt-2"
            >
                <nuxt-link
                    :to="'/kill/'+killmail.id"
                    style="text-decoration: none; color: #ccc;"
                >
                    <b-card
                        :img-src="EVEONLINE_IMAGE+'types/'+killmail.victim.ship.id+'/render?size=128'"
                        :img-alt="killmail.victim.ship.name"
                        no-body
                    >
                        <b-card-text class="text-center mt-2 mb-1">
                            {{killmail.victim.ship.name}}
                            <br />
                            {{AbbreviateNumber(killmail.totalValue)}} ISK
                            <br />
                            <span v-if="killmail.victim.character">{{killmail.victim.character.name}}</span>
                            <span v-else>{{killmail.victim.corporation.name}}</span>
                        </b-card-text>
                    </b-card>
                </nuxt-link>
            </b-col>
        </b-row>
        <hr
            style="background-color: white"
            v-if="mv.length > 0"
        />
    </div>
</template>



<script>
import { AbbreviateNumber } from "../util/abbreviate";
import { EVEONLINE_IMAGE } from "../util/const/urls";

export default {
    name: "KillmailHighlight",
    props: ["mv"],
    data() {
        return {
            EVEONLINE_IMAGE: EVEONLINE_IMAGE,
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