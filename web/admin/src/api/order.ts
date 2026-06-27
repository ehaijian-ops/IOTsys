import http from './index'

// ========== 订单管理 ==========
export interface OrderQuery {
  order_sn?: string
  order_type?: string
  status?: string
  device_id?: string
  user_id?: string
  start_date?: string
  end_date?: string
  page?: number
  page_size?: number
}

export function getOrders(params: OrderQuery) {
  return http.get('/orders', { params })
}
export function getOrder(id: string) {
  return http.get(`/orders/${id}`)
}
export function createOrder(data: any) {
  return http.post('/orders', data)
}
export function startCharging(id: string) {
  return http.put(`/orders/${id}/start`)
}
export function endOrder(id: string, data: { amount: number; energy_kwh: number }) {
  return http.put(`/orders/${id}/end`, data)
}
export function cancelOrder(id: string) {
  return http.put(`/orders/${id}/cancel`)
}
export function deleteOrder(id: string) {
  return http.delete(`/orders/${id}`)
}
export function getChargeCurve(id: string) {
  return http.get(`/orders/${id}/curve`)
}
export function refundOrder(id: string, data: { amount: number; reason: string }) {
  return http.post(`/orders/${id}/refund`, data)
}
export function processRefund(orderId: string, refundId: string, data: { status: string; trade_no?: string }) {
  return http.put(`/orders/${orderId}/refund/${refundId}`, data)
}
