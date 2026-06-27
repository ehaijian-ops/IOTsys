import http from './index'

// IC卡
export function getICCards(params?: any) { return http.get('/cards/ic', { params }) }
export function createICCard(data: { card_no: string; card_uid?: string }) { return http.post('/cards/ic', data) }
export function getICCard(id: string) { return http.get(`/cards/ic/${id}`) }
export function updateICCard(id: string, data: any) { return http.put(`/cards/ic/${id}`, data) }
export function rechargeICCard(id: string, data: { amount: number; remark?: string }) { return http.post(`/cards/ic/${id}/recharge`, data) }
export function bindICCard(id: string, data: { user_id: number }) { return http.post(`/cards/ic/${id}/bind`, data) }
export function reportLostICCard(id: string) { return http.post(`/cards/ic/${id}/lost`) }
export function deleteICCard(id: string) { return http.delete(`/cards/ic/${id}`) }
export function batchImportICCards(data: { card_nos: string[] }) { return http.post('/cards/ic/batch-import', data) }
// 流量卡
export function getTrafficCards(params?: any) { return http.get('/cards/traffic', { params }) }
export function createTrafficCard(data: any) { return http.post('/cards/traffic', data) }
export function updateTrafficCard(id: string, data: any) { return http.put(`/cards/traffic/${id}`, data) }
export function deleteTrafficCard(id: string) { return http.delete(`/cards/traffic/${id}`) }
export function bindTrafficCard(id: string, data: { device_id: string }) { return http.post(`/cards/traffic/${id}/bind`, data) }
// 月卡
export function getMonthlyCards(params?: any) { return http.get('/cards/monthly', { params }) }
