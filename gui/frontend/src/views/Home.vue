<template>
  <div class="home">
    <!-- 网络工具 -->
    <div id="network" class="category-section">
      <h2 class="category-title">Network</h2>
      <div class="category-subtitle">网络诊断工具集</div>
      
      <el-row :gutter="16" class="tools-grid">
        <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="tool in networkTools" :key="tool.name">
          <div class="tool-card" @click="$router.push(tool.route)">
            <div class="tool-icon-wrapper" :class="tool.iconBg">
              <el-icon class="tool-icon">
                <component :is="tool.icon" />
              </el-icon>
            </div>
            <div class="tool-info">
              <h3 class="tool-name">{{ tool.name }}</h3>
              <div class="tool-desc">{{ tool.description }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- 进程管理 -->
    <div id="process" class="category-section">
      <h2 class="category-title">Process</h2>
      <div class="category-subtitle">进程管理工具</div>
      
      <el-row :gutter="16" class="tools-grid">
        <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="tool in processTools" :key="tool.name">
          <div class="tool-card" @click="$router.push(tool.route)">
            <div class="tool-icon-wrapper" :class="tool.iconBg">
              <el-icon class="tool-icon">
                <component :is="tool.icon" />
              </el-icon>
            </div>
            <div class="tool-info">
              <h3 class="tool-name">{{ tool.name }}</h3>
              <div class="tool-desc">{{ tool.description }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>
    
    <!-- 文件系统 -->
    <div id="fs" class="category-section">
      <h2 class="category-title">File System</h2>
      <div class="category-subtitle">文件系统工具集</div>
      
      <el-row :gutter="16" class="tools-grid">
        <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="tool in fsTools" :key="tool.name">
          <div class="tool-card" @click="$router.push(tool.route)">
            <div class="tool-icon-wrapper" :class="tool.iconBg">
              <el-icon class="tool-icon">
                <component :is="tool.icon" />
              </el-icon>
            </div>
            <div class="tool-info">
              <h3 class="tool-name">{{ tool.name }}</h3>
              <div class="tool-desc">{{ tool.description }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- 格式化 -->
    <div id="fmt" class="category-section">
      <h2 class="category-title">Format</h2>
      <div class="category-subtitle">格式化数据文件或文本内容</div>
      
      <el-row :gutter="16" class="tools-grid">
        <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="tool in fmtTools" :key="tool.name">
          <div class="tool-card" @click="$router.push(tool.route)">
            <div class="tool-icon-wrapper" :class="tool.iconBg">
              <el-icon class="tool-icon">
                <component :is="tool.icon" />
              </el-icon>
            </div>
            <div class="tool-info">
              <h3 class="tool-name">{{ tool.name }}</h3>
              <div class="tool-desc">{{ tool.description }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>

    <!-- 文本处理 -->
    <div id="text" class="category-section">
      <h2 class="category-title">Text</h2>
      <div class="category-subtitle">文本处理工具</div>
      
      <el-row :gutter="16" class="tools-grid">
        <el-col :xs="24" :sm="12" :md="8" :lg="6" v-for="tool in textTools" :key="tool.name">
          <div class="tool-card" @click="$router.push(tool.route)">
            <div class="tool-icon-wrapper" :class="tool.iconBg">
              <el-icon class="tool-icon">
                <component :is="tool.icon" />
              </el-icon>
            </div>
            <div class="tool-info">
              <h3 class="tool-name">{{ tool.name }}</h3>
              <div class="tool-desc">{{ tool.description }}</div>
            </div>
          </div>
        </el-col>
      </el-row>
    </div>
  </div>
</template>

<script setup lang="ts">
import {
  Monitor,
  Connection,
  Search,
  Location,
  Timer,
  Link,
  Lock,
  DCaret,
  List,
  Share,
  InfoFilled,
  CircleClose,
  Folder,
  Document,
  DocumentCopy,
  Download,
  Upload,
  Files,
  Sort,
  Grid,
  ScaleToOriginal,
  Crop,
  Edit,
  Finished,
  Filter
} from '@element-plus/icons-vue'

// 网络工具
const networkTools = [
  {
    name: 'Ping',
    description: '执行Ping测试',
    icon: 'Monitor',
    route: '/network/ping',
    iconBg: 'bg-green'
  },
  {
    name: 'Port Scan',
    description: '执行端口扫描',
    icon: 'Search',
    route: '/network/portscan',
    iconBg: 'bg-orange'
  },
  {
    name: 'DNS',
    description: '执行DNS查询',
    icon: 'Location',
    route: '/network/dns',
    iconBg: 'bg-blue'
  },
  {
    name: 'IP Info',
    description: '获取IP地址信息',
    icon: 'InfoFilled',
    route: '/network/ipinfo',
    iconBg: 'bg-purple'
  },
  {
    name: 'Speed Test',
    description: '执行网络速度测试',
    icon: 'Timer',
    route: '/network/speedtest',
    iconBg: 'bg-red'
  },
  {
    name: 'Traceroute',
    description: '执行路由跟踪',
    icon: 'DCaret',
    route: '/network/traceroute',
    iconBg: 'bg-indigo'
  },
  {
    name: 'Certificate',
    description: '证书的检查与生成',
    icon: 'Lock',
    route: '/network/cert',
    iconBg: 'bg-teal'
  },
  {
    name: 'Sniff',
    description: '执行网络抓包',
    icon: 'Connection',
    route: '/network/sniff',
    iconBg: 'bg-cyan'
  }
]

// 进程管理工具
const processTools = [
  {
    name: 'List',
    description: '列出系统进程',
    icon: 'List',
    route: '/process/list',
    iconBg: 'bg-blue'
  },
  {
    name: 'Tree',
    description: '以树形结构显示进程',
    icon: 'Share',
    route: '/process/tree',
    iconBg: 'bg-green'
  },
  {
    name: 'Info',
    description: '显示进程详情',
    icon: 'InfoFilled',
    route: '/process/info',
    iconBg: 'bg-purple'
  },
  {
    name: 'Kill',
    description: '终止指定进程',
    icon: 'CircleClose',
    route: '/process/kill',
    iconBg: 'bg-red'
  },
  {
    name: 'Children',
    description: '列出指定进程的子进程',
    icon: 'Connection',
    route: '/process/children',
    iconBg: 'bg-orange'
  }
]

// 文件系统工具
const fsTools = [
  {
    name: 'Compress',
    description: '压缩或解压缩文件',
    icon: 'Files',
    route: '/fs/compress',
    iconBg: 'bg-blue'
  },
  {
    name: 'Find',
    description: '搜索文件和目录',
    icon: 'Search',
    route: '/fs/find',
    iconBg: 'bg-green'
  },
  {
    name: 'Split',
    description: '大文件/目录的分片',
    icon: 'Crop',
    route: '/fs/split',
    iconBg: 'bg-purple'
  },
  {
    name: 'Tree',
    description: '显示目录结构',
    icon: 'Share',
    route: '/fs/tree',
    iconBg: 'bg-teal'
  }
]

// 格式化工具
const fmtTools = [
  {
    name: 'Format',
    description: '格式化数据（JSON/XML/YAML）',
    icon: 'ScaleToOriginal',
    route: '/fmt/fmt',
    iconBg: 'bg-blue'
  }
]

// 文本处理工具
const textTools = [
  {
    name: 'Grep',
    description: '文本搜索',
    icon: 'Search',
    route: '/text/grep',
    iconBg: 'bg-green'
  },
  {
    name: 'Replace',
    description: '文本替换',
    icon: 'Edit',
    route: '/text/replace',
    iconBg: 'bg-red'
  },
  {
    name: 'Filter',
    description: '文本过滤',
    icon: 'Filter',
    route: '/text/filter',
    iconBg: 'bg-purple'
  }
]
</script>

<style scoped>
@import '@/styles/common.css';
</style> 