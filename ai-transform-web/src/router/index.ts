/*
import { createRouter,createWebHashHistory,RouteRecordRaw} from 'vue-router';
import Home from '../views/Home/Home.vue'
import Internal from '../views/Internal/Internal.vue'
import Applications from '../views/Applications/Applications.vue'
import Deploy from '../views/Deploy/Deploy.vue'

const routes:Array<RouteRecordRaw> = [
    {path: '/',name:"home", component: Home},
    {path: '/apps',name:"app" ,component: Applications},
    {path: '/internal',name:"internal" ,component: Internal,children:[
        {
            name:"app",path:"/apps",component:Applications
        },
        {
            name:"deploy",path:"/deploy",component: Deploy
        },
    ]},
]


const router = createRouter({
    history:createWebHashHistory(),
    routes:routes,
})
export default router
*/