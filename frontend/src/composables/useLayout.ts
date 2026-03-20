import { ref, onMounted, onUnmounted } from 'vue'

/** 宽屏阈值 768px — 宽屏显示侧边栏，窄屏显示底部 TabBar */
const WIDE = 768

export function useLayout() {
  const isWide = ref(window.innerWidth >= WIDE)

  function onResize() { isWide.value = window.innerWidth >= WIDE }

  onMounted(() => window.addEventListener('resize', onResize))
  onUnmounted(() => window.removeEventListener('resize', onResize))

  return { isWide }
}
