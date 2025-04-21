<template>
  <div class="fs-tools">
    <h1 class="page-title">File System</h1>
    <p class="page-description">文件系统工具集</p>

    <el-row :gutter="30" class="tools-grid">
      <el-col :span="8" v-for="tool in fsTools" :key="tool.name">
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
import { Files, Search, Crop, Share } from "@element-plus/icons-vue";

const router = useRouter();
const route = useRoute();

const navigateTo = (path: string): void => {
  router.push(path);
};

const isActive = (path: string): boolean => {
  return route.path === path;
};

const fsTools = [
  {
    name: "Compress",
    description: "压缩或解压缩文件",
    icon: "Files",
    route: "/fs/compress",
    iconBg: "bg-blue",
  },
  {
    name: "Find",
    description: "搜索文件和目录",
    icon: "Search",
    route: "/fs/find",
    iconBg: "bg-green",
  },
  {
    name: "Split",
    description: "大文件/目录的分片",
    icon: "Crop",
    route: "/fs/split",
    iconBg: "bg-purple",
  },
  {
    name: "Tree",
    description: "显示目录结构",
    icon: "Share",
    route: "/fs/tree",
    iconBg: "bg-teal",
  },
];
</script>

<style scoped>
@import "@/styles/common.css";
</style>
