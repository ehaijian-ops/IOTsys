import http from './index'

// 站点完整信息
export interface Site {
  id: string
  name: string
  address: string
  latitude: number
  longitude: number
  contact: string
  phone: string
  status: string
  device_count: number
  created_at: string
  updated_at: string
}

// 站点简要（供下拉选择）
export interface SiteBrief {
  id: string
  name: string
}

// 创建站点请求
export interface CreateSiteRequest {
  name: string
  address: string
  contact: string
  phone: string
  latitude: number
  longitude: number
  status: string
}

// 更新站点请求
export interface UpdateSiteRequest {
  name?: string
  address?: string
  contact?: string
  phone?: string
  latitude?: number
  longitude?: number
  status?: string
}

// 获取站点列表（CRUD用）
export function getSitesManage() {
  return http.get<any, { code: string; data: Site[] }>('/sites/manage')
}

// 获取站点详情
export function getSiteDetail(id: string) {
  return http.get<any, { code: string; data: Site }>(`/sites/manage/${id}`)
}

// 创建站点
export function createSite(data: CreateSiteRequest) {
  return http.post('/sites/manage', data)
}

// 更新站点
export function updateSite(id: string, data: UpdateSiteRequest) {
  return http.put(`/sites/manage/${id}`, data)
}

// 删除站点
export function deleteSite(id: string) {
  return http.delete(`/sites/manage/${id}`)
}

// 获取站点列表（简要，供下拉选择）
export function getSitesBrief() {
  return http.get<any, { code: string; data: SiteBrief[] }>('/sites')
}
