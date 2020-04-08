<template>
    <v-card tile>
        <v-list-item v-for="network in list" two-line :key="network.name">
            <v-list-item-content>
                <v-list-item-title>{{network.name}}</v-list-item-title>
                <v-list-item-subtitle>
                    {{network.running ? 'online' : 'offline'}}
                </v-list-item-subtitle>
            </v-list-item-content>
        </v-list-item>
    </v-card>
</template>

<script lang="ts">
    import Vue from 'vue';
    import Component from "vue-class-component";
    import {Network} from "@/api/api";
    import {client} from "@/api";

    @Component({})
    export default class Networks extends Vue {
        loading = false;
        list: Array<Network> = [];

        mounted() {
            this.fetch();
        }

        async fetch() {
            try {
                this.list = await client.networks()
            } catch (e) {
                this.$store.commit('error', e);
            }
        }
    }
</script>
