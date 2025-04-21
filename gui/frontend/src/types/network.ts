export interface PingResult {
  success: boolean;
  avgLatency?: string;
  packetLoss?: string;
  error?: string;
  outputLines?: string[];
}

export interface DNSResult {
  success: boolean;
  domain?: string;
  ipList?: string[];
  error?: string;
  serverUsed?: string;
}

export interface SpeedTestConfig {
  host: string;
  port: number;
  dataSize: number;
}
