<template>
    <Loading v-if="$apollo.queries.killmail.loading" />
    <Error v-else-if="$apollo.queries.killmail.error" />
    <b-container v-else>
        <b-row>
            <b-col lg="12">
                <h2>Killmail {{killmail.id}} Overview</h2>
                <hr style="background-color: white" />
            </b-col>
        </b-row>
        <b-row>
            <b-col md="8">
                <b-row>
                    <b-col lg="6">
                        <FittingWheel :victim="killmail.victim" />
                    </b-col>
                    <b-col lg="6">
                        <VictimInfo
                            :killmail="killmail"
                            class="mt-2"
                        />
                    </b-col>
                </b-row>
                <h3>Item(s) Dropped/Destroyed</h3>
                <hr style="background-color:white;" />
                <ItemDetail :victim="killmail.victim" />
            </b-col>
            <b-col
                md="4"
                class="p-1 mt-1"
            >
                <Attackers :killmail="killmail" />
            </b-col>
        </b-row>
    </b-container>
</template>

<script>
import { KILLMAIL } from "../../util/queries";

import FittingWheel from "../../components/kill/FittingWheel";
import VictimInfo from "../../components/kill/VictimInfo";
import ItemDetail from "../../components/kill/ItemDetail";
import Attackers from "../../components/kill/Attackers";

export default {
    data() {
        return {
            killmail: {},
            error: "",
        };
    },
    apollo: {
        killmail: {
            query: KILLMAIL,
            variables() {
                return {
                    id: this.$router.currentRoute.params.id,
                };
            },
            result(result, error) {
                this.killmail = result.data.killmail;
            },
            error(result, error) {
                this.error = JSON.stringify(result.message);
            },
        },
    },
};
</script>

<style>
</style>