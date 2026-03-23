import axios from 'axios'

const http = axios.create({
  baseURL: '/api/v1',
  timeout: 10000,
})

// 响应拦截：统一错误处理
http.interceptors.response.use(
  res => res.data,
  err => Promise.reject(err)
)

export default http

// ── 系统 ──────────────────────────────────────────────────
export const systemApi = {
  info:     ()                            => http.get('/system/info'),
  snapshot: ()                            => http.get('/system/metrics'),
  history:  (type: string, range: string) => http.get('/metrics/history', { params: { type, range } }),
}

// ── 微应用（只有列表 + 启停重启） ─────────────────────────
export const appsApi = {
  list:    ()            => http.get('/apps'),
  start:   (id: string)  => http.post(`/apps/${id}/start`),
  stop:    (id: string)  => http.post(`/apps/${id}/stop`),
  restart: (id: string)  => http.post(`/apps/${id}/restart`),
}

// ── 审计 ──────────────────────────────────────────────────
export const auditApi = {
  query: (params?: Record<string, string>) => http.get('/audit', { params }),
}
