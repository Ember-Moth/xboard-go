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
const showNoticeModal = ref(false)
const currentNotice = ref<Notice | null>(null)

const fetchNotices = async () => {
  try {
    const res = await api.get('/api/v1/notices')
    notices.value = res.data.data || []
  } catch (e) {}
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

const usedTrafficGB = computed(() => (userStore.usedTraffic / 1024 / 1024 / 1024).toFixed(2))
const totalTrafficGB = computed(() => (userStore.totalTraffic / 1024 / 1024 / 1024).toFixed(0))

const openNotice = (notice: Notice) => {
  currentNotice.value = notice
  showNoticeModal.value = true
}
</script>

<template>
  <div class="space-y-4 pb-20 lg:pb-0">
    <!-- Welcome Card -->
    <div class="bg-gradient-to-r from-indigo-500 to-purple-600 rounded-2xl p-5 text-white">
      <div class="flex items-center justify-between">
        <div>
          <h1 class="text-lg font-semibold">欢迎回来 👋</h1>
          <p class="text-white/80 text-sm mt-1">{{ userStore.user?.email }}</p>
        </div>
        <div class="w-12 h-12 rounded-full bg-white/20 flex items-center justify-center text-2xl font-bold">
          {{ userStore.user?.email?.charAt(0).toUpperCase() }}
        </div>
      </div>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-2 gap-3">
      <!-- Traffic -->
      <div class="bg-white rounded-2xl p-4 shadow-sm">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-8 h-8 rounded-lg bg-blue-100 flex items-center justify-center">
            <svg class="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16V4m0 0L3 8m4-4l4 4m6 0v12m0 0l4-4m-4 4l-4-4"/>
            </svg>
          </div>
          <span class="text-xs text-gray-500">流量</span>
        </div>
        <p class="text-xl font-bold text-gray-900">{{ usedTrafficGB }}<span class="text-sm font-normal text-gray-400"> / {{ totalTrafficGB }} GB</span></p>
        <div class="mt-2 h-1.5 bg-gray-100 rounded-full overflow-hidden">
          <div class="h-full bg-gradient-to-r from-blue-400 to-blue-600 rounded-full transition-all" :style="{ width: `${Math.min(userStore.trafficPercent, 100)}%` }"/>
        </div>
      </div>

      <!-- Expire -->
      <div class="bg-white rounded-2xl p-4 shadow-sm">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-8 h-8 rounded-lg bg-green-100 flex items-center justify-center">
            <svg class="w-4 h-4 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 7V3m8 4V3m-9 8h10M5 21h14a2 2 0 002-2V7a2 2 0 00-2-2H5a2 2 0 00-2 2v12a2 2 0 002 2z"/>
            </svg>
          </div>
          <span class="text-xs text-gray-500">到期</span>
        </div>
        <p class="text-xl font-bold text-gray-900">{{ expireDate }}</p>
        <p class="text-xs text-gray-400 mt-1">
          <template v-if="daysLeft >= 0">剩余 {{ daysLeft }} 天</template>
          <template v-else>永久有效</template>
        </p>
      </div>

      <!-- Balance -->
      <div class="bg-white rounded-2xl p-4 shadow-sm">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-8 h-8 rounded-lg bg-yellow-100 flex items-center justify-center">
            <svg class="w-4 h-4 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 8c-1.657 0-3 .895-3 2s1.343 2 3 2 3 .895 3 2-1.343 2-3 2m0-8c1.11 0 2.08.402 2.599 1M12 8V7m0 1v8m0 0v1m0-1c-1.11 0-2.08-.402-2.599-1M21 12a9 9 0 11-18 0 9 9 0 0118 0z"/>
            </svg>
          </div>
          <span class="text-xs text-gray-500">余额</span>
        </div>
        <p class="text-xl font-bold text-gray-900">¥{{ ((userStore.user?.balance ?? 0) / 100).toFixed(2) }}</p>
      </div>

      <!-- Plan -->
      <div class="bg-white rounded-2xl p-4 shadow-sm">
        <div class="flex items-center gap-2 mb-3">
          <div class="w-8 h-8 rounded-lg bg-purple-100 flex items-center justify-center">
            <svg class="w-4 h-4 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"/>
            </svg>
          </div>
          <span class="text-xs text-gray-500">套餐</span>
        </div>
        <p class="text-lg font-bold text-gray-900 truncate">{{ userStore.user?.plan?.name || '无套餐' }}</p>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="bg-white rounded-2xl p-4 shadow-sm">
      <h3 class="text-sm font-medium text-gray-900 mb-3">快捷操作</h3>
      <div class="grid grid-cols-4 gap-2">
        <RouterLink to="/subscribe" class="flex flex-col items-center gap-1.5 p-3 rounded-xl hover:bg-gray-50 transition-colors">
          <div class="w-10 h-10 rounded-xl bg-indigo-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 16v1a3 3 0 003 3h10a3 3 0 003-3v-1m-4-8l-4-4m0 0L8 8m4-4v12"/>
            </svg>
          </div>
          <span class="text-xs text-gray-600">导入订阅</span>
        </RouterLink>
        <RouterLink to="/plans" class="flex flex-col items-center gap-1.5 p-3 rounded-xl hover:bg-gray-50 transition-colors">
          <div class="w-10 h-10 rounded-xl bg-green-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z"/>
            </svg>
          </div>
          <span class="text-xs text-gray-600">购买套餐</span>
        </RouterLink>
        <RouterLink to="/tickets" class="flex flex-col items-center gap-1.5 p-3 rounded-xl hover:bg-gray-50 transition-colors">
          <div class="w-10 h-10 rounded-xl bg-orange-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-orange-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 10h.01M12 10h.01M16 10h.01M9 16H5a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v8a2 2 0 01-2 2h-5l-5 5v-5z"/>
            </svg>
          </div>
          <span class="text-xs text-gray-600">联系客服</span>
        </RouterLink>
        <RouterLink to="/knowledge" class="flex flex-col items-center gap-1.5 p-3 rounded-xl hover:bg-gray-50 transition-colors">
          <div class="w-10 h-10 rounded-xl bg-blue-100 flex items-center justify-center">
            <svg class="w-5 h-5 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 6.253v13m0-13C10.832 5.477 9.246 5 7.5 5S4.168 5.477 3 6.253v13C4.168 18.477 5.754 18 7.5 18s3.332.477 4.5 1.253m0-13C13.168 5.477 14.754 5 16.5 5c1.747 0 3.332.477 4.5 1.253v13C19.832 18.477 18.247 18 16.5 18c-1.746 0-3.332.477-4.5 1.253"/>
            </svg>
          </div>
          <span class="text-xs text-gray-600">使用教程</span>
        </RouterLink>
      </div>
    </div>

    <!-- Traffic Details -->
    <div class="bg-white rounded-2xl p-4 shadow-sm">
      <h3 class="text-sm font-medium text-gray-900 mb-3">流量详情</h3>
      <div class="grid grid-cols-2 gap-3">
        <div class="flex items-center gap-3 p-3 bg-blue-50 rounded-xl">
          <svg class="w-5 h-5 text-blue-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 11l5-5m0 0l5 5m-5-5v12"/>
          </svg>
          <div>
            <p class="text-xs text-gray-500">上传</p>
            <p class="font-semibold text-gray-900">{{ formatBytes(userStore.user?.u ?? 0) }}</p>
          </div>
        </div>
        <div class="flex items-center gap-3 p-3 bg-green-50 rounded-xl">
          <svg class="w-5 h-5 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M17 13l-5 5m0 0l-5-5m5 5V6"/>
          </svg>
          <div>
            <p class="text-xs text-gray-500">下载</p>
            <p class="font-semibold text-gray-900">{{ formatBytes(userStore.user?.d ?? 0) }}</p>
          </div>
        </div>
      </div>
    </div>

    <!-- Notices -->
    <div v-if="notices.length > 0" class="bg-white rounded-2xl p-4 shadow-sm">
      <h3 class="text-sm font-medium text-gray-900 mb-3">公告</h3>
      <div class="space-y-2">
        <div v-for="notice in notices.slice(0, 3)" :key="notice.id" @click="openNotice(notice)" class="flex items-center gap-3 p-3 bg-gray-50 rounded-xl cursor-pointer hover:bg-gray-100 transition-colors">
          <div class="w-8 h-8 rounded-lg bg-indigo-100 flex items-center justify-center flex-shrink-0">
            <svg class="w-4 h-4 text-indigo-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M11 5.882V19.24a1.76 1.76 0 01-3.417.592l-2.147-6.15M18 13a3 3 0 100-6M5.436 13.683A4.001 4.001 0 017 6h1.832c4.1 0 7.625-1.234 9.168-3v14c-1.543-1.766-5.067-3-9.168-3H7a3.988 3.988 0 01-1.564-.317z"/>
            </svg>
          </div>
          <div class="flex-1 min-w-0">
            <p class="text-sm font-medium text-gray-900 truncate">{{ notice.title }}</p>
            <p class="text-xs text-gray-400">{{ notice.created_at ? dayjs.unix(notice.created_at).format('MM-DD') : '' }}</p>
          </div>
          <svg class="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 5l7 7-7 7"/>
          </svg>
        </div>
      </div>
    </div>

    <!-- Notice Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showNoticeModal" class="fixed inset-0 z-50 flex items-end lg:items-center justify-center p-4" @click.self="showNoticeModal = false">
          <div class="fixed inset-0 bg-black/50" @click="showNoticeModal = false"/>
          <div class="relative bg-white rounded-t-2xl lg:rounded-2xl w-full max-w-lg max-h-[80vh] overflow-hidden">
            <div class="sticky top-0 bg-white border-b border-gray-100 px-4 py-3 flex items-center justify-between">
              <h3 class="font-semibold text-gray-900">{{ currentNotice?.title }}</h3>
              <button @click="showNoticeModal = false" class="p-1 rounded-lg hover:bg-gray-100">
                <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/>
                </svg>
              </button>
            </div>
            <div class="p-4 overflow-y-auto prose prose-sm max-w-none" v-html="currentNotice?.content"/>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.modal-enter-active, .modal-leave-active {
  transition: all 0.3s ease;
}
.modal-enter-from, .modal-leave-to {
  opacity: 0;
}
.modal-enter-from > div:last-child, .modal-leave-to > div:last-child {
  transform: translateY(100%);
}
@media (min-width: 1024px) {
  .modal-enter-from > div:last-child, .modal-leave-to > div:last-child {
    transform: scale(0.95);
  }
}
</style>
