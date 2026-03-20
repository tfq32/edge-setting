<template>
  <AppLayout>
    <div class="page">
      <div class="top-bar">
        <div class="top-bar__title">微应用</div>
        <div class="search-box">
          <span class="search-icon">🔍</span>
          <input v-model="keyword" class="search-input" placeholder="搜索应用名称…" />
        </div>
        <div class="filter-pills">
          <button v-for="f in filterOpts" :key="f.key"
            class="filter-pill" :class="{ active: activeFilter === f.key }"
            @click="activeFilter = f.key">
            {{ f.label }}&nbsp;<span class="pill-cnt">{{ f.count }}</span>
          </button>
        </div>
      </div>

      <van-pull-refresh v-model="refreshing" @refresh="onRefresh" class="pull-area">
        <div class="content">
          <div v-if="apps.loading" class="tip">加载中…</div>
          <div v-else-if="filtered.length === 0" class="tip">暂无匹配应用</div>
          <AppCard v-for="app in filtered" :key="app.id" :app="app" class="card-gap"
            @click="$router.push(`/apps/${app.id}`)" />
        </div>
      </van-pull-refresh>
    </div>
  </AppLayout>
</template>

<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useAppsStore } from '@/stores/apps'
import AppLayout from '@/components/AppLayout.vue'
import AppCard from '@/components/AppCard.vue'

const apps = useAppsStore()
const keyword = ref('')
const activeFilter = ref('all')
const refreshing = ref(false)

const filterOpts = computed(() => [
  { key:'all',     label:'全部',   count: apps.list.length },
  { key:'running', label:'运行中', count: apps.list.filter(a=>a.status==='running').length },
  { key:'stopped', label:'已停止', count: apps.list.filter(a=>a.status==='stopped').length },
])

const filtered = computed(() =>
  apps.list
    .filter(a => activeFilter.value === 'all' || a.status === activeFilter.value)
    .filter(a => !keyword.value || a.id.includes(keyword.value) || a.name?.includes(keyword.value))
)

async function onRefresh() { await apps.fetchList(); refreshing.value = false }
onMounted(() => apps.fetchList())
</script>

<style scoped>
.page { height: 100%; display: flex; flex-direction: column; overflow: hidden; }
.top-bar {
  padding: 14px 18px 10px; background: rgba(255,255,255,0.92);
  border-bottom: 1px solid var(--border-subtle); flex-shrink: 0;
}
.top-bar__title { font-size: 20px; font-weight: 700; color: var(--text); margin-bottom: 10px; }

.search-box {
  display: flex; align-items: center; gap: 8px;
  background: var(--primary-soft); border: 1px solid var(--border);
  border-radius: 12px; padding: 8px 13px; margin-bottom: 10px;
}
.search-icon  { font-size: 15px; }
.search-input {
  flex: 1; border: none; background: transparent;
  font-size: 14px; color: var(--text); outline: none;
}
.search-input::placeholder { color: var(--muted-text); }

.filter-pills { display: flex; gap: 8px; }
.filter-pill {
  padding: 6px 12px; border-radius: 20px; font-size: 13px; font-weight: 600;
  border: 1px solid var(--border); background: var(--primary-soft); color: var(--sub-text); cursor: pointer;
  display: flex; align-items: center; gap: 3px; transition: all .15s;
}
.filter-pill.active { background: var(--primary); color: #fff; border-color: var(--primary); }
.pill-cnt {
  font-size: 11px; background: rgba(255,255,255,0.28); border-radius: 8px;
  padding: 0 5px; line-height: 1.6;
}

.pull-area { flex: 1; overflow-y: auto; }
.content { padding: 10px 14px 0; }
.card-gap { margin-bottom: 10px; }
.tip { font-size: 14px; color: var(--muted-text); text-align: center; padding: 40px 0; }
</style>
