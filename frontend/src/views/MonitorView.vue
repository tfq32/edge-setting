<template>
  <AppLayout>
    <div class="page">
      <div class="top-bar">
        <span class="top-bar__title">系统监控</span>
        <span class="ws-dot" :class="`ws--${metrics.wsStatus}`" />
      </div>
      <div class="metric-tabs">
        <button v-for="t in TABS" :key="t.key"
          class="m-tab" :class="{ active: activeTab === t.key }"
          @click="switchTab(t.key)">{{ t.label }}</button>
      </div>

      <van-pull-refresh v-model="refreshing" @refresh="onRefresh" class="pull-area">
        <div class="content">
          <!-- 大仪表 -->
          <div class="card gauge-card">
            <div class="gauge-wrap">
              <svg viewBox="0 0 72 72" width="84" height="84">
                <circle cx="36" cy="36" r="28" fill="none" :stroke="trackColor" stroke-width="8"/>
                <circle cx="36" cy="36" r="28" fill="none" :stroke="ringColor"
                  stroke-width="8"
                  :stroke-dasharray="`${(currentPct/100)*(2*Math.PI*28)} ${2*Math.PI*28}`"
                  stroke-linecap="round" transform="rotate(-90 36 36)"/>
              </svg>
              <div class="gauge-center-val">{{ currentPct.toFixed(0) }}%</div>
            </div>
            <div class="gauge-info">
              <div class="gauge-label">{{ currentTab?.label }}</div>
              <div class="gauge-main-val">{{ currentDisplay }}</div>
              <div class="gauge-status-badge" :class="statusClass">{{ statusText }}</div>
            </div>
          </div>

          <!-- 趋势图 -->
          <div class="card chart-card">
            <div class="chart-head">
              <span class="chart-title">历史趋势</span>
              <div class="range-pills">
                <button v-for="r in RANGES" :key="r"
                  class="range-pill" :class="{ active: activeRange === r }"
                  @click="changeRange(r)">{{ r }}</button>
              </div>
            </div>
            <EChartsWrapper :option="chartOption" height="150px" />
          </div>

          <!-- CPU 核心 -->
          <div class="card" v-if="activeTab === 'cpu' && snap?.cpu_cores?.length">
            <div class="card-title">核心占用</div>
            <div v-for="(v, i) in snap.cpu_cores" :key="i" class="core-row">
              <span class="core-lbl">核心 {{ i }}</span>
              <div class="core-track">
                <div class="core-fill" :style="{ width: v + '%', background: coreColor(v) }"/>
              </div>
              <span class="core-val">{{ v.toFixed(0) }}%</span>
            </div>
          </div>

          <!-- 内存详情 -->
          <div class="card" v-if="activeTab === 'mem' && snap">
            <div class="card-title">内存详情</div>
            <div v-for="r in memRows" :key="r.label" class="info-row">
              <span class="info-key">{{ r.label }}</span>
              <span class="info-val">{{ r.value }}</span>
            </div>
          </div>
        </div>
      </van-pull-refresh>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, watch } from 'vue'
import { useMetricsStore } from '@/stores/metrics'
import { systemApi } from '@/api'
import AppLayout from '@/components/AppLayout.vue'
import EChartsWrapper from '@/components/EChartsWrapper.vue'

const metrics = useMetricsStore()
const snap = computed(() => metrics.latest)
const refreshing = ref(false)

const TABS = [
  { key:'cpu', label:'CPU' }, { key:'mem', label:'内存' },
  { key:'disk', label:'磁盘' }, { key:'net', label:'网络' },
]
const RANGES = ['1h','6h','24h','7d']
const activeTab   = ref('cpu')
const activeRange = ref('1h')
const histData    = ref<{ts:number;val:number}[]>([])

const currentTab = computed(() => TABS.find(t => t.key === activeTab.value))
const currentPct = computed(() => {
  if (!snap.value) return 0
  return ({cpu:snap.value.cpu, mem:snap.value.mem_pct, disk:snap.value.disk_pct,
    net:Math.min((snap.value.net_in/1048576)*10,100)})[activeTab.value] ?? 0
})
const currentDisplay = computed(() => {
  const s = snap.value; if (!s) return '—'
  return ({
    cpu:  `${s.cpu.toFixed(1)} %`,
    mem:  `${(s.mem_used/1073741824).toFixed(1)} / ${(s.mem_total/1073741824).toFixed(1)} GB`,
    disk: `${s.disk_pct.toFixed(1)} %`,
    net:  `↓ ${(s.net_in/1024).toFixed(0)} KB/s`,
  })[activeTab.value] ?? '—'
})
const ringColor   = computed(() => currentPct.value > 85 ? 'var(--danger)' : currentPct.value > 70 ? 'var(--warning)' : 'var(--primary)')
const trackColor  = computed(() => currentPct.value > 85 ? 'var(--danger-soft)' : 'var(--primary-soft)')
const statusClass = computed(() => currentPct.value > 85 ? 'st--danger' : currentPct.value > 70 ? 'st--warn' : 'st--ok')
const statusText  = computed(() => currentPct.value > 85 ? '告警' : currentPct.value > 70 ? '警告' : '正常')

const memRows = computed(() => {
  const s = snap.value; if (!s) return []
  const gb = (b:number) => (b/1073741824).toFixed(2) + ' GB'
  return [
    {label:'已使用', value:gb(s.mem_used)},
    {label:'总量',   value:gb(s.mem_total)},
    {label:'可用',   value:gb(s.mem_total - s.mem_used)},
    {label:'Swap',   value:`${gb(s.swap_used)} / ${gb(s.swap_total)}`},
  ]
})

const chartOption = computed(() => ({
  grid: { top:8, right:8, bottom:28, left:46 },
  xAxis: { type:'time',
    axisLabel:{
      fontSize:11, color:'#9ca3af',
      // 1h 约 20 个点，自动限制最多 5 个刻度避免重叠
      hideOverlap: true,
      formatter:(val:number) => {
        const d = new Date(val)
        return d.getHours().toString().padStart(2,'0') + ':' + d.getMinutes().toString().padStart(2,'0')
      },
    },
    axisLine:{show:false}, splitLine:{show:false},
    maxInterval: activeRange.value === '1h' ? 3600000/4 : activeRange.value === '6h' ? 3600000*1.5 : 3600000*6,
  },
  yAxis: { type:'value', min:0, max:100,
    axisLabel:{fontSize:11,color:'#9ca3af',formatter:'{value}%'},
    splitLine:{lineStyle:{color:'rgba(66,170,245,0.08)'}} },
  series:[{ type:'line', smooth:true, symbol:'none',
    data: histData.value.map(d=>[d.ts,d.val]),
    lineStyle:{color:'var(--primary)',width:2.5},
    areaStyle:{color:{type:'linear',x:0,y:0,x2:0,y2:1,
      colorStops:[{offset:0,color:'rgba(66,170,245,0.28)'},{offset:1,color:'rgba(66,170,245,0)'}]}},
  }],
  tooltip:{trigger:'axis',formatter:(p:any)=>`${new Date(p[0].value[0]).toLocaleTimeString()}: ${p[0].value[1].toFixed(1)}%`},
}))

watch(() => metrics.latest, (s) => {
  if (!s) return
  const val = ({cpu:s.cpu,mem:s.mem_pct,disk:s.disk_pct,net:s.net_in/1024})[activeTab.value] ?? s.cpu
  histData.value.push({ts:s.ts,val})
  if (histData.value.length > 1200) histData.value.shift()
})

function coreColor(v:number) { return v>85?'var(--danger)':v>60?'var(--warning)':'var(--primary)' }
async function loadHistory() {
  try {
    const res:any = await systemApi.history(activeTab.value, activeRange.value)
    if (!res?.data) return
    histData.value = res.data.map((r:any)=>({
      ts:r.ts, val:({cpu:r.cpu,mem:r.mem_pct,disk:r.disk_pct,net:r.net_in/1024})[activeTab.value]??r.cpu,
    }))
  } catch {}
}
function switchTab(key:string) { activeTab.value = key; histData.value = []; loadHistory() }
function changeRange(r:string) { activeRange.value = r; loadHistory() }
async function onRefresh() { await metrics.fetchSnapshot(); await loadHistory(); refreshing.value = false }
onMounted(async () => { metrics.connectWS(); await metrics.fetchSnapshot(); await loadHistory() })
onUnmounted(() => metrics.disconnectWS())
</script>

<style scoped>
.page { height: 100%; display: flex; flex-direction: column; overflow: hidden; }
.top-bar {
  padding: 14px 18px 10px; background: rgba(255,255,255,0.92);
  border-bottom: 1px solid var(--border-subtle);
  display: flex; align-items: center; justify-content: space-between; flex-shrink: 0;
}
.top-bar__title { font-size: 20px; font-weight: 700; color: var(--text); }
.ws-dot { width: 9px; height: 9px; border-radius: 50%; }
.ws--open       { background: var(--success); animation: blink 2s infinite; }
.ws--connecting { background: var(--warning); }
.ws--closed     { background: var(--danger); }

.metric-tabs { display: flex; gap: 8px; padding: 10px 14px 8px; flex-shrink: 0; }
.m-tab {
  padding: 7px 14px; border-radius: 20px; font-size: 13px; font-weight: 600;
  border: 1px solid var(--border); background: var(--primary-soft); color: var(--sub-text); cursor: pointer;
  transition: all .15s;
}
.m-tab.active { background: var(--primary); color: #fff; border-color: var(--primary); }

.pull-area { flex: 1; overflow-y: auto; }
.content   { padding: 8px 14px 0; }

.card {
  background: var(--bg-white); border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle); box-shadow: 0 2px 12px var(--shadow);
  padding: 14px 16px; margin-bottom: 12px;
}
.card-title { font-size: 14px; font-weight: 700; color: var(--text); margin-bottom: 12px; }

/* 仪表卡 */
.gauge-card { display: flex; align-items: center; gap: 20px; }
.gauge-wrap { position: relative; display: flex; align-items: center; justify-content: center; flex-shrink: 0; }
.gauge-center-val { position: absolute; font-size: 17px; font-weight: 700; color: var(--text); }
.gauge-label     { font-size: 12px; color: var(--muted-text); margin-bottom: 5px; }
.gauge-main-val  { font-size: 22px; font-weight: 700; color: var(--text); margin-bottom: 8px; }
.gauge-status-badge {
  display: inline-flex; align-items: center; gap: 5px;
  font-size: 12px; font-weight: 700;
  padding: 4px 10px; border-radius: 8px;
}
.st--ok     { background: var(--success-soft); color: var(--success); }
.st--warn   { background: var(--warning-soft); color: var(--warning); }
.st--danger { background: var(--danger-soft);  color: var(--danger); }

/* 图表 */
.chart-head { display: flex; justify-content: space-between; align-items: center; margin-bottom: 10px; }
.chart-title { font-size: 14px; font-weight: 700; color: var(--text); }
.range-pills { display: flex; gap: 5px; }
.range-pill {
  font-size: 12px; padding: 4px 9px; border-radius: 8px;
  border: none; background: var(--primary-soft); color: var(--sub-text); cursor: pointer;
}
.range-pill.active { background: var(--primary); color: #fff; }

/* 核心 */
.core-row { display: flex; align-items: center; gap: 10px; margin-bottom: 10px; }
.core-lbl  { font-size: 13px; color: var(--muted-text); width: 48px; }
.core-track { flex: 1; height: 6px; border-radius: 3px; background: var(--primary-soft); }
.core-fill  { height: 100%; border-radius: 3px; transition: width .4s; }
.core-val   { font-size: 13px; color: var(--text); font-weight: 600; width: 38px; text-align: right; }

/* 信息行 */
.info-row { display: flex; justify-content: space-between; padding: 9px 0; border-bottom: 1px solid var(--border-subtle); }
.info-row:last-child { border-bottom: none; }
.info-key { font-size: 13px; color: var(--muted-text); }
.info-val { font-size: 13px; color: var(--text); font-family: monospace; font-weight: 600; }
</style>
