<template>
  <div class="network-tool">
    <div class="tool-content">
      <div v-if="currentTool === 'ping'" class="tool-section">
        <el-form
          :model="pingParams"
          :rules="pingRules"
          ref="pingForm"
          class="ping-form"
        >
          <el-form-item label="目标主机/IP" prop="target">
            <el-input
              v-model="pingParams.target"
              placeholder="请输入主机名或IP地址"
            ></el-input>
          </el-form-item>
          <el-form-item label="Ping次数">
            <el-input-number
              v-model="pingParams.count"
              :min="1"
            ></el-input-number>
          </el-form-item>
          <el-form-item label="请求间隔(秒)">
            <el-input-number
              v-model="pingParams.interval"
              :min="0.1"
              :max="10"
              :step="0.1"
              :precision="1"
            ></el-input-number>
          </el-form-item>
          <el-form-item>
            <el-button
              type="primary"
              @click="submitPingForm"
              :loading="isLoading"
              v-if="!isLoading"
              >执行Ping测试</el-button
            >
            <el-button type="warning" @click="stopPing" v-else
              >停止测试</el-button
            >
            <el-button @click="resetPingForm" :disabled="isLoading"
              >重置</el-button
            >
          </el-form-item>
        </el-form>

        <!-- 执行日志区域 -->
        <div class="log-section">
          <div class="log-header">
            <h3>执行日志</h3>
            <div class="log-actions">
              <el-button type="primary" size="small" @click="clearLogs"
                >清空日志</el-button
              >
              <el-switch
                v-model="autoScroll"
                active-text="自动滚动"
                inactive-text="手动滚动"
              />
            </div>
          </div>
          <div class="log-content">
            <el-scrollbar ref="logScrollbar" class="log-scrollbar" always>
              <div class="system-log-container">
                <pre
                  v-for="(log, index) in systemLogs"
                  :key="index"
                  :class="log.type"
                  >{{ log.message }}</pre
                >
              </div>
            </el-scrollbar>
          </div>
        </div>
      </div>

      <div v-else class="tool-placeholder">
        <el-empty description="工具开发中，敬请期待">
          <template #image>
            <el-icon class="el-icon--big"><Tools /></el-icon>
          </template>
          <template #description>
            <p>{{ toolInfo?.name || currentTool }} 工具正在开发中，敬请期待</p>
          </template>
          <el-button type="primary" @click="$router.push('/network')"
            >返回网络工具列表</el-button
          >
        </el-empty>
      </div>
    </div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, reactive, nextTick } from "vue";
import { useRoute, useRouter } from "vue-router";
import {
  pingHost,
  stopPingTest,
  type PingResult,
  type PingOptions,
} from "../../services/networkService";
import type { FormInstance, FormRules } from "element-plus";
import { Tools } from "@element-plus/icons-vue";
import { ElMessage } from "element-plus";
import { EventsOn, EventsOff } from "../../../wailsjs/runtime/runtime";

const route = useRoute();
const router = useRouter();
const currentTool = computed(() => route.params.tool as string);
const isLoading = ref(false);
const pingResult = ref<PingResult | null>(null);
const pingForm = ref<FormInstance>();

interface SystemLog {
  message: string;
  type: "info" | "error" | "success" | "output";
}

// Ping工具参数
const pingParams = ref<PingOptions>({
  target: "",
  count: 4,
  interval: 1.0,
});

// Ping表单验证规则
const pingRules = reactive<FormRules>({
  target: [
    { required: true, message: "请输入目标主机或IP地址", trigger: "blur" },
    { min: 1, max: 255, message: "长度应在1-255个字符之间", trigger: "blur" },
  ],
});

const systemLogs = ref<SystemLog[]>([]);
const autoScroll = ref(true);
const logScrollbar = ref();

// 监听ping输出事件
onMounted(() => {
  EventsOn("ping:output", (output: string) => {
    addSystemLog(output, "output");
  });
});

// 清理事件监听
onUnmounted(() => {
  EventsOff("ping:output");
});

// 添加系统日志
const addSystemLog = (
  message: string,
  type: "info" | "error" | "success" | "output" = "info"
) => {
  const log: SystemLog = {
    message,
    type,
  };
  systemLogs.value.push(log);

  // 如果开启了自动滚动，等待DOM更新后滚动到底部
  if (autoScroll.value) {
    nextTick(() => {
      const scrollbar = logScrollbar.value?.wrap;
      if (scrollbar) {
        scrollbar.scrollTop = scrollbar.scrollHeight;
      }
    });
  }
};

// 清空日志
const clearLogs = () => {
  systemLogs.value = [];
};

// 提交Ping表单
const submitPingForm = () => {
  if (!pingForm.value) return;

  pingForm.value.validate(async (valid) => {
    if (valid) {
      await runPingTest();
    }
  });
};

// 停止Ping测试
const stopPing = async () => {
  try {
    await stopPingTest();
    addSystemLog("已停止Ping测试", "info");
  } catch (error) {
    console.error("停止Ping测试失败:", error);
    const errorMessage = error instanceof Error ? error.message : String(error);
    addSystemLog(`停止测试失败: ${errorMessage}`, "error");
  }
};

// 执行Ping测试
const runPingTest = async () => {
  if (!pingParams.value.target) {
    ElMessage.warning("请输入目标主机或IP地址");
    return;
  }

  isLoading.value = true;
  addSystemLog(
    `开始对 ${pingParams.value.target} 执行Ping测试 (间隔: ${pingParams.value.interval}秒)`,
    "info"
  );

  try {
    pingResult.value = await pingHost(pingParams.value);

    if (!pingResult.value.success) {
      addSystemLog(`Ping测试失败: ${pingResult.value.error}`, "error");
      ElMessage.error({
        message: pingResult.value.error,
        duration: 5000,
        showClose: true,
      });
    }
  } catch (error) {
    console.error("Ping测试失败:", error);
    const errorMessage = error instanceof Error ? error.message : String(error);
    addSystemLog(`执行出错: ${errorMessage}`, "error");
    ElMessage.error({
      message: `执行出错: ${errorMessage}`,
      duration: 5000,
      showClose: true,
    });
  } finally {
    isLoading.value = false;
  }
};

// 清空表单和结果
const resetPingForm = () => {
  if (pingForm.value) {
    pingForm.value.resetFields();
  }
  pingResult.value = null;
  addSystemLog("已重置所有测试结果", "info");
};

// 获取当前工具信息
const toolInfo = computed(() => {
  return networkTools.find((tool) => tool.route === currentTool.value);
});

// 网络工具信息
const networkTools = [
  {
    name: "Ping",
    description: "执行Ping测试",
    route: "ping",
  },
  // ... 其他工具配置 ...
];
</script>

<style scoped>
.network-tool {
  height: 100vh;
  padding: 20px;
  display: flex;
  flex-direction: column;
}

.tool-content {
  flex: 1;
  display: flex;
  flex-direction: column;
}

.tool-section {
  flex: 1;
  background: white;
  border-radius: 12px;
  padding: 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  display: flex;
  flex-direction: column;
}

.ping-form {
  margin-bottom: 20px;
}

.log-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  background: #f8f9fa;
  border-radius: 8px;
  padding: 20px;
  min-height: 0; /* 重要：允许flex子元素收缩 */
  overflow: hidden; /* 防止内容溢出 */
}

.log-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 15px;
}

.log-header h3 {
  margin: 0;
  font-size: 18px;
  color: #1a1a1a;
}

.log-actions {
  display: flex;
  gap: 15px;
  align-items: center;
}

.log-content {
  flex: 1;
  min-height: 0; /* 重要：允许flex子元素收缩 */
  position: relative;
}

.log-scrollbar {
  height: 100% !important;
}

.log-scrollbar :deep(.el-scrollbar__wrap) {
  overflow-x: hidden;
}

.system-log-container {
  padding: 10px;
  font-family: monospace;
  font-size: 13px;
  line-height: 1.4;
}

.system-log-container pre {
  margin: 2px 0;
  padding: 3px 0;
  white-space: pre-wrap;
  word-wrap: break-word;
}

.system-log-container .info {
  color: #909399;
}

.system-log-container .error {
  color: #f56c6c;
}

.system-log-container .success {
  color: #67c23a;
}

.system-log-container .output {
  color: #303133;
}

.tool-placeholder {
  flex: 1;
  background: white;
  border-radius: 12px;
  padding: 60px 24px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  text-align: center;
  display: flex;
  flex-direction: column;
  justify-content: center;
}

.el-icon--big {
  font-size: 60px;
  color: #909399;
}
</style>
