import { defineStore } from 'pinia'
import { ref } from 'vue'
import { appsApi } from '@/api'
import { showToast, showConfirmDialog } from 'vant'

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
    } catch (e: any) {
      showToast({ message: '获取微应用列表失败', type: 'fail' })
    } finally {
      loading.value = false
    }
  }

  async function fetchDetail(id: string): Promise<AppInfo | null> {
    try {
      const res: any = await appsApi.detail(id)
      return res?.data ?? null
    } catch {
      return null
    }
  }

  async function start(id: string) {
    try {
      await appsApi.start(id)
      showToast({ message: '启动成功', type: 'success' })
      await fetchList()
    } catch (e: any) {
      showToast({ message: e.response?.data?.message ?? '启动失败', type: 'fail' })
    }
  }

  async function stop(id: string, name: string) {
    try {
      await showConfirmDialog({
        title: '确认停止',
        message: `停止「${name}」将影响相关服务，确定继续？`,
        confirmButtonText: '停止',
        cancelButtonText: '取消',
      })
    } catch { return }

    try {
      await appsApi.stop(id)
      showToast({ message: '停止成功', type: 'success' })
      await fetchList()
    } catch (e: any) {
      showToast({ message: e.response?.data?.message ?? '停止失败', type: 'fail' })
    }
  }

  async function restart(id: string, name: string) {
    try {
      await showConfirmDialog({
        title: '确认重启',
        message: `重启「${name}」将短暂中断服务，确定继续？`,
        confirmButtonText: '重启',
        cancelButtonText: '取消',
      })
    } catch { return }

    try {
      await appsApi.restart(id)
      showToast({ message: '重启成功', type: 'success' })
      await fetchList()
    } catch (e: any) {
      showToast({ message: e.response?.data?.message ?? '重启失败', type: 'fail' })
    }
  }

  return { list, loading, fetchList, fetchDetail, start, stop, restart }
})
