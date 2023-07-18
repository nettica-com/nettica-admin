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
    path: '/hosts',
    name: 'hosts',
    component: function () {
      return import(/* webpackChunkName: "Hosts" */ '../views/Hosts.vue')
    },
    meta: {
      requiresAuth: true
    }
  },
  {
    path: '/mesh',
    name: 'mesh',
    component: function () {
      return import(/* webpackChunkName: "Mesh" */ '../views/Mesh.vue')
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
    path: '/login',
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
    if (store.getters["auth/isAuthenticated"]) {
      next()
      return
    }
    //next(window.location.origin)
  } else {
    next()
  }
})

export default router
