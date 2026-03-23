import { defineStore } from 'pinia'
import { ref } from 'vue'
import { appsApi } from '@/api'
import { showToast } from 'vant'

export type AppStatus = 'running' | 'stopped' | 'error'

export interface AppInfo {
  id:         string
  name:       string
  type:       'system' | 'user'
  has_ui:     boolean
  status:     AppStatus
  pid:        number
  cpu:        number
  mem:        number
  ports:      number[]
  version:    string
  start_time: number
  work_dir:   string
}

export const useAppsStore = defineStore('apps', () => {
  const list    = ref<AppInfo[]>([])
  const loading = ref(false)

  async function fetchList() {
    loading.value = true
    try {
      const res: any = await appsApi.list()
      if (res?.data) list.value = res.data
    } catch {
      showToast({ message: '获取微应用列表失败', type: 'fail' })
    } finally {
      loading.value = false
    }
  }

  return { list, loading, fetchList }
})
