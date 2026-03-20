<template>
  <div class="tab-bar">
    <router-link v-for="tab in tabs" :key="tab.to" :to="tab.to"
      class="tab-item" :class="{ 'tab-item--active': isActive(tab.to) }">
      <span class="tab-icon">{{ tab.icon }}</span>
      <span class="tab-label">{{ tab.label }}</span>
    </router-link>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
const route = useRoute()
const tabs = [
  { to:'/',        label:'首页',   icon:'◈' },
  { to:'/monitor', label:'监控',   icon:'◉' },
  { to:'/apps',    label:'应用',   icon:'⊞' },
  { to:'/about',   label:'系统',   icon:'◎' },
]
function isActive(path:string) {
  return path === '/' ? route.path === '/' : route.path.startsWith(path)
}
</script>

<style scoped>
.tab-bar {
  display: flex;
  border-top: 1px solid var(--border-subtle);
  background: rgba(255,255,255,0.96);
  padding: 8px 0 env(safe-area-inset-bottom, 10px);
  flex-shrink: 0;
}
.tab-item {
  flex: 1; display: flex; flex-direction: column; align-items: center; gap: 3px;
  text-decoration: none; padding: 4px 0;
}
.tab-icon  { font-size: 22px; color: var(--muted-text); line-height: 1.2; transition: color .15s; }
.tab-label { font-size: 12px; color: var(--muted-text); font-weight: 400; transition: color .15s; }
.tab-item--active .tab-icon,
.tab-item--active .tab-label { color: var(--primary); }
.tab-item--active .tab-label { font-weight: 700; }
</style>
