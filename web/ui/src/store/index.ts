import Vue from 'vue'
import Vuex from 'vuex'

Vue.use(Vuex);


export default new Vuex.Store({
    state: {},
    mutations: {
        error(state, e) {
            console.error(e);
        }
    },
    actions: {},
    modules: {}
})
