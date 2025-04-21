<template>
  <div class="process-tool">
    <div class="page-header">
      <h1 class="page-title">{{ toolInfo?.name || currentTool }}</h1>
      <p class="page-description">{{ toolInfo?.description || '进程管理工具' }}</p>
    </div>

    <div class="tool-content">
      <div class="tool-placeholder">
        <el-empty description="工具开发中，敬请期待"></el-empty>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed } from 'vue'
import { useRoute } from 'vue-router'

const route = useRoute()
const currentTool = computed(() => route.params.tool as string)

// 进程管理工具信息
const processTools = [
  {
    name: 'List',
    description: '列出系统进程',
    route: 'list'
  },
  {
    name: 'Tree',
    description: '以树形结构显示进程',
    route: 'tree'
  },
  {
    name: 'Info',
    description: '显示进程详情',
    route: 'info'
  },
  {
    name: 'Kill',
    description: '终止指定进程',
    route: 'kill'
  },
  {
    name: 'Children',
    description: '列出指定进程的子进程',
    route: 'children'
  }
]

// 获取当前工具信息
const toolInfo = computed(() => {
  return processTools.find(tool => tool.route === currentTool.value)
})
</script>

<style scoped>
.process-tool {
  padding: 20px;
  max-width: 1200px;
  margin: 0 auto;
}

.page-header {
  margin-bottom: 40px;
}

.page-title {
  font-size: 28px;
  font-weight: 600;
  margin: 0;
  color: #1a1a1a;
}

.page-description {
  color: #666;
  margin-top: 8px;
  font-size: 16px;
}

.tool-placeholder {
  background: white;
  border-radius: 12px;
  padding: 60px 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  text-align: center;
}
</style> 