import http from './index'

// 代理商
export function getAgents(params?: any) { return http.get('/agents', { params }) }
export function createAgent(data: any) { return http.post('/agents', data) }
export function getAgent(id: string) { return http.get(`/agents/${id}`) }
export function updateAgent(id: string, data: any) { return http.put(`/agents/${id}`, data) }
export function deleteAgent(id: string) { return http.delete(`/agents/${id}`) }
// 运营商
export function getOperators(params?: any) { return http.get('/operators', { params }) }
export function createOperator(data: any) { return http.post('/operators', data) }
export function getOperator(id: string) { return http.get(`/operators/${id}`) }
export function updateOperator(id: string, data: any) { return http.put(`/operators/${id}`, data) }
export function deleteOperator(id: string) { return http.delete(`/operators/${id}`) }
