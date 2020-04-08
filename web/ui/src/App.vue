<template>
    <v-app>
        <v-app-bar
                app
                color="primary"
                dark
        >

        </v-app-bar>

        <v-content>
            <router-view/>
        </v-content>
    </v-app>
</template>

<script lang="ts">
    import Vue from 'vue';
    import HelloWorld from './components/HelloWorld.vue';
    import Component from "vue-class-component";

    @Component({components: {HelloWorld}})
    export default class App extends Vue {

    }
</script>
