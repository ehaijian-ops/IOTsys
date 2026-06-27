import http from './index'

// 故障
export function getFaults(params?: any) { return http.get('/maintenance/faults', { params }) }
export function createFault(data: any) { return http.post('/faults', data) }
export function handleFault(id: string, data: { result: string }) { return http.put(`/maintenance/faults/${id}`, data) }
// 定时任务
export function getTasks(params?: any) { return http.get('/maintenance/tasks', { params }) }
export function createTask(data: any) { return http.post('/maintenance/tasks', data) }
export function updateTask(id: string, data: any) { return http.put(`/maintenance/tasks/${id}`, data) }
export function deleteTask(id: string) { return http.delete(`/maintenance/tasks/${id}`) }
export function getTaskLogs(id: string, params?: any) { return http.get(`/maintenance/tasks/${id}/logs`, { params }) }
// 下载
export function getDownloads(params?: any) { return http.get('/maintenance/downloads', { params }) }
export function createDownload(data: any) { return http.post('/maintenance/downloads', data) }
