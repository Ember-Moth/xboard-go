<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'
import { useUserStore } from '@/stores/user'
import api from '@/api'

const userStore = useUserStore()
const servers = ref<any[]>([])
const loading = ref(false)
const copied = ref('')
const activeTab = ref('clients')

const subscribeUrl = computed(() => `${window.location.origin}/api/v1/client/subscribe?token=${userStore.user?.token}`)

const clients = [
  { name: 'Shadowrocket', icon: '🚀', format: 'shadowrocket', platform: 'iOS', color: 'bg-blue-500' },
  { name: 'Quantumult X', icon: '🎯', format: 'quantumultx', platform: 'iOS', color: 'bg-purple-500' },
  { name: 'Surge', icon: '🌊', format: 'surge', platform: 'iOS/macOS', color: 'bg-cyan-500' },
  { name: 'Stash', icon: '📦', format: 'stash', platform: 'iOS/macOS', color: 'bg-orange-500' },
  { name: 'Loon', icon: '🎈', format: 'loon', platform: 'iOS', color: 'bg-pink-500' },
  { name: 'sing-box', icon: '📱', format: 'singbox', platform: '全平台', color: 'bg-green-500' },
  { name: 'Clash Verge', icon: '🔥', format: 'clash', platform: 'Win/Mac/Linux', color: 'bg-red-500' },
  { name: 'v2rayN', icon: '🔷', format: 'v2rayn', platform: 'Windows', color: 'bg-blue-600' },
  { name: 'Hiddify', icon: '🔒', format: 'hiddify', platform: '全平台', color: 'bg-indigo-500' },
]

const copyUrl = async (format?: string) => {
  let url = subscribeUrl.value
  if (format) url += `&format=${format}`
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
  } catch (e) {} finally { loading.value = false }
}

onMounted(fetchServers)
</script>

<template>
  <div class="space-y-4 pb-20 lg:pb-0">
    <!-- Header -->
    <div class="bg-gradient-to-r from-indigo-500 to-purple-600 rounded-2xl p-5 text-white">
      <h1 class="text-lg font-semibold">订阅管理</h1>
      <p class="text-white/80 text-sm mt-1">选择客户端一键导入</p>
    </div>

    <!-- Subscribe URL -->
    <div class="bg-white rounded-2xl p-4 shadow-sm">
      <div class="flex items-center gap-2 mb-3">
        <svg class="w-5 h-5 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"/>
        </svg>
        <span class="text-sm font-medium text-gray-900">订阅链接</span>
      </div>
      <div class="flex gap-2">
        <div class="flex-1 bg-gray-50 rounded-xl px-3 py-2.5 font-mono text-xs text-gray-500 truncate">{{ subscribeUrl }}</div>
        <button @click="copyUrl()" class="px-4 py-2.5 bg-indigo-500 text-white rounded-xl text-sm font-medium hover:bg-indigo-600 transition-colors flex-shrink-0">
          {{ copied === 'default' ? '✓' : '复制' }}
        </button>
      </div>
      <p class="text-xs text-amber-600 mt-2 flex items-center gap-1">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z"/>
        </svg>
        请勿泄露此链接
      </p>
    </div>

    <!-- Tabs -->
    <div class="flex gap-1 bg-gray-100 p-1 rounded-xl">
      <button @click="activeTab = 'clients'" :class="activeTab === 'clients' ? 'bg-white shadow-sm' : ''" class="flex-1 px-3 py-2 rounded-lg text-sm font-medium transition-all">客户端</button>
      <button @click="activeTab = 'servers'" :class="activeTab === 'servers' ? 'bg-white shadow-sm' : ''" class="flex-1 px-3 py-2 rounded-lg text-sm font-medium transition-all">节点 ({{ servers.length }})</button>
    </div>

    <!-- Clients -->
    <div v-if="activeTab === 'clients'" class="grid grid-cols-3 gap-3">
      <button v-for="client in clients" :key="client.name" @click="importToClient(client.format)" class="bg-white rounded-2xl p-4 shadow-sm hover:shadow-md transition-all active:scale-95">
        <div :class="client.color" class="w-12 h-12 rounded-xl flex items-center justify-center text-white text-xl mx-auto mb-2 shadow-lg">{{ client.icon }}</div>
        <p class="text-sm font-medium text-gray-900 text-center">{{ client.name }}</p>
        <p class="text-xs text-gray-400 text-center mt-0.5">{{ client.platform }}</p>
        <p v-if="copied === client.format" class="text-xs text-green-500 text-center mt-1">✓ 已复制</p>
      </button>
    </div>

    <!-- Servers -->
    <div v-if="activeTab === 'servers'" class="bg-white rounded-2xl shadow-sm overflow-hidden">
      <div v-if="loading" class="p-8 text-center text-gray-400">
        <div class="animate-spin w-6 h-6 border-2 border-indigo-500 border-t-transparent rounded-full mx-auto mb-2"/>
        加载中...
      </div>
      <div v-else-if="servers.length === 0" class="p-8 text-center text-gray-400">
        <svg class="w-12 h-12 mx-auto mb-2 text-gray-300" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="1.5" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
        </svg>
        暂无节点
      </div>
      <div v-else class="divide-y divide-gray-50">
        <div v-for="server in servers" :key="server.id" class="p-4 flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-gradient-to-br from-indigo-400 to-purple-500 flex items-center justify-center text-white font-bold text-sm shadow">
            {{ server.type?.charAt(0).toUpperCase() }}
          </div>
          <div class="flex-1 min-w-0">
            <p class="font-medium text-gray-900 text-sm truncate">{{ server.name }}</p>
            <p class="text-xs text-gray-400 flex items-center gap-2">
              <span class="flex items-center gap-1"><span class="w-1.5 h-1.5 rounded-full bg-green-400"/>{{ server.type }}</span>
              <span>{{ server.rate }}x</span>
            </p>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
