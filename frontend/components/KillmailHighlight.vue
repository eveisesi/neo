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
                <b-card
                    :img-src="EVEONLINE_IMAGE+'types/'+killmail.victim.ship.id+'/render?size=128'"
                    :img-alt="killmail.victim.ship.name"
                    no-body
                >
                    <b-card-text class="text-center mt-2 mb-1">
                        <nuxt-link :to="'/ships/'+ killmail.victim.ship.id">{{killmail.victim.ship.name}}</nuxt-link>
                        <br />
                        <nuxt-link :to="'/kill/'+killmail.id">{{AbbreviateNumber(killmail.totalValue)}} ISK</nuxt-link>
                        <br />
                        <nuxt-link
                            v-if="killmail.victim.character"
                            :to="'/characters/' + killmail.victim.character.id"
                        >{{killmail.victim.character.name}}</nuxt-link>
                        <nuxt-link
                            v-else
                            :to="'/corporations/' + killmail.victim.corporation.id"
                        >{{killmail.victim.corporation.name}}</nuxt-link>
                    </b-card-text>
                </b-card>
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