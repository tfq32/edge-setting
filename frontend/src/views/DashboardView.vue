<template>
  <AppLayout>
    <div class="page">
      <div class="top-bar">
        <div>
          <div class="top-bar__title">设置</div>
        </div>
        <div class="uptime" v-if="snap">运行 {{ formatUptime(snap.uptime) }}</div>
      </div>

      <van-pull-refresh v-model="refreshing" @refresh="onRefresh" class="pull-area">
        <div class="content">
          <!-- 2×2 指标卡 -->
          <div class="metric-grid">
            <MetricCard label="CPU 占用" :pct="snap?.cpu ?? 0" :value="snap?.cpu ?? 0" :warn="80" :danger="90"
              @click="$router.push('/monitor')" />
            <MetricCard label="内存使用" :pct="snap?.mem_pct ?? 0" :value="snap?.mem_pct ?? 0" :warn="85" :danger="92"
              @click="$router.push('/monitor')" />
            <MetricCard label="磁盘使用" :pct="snap?.disk_pct ?? 0" :value="snap?.disk_pct ?? 0" :warn="85" :danger="90"
              @click="$router.push('/monitor')" />
            <MetricCard label="网络流入" :pct="Math.min((snap?.net_in??0)/1048576*10,100)"
              :value="snap?.net_in??0" unit=" MB/s" :warn="80" :danger="95"
              @click="$router.push('/monitor')" />
          </div>

          <!-- 微应用状态 -->
          <div class="card">
            <div class="card-header">
              <span class="card-title">微应用状态</span>
              <span class="view-all" @click="$router.push('/apps')">查看全部 →</span>
            </div>

            <div v-if="apps.loading" class="tip">加载中…</div>
            <template v-else>
              <!-- 统计行：只有运行中 + 已停止 -->
              <div class="stat-row">
                <div class="stat-chip chip--success">
                  <span class="stat-num">{{ runningCount }}</span>
                  <span class="stat-lbl">运行中</span>
                </div>
                <div class="stat-chip chip--danger">
                  <span class="stat-num">{{ stoppedCount }}</span>
                  <span class="stat-lbl">已停止</span>
                </div>
              </div>
              <!-- 应用行 -->
              <div class="app-rows">
                <div v-for="app in apps.list.slice(0, 5)" :key="app.id"
                  class="app-row" @click="$router.push(`/apps/${app.id}`)">
                  <span class="status-dot" :class="`dot--${app.status}`" />
                  <span class="app-name">{{ app.id }}</span>
                  <span class="app-cpu">{{ app.cpu.toFixed(1) }}%</span>
                  <span class="app-mem">{{ formatMem(app.mem) }}</span>
                </div>
              </div>
            </template>
          </div>
        </div>
      </van-pull-refresh>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useMetricsStore } from '@/stores/metrics'
import { useAppsStore } from '@/stores/apps'
import AppLayout from '@/components/AppLayout.vue'
import MetricCard from '@/components/MetricCard.vue'

const metrics = useMetricsStore()
const apps    = useAppsStore()
const snap    = computed(() => metrics.latest)
const refreshing = ref(false)

const runningCount = computed(() => apps.list.filter(a => a.status === 'running').length)
const stoppedCount = computed(() => apps.list.filter(a => a.status === 'stopped').length)

function formatUptime(s: number) {
  const d = Math.floor(s / 86400), h = Math.floor((s % 86400) / 3600)
  return d > 0 ? `${d}天 ${h}时` : `${h}时`
}
function formatMem(b: number) { return b ? (b / 1048576).toFixed(0) + 'MB' : '—' }

async function onRefresh() {
  await Promise.all([metrics.fetchSnapshot(), apps.fetchList()])
  refreshing.value = false
}
onMounted(async () => {
  metrics.connectWS()
  await metrics.fetchSnapshot()
  await apps.fetchList()
})
onUnmounted(() => metrics.disconnectWS())
</script>

<style scoped>
.page { height: 100%; display: flex; flex-direction: column; overflow: hidden; }

.top-bar {
  padding: 14px 18px 10px;
  background: rgba(255,255,255,0.92);
  border-bottom: 1px solid var(--border-subtle);
  display: flex; align-items: center; justify-content: space-between;
  flex-shrink: 0;
}
.top-bar__title { font-size: 20px; font-weight: 700; color: var(--text); }
.uptime { font-size: 13px; color: var(--primary); font-weight: 600; }

.pull-area { flex: 1; overflow-y: auto; }
.content { padding: 12px 14px 0; }

.metric-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-bottom: 12px; }

/* 微应用状态卡 */
.card {
  background: var(--bg-white); border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle); box-shadow: 0 2px 12px var(--shadow);
  padding: 14px 16px; margin-bottom: 14px;
}
.card-header { display: flex; justify-content: space-between; align-items: center; margin-bottom: 14px; }
.card-title  { font-size: 16px; font-weight: 700; color: var(--text); }
.view-all    { font-size: 13px; color: var(--primary); cursor: pointer; }

.stat-row { display: grid; grid-template-columns: 1fr 1fr; gap: 10px; margin-bottom: 14px; }
.stat-chip { border-radius: var(--radius-md); padding: 12px 0; text-align: center; }
.chip--success { background: var(--success-soft); }
.chip--danger  { background: var(--danger-soft);  }
.stat-num { display: block; font-size: 28px; font-weight: 700; }
.chip--success .stat-num { color: var(--success); }
.chip--danger  .stat-num { color: var(--danger);  }
.stat-lbl { display: block; font-size: 12px; margin-top: 3px; opacity: .85; }
.chip--success .stat-lbl { color: var(--success); }
.chip--danger  .stat-lbl { color: var(--danger);  }

.app-rows { display: flex; flex-direction: column; }
.app-row {
  display: flex; align-items: center; gap: 10px;
  padding: 10px 0; border-bottom: 1px solid var(--border-subtle); cursor: pointer;
}
.app-row:last-child { border-bottom: none; }

.status-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.dot--running { background: var(--success); animation: blink 2.4s infinite; }
.dot--stopped { background: var(--muted-text); }

.app-name { font-size: 13px; color: var(--text); font-family: monospace; flex: 1;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.app-cpu  { font-size: 12px; color: var(--muted-text); flex-shrink: 0; }
.app-mem  { font-size: 12px; color: var(--muted-text); width: 46px; text-align: right; flex-shrink: 0; }
.tip      { font-size: 13px; color: var(--muted-text); text-align: center; padding: 20px 0; }
</style>
