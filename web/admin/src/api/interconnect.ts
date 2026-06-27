import http from './index'

// 互联机构
export function getInterconnectOrgs(params?: any) { return http.get('/interconnect/orgs', { params }) }
export function createInterconnectOrg(data: any) { return http.post('/interconnect/orgs', data) }
export function getInterconnectOrg(id: string) { return http.get(`/interconnect/orgs/${id}`) }
export function updateInterconnectOrg(id: string, data: any) { return http.put(`/interconnect/orgs/${id}`, data) }
export function deleteInterconnectOrg(id: string) { return http.delete(`/interconnect/orgs/${id}`) }
// 互联密钥
export function getInterconnectKeys(params?: any) { return http.get('/interconnect/keys', { params }) }
export function createInterconnectKey(data: any) { return http.post('/interconnect/keys', data) }
export function updateInterconnectKey(id: string, data: any) { return http.put(`/interconnect/keys/${id}`, data) }
export function deleteInterconnectKey(id: string) { return http.delete(`/interconnect/keys/${id}`) }
