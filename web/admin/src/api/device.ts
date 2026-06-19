import http from './index'

// 设备相关 API
export interface Device {
  id: string
  sn: string
  device_type: string
  protocol: string
  vendor: string
  model: string
  site_id: string
  install_location: string
  firmware_version: string
  status: string
  last_online_at: string
  created_at: string
  updated_at: string
}

export interface DeviceDetail extends Device {
  is_online: boolean
  realtime_data: Record<string, string>
}

export interface DeviceQuery {
  device_type?: string
  protocol?: string
  status?: string
  site_id?: string
  keyword?: string
  page?: number
  page_size?: number
}

export interface PaginatedResponse<T> {
  code: string
  data: T[]
  total: number
  page: number
}

// 获取设备列表
export function getDevices(params: DeviceQuery) {
  return http.get<any, PaginatedResponse<Device>>('/devices', { params })
}

// 获取设备详情
export function getDevice(id: string) {
  return http.get<any, { code: string; data: DeviceDetail }>(`/devices/${id}`)
}

// 创建设备
export function createDevice(data: Partial<Device>) {
  return http.post('/devices', data)
}

// 更新设备
export function updateDevice(id: string, data: Partial<Device>) {
  return http.put(`/devices/${id}`, data)
}

// 删除设备
export function deleteDevice(id: string) {
  return http.delete(`/devices/${id}`)
}

// 下发指令
export function sendCommand(deviceId: string, cmdType: string, params: Record<string, any>) {
  return http.post(`/devices/${deviceId}/commands`, { cmd_type: cmdType, params })
}

// 获取设备指令列表
export function getDeviceCommands(deviceId: string, params: { page?: number; page_size?: number }) {
  return http.get(`/devices/${deviceId}/commands`, { params })
}
