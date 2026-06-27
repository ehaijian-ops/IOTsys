import http from './index'

// 钱包
export function getWallet(userId: string) {
  return http.get('/finance/wallet', { params: { user_id: userId } })
}
// 充值
export function recharge(data: { user_id: number; amount: number; bonus_amount?: number; pay_method: string; trade_no?: string }) {
  return http.post('/recharge', data)
}
export function adminRecharge(data: { user_id: number; amount: number; remark?: string }) {
  return http.post('/finance/recharge', data)
}
export function getRecharges(params?: any) {
  return http.get('/finance/recharges', { params })
}
// 提现
export function applyWithdraw(data: { amount: number; bank_name?: string; bank_card_no?: string; bank_account?: string }) {
  return http.post('/withdraw', data)
}
export function getWithdraws(params?: any) {
  return http.get('/finance/withdraws', { params })
}
export function processWithdraw(id: string, data: { status: string; actual_amount?: number; remark?: string }) {
  return http.put(`/finance/withdraws/${id}`, data)
}
// 分成
export function getSplits(params?: any) {
  return http.get('/finance/splits', { params })
}
