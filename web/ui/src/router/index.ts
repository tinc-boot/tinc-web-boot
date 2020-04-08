import Vue from 'vue'
import VueRouter, {RouteConfig} from 'vue-router'
import Networks from '../views/Networks.vue'

Vue.use(VueRouter);

const routes: Array<RouteConfig> = [
    {
        path: '/',
        name: 'networks',
        component: Networks
    }
];

const router = new VueRouter({
    routes
});

export default router
