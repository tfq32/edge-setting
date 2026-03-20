<template>
  <div class="app-card" :class="`app-card--${app.status}`" @click="$emit('click')">
    <div class="ac-row">
      <div class="ac-dot" :class="`ac-dot--${app.status}`" />
      <span class="ac-name">{{ app.id }}</span>
      <span class="ac-tag ac-tag--type" :class="app.type === 'system' ? 'ac-tag--sys' : 'ac-tag--usr'">
        {{ app.type === 'system' ? '系统' : '用户' }}
      </span>
      <span class="ac-tag" :class="`ac-tag--${app.status}`">
        {{ statusLabel[app.status] }}
      </span>
    </div>
    <div class="ac-metrics">
      <span>CPU {{ app.cpu.toFixed(1) }}%</span>
      <span>内存 {{ formatMem(app.mem) }}</span>
      <span class="ac-ver">{{ app.version }}</span>
    </div>
    <div v-if="app.status === 'running'" class="ac-bar-row">
      <div class="ac-bar">
        <div class="ac-bar-inner" :style="{ width: Math.min(app.cpu * 5, 100) + '%', background: cpuColor }" />
      </div>
      <span class="ac-bar-val">{{ app.cpu.toFixed(1) }}%</span>
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'
import type { AppInfo } from '@/stores/apps'

const props = defineProps<{ app: AppInfo }>()
defineEmits<{ click:[] }>()

const statusLabel: Record<string,string> = { running:'运行中', stopped:'已停止', error:'异常' }
function formatMem(b:number) { return b ? (b/1048576).toFixed(0)+' MB' : '—' }
const cpuColor = computed(() => {
  const v = props.app.cpu
  return v > 80 ? 'var(--danger)' : v > 50 ? 'var(--warning)' : 'var(--primary)'
})
</script>

<style scoped>
.app-card {
  background: var(--bg-white); border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle); border-left: 3px solid var(--muted-text);
  box-shadow: 0 1px 6px var(--shadow); padding: 12px 14px; cursor: pointer;
}
.app-card:active { opacity: .8; }
.app-card--running { border-left-color: var(--success); }
.app-card--stopped { border-left-color: var(--muted-text); opacity: .72; }

.ac-dot { width: 8px; height: 8px; border-radius: 50%; flex-shrink: 0; }
.ac-dot--running { background: var(--success); animation: blink 2.4s infinite; }
.ac-dot--stopped { background: var(--muted-text); }

.ac-row { display: flex; align-items: center; gap: 8px; margin-bottom: 7px; }
.ac-name {
  font-size: 14px; font-weight: 700; color: var(--text);
  font-family: monospace; flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
}
.ac-tag { font-size: 11px; padding: 2px 6px; border-radius: 5px; flex-shrink: 0; }
.ac-tag--sys { background: rgba(139,92,246,.08); color: var(--system); }
.ac-tag--usr { background: var(--primary-soft); color: var(--primary-light); }
.ac-tag--running { background: var(--success-soft); color: var(--success); }
.ac-tag--stopped { background: var(--danger-soft);  color: var(--danger); }

.ac-metrics { display: flex; gap: 12px; font-size: 13px; color: var(--muted-text); }
.ac-ver     { margin-left: auto; font-family: monospace; }

.ac-bar-row { display: flex; align-items: center; gap: 8px; margin-top: 8px; }
.ac-bar     { flex: 1; height: 4px; border-radius: 2px; background: var(--primary-soft); }
.ac-bar-inner { height: 100%; border-radius: 2px; transition: width .5s; }
.ac-bar-val { font-size: 12px; color: var(--muted-text); width: 36px; text-align: right; }
</style>
