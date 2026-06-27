import http from './index'

// 站点完整信息（同步后端 16 字段模型）
export interface Site {
  id: string | number
  name: string
  address: string
  province: string
  city: string
  latitude: number
  longitude: number
  contact: string
  phone: string
  agent_id?: string | number
  operator_id?: string | number
  agent_name?: string
  operator_name?: string
  commission_rate: number
  billing_scheme_id?: string | number
  billing_scheme_name?: string
  business_hours: string
  image_url: string
  status: string
  device_count: number
  created_at: string
  updated_at: string
}

// 站点简要（供下拉选择）
export interface SiteBrief {
  id: string | number
  name: string
}

// 创建站点请求
export interface CreateSiteRequest {
  name: string
  address: string
  province: string
  city: string
  contact: string
  phone: string
  latitude: number
  longitude: number
  agent_id?: string | number
  operator_id?: string | number
  commission_rate: number
  billing_scheme_id?: string | number
  business_hours: string
  image_url: string
  status: string
}

// 更新站点请求
export interface UpdateSiteRequest {
  name?: string
  address?: string
  province?: string
  city?: string
  contact?: string
  phone?: string
  latitude?: number
  longitude?: number
  agent_id?: string | number
  operator_id?: string | number
  commission_rate?: number
  billing_scheme_id?: string | number
  business_hours?: string
  image_url?: string
  status?: string
}

// 列表查询参数
export interface SiteListParams {
  page?: number
  page_size?: number
  keyword?: string
  status?: string
  province?: string
  city?: string
  agent_id?: string | number
}

// 分页响应
export interface SiteListResponse {
  items: Site[]
  total: number
  page: number
  page_size: number
}

// 获取站点列表（分页）
export function getSitesManage(params?: SiteListParams) {
  return http.get<any, { code: string; data: SiteListResponse }>('/sites/manage', { params })
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
