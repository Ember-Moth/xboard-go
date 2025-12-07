<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'

interface InviteCode {
  id: number
  code: string
  status: boolean
  pv: number
  created_at: number
}

interface InviteStats {
  invited_count: number
  total_commission: number
  commission_balance: number
}

const codes = ref<InviteCode[]>([])
const stats = ref<InviteStats>({
  invited_count: 0,
  total_commission: 0,
  commission_balance: 0
})
const loading = ref(false)
const generating = ref(false)
const withdrawing = ref(false)
const copied = ref(false)

const inviteUrl = computed(() => {
  if (codes.value.length === 0) return ''
  return `${window.location.origin}/register?code=${codes.value[0].code}`
})

const fetchData = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v1/user/invite')
    codes.value = res.data.data.codes || []
    stats.value = res.data.data.stats || stats.value
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const generateCode = async () => {
  generating.value = true
  try {
    await api.post('/api/v1/user/invite/generate')
    await fetchData()
  } catch (e) {
    console.error(e)
  } finally {
    generating.value = false
  }
}

const withdraw = async () => {
  if (stats.value.commission_balance <= 0) return
  withdrawing.value = true
  try {
    await api.post('/api/v1/user/invite/withdraw')
    await fetchData()
    alert('已成功转入账户余额')
  } catch (e: any) {
    alert(e.response?.data?.error || '提现失败')
  } finally {
    withdrawing.value = false
  }
}

const copyUrl = async () => {
  try {
    await navigator.clipboard.writeText(inviteUrl.value)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  } catch (e) {
    const input = document.createElement('input')
    input.value = inviteUrl.value
    document.body.appendChild(input)
    input.select()
    document.execCommand('copy')
    document.body.removeChild(input)
    copied.value = true
    setTimeout(() => copied.value = false, 2000)
  }
}

const formatMoney = (amount: number) => {
  return (amount / 100).toFixed(2)
}

onMounted(fetchData)
</script>

<template>
  <div class="space-y-6 animate-fade-in">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">邀请返利</h1>
        <p class="text-gray-500 mt-1">邀请好友注册，获得佣金奖励</p>
      </div>
    </div>

    <!-- Stats Cards -->
    <div class="grid grid-cols-1 md:grid-cols-3 gap-4">
      <div class="bg-gradient-to-br from-blue-50 to-blue-100 rounded-2xl p-6">
        <div class="flex items-center justify-between mb-4">
          <div class="w-12 h-12 rounded-xl bg-blue-500 flex items-center justify-center shadow-lg">
            <span class="text-2xl text-white">👥</span>
          </div>
        </div>
        <div class="text-3xl font-bold text-gray-900 mb-1">{{ stats.invited_count }}</div>
        <div class="text-sm text-gray-600">邀请人数</div>
      </div>
      
      <div class="bg-gradient-to-br from-green-50 to-green-100 rounded-2xl p-6">
        <div class="flex items-center justify-between mb-4">
          <div class="w-12 h-12 rounded-xl bg-green-500 flex items-center justify-center shadow-lg">
            <span class="text-2xl text-white">💰</span>
          </div>
        </div>
        <div class="text-3xl font-bold text-gray-900 mb-1">¥{{ formatMoney(stats.total_commission) }}</div>
        <div class="text-sm text-gray-600">累计佣金</div>
      </div>
      
      <div class="bg-gradient-to-br from-orange-50 to-orange-100 rounded-2xl p-6">
        <div class="flex items-center justify-between mb-4">
          <div class="w-12 h-12 rounded-xl bg-orange-500 flex items-center justify-center shadow-lg">
            <span class="text-2xl text-white">🎁</span>
          </div>
          <button
            v-if="stats.commission_balance > 0"
            @click="withdraw"
            :disabled="withdrawing"
            class="px-3 py-1.5 bg-orange-500 text-white rounded-lg text-sm hover:bg-orange-600 transition disabled:opacity-50"
          >
            {{ withdrawing ? '处理中...' : '转入余额' }}
          </button>
        </div>
        <div class="text-3xl font-bold text-gray-900 mb-1">¥{{ formatMoney(stats.commission_balance) }}</div>
        <div class="text-sm text-gray-600">可提现佣金</div>
      </div>
    </div>

    <!-- Invite Link Card -->
    <div class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
      <div class="flex items-center gap-3 mb-6">
        <div class="w-10 h-10 rounded-xl bg-primary-100 flex items-center justify-center">
          <span class="text-xl">🔗</span>
        </div>
        <div>
          <h2 class="font-semibold text-gray-900">邀请链接</h2>
          <p class="text-sm text-gray-500">分享链接给好友，好友注册并购买套餐后获得佣金</p>
        </div>
      </div>
      
      <div v-if="codes.length > 0" class="space-y-4">
        <div class="flex gap-3">
          <div class="flex-1 bg-gray-50 rounded-xl px-4 py-3 font-mono text-sm text-gray-600 truncate">
            {{ inviteUrl }}
          </div>
          <button
            @click="copyUrl"
            class="px-6 py-3 bg-primary-500 text-white rounded-xl font-medium hover:bg-primary-600 transition-colors whitespace-nowrap"
          >
            {{ copied ? '✓ 已复制' : '复制链接' }}
          </button>
        </div>
        
        <!-- Share Buttons -->
        <div class="flex items-center gap-3 pt-2">
          <span class="text-sm text-gray-500">分享到：</span>
          <button class="w-10 h-10 rounded-xl bg-green-500 text-white flex items-center justify-center hover:bg-green-600 transition-colors">
            <span>💬</span>
          </button>
          <button class="w-10 h-10 rounded-xl bg-blue-500 text-white flex items-center justify-center hover:bg-blue-600 transition-colors">
            <span>📱</span>
          </button>
          <button class="w-10 h-10 rounded-xl bg-sky-500 text-white flex items-center justify-center hover:bg-sky-600 transition-colors">
            <span>✈️</span>
          </button>
        </div>
      </div>

      <div v-else class="text-center py-8">
        <div class="w-16 h-16 rounded-2xl bg-gray-100 flex items-center justify-center mx-auto mb-4">
          <span class="text-3xl">🎫</span>
        </div>
        <p class="text-gray-500 mb-4">您还没有邀请码</p>
        <button
          @click="generateCode"
          :disabled="generating"
          class="px-6 py-3 bg-primary-500 text-white rounded-xl font-medium hover:bg-primary-600 transition disabled:opacity-50"
        >
          {{ generating ? '生成中...' : '生成邀请码' }}
        </button>
      </div>
    </div>

    <!-- How it works -->
    <div class="bg-white rounded-2xl shadow-sm border border-gray-100 p-6">
      <h2 class="font-semibold text-gray-900 mb-6">邀请流程</h2>
      <div class="grid grid-cols-1 md:grid-cols-4 gap-4">
        <div class="text-center">
          <div class="w-12 h-12 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center mx-auto mb-3 text-xl font-bold">1</div>
          <p class="text-sm font-medium text-gray-900">分享链接</p>
          <p class="text-xs text-gray-500 mt-1">复制邀请链接分享给好友</p>
        </div>
        <div class="text-center">
          <div class="w-12 h-12 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center mx-auto mb-3 text-xl font-bold">2</div>
          <p class="text-sm font-medium text-gray-900">好友注册</p>
          <p class="text-xs text-gray-500 mt-1">好友通过链接注册账号</p>
        </div>
        <div class="text-center">
          <div class="w-12 h-12 rounded-full bg-primary-100 text-primary-600 flex items-center justify-center mx-auto mb-3 text-xl font-bold">3</div>
          <p class="text-sm font-medium text-gray-900">好友购买</p>
          <p class="text-xs text-gray-500 mt-1">好友购买任意套餐</p>
        </div>
        <div class="text-center">
          <div class="w-12 h-12 rounded-full bg-green-100 text-green-600 flex items-center justify-center mx-auto mb-3 text-xl font-bold">✓</div>
          <p class="text-sm font-medium text-gray-900">获得佣金</p>
          <p class="text-xs text-gray-500 mt-1">自动获得订单佣金</p>
        </div>
      </div>
    </div>

    <!-- Invite Codes List -->
    <div v-if="codes.length > 0" class="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
      <div class="p-6 border-b border-gray-100 flex items-center justify-between">
        <div class="flex items-center gap-3">
          <div class="w-10 h-10 rounded-xl bg-purple-100 flex items-center justify-center">
            <span class="text-xl">🎫</span>
          </div>
          <div>
            <h2 class="font-semibold text-gray-900">我的邀请码</h2>
            <p class="text-sm text-gray-500">共 {{ codes.length }} 个邀请码</p>
          </div>
        </div>
        <button
          @click="generateCode"
          :disabled="generating"
          class="px-4 py-2 text-primary-600 hover:bg-primary-50 rounded-lg transition text-sm font-medium"
        >
          {{ generating ? '生成中...' : '+ 生成新邀请码' }}
        </button>
      </div>

      <div class="divide-y divide-gray-100">
        <div v-for="code in codes" :key="code.id" class="p-4 hover:bg-gray-50 transition-colors flex items-center justify-between">
          <div class="flex items-center gap-4">
            <div class="w-10 h-10 rounded-xl bg-gray-100 flex items-center justify-center font-mono text-sm">
              {{ code.code.slice(0, 2) }}
            </div>
            <div>
              <p class="font-mono text-sm text-gray-900">{{ code.code }}</p>
              <p class="text-xs text-gray-500">
                创建于 {{ new Date(code.created_at * 1000).toLocaleDateString() }}
              </p>
            </div>
          </div>
          <div class="flex items-center gap-4">
            <div class="text-right">
              <p class="text-sm text-gray-600">{{ code.pv }} 次访问</p>
            </div>
            <span
              :class="code.status ? 'bg-gray-100 text-gray-600' : 'bg-green-100 text-green-600'"
              class="px-3 py-1 rounded-full text-xs font-medium"
            >
              {{ code.status ? '已使用' : '可用' }}
            </span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
