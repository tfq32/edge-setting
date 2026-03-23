<template>
  <div ref="el" class="chart-wrap" />
</template>

<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import * as echarts from 'echarts'

const props = defineProps<{
  option: echarts.EChartsOption
  height?: string
}>()

const el = ref<HTMLElement>()
let chart: echarts.ECharts | null = null
let ro: ResizeObserver | null = null

onMounted(() => {
  if (!el.value) return
  el.value.style.height = props.height ?? '180px'
  chart = echarts.init(el.value, null, { renderer: 'svg' })
  chart.setOption(props.option)

  ro = new ResizeObserver(() => chart?.resize())
  ro.observe(el.value)
})

onUnmounted(() => {
  ro?.disconnect()
  chart?.dispose()
})

watch(() => props.option, (opt) => {
  chart?.setOption(opt, { notMerge: false })
}, { deep: true })
</script>

<style scoped>
.chart-wrap { width: 100%; }
</style>
