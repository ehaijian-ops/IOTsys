import http from './index'

// 角色
export function getRoles(params?: any) { return http.get('/system/roles', { params }) }
export function createRole(data: any) { return http.post('/system/roles', data) }
export function updateRole(id: string, data: any) { return http.put(`/system/roles/${id}`, data) }
export function deleteRole(id: string) { return http.delete(`/system/roles/${id}`) }
// 菜单
export function getMenuTree() { return http.get('/system/menus/tree') }
export { getMenuTree as getMenus }
export function createMenu(data: any) { return http.post('/system/menus', data) }
export function updateMenu(id: string, data: any) { return http.put(`/system/menus/${id}`, data) }
export function deleteMenu(id: string) { return http.delete(`/system/menus/${id}`) }
// 日志
export function getLoginLogs(params?: any) { return http.get('/system/login-logs', { params }) }
export function getOperationLogs(params?: any) { return http.get('/system/operation-logs', { params }) }
export { getOperationLogs as getSystemLogs }
