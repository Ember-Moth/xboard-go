<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'

const loading = ref(false)
const overview = ref<any>({})
const serverTraffic = ref<any[]>([])
const dailyStats = ref<any[]>([])

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const fetchData = async () => {
  loading.value = true
  try {
    const [overviewRes, serverRes, dailyRes] = await Promise.all([
      api.get('/api/v2/admin/traffic/overview'),
      api.get('/api/v2/admin/traffic/servers'),
      api.get('/api/v2/admin/traffic/daily?days=30')
    ])
    overview.value = overviewRes.data.data || {}
    serverTraffic.value = serverRes.data.data || []
    dailyStats.value = dailyRes.data.data || []
  } catch (e) {} finally { loading.value = false }
}

// 计算饼图数据
const pieData = computed(() => {
  const total = overview.value.total_traffic || 0
  if (total === 0) return []
  return [
    { name: '上传', value: overview.value.total_upload || 0, percent: overview.value.upload_percent || 0, color: '#6366f1' },
    { name: '下载', value: overview.value.total_download || 0, percent: overview.value.download_percent || 0, color: '#22c55e' }
  ]
})

// 计算节点流量占比
const serverPieData = computed(() => {
  const total = serverTraffic.value.reduce((sum, s) => sum + (s.total || 0), 0)
  if (total === 0) return []
  const colors = ['#6366f1', '#22c55e', '#f59e0b', '#ef4444', '#8b5cf6', '#06b6d4', '#ec4899', '#84cc16']
  return serverTraffic.value.map((s, i) => ({
    name: s.server_name,
    value: s.total,
    percent: (s.total / total * 100).toFixed(1),
    color: colors[i % colors.length]
  })).sort((a, b) => b.value - a.value).slice(0, 8)
})

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">流量统计</h1>
        <p class="text-gray-500 text-sm mt-1">查看系统流量使用情况</p>
      </div>
      <button @click="fetchData" :disabled="loading" class="px-4 py-2 text-sm text-gray-600 hover:bg-gray-100 rounded-xl transition-colors">
        {{ loading ? '加载中...' : '刷新' }}
      </button>
    </div>

    <!-- Overview Cards -->
    <div class="grid grid-cols-2 lg:grid-cols-4 gap-4">
      <div class="bg-white rounded-2xl p-5 shadow-sm">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-xl bg-indigo-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"/>
            </svg>
          </div>
          <span class="text-sm text-gray-500">总流量</span>
        </div>
        <p class="text-2xl font-bold text-gray-900">{{ formatBytes(overview.total_traffic || 0) }}</p>
      </div>
      <div class="bg-white rounded-2xl p-5 shadow-sm">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-xl bg-blue-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 11l5-5m0 0l5 5m-5-5v12"/>
            </svg>
          </div>
          <span class="text-sm text-gray-500">上传</span>
        </div>
        <p class="text-2xl font-bold text-blue-600">{{ formatBytes(overview.total_upload || 0) }}</p>
      </div>
      <div class="bg-white rounded-2xl p-5 shadow-sm">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-xl bg-green-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 13l-5 5m0 0l-5-5m5 5V6"/>
            </svg>
          </div>
          <span class="text-sm text-gray-500">下载</span>
        </div>
        <p class="text-2xl font-bold text-green-600">{{ formatBytes(overview.total_download || 0) }}</p>
      </div>
      <div class="bg-white rounded-2xl p-5 shadow-sm">
        <div class="flex items-center gap-3 mb-3">
          <div class="w-10 h-10 rounded-xl bg-purple-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4.354a4 4 0 110 5.292M15 21H3v-1a6 6 0 0112 0v1zm0 0h6v-1a6 6 0 00-9-5.197M13 7a4 4 0 11-8 0 4 4 0 018 0z"/>
            </svg>
          </div>
          <span class="text-sm text-gray-500">活跃用户</span>
        </div>
        <p class="text-2xl font-bold text-purple-600">{{ overview.active_users || 0 }}</p>
      </div>
    </div>

    <!-- Charts Row -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <!-- Upload/Download Pie -->
      <div class="bg-white rounded-2xl p-6 shadow-sm">
        <h3 class="text-lg font-semibold mb-4">上传/下载占比</h3>
        <div class="flex items-center justify-center">
          <div class="relative w-48 h-48">
            <svg viewBox="0 0 100 100" class="w-full h-full -rotate-90">
              <circle cx="50" cy="50" r="40" fill="none" stroke="#e5e7eb" stroke-width="20"/>
              <circle v-if="pieData.length > 0" cx="50" cy="50" r="40" fill="none" :stroke="pieData[0].color" stroke-width="20" :stroke-dasharray="`${pieData[0].percent * 2.51} 251`"/>
            </svg>
            <div class="absolute inset-0 flex flex-col items-center justify-center">
              <span class="text-2xl font-bold text-gray-900">{{ formatBytes(overview.total_traffic || 0) }}</span>
              <span class="text-sm text-gray-500">总流量</span>
            </div>
          </div>
        </div>
        <div class="flex justify-center gap-6 mt-4">
          <div v-for="item in pieData" :key="item.name" class="flex items-center gap-2">
            <div class="w-3 h-3 rounded-full" :style="{ backgroundColor: item.color }"/>
            <span class="text-sm text-gray-600">{{ item.name }} {{ item.percent.toFixed(1) }}%</span>
          </div>
        </div>
      </div>

      <!-- Server Traffic Pie -->
      <div class="bg-white rounded-2xl p-6 shadow-sm">
        <h3 class="text-lg font-semibold mb-4">节点流量占比</h3>
        <div class="space-y-3">
          <div v-for="item in serverPieData" :key="item.name" class="flex items-center gap-3">
            <div class="w-3 h-3 rounded-full flex-shrink-0" :style="{ backgroundColor: item.color }"/>
            <div class="flex-1 min-w-0">
              <div class="flex items-center justify-between mb-1">
                <span class="text-sm text-gray-700 truncate">{{ item.name }}</span>
                <span class="text-sm text-gray-500">{{ item.percent }}%</span>
              </div>
              <div class="h-2 bg-gray-100 rounded-full overflow-hidden">
                <div class="h-full rounded-full transition-all" :style="{ width: `${item.percent}%`, backgroundColor: item.color }"/>
              </div>
            </div>
            <span class="text-sm text-gray-500 flex-shrink-0">{{ formatBytes(item.value) }}</span>
          </div>
          <div v-if="serverPieData.length === 0" class="text-center text-gray-400 py-8">暂无数据</div>
        </div>
      </div>
    </div>

    <!-- Top Users -->
    <div class="bg-white rounded-2xl p-6 shadow-sm">
      <h3 class="text-lg font-semibold mb-4">用户流量排行</h3>
      <div class="overflow-x-auto">
        <table class="w-full">
          <thead>
            <tr class="text-left text-sm text-gray-500 border-b border-gray-100">
              <th class="pb-3 font-medium">排名</th>
              <th class="pb-3 font-medium">用户</th>
              <th class="pb-3 font-medium text-right">上传</th>
              <th class="pb-3 font-medium text-right">下载</th>
              <th class="pb-3 font-medium text-right">总计</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="(user, index) in (overview.top_users || [])" :key="user.user_id" class="border-b border-gray-50 hover:bg-gray-50">
              <td class="py-3">
                <span :class="['inline-flex items-center justify-center w-6 h-6 rounded-full text-xs font-medium', index < 3 ? 'bg-indigo-100 text-indigo-600' : 'bg-gray-100 text-gray-600']">
                  {{ index + 1 }}
                </span>
              </td>
              <td class="py-3 text-sm text-gray-900">{{ user.email }}</td>
              <td class="py-3 text-sm text-gray-500 text-right">{{ formatBytes(user.upload) }}</td>
              <td class="py-3 text-sm text-gray-500 text-right">{{ formatBytes(user.download) }}</td>
              <td class="py-3 text-sm font-medium text-gray-900 text-right">{{ formatBytes(user.total) }}</td>
            </tr>
            <tr v-if="!overview.top_users?.length">
              <td colspan="5" class="py-8 text-center text-gray-400">暂无数据</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>
  </div>
</template>
