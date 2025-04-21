<template>
  <div class="format-tools">
    <h1 class="page-title">Format</h1>
    <p class="page-description">数据格式化和转换工具</p>
    
    <el-row :gutter="30">
      <el-col :span="8" v-for="tool in formatTools" :key="tool.name">
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
import { Document, Tickets, Connection } from '@element-plus/icons-vue'

const selectedTool = ref('')

const formatTools = [
  {
    name: 'JSON格式化',
    description: 'JSON数据的格式化、压缩和验证',
    icon: Document,
    color: '#409EFF'
  },
  {
    name: 'XML格式化',
    description: 'XML文档的格式化、压缩和验证',
    icon: Tickets,
    color: '#67C23A'
  },
  {
    name: 'YAML格式化',
    description: 'YAML数据的格式化和转换',
    icon: Connection,
    color: '#E6A23C'
  }
]
</script>

<style scoped>
@import '@/styles/common.css';
.format-tools {
  padding: 16px;
  max-width: 1280px;
  margin: 0 auto;
}
</style> 