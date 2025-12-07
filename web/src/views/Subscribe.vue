<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import api from '@/api'

const userStore = useUserStore()
const servers = ref<any[]>([])
const loading = ref(false)
const copied = ref('')
const activeTab = ref('clients')

const subscribeUrl = computed(() => {
  return `${window.location.origin}/api/v1/client/subscribe?token=${userStore.user?.token}`
})

// 客户端分类
const clientCategories = [
  {
    name: 'iOS',
    clients: [
      { name: 'Shadowrocket', icon: '🚀', format: 'shadowrocket', color: 'from-blue-400 to-blue-600' },
      { name: 'Quantumult X', icon: '🎯', format: 'quantumultx', color: 'from-purple-400 to-purple-600' },
      { name: 'Surge', icon: '🌊', format: 'surge', color: 'from-cyan-400 to-cyan-600' },
      { name: 'Stash', icon: '📦', format: 'stash', color: 'from-orange-400 to-orange-600' },
      { name: 'Loon', icon: '🎈', format: 'loon', color: 'from-pink-400 to-pink-600' },
    ]
  },
  {
    name: 'Android',
    clients: [
      { name: 'sing-box', icon: '📱', format: 'singbox', color: 'from-green-400 to-green-600' },
      { name: 'Clash Meta', icon: '⚡', format: 'clashmeta', color: 'from-yellow-400 to-yellow-600' },
      { name: 'v2rayNG', icon: '🔷', format: 'v2rayng', color: 'from-blue-400 to-blue-600' },
      { name: 'Surfboard', icon: '🏄', format: 'surfboard', color: 'from-teal-400 to-teal-600' },
      { name: 'Hiddify', icon: '🔒', format: 'hiddify', color: 'from-indigo-400 to-indigo-600' },
    ]
  },
  {
    name: 'Windows',
    clients: [
      { name: 'Clash Verge', icon: '🔥', format: 'clash', color: 'from-red-400 to-red-600' },
      { name: 'v2rayN', icon: '🔷', format: 'v2rayn', color: 'from-blue-400 to-blue-600' },
      { name: 'sing-box', icon: '📦', format: 'singbox', color: 'from-green-400 to-green-600' },
      { name: 'Hiddify', icon: '🔒', format: 'hiddify', color: 'from-indigo-400 to-indigo-600' },
    ]
  },
  {
    name: 'macOS',
    clients: [
      { name: 'Surge', icon: '🌊', format: 'surge', color: 'from-cyan-400 to-cyan-600' },
      { name: 'Clash Verge', icon: '🔥', format: 'clash', color: 'from-red-400 to-red-600' },
      { name: 'sing-box', icon: '📦', format: 'singbox', color: 'from-green-400 to-green-600' },
      { name: 'Stash', icon: '📦', format: 'stash', color: 'from-orange-400 to-orange-600' },
    ]
  },
  {
    name: 'Linux',
    clients: [
      { name: 'Clash Meta', icon: '⚡', format: 'clashmeta', color: 'from-yellow-400 to-yellow-600' },
      { name: 'sing-box', icon: '📦', format: 'singbox', color: 'from-green-400 to-green-600' },
    ]
  }
]

// 热门客户端（横向展示）
const popularClients = [
  { name: 'Shadowrocket', icon: '🚀', format: 'shadowrocket', bg: 'bg-gradient-to-br from-blue-50 to-blue-100', iconBg: 'bg-blue-500' },
  { name: 'Surge', icon: '🌊', format: 'surge', bg: 'bg-gradient-to-br from-cyan-50 to-cyan-100', iconBg: 'bg-cyan-500' },
  { name: 'Stash', icon: '📦', format: 'stash', bg: 'bg-gradient-to-br from-orange-50 to-orange-100', iconBg: 'bg-orange-500' },
  { name: 'Quantumult X', icon: '🎯', format: 'quantumultx', bg: 'bg-gradient-to-br from-purple-50 to-purple-100', iconBg: 'bg-purple-500' },
  { name: 'Hiddify', icon: '🔒', format: 'hiddify', bg: 'bg-gradient-to-br from-indigo-50 to-indigo-100', iconBg: 'bg-indigo-500' },
  { name: 'sing-box', icon: '📱', format: 'singbox', bg: 'bg-gradient-to-br from-green-50 to-green-100', iconBg: 'bg-green-500' },
  { name: 'Loon', icon: '🎈', format: 'loon', bg: 'bg-gradient-to-br from-pink-50 to-pink-100', iconBg: 'bg-pink-500' },
]

const copyUrl = async (format?: string) => {
  let url = subscribeUrl.value
  if (format) {
    url += `&format=${format}`
  }
  
  try {
    await navigator.clipboard.writeText(url)
    copied.value = format || 'default'
    setTimeout(() => copied.value = '', 2000)
  } catch (e) {
    const input = document.createElement('input')
    input.value = url
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
    copied.value = format || 'default'
    setTimeout(() => copied.value = '', 2000)
  }
}

const importToClient = (format: string) => {
  const url = subscribeUrl.value + `&format=${format}`
  // 尝试使用 URL Scheme 导入
  const schemes: Record<string, string> = {
    shadowrocket: `shadowrocket://add/sub://${btoa(url)}?remark=${encodeURIComponent('订阅')}`,
    quantumultx: `quantumult-x:///add-resource?remote-resource=${encodeURIComponent(JSON.stringify({ server_remote: [url + ', tag=订阅'] }))}`,
    clash: `clash://install-config?url=${encodeURIComponent(url)}`,
    clashmeta: `clash://install-config?url=${encodeURIComponent(url)}`,
    surge: `surge:///install-config?url=${encodeURIComponent(url)}`,
    stash: `stash://install-config?url=${encodeURIComponent(url)}`,
    loon: `loon://import?sub=${encodeURIComponent(url)}`,
    singbox: `sing-box://import-remote-profile?url=${encodeURIComponent(url)}`,
  }
  
  if (schemes[format]) {
    window.location.href = schemes[format]
  } else {
    copyUrl(format)
  }
}

const fetchServers = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v1/user/subscribe')
    servers.value = res.data.data.servers || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchServers)
</script>

<template>
  <div class="space-y-6 animate-fade-in">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">订阅管理</h1>
        <p class="text-gray-500 mt-1">选择客户端导入订阅</p>
      </div>
    </div>

    <!-- Popular Clients - 横向滚动 -->
    <div class="bg-gradient-to-r from-slate-50 to-blue-50 rounded-2xl p-6">
      <div class="flex items-center gap-4 overflow-x-auto pb-2 scrollbar-hide">
        <button
          v-for="client in popularClients"
          :key="client.name"
          @click="importToClient(client.format)"
          :class="client.bg"
          class="flex-shrink-0 flex flex-col items-center gap-2 px-6 py-4 rounded-2xl hover:shadow-lg transition-all duration-300 hover:-translate-y-1 min-w-[100px]"
        >
          <div :class="client.iconBg" class="w-12 h-12 rounded-xl flex items-center justify-center text-white text-xl shadow-lg">
            {{ client.icon }}
          </div>
          <span class="text-sm font-medium text-gray-700 whitespace-nowrap">{{ client.name }}</span>
        </button>
      </div>
    </div>

    <!-- Subscribe URL Card -->
    <div class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
      <div class="flex items-center gap-3 mb-4">
        <div class="w-10 h-10 rounded-xl bg-primary-100 flex items-center justify-center">
          <span class="text-xl">🔗</span>
        </div>
        <div>
          <h2 class="font-semibold text-gray-900">通用订阅链接</h2>
          <p class="text-sm text-gray-500">复制链接手动导入客户端</p>
        </div>
      </div>
      
      <div class="flex gap-3">
        <div class="flex-1 bg-gray-50 rounded-xl px-4 py-3 font-mono text-sm text-gray-600 truncate">
          {{ subscribeUrl }}
        </div>
        <button 
          @click="copyUrl()" 
          class="px-6 py-3 bg-primary-500 text-white rounded-xl font-medium hover:bg-primary-600 transition-colors whitespace-nowrap"
        >
          {{ copied === 'default' ? '✓ 已复制' : '复制链接' }}
        </button>
      </div>
      
      <div class="mt-4 flex items-center gap-2 text-sm text-amber-600 bg-amber-50 rounded-xl px-4 py-3">
        <span>⚠️</span>
        <span>请勿泄露此链接，如已泄露请在设置中重置</span>
      </div>
    </div>

    <!-- Tabs -->
    <div class="flex gap-2 bg-gray-100 p-1 rounded-xl w-fit">
      <button 
        @click="activeTab = 'clients'"
        :class="activeTab === 'clients' ? 'bg-white shadow-sm' : 'hover:bg-gray-200'"
        class="px-4 py-2 rounded-lg text-sm font-medium transition-all"
      >
        按平台选择
      </button>
      <button 
        @click="activeTab = 'servers'"
        :class="activeTab === 'servers' ? 'bg-white shadow-sm' : 'hover:bg-gray-200'"
        class="px-4 py-2 rounded-lg text-sm font-medium transition-all"
      >
        节点列表
      </button>
    </div>

    <!-- Clients by Platform -->
    <div v-if="activeTab === 'clients'" class="space-y-6">
      <div v-for="category in clientCategories" :key="category.name" class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
        <h3 class="font-semibold text-gray-900 mb-4 flex items-center gap-2">
          <span class="w-2 h-2 rounded-full bg-primary-500"></span>
          {{ category.name }}
        </h3>
        <div class="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-5 gap-3">
          <button
            v-for="client in category.clients"
            :key="client.name"
            @click="importToClient(client.format)"
            class="group flex flex-col items-center gap-2 p-4 rounded-xl bg-gray-50 hover:bg-gray-100 transition-all duration-200 hover:shadow-md"
          >
            <div :class="`bg-gradient-to-br ${client.color}`" class="w-12 h-12 rounded-xl flex items-center justify-center text-white text-xl shadow-md group-hover:scale-110 transition-transform">
              {{ client.icon }}
            </div>
            <span class="text-sm font-medium text-gray-700">{{ client.name }}</span>
            <span 
              v-if="copied === client.format" 
              class="text-xs text-green-600 bg-green-100 px-2 py-0.5 rounded-full"
            >
              已复制
            </span>
          </button>
        </div>
      </div>
    </div>

    <!-- Server List -->
    <div v-if="activeTab === 'servers'" class="bg-white rounded-2xl shadow-sm border border-gray-100">
      <div class="p-6 border-b border-gray-100 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-green-100 flex items-center justify-center">
            <span class="text-xl">🌐</span>
          </div>
          <div>
            <h2 class="font-semibold text-gray-900">可用节点</h2>
            <p class="text-sm text-gray-500">共 {{ servers.length }} 个节点</p>
          </div>
        </div>
        <button 
          @click="fetchServers" 
          :disabled="loading"
          class="px-4 py-2 text-sm text-primary-600 hover:bg-primary-50 rounded-lg transition-colors"
        >
          {{ loading ? '刷新中...' : '刷新列表' }}
        </button>
      </div>
      
      <div v-if="loading" class="p-12 text-center text-gray-500">
        <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full mx-auto mb-3"></div>
        加载中...
      </div>
      
      <div v-else-if="servers.length === 0" class="p-12 text-center text-gray-500">
        <span class="text-4xl mb-3 block">📭</span>
        暂无可用节点
      </div>
      
      <div v-else class="divide-y divide-gray-100">
        <div 
          v-for="server in servers" 
          :key="server.id"
          class="p-4 hover:bg-gray-50 transition-colors flex items-center justify-between"
        >
          <div class="flex items-center gap-4">
            <div class="w-12 h-12 rounded-xl bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center text-white font-bold shadow-md">
              {{ server.type?.charAt(0).toUpperCase() }}
            </div>
            <div>
              <p class="font-medium text-gray-900">{{ server.name }}</p>
              <p class="text-sm text-gray-500">
                <span class="inline-flex items-center gap-1">
                  <span class="w-1.5 h-1.5 rounded-full bg-green-500"></span>
                  {{ server.type }}
                </span>
                <span class="mx-2">·</span>
                <span>{{ server.rate }}x 倍率</span>
              </p>
            </div>
          </div>
          <div class="flex items-center gap-2">
            <span 
              v-for="tag in (server.tags || [])" 
              :key="tag"
              class="px-2 py-1 text-xs rounded-full bg-blue-100 text-blue-600"
            >
              {{ tag }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<style scoped>
.scrollbar-hide::-webkit-scrollbar {
  display: none;
}
.scrollbar-hide {
  -ms-overflow-style: none;
  scrollbar-width: none;
}
</style>
