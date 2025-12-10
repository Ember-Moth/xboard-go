<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

interface TrafficStats {
  total_upload: number
  total_download: number
  total_traffic: number
  active_users: number
  over_traffic_users: number
  upload_gb: number
  download_gb: number
  total_gb: number
}

interface WarningUser {
  id: number
  email: string
  upload: number
  download: number
  total_used: number
  transfer_enable: number
  usage_percent: number
  is_over_limit: boolean
  total_gb: number
  limit_gb: number
}

const stats = ref<TrafficStats>({
  total_upload: 0,
  total_download: 0,
  total_traffic: 0,
  active_users: 0,
  over_traffic_users: 0,
  upload_gb: 0,
  download_gb: 0,
  total_gb: 0,
})

const warningUsers = ref<WarningUser[]>([])
const loading = ref(false)
const threshold = ref(80)
const showResetModal = ref(false)

const fetchStats = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/traffic/stats')
    stats.value = res.data.data
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchWarnings = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/traffic/warnings', {
      params: { threshold: threshold.value }
    })
    warningUsers.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const resetUserTraffic = async (userId: number) => {
  if (!confirm('确定要重置该用户的流量吗？')) return
  try {
    await api.post(`/api/v2/admin/traffic/reset/${userId}`)
    fetchWarnings()
    fetchStats()
  } catch (e: any) {
    alert(e.response?.data?.error || '重置失败')
  }
}

const resetAllTraffic = async () => {
  if (!confirm('确定要重置所有用户的流量吗？此操作不可恢复！')) return
  try {
    const res = await api.post('/api/v2/admin/traffic/reset-all')
    alert(`已重置 ${res.data.count} 个用户的流量`)
    showResetModal.value = false
    fetchWarnings()
    fetchStats()
  } catch (e: any) {
    alert(e.response?.data?.error || '重置失败')
  }
}

const sendWarning = async (userId: number) => {
  try {
    await api.post(`/api/v2/admin/traffic/warning/${userId}`)
    alert('预警通知已发送')
  } catch (e: any) {
    alert(e.response?.data?.error || '发送失败')
  }
}

const batchSendWarnings = async () => {
  if (!confirm(`确定要向所有流量使用超过 ${threshold.value}% 的用户发送预警通知吗？`)) return
  try {
    const res = await api.post(`/api/v2/admin/traffic/warnings/send?threshold=${threshold.value}`)
    alert(`已发送 ${res.data.success}/${res.data.total} 条通知`)
  } catch (e: any) {
    alert(e.response?.data?.error || '发送失败')
  }
}

const autobanUsers = async () => {
  if (!confirm('确定要自动封禁所有超流量用户吗？')) return
  try {
    const res = await api.post('/api/v2/admin/traffic/autoban')
    alert(`已封禁 ${res.data.count} 个用户`)
    fetchWarnings()
    fetchStats()
  } catch (e: any) {
    alert(e.response?.data?.error || '操作失败')
  }
}

const getProgressColor = (percent: number) => {
  if (percent >= 100) return 'bg-red-500'
  if (percent >= 90) return 'bg-orange-500'
  if (percent >= 80) return 'bg-yellow-500'
  return 'bg-green-500'
}

onMounted(() => {
  fetchStats()
  fetchWarnings()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">流量管理</h1>
        <p class="text-gray-500 mt-1">监控和管理用户流量使用情况</p>
      </div>
      <div class="flex gap-3">
        <button @click="fetchStats(); fetchWarnings()" class="px-4 py-2 text-gray-600 hover:bg-gray-100 rounded-xl transition-colors">
          <svg class="w-4 h-4 inline mr-1" :class="loading ? 'animate-spin' : ''" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15"/>
          </svg>
          刷新
        </button>
        <button @click="showResetModal = true" class="px-4 py-2 bg-red-500 text-white rounded-xl hover:bg-red-600 transition-colors">
          重置所有流量
        </button>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-gradient-to-br from-blue-500 to-blue-600 rounded-2xl p-6 text-white shadow-lg">
        <div class="flex items-center justify-between mb-2">
          <span class="text-blue-100 text-sm">总流量</span>
          <svg class="w-8 h-8 text-blue-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/>
          </svg>
        </div>
        <p class="text-3xl font-bold">{{ stats.total_gb.toFixed(2) }} GB</p>
        <p class="text-blue-100 text-sm mt-1">↑ {{ stats.upload_gb.toFixed(2) }} GB / ↓ {{ stats.download_gb.toFixed(2) }} GB</p>
      </div>

      <div class="bg-gradient-to-br from-green-500 to-green-600 rounded-2xl p-6 text-white shadow-lg">
        <div class="flex items-center justify-between mb-2">
          <span class="text-green-100 text-sm">活跃用户</span>
          <svg class="w-8 h-8 text-green-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 20h5v-2a3 3 0 00-5.356-1.857M17 20H7m10 0v-2c0-.656-.126-1.283-.356-1.857M7 20H2v-2a3 3 0 015.356-1.857M7 20v-2c0-.656.126-1.283.356-1.857m0 0a5.002 5.002 0 019.288 0M15 7a3 3 0 11-6 0 3 3 0 016 0zm6 3a2 2 0 11-4 0 2 2 0 014 0zM7 10a2 2 0 11-4 0 2 2 0 014 0z"/>
          </svg>
        </div>
        <p class="text-3xl font-bold">{{ stats.active_users }}</p>
        <p class="text-green-100 text-sm mt-1">正在使用服务</p>
      </div>

      <div class="bg-gradient-to-br from-orange-500 to-orange-600 rounded-2xl p-6 text-white shadow-lg">
        <div class="flex items-center justify-between mb-2">
          <span class="text-orange-100 text-sm">超流量用户</span>
          <svg class="w-8 h-8 text-orange-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
          </svg>
        </div>
        <p class="text-3xl font-bold">{{ stats.over_traffic_users }}</p>
        <p class="text-orange-100 text-sm mt-1">需要处理</p>
      </div>

      <div class="bg-gradient-to-br from-purple-500 to-purple-600 rounded-2xl p-6 text-white shadow-lg">
        <div class="flex items-center justify-between mb-2">
          <span class="text-purple-100 text-sm">预警阈值</span>
          <svg class="w-8 h-8 text-purple-200" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M15 17h5l-1.405-1.405A2.032 2.032 0 0118 14.158V11a6.002 6.002 0 00-4-5.659V5a2 2 0 10-4 0v.341C7.67 6.165 6 8.388 6 11v3.159c0 .538-.214 1.055-.595 1.436L4 17h5m6 0v1a3 3 0 11-6 0v-1m6 0H9"/>
          </svg>
        </div>
        <div class="flex items-center gap-2">
          <input v-model.number="threshold" type="number" min="50" max="100" step="5" class="w-20 px-2 py-1 bg-white/20 border border-white/30 rounded-lg text-white text-xl font-bold" />
          <span class="text-2xl font-bold">%</span>
        </div>
        <button @click="fetchWarnings" class="text-purple-100 text-sm mt-1 hover:text-white">点击更新</button>
      </div>
    </div>

    <!-- Actions -->
    <div class="bg-white rounded-2xl p-6 shadow-sm">
      <h2 class="text-lg font-semibold mb-4">批量操作</h2>
      <div class="flex flex-wrap gap-3">
        <button @click="batchSendWarnings" class="px-4 py-2 bg-yellow-500 text-white rounded-xl hover:bg-yellow-600 transition-colors flex items-center gap-2">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 8l7.89 5.26a2 2 0 002.22 0L21 8M5 19h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v10a2 2 0 002 2z"/>
          </svg>
          批量发送预警
        </button>
        <button @click="autobanUsers" class="px-4 py-2 bg-red-500 text-white rounded-xl hover:bg-red-600 transition-colors flex items-center gap-2">
          <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M18.364 18.364A9 9 0 005.636 5.636m12.728 12.728A9 9 0 015.636 5.636m12.728 12.728L5.636 5.636"/>
          </svg>
          自动封禁超流量用户
        </button>
      </div>
      <p class="text-sm text-gray-500 mt-3">
        ⚠️ 注意：流量统计采用平均分配算法，单用户流量可能不够精确。详见文档说明。
      </p>
    </div>

    <!-- Warning Users Table -->
    <div class="bg-white rounded-2xl shadow-sm overflow-hidden">
      <div class="px-6 py-4 border-b border-gray-200 flex items-center justify-between">
        <h2 class="text-lg font-semibold">流量预警用户 (≥{{ threshold }}%)</h2>
        <span class="text-sm text-gray-500">共 {{ warningUsers.length }} 个用户</span>
      </div>

      <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

      <div v-else-if="warningUsers.length === 0" class="text-center py-12 text-gray-500">
        <svg class="w-16 h-16 mx-auto text-gray-300 mb-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z"/>
        </svg>
        <p>暂无预警用户</p>
      </div>

      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50 border-b border-gray-200">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">已用流量</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">总流量</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">使用率</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="user in warningUsers" :key="user.id" class="hover:bg-gray-50">
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-gradient-to-br from-indigo-400 to-indigo-600 flex items-center justify-center text-white text-sm font-medium">
                    {{ user.email.charAt(0).toUpperCase() }}
                  </div>
                  <div>
                    <div class="font-medium text-gray-900">{{ user.email }}</div>
                    <div class="text-xs text-gray-500">ID: {{ user.id }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 text-sm">{{ user.total_gb.toFixed(2) }} GB</td>
              <td class="px-6 py-4 text-sm">{{ user.limit_gb.toFixed(2) }} GB</td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-2">
                  <div class="flex-1 h-2 bg-gray-200 rounded-full overflow-hidden">
                    <div :class="['h-full transition-all', getProgressColor(user.usage_percent)]" :style="{ width: Math.min(user.usage_percent, 100) + '%' }"></div>
                  </div>
                  <span class="text-sm font-medium" :class="user.is_over_limit ? 'text-red-600' : 'text-yellow-600'">
                    {{ user.usage_percent.toFixed(1) }}%
                  </span>
                </div>
              </td>
              <td class="px-6 py-4">
                <span v-if="user.is_over_limit" class="px-2 py-1 bg-red-100 text-red-600 rounded-full text-xs">超流量</span>
                <span v-else class="px-2 py-1 bg-yellow-100 text-yellow-600 rounded-full text-xs">预警</span>
              </td>
              <td class="px-6 py-4 text-right space-x-2">
                <button @click="sendWarning(user.id)" class="text-yellow-600 hover:text-yellow-700 text-sm">发送预警</button>
                <button @click="resetUserTraffic(user.id)" class="text-indigo-600 hover:text-indigo-700 text-sm">重置流量</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- Reset All Modal -->
    <Teleport to="body">
      <div v-if="showResetModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showResetModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-md p-6">
          <div class="text-center">
            <div class="w-16 h-16 bg-red-100 rounded-full flex items-center justify-center mx-auto mb-4">
              <svg class="w-8 h-8 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
              </svg>
            </div>
            <h3 class="text-lg font-bold mb-2">重置所有用户流量</h3>
            <p class="text-gray-600 mb-6">此操作将重置所有用户的上传和下载流量为 0，此操作不可恢复！</p>
            <div class="flex gap-3">
              <button @click="showResetModal = false" class="flex-1 px-4 py-2.5 border border-gray-200 text-gray-600 rounded-xl hover:bg-gray-50">取消</button>
              <button @click="resetAllTraffic" class="flex-1 px-4 py-2.5 bg-red-500 text-white rounded-xl hover:bg-red-600">确认重置</button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
