<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import api from '@/api'
import dayjs from 'dayjs'

const userStore = useUserStore()

interface Notice {
  id: number
  title: string
  content: string
  created_at?: number
}

const notices = ref<Notice[]>([])
const currentNoticeIndex = ref(0)

const fetchNotices = async () => {
  try {
    const res = await api.get('/api/v1/notices')
    notices.value = res.data.data || []
  } catch (e) {
    console.error(e)
  }
}

onMounted(fetchNotices)

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const expireDate = computed(() => {
  if (!userStore.user?.expired_at) return '永久'
  return dayjs.unix(userStore.user.expired_at).format('YYYY-MM-DD')
})

const daysLeft = computed(() => {
  if (!userStore.user?.expired_at) return -1
  const now = dayjs()
  const expire = dayjs.unix(userStore.user.expired_at)
  return expire.diff(now, 'day')
})

const usedTrafficGB = computed(() => {
  return (userStore.usedTraffic / 1024 / 1024 / 1024).toFixed(2)
})

const totalTrafficGB = computed(() => {
  return (userStore.totalTraffic / 1024 / 1024 / 1024).toFixed(0)
})

const prevNotice = () => {
  if (currentNoticeIndex.value > 0) {
    currentNoticeIndex.value--
  }
}

const nextNotice = () => {
  if (currentNoticeIndex.value < notices.value.length - 1) {
    currentNoticeIndex.value++
  }
}
</script>

<template>
  <div class="space-y-6 animate-fade-in">
    <!-- Header with Navigation -->
    <div class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
      <div class="flex items-center justify-between mb-4">
        <div>
          <h1 class="text-xl font-bold text-gray-900">欢迎回来</h1>
          <p class="text-gray-500 text-sm mt-1">查看您的服务和使用情况</p>
        </div>
        
        <!-- Navigation Tabs -->
        <div class="flex items-center gap-1 bg-gray-100 p-1 rounded-xl">
          <RouterLink to="/dashboard" class="px-4 py-2 bg-white rounded-lg text-sm font-medium shadow-sm flex items-center gap-2">
            <span>📊</span> 仪表盘
          </RouterLink>
          <RouterLink to="/plans" class="px-4 py-2 hover:bg-gray-200 rounded-lg text-sm font-medium text-gray-600 flex items-center gap-2">
            <span>🛒</span> 商店
          </RouterLink>
          <RouterLink to="/invite" class="px-4 py-2 hover:bg-gray-200 rounded-lg text-sm font-medium text-gray-600 flex items-center gap-2">
            <span>👥</span> 邀请
          </RouterLink>
          <RouterLink to="/knowledge" class="px-4 py-2 hover:bg-gray-200 rounded-lg text-sm font-medium text-gray-600 flex items-center gap-2">
            <span>➕</span> 更多
          </RouterLink>
        </div>
      </div>
      
      <!-- Email -->
      <div class="flex items-center gap-2 text-gray-500 text-sm">
        <span>📧</span>
        <span>{{ userStore.user?.email }}</span>
      </div>
    </div>

    <!-- Notice Card -->
    <div v-if="notices.length > 0" class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
      <div class="flex items-center justify-between mb-4">
        <h2 class="font-semibold text-gray-900">网站公告</h2>
        <span class="text-sm text-gray-500">第 {{ currentNoticeIndex + 1 }} 条，共 {{ notices.length }} 条</span>
      </div>
      
      <div class="bg-gray-50 rounded-xl p-4">
        <h3 class="font-medium text-gray-900 mb-2">{{ notices[currentNoticeIndex]?.title }}</h3>
        <p class="text-sm text-gray-500">{{ notices[currentNoticeIndex]?.created_at ? dayjs.unix(notices[currentNoticeIndex].created_at!).format('YYYY/MM/DD') : '' }}</p>
      </div>
      
      <div class="flex items-center justify-end gap-2 mt-4">
        <button 
          @click="prevNotice" 
          :disabled="currentNoticeIndex === 0"
          class="px-3 py-1.5 text-sm border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          ‹ 上一条
        </button>
        <button class="px-3 py-1.5 text-sm border border-primary-500 text-primary-600 rounded-lg hover:bg-primary-50">
          ⊙ 查看详情
        </button>
        <button 
          @click="nextNotice"
          :disabled="currentNoticeIndex === notices.length - 1"
          class="px-3 py-1.5 text-sm border border-gray-200 rounded-lg hover:bg-gray-50 disabled:opacity-50 disabled:cursor-not-allowed"
        >
          下一条 ›
        </button>
      </div>
    </div>

    <!-- Plan Info Card -->
    <div class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
      <h2 class="font-semibold text-gray-900 mb-4">套餐信息</h2>
      
      <div class="grid grid-cols-3 gap-6 mb-6">
        <div>
          <p class="text-sm text-gray-500 mb-1">套餐名称</p>
          <p class="font-medium text-gray-900">{{ userStore.user?.plan?.name || '无套餐' }}</p>
        </div>
        <div>
          <p class="text-sm text-gray-500 mb-1">到期时间</p>
          <p class="font-medium text-gray-900">{{ expireDate }}</p>
        </div>
        <div>
          <p class="text-sm text-gray-500 mb-1">套餐流量</p>
          <p class="font-medium text-gray-900">{{ totalTrafficGB }} GB</p>
        </div>
      </div>
      
      <div class="flex items-center gap-3">
        <RouterLink to="/subscribe" class="px-4 py-2 bg-primary-500 text-white rounded-xl text-sm font-medium hover:bg-primary-600 transition-colors flex items-center gap-2">
          <span>📤</span> 导入订阅
        </RouterLink>
        <RouterLink to="/plans" class="px-4 py-2 border border-gray-200 rounded-xl text-sm font-medium hover:bg-gray-50 transition-colors flex items-center gap-2">
          <span>🛒</span> 续费套餐
        </RouterLink>
        <RouterLink to="/tickets" class="px-4 py-2 border border-gray-200 rounded-xl text-sm font-medium hover:bg-gray-50 transition-colors flex items-center gap-2">
          <span>💬</span> 工单支持
        </RouterLink>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
      <!-- Traffic Used -->
      <div class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 hover:shadow-md transition-shadow">
        <div class="flex items-start justify-between mb-4">
          <div class="w-12 h-12 rounded-xl bg-blue-100 flex items-center justify-center">
            <span class="text-2xl">📊</span>
          </div>
          <span class="text-xs text-gray-400">本月</span>
        </div>
        <p class="text-2xl font-bold text-gray-900 mb-1">{{ usedTrafficGB }} GB</p>
        <p class="text-sm text-gray-500">已用流量</p>
        <div class="mt-3 h-2 bg-gray-100 rounded-full overflow-hidden">
          <div 
            class="h-full bg-gradient-to-r from-blue-400 to-blue-600 rounded-full transition-all duration-500"
            :style="{ width: `${Math.min(userStore.trafficPercent, 100)}%` }"
          ></div>
        </div>
      </div>

      <!-- Expire -->
      <div class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 hover:shadow-md transition-shadow">
        <div class="flex items-start justify-between mb-4">
          <div class="w-12 h-12 rounded-xl bg-green-100 flex items-center justify-center">
            <span class="text-2xl">📅</span>
          </div>
          <span class="text-xs text-gray-400">有效期</span>
        </div>
        <p class="text-2xl font-bold text-gray-900 mb-1">{{ expireDate }}</p>
        <p class="text-sm text-gray-500">
          <template v-if="daysLeft >= 0">剩余 {{ daysLeft }} 天</template>
          <template v-else>永久有效</template>
        </p>
      </div>

      <!-- Balance -->
      <div class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 hover:shadow-md transition-shadow">
        <div class="flex items-start justify-between mb-4">
          <div class="w-12 h-12 rounded-xl bg-yellow-100 flex items-center justify-center">
            <span class="text-2xl">💰</span>
          </div>
          <span class="text-xs text-gray-400">可用</span>
        </div>
        <p class="text-2xl font-bold text-gray-900 mb-1">¥{{ ((userStore.user?.balance ?? 0) / 100).toFixed(2) }}</p>
        <p class="text-sm text-gray-500">账户余额</p>
      </div>

      <!-- Help -->
      <RouterLink to="/knowledge" class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6 hover:shadow-md transition-shadow group">
        <div class="flex items-start justify-between mb-4">
          <div class="w-12 h-12 rounded-xl bg-purple-100 flex items-center justify-center">
            <span class="text-2xl">📚</span>
          </div>
          <span class="text-gray-400 group-hover:translate-x-1 transition-transform">›</span>
        </div>
        <p class="text-lg font-bold text-gray-900 mb-1">查看帮助</p>
        <p class="text-sm text-gray-500">使用教程和常见问题</p>
      </RouterLink>
    </div>

    <!-- Quick Stats Row -->
    <div class="grid grid-cols-2 md:grid-cols-4 gap-4">
      <div class="bg-gradient-to-br from-blue-50 to-blue-100 rounded-2xl p-4">
        <div class="flex items-center gap-3">
          <span class="text-2xl">⬆️</span>
          <div>
            <p class="text-sm text-gray-600">上传流量</p>
            <p class="font-semibold text-gray-900">{{ formatBytes(userStore.user?.u ?? 0) }}</p>
          </div>
        </div>
      </div>
      <div class="bg-gradient-to-br from-green-50 to-green-100 rounded-2xl p-4">
        <div class="flex items-center gap-3">
          <span class="text-2xl">⬇️</span>
          <div>
            <p class="text-sm text-gray-600">下载流量</p>
            <p class="font-semibold text-gray-900">{{ formatBytes(userStore.user?.d ?? 0) }}</p>
          </div>
        </div>
      </div>
      <div class="bg-gradient-to-br from-purple-50 to-purple-100 rounded-2xl p-4">
        <div class="flex items-center gap-3">
          <span class="text-2xl">🔄</span>
          <div>
            <p class="text-sm text-gray-600">重置日期</p>
            <p class="font-semibold text-gray-900">每月1日</p>
          </div>
        </div>
      </div>
      <div class="bg-gradient-to-br from-orange-50 to-orange-100 rounded-2xl p-4">
        <div class="flex items-center gap-3">
          <span class="text-2xl">⚡</span>
          <div>
            <p class="text-sm text-gray-600">在线设备</p>
            <p class="font-semibold text-gray-900">{{ userStore.user?.device_limit || '不限' }}</p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
