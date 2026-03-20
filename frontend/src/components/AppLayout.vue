<!-- 响应式布局容器：宽屏=侧边栏+内容；窄屏=内容+底部TabBar -->
<template>
  <div class="app-layout">
    <!-- 宽屏侧边栏 -->
    <aside v-if="isWide" class="sidebar">
      <div class="sidebar__brand">
        <div class="brand-icon">⚙</div>
      </div>
      <nav class="sidebar__nav">
        <router-link
          v-for="item in NAV"
          :key="item.to"
          :to="item.to"
          class="nav-item"
          :class="{ 'nav-item--active': isActive(item.to) }"
          :title="item.label"
        >
          <span class="nav-icon">{{ item.icon }}</span>
          <span class="nav-label">{{ item.label }}</span>
        </router-link>
      </nav>
    </aside>

    <!-- 主内容区 -->
    <div class="layout-main">
      <slot />
      <!-- 窄屏底部 TabBar -->
      <TabBar v-if="!isWide" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { useRoute } from 'vue-router'
import { useLayout } from '@/composables/useLayout'
import TabBar from '@/components/TabBar.vue'

const route = useRoute()
const { isWide } = useLayout()

const NAV = [
  { to: '/',        label: '首页',   icon: '◈' },
  { to: '/monitor', label: '监控',   icon: '◉' },
  { to: '/apps',    label: '应用',   icon: '⊞' },
  { to: '/about',   label: '系统',   icon: '◎' },
]

function isActive(path: string) {
  return path === '/' ? route.path === '/' : route.path.startsWith(path)
}
</script>

<style scoped>
.app-layout {
  height: 100%;
  display: flex;
  overflow: hidden;
}

/* ── 侧边栏（宽屏） ─────────────────────────────────────── */
.sidebar {
  width: 68px;
  flex-shrink: 0;
  background: rgba(255,255,255,0.97);
  border-right: 1px solid var(--border-subtle);
  display: flex;
  flex-direction: column;
  padding: 14px 0;
  z-index: 10;
}
.sidebar__brand {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 4px 0 18px;
}
.brand-icon {
  width: 36px;
  height: 36px;
  border-radius: 10px;
  background: var(--primary);
  display: flex;
  align-items: center;
  justify-content: center;
  font-size: 18px;
  color: #fff;
  box-shadow: 0 3px 10px rgba(66,170,245,0.38);
}

.sidebar__nav { display: flex; flex-direction: column; }
.nav-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  padding: 10px 0;
  text-decoration: none;
  position: relative;
  border-right: 3px solid transparent;
  transition: background .15s;
}
.nav-icon  { font-size: 20px; color: var(--muted-text); transition: color .15s; line-height: 1.2; }
.nav-label { font-size: 10px; color: var(--muted-text); font-weight: 400; transition: color .15s; }
.nav-item--active {
  background: var(--primary-soft);
  border-right-color: var(--primary);
}
.nav-item--active .nav-icon,
.nav-item--active .nav-label { color: var(--primary); }
.nav-item--active .nav-label { font-weight: 700; }

/* ── 主内容 ──────────────────────────────────────────────── */
.layout-main {
  flex: 1;
  display: flex;
  flex-direction: column;
  overflow: hidden;
  min-width: 0;
}
</style>
