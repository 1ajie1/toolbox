<template>
  <div class="network-tools">
    <h1 class="page-title">Network</h1>
    <p class="page-description">网络诊断工具集</p>

    <el-row :gutter="30" class="tools-grid">
      <el-col :span="8" v-for="tool in networkTools" :key="tool.name">
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
  Monitor,
  Connection,
  Search,
  Location,
  Timer,
  DCaret,
  Lock,
  InfoFilled,
} from "@element-plus/icons-vue";

const router = useRouter();
const route = useRoute();

const navigateTo = (path: string): void => {
  router.push(path);
};

const isActive = (path: string): boolean => {
  return route.path === path;
};

const networkTools = [
  {
    name: "Ping",
    description: "执行Ping测试",
    icon: "Monitor",
    route: "/network/ping",
    iconBg: "bg-green",
  },
  {
    name: "Port Scan",
    description: "执行端口扫描",
    icon: "Search",
    route: "/network/portscan",
    iconBg: "bg-orange",
  },
  {
    name: "DNS",
    description: "执行DNS查询",
    icon: "Location",
    route: "/network/dns",
    iconBg: "bg-blue",
  },
  {
    name: "IP Info",
    description: "获取IP地址信息",
    icon: "InfoFilled",
    route: "/network/ipinfo",
    iconBg: "bg-purple",
  },
  {
    name: "Speed Test",
    description: "执行网络速度测试",
    icon: "Timer",
    route: "/network/speedtest",
    iconBg: "bg-red",
  },
  {
    name: "Traceroute",
    description: "执行路由跟踪",
    icon: "DCaret",
    route: "/network/traceroute",
    iconBg: "bg-indigo",
  },
  {
    name: "Certificate",
    description: "证书的检查与生成",
    icon: "Lock",
    route: "/network/cert",
    iconBg: "bg-teal",
  },
  {
    name: "Sniff",
    description: "执行网络抓包",
    icon: "Connection",
    route: "/network/sniff",
    iconBg: "bg-cyan",
  },
];
</script>

<style scoped>
@import "@/styles/common.css";
</style>
