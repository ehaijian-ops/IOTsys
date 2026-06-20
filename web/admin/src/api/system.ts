import http from './index'

export interface ServiceStatus {
  status: string        // running / stopped / disabled / error
  details: string
}

export interface ServerInfo {
  name: string
  version: string
  uptime: string
  start_time: string
  env: string
  port: number
}

export interface ServicesInfo {
  mysql: ServiceStatus
  redis: ServiceStatus
  mongodb: ServiceStatus
  kafka: ServiceStatus
  tcp_server: ServiceStatus
  sse_hub: ServiceStatus
}

export interface ResourceInfo {
  goroutines: number
  heap_alloc_mb: string
  num_cpu: number
  tcp_connections: number
  sse_clients: number
}

export interface SystemStatus {
  server: ServerInfo
  services: ServicesInfo
  resources: ResourceInfo
  checked_at: string
}

export function getSystemStatus(): Promise<SystemStatus> {
  return http.get('/system/status') as any
}
