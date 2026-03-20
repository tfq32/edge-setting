<template>
  <AppLayout>
    <div class="page">
      <!-- 顶栏：只读，无操作按钮 -->
      <div class="top-bar">
        <button class="back-btn" @click="$router.back()">‹</button>
        <div class="top-bar__info">
          <div class="app-name">{{ app?.id ?? route.params.id }}</div>
          <div class="app-status" :class="`st--${app?.status}`">
            {{ statusLabel[app?.status ?? ''] ?? '加载中' }}
          </div>
        </div>
        <div class="status-tag" :class="`tag--${app?.status}`">
          {{ statusLabel[app?.status ?? ''] ?? '—' }}
        </div>
      </div>

      <!-- Tab：详情 / 日志 -->
      <div class="tab-strip">
        <button v-for="t in ['详情','日志']" :key="t"
          class="tab-btn" :class="{ active: activeTab === t }"
          @click="activeTab = t">{{ t }}</button>
      </div>

      <van-pull-refresh v-model="refreshing" @refresh="onRefresh" class="pull-area">
        <div class="content">

          <!-- ① 详情 -->
          <template v-if="activeTab === '详情'">
            <div class="card" v-if="app">
              <div v-for="row in infoRows" :key="row.label" class="info-row">
                <span class="info-key">{{ row.label }}</span>
                <span class="info-val">{{ row.value }}</span>
              </div>
            </div>
            <div class="tip" v-else>加载中…</div>

            <div class="card" v-if="app?.status === 'running'">
              <div class="card-title">资源占用</div>
              <div class="res-row">
                <span class="res-lbl">CPU</span>
                <div class="res-track"><div class="res-fill"
                  :style="{ width: Math.min((app?.cpu??0)*2,100)+'%', background: cpuColor }"/></div>
                <span class="res-val">{{ app?.cpu.toFixed(1) }}%</span>
              </div>
              <div class="res-row">
                <span class="res-lbl">内存</span>
                <div class="res-track"><div class="res-fill"
                  :style="{ width: Math.min((app?.mem??0)/536870912*100,100)+'%', background:'var(--system)' }"/></div>
                <span class="res-val">{{ formatMem(app?.mem??0) }}</span>
              </div>
            </div>
          </template>

          <!-- ② 日志 -->
          <template v-else>
            <!-- 中文 Level Pills（替代 select） -->
            <div class="level-pills">
              <button v-for="lv in LEVELS" :key="lv.key"
                class="level-pill" :class="{ active: logLevel === lv.key, [`pill--${lv.key}`]: logLevel === lv.key }"
                @click="logLevel = lv.key; loadLogs()">{{ lv.label }}</button>
            </div>
            <!-- 关键字输入 -->
            <div class="search-box">
              <span class="search-icon">🔍</span>
              <input v-model="logKeyword" class="search-input" placeholder="搜索日志内容…"
                @keyup.enter="loadLogs" />
            </div>
            <!-- 日志卡片 -->
            <div v-if="logs.length === 0" class="tip">暂无日志</div>
            <div v-else class="log-list">
              <div v-for="log in logs" :key="log.id" class="log-card">
                <div class="log-card__head">
                  <span class="lv-badge" :class="`lv--${log.level}`">{{ lvZh[log.level] ?? log.level }}</span>
                  <span class="log-ts">{{ formatTs(log.ts) }}</span>
                </div>
                <div class="log-msg">{{ log.content }}</div>
              </div>
            </div>
          </template>

        </div>
      </van-pull-refresh>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useRoute } from 'vue-router'
import { useAppsStore } from '@/stores/apps'
import { appsApi } from '@/api'
import AppLayout from '@/components/AppLayout.vue'
import type { AppInfo } from '@/stores/apps'

const route = useRoute()
const appsStore = useAppsStore()
const id = route.params.id as string

const app      = ref<AppInfo | null>(null)
const activeTab = ref('详情')
const logs      = ref<any[]>([])
const logLevel  = ref('')       // 空串 = 全部
const logKeyword = ref('')
const refreshing = ref(false)

const statusLabel: Record<string,string> = { running:'运行中', stopped:'已停止', error:'异常' }

// 中文 Level 对照
const lvZh: Record<string,string> = { INFO:'信息', WARN:'警告', ERROR:'错误' }
const LEVELS = [
  { key:'',      label:'全部' },
  { key:'INFO',  label:'信息' },
  { key:'WARN',  label:'警告' },
  { key:'ERROR', label:'错误' },
]

const cpuColor = computed(() => {
  const v = app.value?.cpu ?? 0
  return v > 80 ? 'var(--danger)' : v > 50 ? 'var(--warning)' : 'var(--primary)'
})

const infoRows = computed(() => {
  if (!app.value) return []
  return [
    { label:'PID',    value: app.value.pid || '—' },
    { label:'版本',   value: app.value.version },
    { label:'端口',   value: app.value.ports?.join(', ') || '—' },
    { label:'类型',   value: app.value.type === 'system' ? '系统应用' : '用户应用' },
    { label:'启动时间', value: app.value.start_time ? new Date(app.value.start_time).toLocaleString() : '—' },
    { label:'工作目录', value: app.value.work_dir },
  ]
})

function formatMem(b:number) { return b ? (b/1048576).toFixed(0)+' MB' : '—' }
function formatTs(ts:number) {
  return new Date(ts).toLocaleTimeString('zh-CN',{hour12:false,hour:'2-digit',minute:'2-digit',second:'2-digit'})
}

async function loadApp()  { app.value = await appsStore.fetchDetail(id) }
async function loadLogs() {
  try {
    const params: Record<string,string> = { limit:'200' }
    if (logLevel.value)   params.level   = logLevel.value
    if (logKeyword.value) params.keyword = logKeyword.value
    const res:any = await appsApi.logs(id, params)
    logs.value = res?.data ?? []
  } catch {}
}

async function onRefresh() {
  await loadApp()
  if (activeTab.value === '日志') await loadLogs()
  refreshing.value = false
}
onMounted(async () => { await loadApp(); await loadLogs() })
</script>

<style scoped>
.page { height: 100%; display: flex; flex-direction: column; overflow: hidden; background: var(--bg); }

/* 顶栏 */
.top-bar {
  padding: 12px 16px 0; background: rgba(255,255,255,0.92);
  border-bottom: 1px solid var(--border-subtle);
  display: flex; align-items: center; gap: 10px; flex-shrink: 0; padding-bottom: 12px;
}
.back-btn { font-size: 28px; color: var(--sub-text); background: none; border: none; cursor: pointer; padding: 0; line-height: 1; flex-shrink: 0; }
.top-bar__info { flex: 1; min-width: 0; }
.app-name { font-size: 15px; font-weight: 700; color: var(--text); font-family: monospace;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.app-status { font-size: 12px; font-weight: 600; margin-top: 2px; }
.st--running { color: var(--success); }
.st--stopped { color: var(--muted-text); }
.st--error   { color: var(--warning); }

.status-tag { padding: 5px 13px; border-radius: 8px; font-size: 13px; font-weight: 700; flex-shrink: 0; }
.tag--running { background: var(--success-soft); color: var(--success); border: 1px solid rgba(16,185,129,.3); }
.tag--stopped { background: var(--danger-soft);  color: var(--danger);  border: 1px solid rgba(239,68,68,.2); }
.tag--error   { background: var(--warning-soft); color: var(--warning); border: 1px solid rgba(245,158,11,.3); }

/* Tab */
.tab-strip { display: flex; gap: 8px; padding: 10px 14px 8px; flex-shrink: 0; }
.tab-btn {
  padding: 7px 20px; border-radius: 20px; font-size: 13px; font-weight: 600;
  border: 1px solid var(--border); background: var(--primary-soft); color: var(--sub-text); cursor: pointer;
  transition: all .15s;
}
.tab-btn.active { background: var(--primary); color: #fff; border-color: var(--primary); }

.pull-area { flex: 1; overflow-y: auto; }
.content { padding: 8px 14px 0; }

/* 卡片 */
.card {
  background: var(--bg-white); border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle); box-shadow: 0 2px 12px var(--shadow);
  overflow: hidden; margin-bottom: 12px;
}
.card-title { font-size: 14px; font-weight: 700; color: var(--text); padding: 12px 14px 8px; }

/* 信息行 */
.info-row { display: flex; justify-content: space-between; align-items: center;
  padding: 12px 14px; border-bottom: 1px solid var(--border-subtle); gap: 12px; }
.info-row:last-child { border-bottom: none; }
.info-key { font-size: 14px; color: var(--muted-text); flex-shrink: 0; }
.info-val { font-size: 14px; color: var(--text); font-family: monospace; font-weight: 600;
  text-align: right; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; max-width: 58%; }

/* 资源 */
.res-row { display: flex; align-items: center; gap: 10px; padding: 0 14px 10px; }
.res-lbl   { font-size: 13px; color: var(--muted-text); width: 32px; }
.res-track { flex: 1; height: 6px; border-radius: 3px; background: var(--primary-soft); }
.res-fill  { height: 100%; border-radius: 3px; transition: width .4s; }
.res-val   { font-size: 13px; color: var(--text); font-weight: 600; width: 52px; text-align: right; }

/* 日志：中文 Level Pills */
.level-pills { display: flex; gap: 7px; margin-bottom: 10px; }
.level-pill {
  padding: 6px 13px; border-radius: 20px; font-size: 13px; font-weight: 600;
  border: 1px solid var(--border); background: var(--primary-soft); color: var(--sub-text); cursor: pointer;
  transition: all .15s;
}
.level-pill.active           { background: var(--primary);  color: #fff; border-color: var(--primary); }
.level-pill.pill--INFO.active  { background: var(--success); border-color: var(--success); }
.level-pill.pill--WARN.active  { background: var(--warning); border-color: var(--warning); }
.level-pill.pill--ERROR.active { background: var(--danger);  border-color: var(--danger); }

/* 搜索框 */
.search-box {
  display: flex; align-items: center; gap: 8px;
  background: var(--primary-soft); border: 1px solid var(--border);
  border-radius: 12px; padding: 8px 13px; margin-bottom: 12px;
}
.search-icon  { font-size: 15px; }
.search-input { flex: 1; border: none; background: transparent; font-size: 14px; color: var(--text); outline: none; }
.search-input::placeholder { color: var(--muted-text); }

/* 日志卡片 */
.log-list { display: flex; flex-direction: column; gap: 8px; padding-bottom: 16px; }
.log-card {
  background: var(--bg-white); border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle); box-shadow: 0 1px 6px var(--shadow);
  padding: 12px 14px;
}
.log-card__head { display: flex; align-items: center; gap: 8px; margin-bottom: 7px; }
.lv-badge {
  font-size: 11px; font-weight: 700; padding: 2px 8px; border-radius: 6px; flex-shrink: 0;
}
.lv--INFO  { background: rgba(16,185,129,.10); color: var(--success); }
.lv--WARN  { background: rgba(245,158,11,.10); color: var(--warning); }
.lv--ERROR { background: rgba(239,68,68,.10);  color: var(--danger); }
.log-ts    { font-size: 12px; color: var(--muted-text); font-family: monospace; }
.log-msg   { font-size: 14px; color: var(--text); line-height: 1.6; word-break: break-all; }

.tip { font-size: 14px; color: var(--muted-text); text-align: center; padding: 32px 0; }
</style>
