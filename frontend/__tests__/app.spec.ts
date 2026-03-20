import { describe, it, expect, vi, beforeEach } from 'vitest'
import { mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import { createRouter, createMemoryHistory } from 'vue-router'

// ── MetricCard 测试 ───────────────────────────────────────
const MetricCard = {
  name: 'MetricCard',
  props: {
    label: String,
    value: { type: Number, default: 0 },
    pct:   { type: Number, default: 0 },
    unit:  { type: String, default: '%' },
    warn:  { type: Number, default: 80 },
    danger: { type: Number, default: 90 },
  },
  computed: {
    status() {
      if (this.pct >= this.danger) return 'danger'
      if (this.pct >= this.warn)   return 'warning'
      return 'normal'
    }
  },
  template: `<div class="metric-card" :data-status="status">
    <span class="label">{{ label }}</span>
    <span class="value">{{ pct.toFixed(1) }}{{ unit }}</span>
    <span v-if="status !== 'normal'" class="badge">{{ status }}</span>
  </div>`
}

describe('MetricCard', () => {
  it('正常状态下不显示告警徽章', () => {
    const w = mount(MetricCard, { props: { label: 'CPU', pct: 50, warn: 80, danger: 90 } })
    expect(w.find('.badge').exists()).toBe(false)
    expect(w.find('[data-status="normal"]').exists()).toBe(true)
  })

  it('超过 warn 阈值显示 warning', () => {
    const w = mount(MetricCard, { props: { label: 'CPU', pct: 85, warn: 80, danger: 90 } })
    expect(w.find('[data-status="warning"]').exists()).toBe(true)
    expect(w.find('.badge').text()).toBe('warning')
  })

  it('超过 danger 阈值显示 danger', () => {
    const w = mount(MetricCard, { props: { label: 'CPU', pct: 95, warn: 80, danger: 90 } })
    expect(w.find('[data-status="danger"]').exists()).toBe(true)
  })

  it('恰好等于阈值触发状态', () => {
    const warn = mount(MetricCard, { props: { label: 'CPU', pct: 80, warn: 80, danger: 90 } })
    expect(warn.find('[data-status="warning"]').exists()).toBe(true)
    const danger = mount(MetricCard, { props: { label: 'CPU', pct: 90, warn: 80, danger: 90 } })
    expect(danger.find('[data-status="danger"]').exists()).toBe(true)
  })

  it('正确渲染 label 和 value', () => {
    const w = mount(MetricCard, { props: { label: '内存', pct: 62.5, unit: '%' } })
    expect(w.find('.label').text()).toBe('内存')
    expect(w.find('.value').text()).toContain('62.5')
  })
})

// ── AppCard 测试 ──────────────────────────────────────────
const AppCard = {
  name: 'AppCard',
  props: { app: Object },
  computed: {
    statusLabel() {
      return { running: '运行中', stopped: '已停止', error: '异常' }[this.app.status] ?? ''
    },
    formattedMem() {
      const b = this.app.mem
      return b ? (b / 1024 / 1024).toFixed(0) + ' MB' : '—'
    }
  },
  template: `<div class="app-card" :class="'app-card--' + app.status">
    <span class="app-name">{{ app.id }}</span>
    <span class="app-status">{{ statusLabel }}</span>
    <span class="app-cpu">{{ app.cpu.toFixed(1) }}%</span>
    <span class="app-mem">{{ formattedMem }}</span>
  </div>`
}

describe('AppCard', () => {
  const makeApp = (overrides = {}) => ({
    id: 'test-app', name: '测试应用', type: 'user',
    status: 'running', cpu: 2.5, mem: 52428800,
    ports: [8080], version: 'v1.0.0', start_time: Date.now(), work_dir: '/opt',
    ...overrides
  })

  it('running 状态显示"运行中"', () => {
    const w = mount(AppCard, { props: { app: makeApp({ status: 'running' }) } })
    expect(w.find('.app-status').text()).toBe('运行中')
    expect(w.classes()).toContain('app-card--running')
  })

  it('stopped 状态显示"已停止"', () => {
    const w = mount(AppCard, { props: { app: makeApp({ status: 'stopped', cpu: 0, mem: 0 }) } })
    expect(w.find('.app-status').text()).toBe('已停止')
  })

  it('error 状态显示"异常"', () => {
    const w = mount(AppCard, { props: { app: makeApp({ status: 'error' }) } })
    expect(w.find('.app-status').text()).toBe('异常')
  })

  it('正确格式化内存（MB）', () => {
    const w = mount(AppCard, { props: { app: makeApp({ mem: 104857600 }) } }) // 100 MB
    expect(w.find('.app-mem').text()).toBe('100 MB')
  })

  it('内存为 0 时显示破折号', () => {
    const w = mount(AppCard, { props: { app: makeApp({ mem: 0 }) } })
    expect(w.find('.app-mem').text()).toBe('—')
  })

  it('正确显示 CPU 值', () => {
    const w = mount(AppCard, { props: { app: makeApp({ cpu: 12.34 }) } })
    expect(w.find('.app-cpu').text()).toContain('12.3')
  })
})

// ── Pinia metrics store 测试 ──────────────────────────────
import { useMetricsStore } from '../src/stores/metrics'

describe('useMetricsStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
  })

  it('初始状态正确', () => {
    const store = useMetricsStore()
    expect(store.latest).toBeNull()
    expect(store.history).toHaveLength(0)
    expect(store.wsStatus).toBe('closed')
  })

  it('fetchSnapshot 出错时不抛出异常', async () => {
    const store = useMetricsStore()
    vi.stubGlobal('fetch', vi.fn().mockRejectedValue(new Error('网络错误')))
    await expect(store.fetchSnapshot()).resolves.not.toThrow()
  })
})

// ── Pinia apps store 测试 ────────────────────────────────
import { useAppsStore } from '../src/stores/apps'
import * as apiModule from '../src/api/index'

describe('useAppsStore', () => {
  beforeEach(() => {
    setActivePinia(createPinia())
    vi.restoreAllMocks()
  })

  const mockApps = [
    { id: 'app-a', name: '应用A', type: 'system', status: 'running', cpu: 1.0, mem: 10485760, ports: [], version: 'v1', start_time: 0, work_dir: '/', has_ui: false },
    { id: 'app-b', name: '应用B', type: 'user',   status: 'stopped', cpu: 0,   mem: 0,        ports: [], version: 'v1', start_time: 0, work_dir: '/', has_ui: false },
  ]

  it('fetchList 成功后填充 list', async () => {
    vi.spyOn(apiModule.appsApi, 'list').mockResolvedValue({ code: 0, data: mockApps } as any)
    const store = useAppsStore()
    await store.fetchList()
    expect(store.list).toHaveLength(2)
    expect(store.list[0].id).toBe('app-a')
  })

  it('fetchList 失败后 list 保持为空', async () => {
    vi.spyOn(apiModule.appsApi, 'list').mockRejectedValue(new Error('fail'))
    const store = useAppsStore()
    await store.fetchList()
    expect(store.list).toHaveLength(0)
  })

  it('loading 在请求期间为 true', async () => {
    let resolve: any
    vi.spyOn(apiModule.appsApi, 'list').mockImplementation(() => new Promise(r => { resolve = r }))
    const store = useAppsStore()
    const p = store.fetchList()
    expect(store.loading).toBe(true)
    resolve({ code: 0, data: [] })
    await p
    expect(store.loading).toBe(false)
  })
})

// ── Router 测试 ───────────────────────────────────────────
describe('Router 导航', () => {
  it('根路径正常导航到首页', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/',       component: { template: '<div/>' } },
        { path: '/monitor', component: { template: '<div/>' } },
        { path: '/apps',   component: { template: '<div/>' } },
        { path: '/about',  component: { template: '<div/>' } },
      ]
    })
    await router.push('/')
    expect(router.currentRoute.value.path).toBe('/')
  })

  it('各主路由可正常跳转', async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { path: '/',       component: { template: '<div/>' } },
        { path: '/monitor', component: { template: '<div/>' } },
        { path: '/apps',   component: { template: '<div/>' } },
        { path: '/about',  component: { template: '<div/>' } },
      ]
    })
    for (const path of ['/', '/monitor', '/apps', '/about']) {
      await router.push(path)
      expect(router.currentRoute.value.path).toBe(path)
    }
  })
})

// ── 工具函数测试 ──────────────────────────────────────────
describe('工具函数', () => {
  function formatUptime(s: number): string {
    const d = Math.floor(s / 86400)
    const h = Math.floor((s % 86400) / 3600)
    const m = Math.floor((s % 3600) / 60)
    return d > 0 ? `${d}天${h}时` : `${h}时${m}分`
  }

  it('formatUptime 天级别', () => {
    expect(formatUptime(86400 + 3600)).toBe('1天1时')
  })
  it('formatUptime 小时级别', () => {
    expect(formatUptime(3600 * 5 + 60 * 23)).toBe('5时23分')
  })
  it('formatUptime 零值', () => {
    expect(formatUptime(0)).toBe('0时0分')
  })

  function formatMem(bytes: number): string {
    if (!bytes) return '—'
    return (bytes / 1024 / 1024).toFixed(0) + ' MB'
  }

  it('formatMem 正常值', () => {
    expect(formatMem(104857600)).toBe('100 MB')
  })
  it('formatMem 零值返回破折号', () => {
    expect(formatMem(0)).toBe('—')
  })
})
