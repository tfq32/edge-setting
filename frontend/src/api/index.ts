import axios from 'axios'

const http = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// ── 响应拦截：统一错误处理 ────────────────────────────────
http.interceptors.response.use(
  res => res.data,
  err => Promise.reject(err)
)

export default http

/* ── API 模块 ────────────────────────────────────────────── */

// 系统
export const systemApi = {
  info:     ()                             => http.get('/system/info'),
  snapshot: ()                             => http.get('/system/metrics'),
  history:  (type: string, range: string)  => http.get('/metrics/history', { params: { type, range } }),
  version:  ()                             => http.get('/version'),
}

// 微应用
export const appsApi = {
  list:       ()                                       => http.get('/apps'),
  detail:     (id: string)                             => http.get(`/apps/${id}`),
  start:      (id: string)                             => http.post(`/apps/${id}/start`),
  stop:       (id: string)                             => http.post(`/apps/${id}/stop`),
  restart:    (id: string)                             => http.post(`/apps/${id}/restart`),
  logs:       (id: string, params?: Record<string, string>) => http.get(`/apps/${id}/logs`, { params }),
  exportLogs: (id: string)                             => axios.get(`/api/v1/apps/${id}/logs/export`, { responseType: 'blob' }),
}

// 审计
export const auditApi = {
  query: (params?: Record<string, string>) => http.get('/audit', { params }),
}
