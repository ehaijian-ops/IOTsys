import http from './index'

// 收费方案
export function getBillingSchemes(params?: any) {
  return http.get('/billing/schemes', { params })
}
export function getBillingScheme(id: string) {
  return http.get(`/billing/schemes/${id}`)
}
export function createBillingScheme(data: any) {
  return http.post('/billing/schemes', data)
}
export function updateBillingScheme(id: string, data: any) {
  return http.put(`/billing/schemes/${id}`, data)
}
export function deleteBillingScheme(id: string) {
  return http.delete(`/billing/schemes/${id}`)
}
export function batchSetPeriods(schemeId: string, data: { periods: any[] }) {
  return http.put(`/billing/schemes/${schemeId}/periods`, data)
}
export function getBillingPeriods(schemeId: string) {
  return http.get(`/billing/schemes/${schemeId}/periods`)
}
export function setBillingPeriods(schemeId: string, periods: any[]) {
  return http.put(`/billing/schemes/${schemeId}/periods`, { periods })
}
// 月卡方案
export function getMonthlySchemes(params?: any) {
  return http.get('/billing/monthly', { params })
}
export { getMonthlySchemes as getMonthlyCardSchemes }
export function createMonthlyScheme(data: any) {
  return http.post('/billing/monthly', data)
}
export { createMonthlyScheme as createMonthlyCardScheme }
export function updateMonthlyScheme(id: string, data: any) {
  return http.put(`/billing/monthly/${id}`, data)
}
export { updateMonthlyScheme as updateMonthlyCardScheme }
export function deleteMonthlyScheme(id: string) {
  return http.delete(`/billing/monthly/${id}`)
}
export { deleteMonthlyScheme as deleteMonthlyCardScheme }
// 充值方案
export function getRechargeSchemes(params?: any) {
  return http.get('/billing/recharges', { params })
}
export function createRechargeScheme(data: any) {
  return http.post('/billing/recharges', data)
}
export function updateRechargeScheme(id: string, data: any) {
  return http.put(`/billing/recharges/${id}`, data)
}
export function deleteRechargeScheme(id: string) {
  return http.delete(`/billing/recharges/${id}`)
}
// 业务配置
export function getConfigs() {
  return http.get('/billing/configs')
}
export function getConfig(key: string) {
  return http.get(`/billing/configs/${key}`)
}
export function setConfig(key: string, data: { value: string; description?: string }) {
  return http.put(`/billing/configs/${key}`, data)
}
