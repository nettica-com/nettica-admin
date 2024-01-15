import Vue from 'vue'
import VueRouter from 'vue-router'
import store from "../store";
import env from "../../env";

Vue.use(VueRouter);

const routes = [
  {
    path: '/join*',
    name: 'join',
    component: function () {
      return import(/* webpackChunkName: "Join" */ '../views/Join.vue')
    },
    meta: {
      requiresAuth: false
    }
  },
  {
    path: '/devices',
    name: 'devices',
    component: function () {
      return import(/* webpackChunkName: "Devices" */ '../views/Devices.vue')
    },
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/networks',
    name: 'networks',
    component: function () {
      return import(/* webpackChunkName: "Network" */ '../views/Network.vue')
    },
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/accounts',
    name: 'accounts',
    component: function () {
      return import(/* webpackChunkName: "Accounts" */ '../views/Accounts.vue')
    },
    meta: {
      requiresAuth: true
    }
  },
  { 
    path: '/services',
    name: 'services',
    component: function () {
      return import(/* webpackChunkName: "Services" */ '../views/Services.vue')
    }
  },
  { 
    path: '/login*',
    name: 'login',
    component: function () {
      return import(/* webpackChunkName: "Login" */ '../views/Login.vue')
    },
    meta: {
      requiresAuth: false
    }
  },
  {
    path: '/',
    name: 'root',
    meta: {
      requiresAuth: false
    }
  },

];

const router = new VueRouter({
  mode: 'history',
  base: process.env.BASE_URL,
  routes
});

router.beforeEach((to, from, next) => {
  if(to.matched.some(record => record.meta.requiresAuth)) {
    store.commit("auth/requiresAuth", true)
    if (store.getters["auth/isAuthenticated"]) {
      next()
      return
    }
    //next(window.location.origin)
  } else {
    store.commit("auth/requiresAuth", false)
    next()
  }
})

export default router
