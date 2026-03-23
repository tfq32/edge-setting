<template>
  <div class="metric-card" :class="statusClass" @click="$emit('click')">
    <div class="mc-header">
      <!-- 左列：标签 + 告警 badge（垂直堆叠，不遮环形图） -->
      <div class="mc-left">
        <span class="mc-label">{{ label }}</span>
        <span v-if="status !== 'normal'" class="mc-badge" :class="`mc-badge--${status}`">
          {{ status === 'warning' ? '警告' : '告警' }}
        </span>
      </div>
      <!-- 右列：环形图 -->
      <svg viewBox="0 0 36 36" width="44" height="44" style="flex-shrink:0">
        <circle cx="18" cy="18" r="14" fill="none" :stroke="trackColor" stroke-width="4" />
        <circle cx="18" cy="18" r="14" fill="none" :stroke="ringColor" stroke-width="4"
          :stroke-dasharray="`${dashLen} ${circumference}`" stroke-linecap="round"
          transform="rotate(-90 18 18)" />
      </svg>
    </div>

    <div class="mc-value">
      {{ displayValue }}<span class="mc-unit">{{ unit }}</span>
    </div>

    <div class="mc-bar">
      <div class="mc-bar-inner" :style="{ width: pct + '%', background: ringColor }" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

const props = withDefaults(defineProps<{
  label:   string
  value:   number
  pct:     number
  unit?:   string
  warn?:   number
  danger?: number
}>(), { unit: '%', warn: 80, danger: 90 })

defineEmits<{ click: [] }>()

const status = computed(() => {
  if (props.pct >= props.danger) return 'danger'
  if (props.pct >= props.warn)   return 'warning'
  return 'normal'
})
const statusClass = computed(() => ({
  'mc--warning': status.value === 'warning',
  'mc--danger':  status.value === 'danger',
}))
const circumference = 2 * Math.PI * 14
const dashLen   = computed(() => (props.pct / 100) * circumference)
const ringColor = computed(() => {
  if (status.value === 'danger')  return 'var(--danger)'
  if (status.value === 'warning') return 'var(--warning)'
  return 'var(--primary)'
})
const trackColor = computed(() => {
  if (status.value === 'danger')  return 'var(--danger-soft)'
  if (status.value === 'warning') return 'var(--warning-soft)'
  return 'var(--primary-soft)'
})
const displayValue = computed(() =>
  props.unit === '%' ? props.pct.toFixed(1) : props.value.toFixed(0)
)
</script>

<style scoped>
.metric-card {
  background: var(--bg-white);
  border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle);
  box-shadow: 0 2px 10px var(--shadow);
  padding: 14px;
  cursor: pointer;
  transition: border-color .2s, box-shadow .2s;
}
.metric-card:active { opacity: .85; }
.mc--warning { border-color: var(--warning-border); }
.mc--danger  { border-color: var(--danger-border); box-shadow: 0 2px 14px var(--danger-soft); }

/* 头部：左列（标签+badge）+ 右列（环形图） */
.mc-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 4px;
}
.mc-left {
  display: flex;
  flex-direction: column;
  align-items: flex-start;
  gap: 5px;
  /* 确保不会被环形图挤压 */
  min-width: 0;
  padding-right: 6px;
}
.mc-label {
  font-size: 13px;
  font-weight: 600;
  color: var(--sub-text);
  line-height: 1.3;
}
/* badge 紧贴在 label 下方，不影响右侧 */
.mc-badge {
  font-size: 11px;
  font-weight: 700;
  padding: 2px 8px;
  border-radius: 5px;
  white-space: nowrap;
}
.mc-badge--warning { background: var(--warning-soft); color: var(--warning); }
.mc-badge--danger  { background: var(--danger-soft);  color: var(--danger); }

/* 数值 */
.mc-value {
  font-size: 28px;
  font-weight: 700;
  color: var(--text);
  line-height: 1;
  margin-bottom: 10px;
}
.mc-unit { font-size: 13px; font-weight: 400; color: var(--sub-text); margin-left: 2px; }

/* 进度条 */
.mc-bar { height: 5px; border-radius: 3px; background: var(--primary-soft); }
.mc-bar-inner { height: 100%; border-radius: 3px; transition: width .5s; }
</style>
