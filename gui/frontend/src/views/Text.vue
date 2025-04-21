import '@/styles/common.css'

<template>
  <div class="text-tools">
    <h1 class="page-title">Text</h1>
    <p class="page-description">文本处理工具集</p>
    
    <el-row :gutter="30">
      <el-col :span="8" v-for="tool in textTools" :key="tool.name">
        <div 
          class="tool-card" 
          :class="{ active: selectedTool === tool.name }"
          @click="selectedTool = tool.name"
        >
          <div class="tool-icon-container" :style="{ backgroundColor: tool.color }">
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
import { ref } from 'vue'
import { Search, Edit, Filter } from '@element-plus/icons-vue'

const selectedTool = ref('')

const textTools = [
  {
    name: '文本搜索',
    description: '在文本中搜索指定内容',
    icon: Search,
    color: '#409EFF'
  },
  {
    name: '文本替换',
    description: '批量替换文本内容',
    icon: Edit,
    color: '#67C23A'
  },
  {
    name: '文本过滤',
    description: '按条件过滤文本行',
    icon: Filter,
    color: '#E6A23C'
  }
]
</script>

<style scoped>
@import '@/styles/common.css';
.text-tools {
  padding: 16px;
  max-width: 1280px;
  margin: 0 auto;
}
</style> 