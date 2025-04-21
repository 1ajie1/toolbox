<template>
  <div class="process-tools">
    <h1 class="page-title">Process</h1>
    <p class="page-description">进程管理工具</p>

    <el-row :gutter="30" class="tools-grid">
      <el-col :span="8" v-for="tool in processTools" :key="tool.name">
        <div
          class="tool-card"
          :class="{ active: isActive(tool.route) }"
          @click="navigateTo(tool.route)"
        >
          <div class="tool-icon-container" :class="tool.iconBg">
            <el-icon><component :is="tool.icon" /></el-icon>
          </div>
          <h3 class="tool-title">{{ tool.name }}</h3>
          <p class="tool-description">{{ tool.description }}</p>
        </div>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from "vue";
import { useRouter, useRoute } from "vue-router";
import {
  List,
  Share,
  InfoFilled,
  CircleClose,
  Connection,
} from "@element-plus/icons-vue";

const router = useRouter();
const route = useRoute();

const navigateTo = (path: string): void => {
  router.push(path);
};

const isActive = (path: string): boolean => {
  return route.path === path;
};

const processTools = [
  {
    name: "List",
    description: "列出系统进程",
    icon: "List",
    route: "/process/list",
    iconBg: "bg-blue",
  },
  {
    name: "Tree",
    description: "以树形结构显示进程",
    icon: "Share",
    route: "/process/tree",
    iconBg: "bg-green",
  },
  {
    name: "Info",
    description: "显示进程详情",
    icon: "InfoFilled",
    route: "/process/info",
    iconBg: "bg-purple",
  },
  {
    name: "Kill",
    description: "终止指定进程",
    icon: "CircleClose",
    route: "/process/kill",
    iconBg: "bg-red",
  },
  {
    name: "Children",
    description: "列出指定进程的子进程",
    icon: "Connection",
    route: "/process/children",
    iconBg: "bg-orange",
  },
];
</script>

<style scoped>
@import "@/styles/common.css";
</style>
