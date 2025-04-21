import { createRouter, createWebHistory, RouteRecordRaw } from "vue-router";
import Home from "../views/Home.vue";
import NetworkTools from "../views/network/NetworkTool.vue";

const routes: Array<RouteRecordRaw> = [
  {
    path: "/",
    name: "home",
    component: Home,
    meta: {
      title: "工具箱 - 首页",
    },
  },
  // 网络工具路由
  {
    path: "/network",
    name: "network",
    component: NetworkTools,
    meta: {
      title: "网络工具",
    },
  },
  {
    path: "/network/:tool",
    name: "network-tool",
    component: () => import("../views/network/NetworkTool.vue"),
    meta: {
      title: "网络工具",
    },
  },
  // 进程管理路由
  {
    path: "/process",
    name: "process",
    component: () => import("../views/Process.vue"),
    meta: {
      title: "进程管理",
    },
  },
  {
    path: "/process/:tool",
    name: "process-tool",
    component: () => import("../views/process/ProcessTool.vue"),
    meta: {
      title: "进程管理工具",
    },
  },
  // 文件系统路由
  {
    path: "/fs",
    name: "fs",
    component: () => import("../views/FileSystem.vue"),
    meta: {
      title: "文件系统",
    },
  },
  {
    path: "/fs/:tool",
    name: "fs-tool",
    component: () => import("../views/fs/FsTool.vue"),
    meta: {
      title: "文件系统工具",
    },
  },
  // 格式化工具路由
  {
    path: "/fmt",
    name: "fmt",
    component: () => import("../views/Format.vue"),
    meta: {
      title: "格式化工具",
    },
  },
  {
    path: "/fmt/:tool",
    name: "fmt-tool",
    component: () => import("../views/fmt/FmtTool.vue"),
    meta: {
      title: "格式化工具",
    },
  },
  // 文本处理路由
  {
    path: "/text",
    name: "text",
    component: () => import("../views/Text.vue"),
    meta: {
      title: "文本处理",
    },
  },
  {
    path: "/text/:tool",
    name: "text-tool",
    component: () => import("../views/text/TextTool.vue"),
    meta: {
      title: "文本处理工具",
    },
  },
];

const router = createRouter({
  history: createWebHistory(),
  routes,
  scrollBehavior(to) {
    if (to.hash) {
      return {
        el: to.hash,
        behavior: "smooth",
      };
    } else {
      return { top: 0 };
    }
  },
});

// 全局前置守卫
router.beforeEach((to, from, next) => {
  // 设置页面标题
  document.title = `${to.meta.title || "工具箱"}`;
  next();
});

export default router;
