<template>
  <AppLayout>
    <div class="page">
      <div class="top-bar">
        <span class="top-bar__title">系统信息</span>
      </div>

      <van-pull-refresh v-model="refreshing" @refresh="onRefresh" class="pull-area">
        <div class="content">
          <div v-if="!sysInfo" class="tip">加载中…</div>
          <template v-else>
            <!-- 节点信息 -->
            <div class="card">
              <div class="group-head">
                <span class="group-icon">🖥</span>
                <span class="group-title">节点信息</span>
              </div>
              <!-- 短文本：左右对齐 -->
              <div class="info-row">
                <span class="info-key">主机名</span>
                <span class="info-val">{{ sysInfo.hostname }}</span>
              </div>
              <div class="info-row">
                <span class="info-key">架构</span>
                <span class="info-val">{{ sysInfo.arch }}</span>
              </div>
              <!-- 长文本：标签在上，值在下 -->
              <div class="info-block">
                <div class="info-block__key">系统运行时长</div>
                <div class="info-block__val">{{ formatSec(sysInfo.uptime_system ?? 0) }}</div>
              </div>
            </div>

            <!-- 系统信息 -->
            <div class="card">
              <div class="group-head">
                <span class="group-icon">⚙</span>
                <span class="group-title">系统信息</span>
              </div>
              <div class="info-block">
                <div class="info-block__key">操作系统</div>
                <div class="info-block__val">{{ sysInfo.platform }} {{ sysInfo.platform_version }}</div>
              </div>
              <div class="info-block last">
                <div class="info-block__key">内核版本</div>
                <div class="info-block__val mono">{{ sysInfo.kernel_version }}</div>
              </div>
            </div>
          </template>
        </div>
      </van-pull-refresh>

    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { systemApi } from '@/api'
import AppLayout from '@/components/AppLayout.vue'

const sysInfo = ref<any>(null)
const refreshing = ref(false)

function formatSec(s: number) {
  const d = Math.floor(s / 86400), h = Math.floor((s % 86400) / 3600), m = Math.floor((s % 3600) / 60)
  return d > 0 ? `${d}天 ${h}时 ${m}分` : `${h}时 ${m}分`
}

async function fetchInfo() {
  try { const res:any = await systemApi.info(); sysInfo.value = res?.data ?? null } catch {}
}
async function onRefresh() { await fetchInfo(); refreshing.value = false }
onMounted(fetchInfo)
</script>

<style scoped>
.page { height: 100%; display: flex; flex-direction: column; overflow: hidden; }
.top-bar {
  padding: 14px 18px 10px; background: rgba(255,255,255,0.92);
  border-bottom: 1px solid var(--border-subtle); flex-shrink: 0;
}
.top-bar__title { font-size: 20px; font-weight: 700; color: var(--text); }

.pull-area { flex: 1; overflow-y: auto; }
.content   { padding: 14px 14px 0; }

/* 卡片 */
.card {
  background: var(--bg-white); border-radius: var(--radius-md);
  border: 1px solid var(--border-subtle); box-shadow: 0 2px 12px var(--shadow);
  overflow: hidden; margin-bottom: 14px;
}

/* 分组标题行 */
.group-head {
  display: flex; align-items: center; gap: 8px;
  background: var(--primary-soft); border-bottom: 1px solid var(--border);
  padding: 9px 16px;
}
.group-icon  { font-size: 15px; }
.group-title { font-size: 13px; font-weight: 700; color: var(--primary); letter-spacing: .3px; }

/* 短文本行：左右对齐 */
.info-row {
  display: flex; justify-content: space-between; align-items: center;
  padding: 13px 16px; border-bottom: 1px solid var(--border-subtle); gap: 12px;
}
.info-key {
  font-size: 14px; color: var(--muted-text); flex-shrink: 0;
}
.info-val {
  font-size: 14px; color: var(--text); font-family: monospace; font-weight: 600;
  overflow: hidden; text-overflow: ellipsis; white-space: nowrap; text-align: right;
}

/* 长文本块：标签在上，值在下 */
.info-block {
  padding: 12px 16px; border-bottom: 1px solid var(--border-subtle);
}
.info-block.last { border-bottom: none; }
.info-block__key {
  font-size: 12px; color: var(--muted-text); margin-bottom: 5px;
}
.info-block__val {
  font-size: 14px; color: var(--text); font-weight: 600; line-height: 1.5;
  word-break: break-all;
}
.info-block__val.mono { font-family: monospace; }

.tip { font-size: 14px; color: var(--muted-text); text-align: center; padding: 40px 0; }
</style>
