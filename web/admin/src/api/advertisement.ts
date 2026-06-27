import http from './index'

// 广告
export function getAds(params?: any) { return http.get('/ads', { params }) }
export function createAd(data: any) { return http.post('/ads', data) }
export function updateAd(id: string, data: any) { return http.put(`/ads/${id}`, data) }
export function deleteAd(id: string) { return http.delete(`/ads/${id}`) }
// 加盟
export function getFranchises(params?: any) { return http.get('/franchises', { params }) }
export function applyFranchise(data: any) { return http.post('/franchises', data) }
export function processFranchise(id: string, data: { status: string }) { return http.put(`/franchises/${id}`, data) }
// 微信用户
export function getWechatUsers(params?: any) { return http.get('/wechat-users', { params }) }
export function freezeWechatUser(id: string) { return http.put(`/wechat-users/${id}/freeze`) }
