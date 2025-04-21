<script setup lang="ts">
// 不需要额外的导入，因为路由组件已经在main.ts中全局注册
import { useRouter } from "vue-router";

const router = useRouter();

// 导航处理函数
const scrollToCategory = (categoryId: string) => {
  // 当前路径
  const currentPath = window.location.pathname;

  if (currentPath === "/" || currentPath === "") {
    // 在首页，直接滚动
    setTimeout(() => {
      const element = document.getElementById(categoryId);
      if (element) {
        element.scrollIntoView({ behavior: "smooth" });
      }
    }, 100);
  } else {
    // 不在首页，先导航到首页
    router.push("/");
    // 等待导航完成后滚动
    setTimeout(() => {
      const element = document.getElementById(categoryId);
      if (element) {
        element.scrollIntoView({ behavior: "smooth" });
      }
    }, 500);
  }
};
</script>

<template>
  <div class="app-wrapper">
    <!-- 固定顶部导航 -->
    <el-header class="app-header fixed-header">
      <div class="header-content">
        <div class="logo">
          <img src="./assets/images/logo-universal.png" alt="Logo" />
          <span class="title">工具箱</span>
        </div>
        <div class="search-box">
          <el-input placeholder="搜索工具..." prefix-icon="Search" />
        </div>
        <div class="header-actions">
          <el-button circle icon="Setting" />
          <el-button circle icon="QuestionFilled" />
        </div>
      </div>
    </el-header>

    <!-- 主体区域 -->
    <div class="main-container">
      <!-- 固定左侧导航 -->
      <div class="fixed-sidebar">
        <el-menu default-active="home" class="app-menu">
          <el-menu-item index="home" @click="$router.push('/')">
            <el-icon><HomeFilled /></el-icon>
            <span>首页</span>
          </el-menu-item>
          <el-menu-item index="network" @click="scrollToCategory('network')">
            <el-icon><Connection /></el-icon>
            <span>网络工具</span>
          </el-menu-item>
          <el-menu-item index="process" @click="scrollToCategory('process')">
            <el-icon><Monitor /></el-icon>
            <span>进程管理</span>
          </el-menu-item>
          <el-menu-item index="fs" @click="scrollToCategory('fs')">
            <el-icon><Folder /></el-icon>
            <span>文件系统</span>
          </el-menu-item>
          <el-menu-item index="fmt" @click="scrollToCategory('fmt')">
            <el-icon><ScaleToOriginal /></el-icon>
            <span>格式化工具</span>
          </el-menu-item>
          <el-menu-item index="text" @click="scrollToCategory('text')">
            <el-icon><Document /></el-icon>
            <span>文本处理</span>
          </el-menu-item>
        </el-menu>
      </div>

      <!-- 内容区域 -->
      <div class="content-area">
        <router-view></router-view>
      </div>
    </div>
  </div>
</template>

<style>
html,
body {
  margin: 0;
  padding: 0;
  height: 100%;
  background-color: #f5f7fa;
}

#app {
  height: 100%;
  font-family: "Nunito", Arial, sans-serif;
}

/* 布局容器 */
.app-wrapper {
  display: flex;
  flex-direction: column;
  height: 100%;
}

/* 顶部导航 */
.fixed-header {
  position: fixed;
  top: 0;
  left: 0;
  right: 0;
  z-index: 100;
  background-color: #fff;
  border-bottom: 1px solid #e4e7ed;
  padding: 0;
  height: 60px;
  line-height: 60px;
}

.header-content {
  max-width: 1400px;
  margin: 0 auto;
  display: flex;
  align-items: center;
  justify-content: space-between;
  height: 100%;
  padding: 0 20px;
}

.logo {
  display: flex;
  align-items: center;
  gap: 10px;
}

.logo img {
  height: 32px;
}

.title {
  font-size: 18px;
  font-weight: 600;
  color: #1a1a1a;
}

.search-box {
  flex: 1;
  max-width: 400px;
  margin: 0 40px;
}

.header-actions {
  display: flex;
  gap: 10px;
}

/* 主体区域布局 */
.main-container {
  display: flex;
  height: calc(100% - 60px);
  margin-top: 60px; /* 为固定的顶部导航条留出空间 */
}

/* 固定侧边栏 */
.fixed-sidebar {
  width: 200px;
  position: fixed;
  top: 60px; /* 顶部导航条高度 */
  left: 0;
  bottom: 0;
  overflow-y: auto; /* 侧边栏内容过多时可滚动 */
  background-color: #fff;
  border-right: 1px solid #e4e7ed;
  z-index: 99;
}

/* 内容区域 */
.content-area {
  flex: 1;
  margin-left: 220px; /* 与侧边栏宽度相同 */
  overflow-y: auto; /* 内容过多时可滚动 */
  padding: 0;
  background-color: #f5f7fa;
}

/* 菜单样式 */
.app-menu {
  border-right: none;
  height: 100%;
}

/* Element Plus 组件样式覆盖 */
.el-menu-item {
  height: 50px;
  line-height: 50px;
}

.el-menu-item [class^="el-icon"] {
  font-size: 18px;
}

.el-input__wrapper {
  background-color: #f5f7fa;
}
</style>
