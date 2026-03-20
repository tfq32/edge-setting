import { defineStore } from 'pinia'
import { ref } from 'vue'
import { systemApi } from '@/api'

export interface Snapshot {
  ts:       number
  cpu:      number
  cpu_cores: number[]
  mem_used:  number
  mem_total: number
  mem_pct:   number
  swap_used: number
  swap_total: number
  disk_pct:  number
  disk_read: number
  disk_write: number
  net_in:    number
  net_out:   number
  load1:     number
  load5:     number
  load15:    number
  uptime:    number
}

export const useMetricsStore = defineStore('metrics', () => {
  const latest   = ref<Snapshot | null>(null)
  const history  = ref<Snapshot[]>([])
  const wsStatus = ref<'connecting' | 'open' | 'closed'>('closed')
  let ws: WebSocket | null = null
  let retryTimer: ReturnType<typeof setTimeout> | null = null
  let retryDelay = 1000

  async function fetchSnapshot() {
    try {
      const res: any = await systemApi.snapshot()
      if (res?.data) latest.value = res.data
    } catch {}
  }

  function connectWS() {
    const token = localStorage.getItem('edge_token') ?? ''
    const proto = location.protocol === 'https:' ? 'wss' : 'ws'
    const url   = `${proto}://${location.host}/api/v1/system/metrics/ws?token=${token}`

    wsStatus.value = 'connecting'
    ws = new WebSocket(url)

    ws.onopen = () => {
      wsStatus.value = 'open'
      retryDelay = 1000
    }

    ws.onmessage = (ev) => {
      try {
        const msg = JSON.parse(ev.data)
        if (msg.type === 'metrics' && msg.data) {
          latest.value = msg.data as Snapshot
          history.value.push(msg.data)
          if (history.value.length > 1200) history.value.shift()
        }
      } catch {}
    }

    ws.onclose = () => {
      wsStatus.value = 'closed'
      // 指数退避重连
      retryTimer = setTimeout(() => {
        retryDelay = Math.min(retryDelay * 2, 30000)
        connectWS()
      }, retryDelay)
    }
  }

  function disconnectWS() {
    if (retryTimer) clearTimeout(retryTimer)
    ws?.close()
    ws = null
    wsStatus.value = 'closed'
  }

  return { latest, history, wsStatus, fetchSnapshot, connectWS, disconnectWS }
})
