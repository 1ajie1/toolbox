// networkService.ts
import { NetworkPing } from "../../wailsjs/go/main/App";
import { network } from "../../wailsjs/go/models";

declare global {
  interface Window {
    runtime?: {
      [key: string]: any;
    };
    go?: {
      main?: {
        App?: {
          NetworkPing?: typeof NetworkPing;
        };
      };
    };
  }
}

export interface PingOptions {
  target: string;
  count: number;
  interval: number;
}

export type PingResult = network.PingResult;

// 检查 Wails 运行时是否已初始化
const isWailsReady = (): boolean => {
  return !!(window.runtime && window.go?.main?.App?.NetworkPing);
};

// 等待 Wails 运行时初始化
const waitForWailsRuntime = async (maxAttempts = 50): Promise<void> => {
  let attempts = 0;

  while (!isWailsReady()) {
    if (attempts >= maxAttempts) {
      throw new Error("等待Wails运行时初始化超时");
    }
    await new Promise((resolve) => setTimeout(resolve, 100));
    attempts++;
  }
};

export async function pingHost(options: PingOptions): Promise<PingResult> {
  try {
    // 添加调试信息
    console.log("Attempting to ping with options:", options);
    console.log("Current runtime status:", {
      hasRuntime: !!window.runtime,
      hasGo: !!window.go,
      hasMain: !!window.go?.main,
      hasApp: !!window.go?.main?.App,
      hasNetworkPing: !!window.go?.main?.App?.NetworkPing,
    });

    // 等待 Wails 运行时初始化
    await waitForWailsRuntime();

    const result = await NetworkPing(options.target, options.count);
    console.log("Ping result:", result);
    return result;
  } catch (error) {
    console.error("Ping执行失败:", error);
    // 返回更详细的错误信息
    return {
      success: false,
      avgLatency: "",
      packetLoss: "",
      error: `执行失败: ${
        error instanceof Error ? error.message : String(error)
      }`,
      outputLines: [],
    };
  }
}

// 停止Ping测试
export async function stopPingTest(): Promise<void> {
  return window.go?.main?.App?.NetworkStopPing();
}
