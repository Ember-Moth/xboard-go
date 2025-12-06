<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

interface Order {
  id: number
  user_id: number
  user_email: string
  trade_no: string
  plan_id: number
  period: string
  total_amount: number
  status: number
  type: number
  created_at: number
}

const orders = ref<Order[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 20
const statusFilter = ref<string>('')

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '待支付', class: 'bg-yellow-100 text-yellow-600' },
  1: { text: '开通中', class: 'bg-blue-100 text-blue-600' },
  2: { text: '已取消', class: 'bg-gray-100 text-gray-600' },
  3: { text: '已完成', class: 'bg-green-100 text-green-600' },
  4: { text: '已折抵', class: 'bg-purple-100 text-purple-600' },
}

const typeMap: Record<number, string> = {
  1: '新购',
  2: '续费',
  3: '升级',
  4: '流量重置',
}

const periodMap: Record<string, string> = {
  monthly: '月付',
  quarterly: '季付',
  half_yearly: '半年付',
  yearly: '年付',
  two_yearly: '两年付',
  three_yearly: '三年付',
  onetime: '一次性',
  reset: '流量重置',
}

const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`
const formatDate = (ts: number) => new Date(ts * 1000).toLocaleString()

const fetchOrders = async () => {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize }
    if (statusFilter.value !== '') {
      params.status = statusFilter.value
    }
    const res = await api.get('/api/v2/admin/orders', { params })
    orders.value = res.data.data || []
    total.value = res.data.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const updateStatus = async (order: Order, status: number) => {
  try {
    await api.put(`/api/v2/admin/order/${order.id}/status`, { status })
    fetchOrders()
  } catch (e: any) {
    alert(e.response?.data?.error || '更新失败')
  }
}

onMounted(fetchOrders)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">订单管理</h1>
        <p class="text-gray-500 mt-1">查看所有订单</p>
      </div>
      <div class="flex items-center gap-4">
        <select v-model="statusFilter" @change="page = 1; fetchOrders()" class="px-4 py-2 border border-gray-200 rounded-xl">
          <option value="">全部状态</option>
          <option value="0">待支付</option>
          <option value="1">开通中</option>
          <option value="2">已取消</option>
          <option value="3">已完成</option>
        </select>
      </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50 border-b border-gray-200">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">订单号</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">类型</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">周期</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">金额</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">时间</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="order in orders" :key="order.id" class="hover:bg-gray-50">
              <td class="px-6 py-4 font-mono text-sm text-gray-500">{{ order.trade_no }}</td>
              <td class="px-6 py-4 text-sm">
                <div>{{ order.user_email || `用户 ${order.user_id}` }}</div>
                <div class="text-xs text-gray-400">ID: {{ order.user_id }}</div>
              </td>
              <td class="px-6 py-4 text-sm">{{ typeMap[order.type] || order.type }}</td>
              <td class="px-6 py-4 text-sm">{{ periodMap[order.period] || order.period }}</td>
              <td class="px-6 py-4 text-sm font-medium">{{ formatPrice(order.total_amount) }}</td>
              <td class="px-6 py-4">
                <span :class="['px-2 py-1 rounded-full text-xs', statusMap[order.status]?.class]">
                  {{ statusMap[order.status]?.text }}
                </span>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">{{ formatDate(order.created_at) }}</td>
              <td class="px-6 py-4 text-right">
                <div v-if="order.status === 0" class="space-x-2">
                  <button @click="updateStatus(order, 3)" class="text-green-600 hover:text-green-700 text-sm">完成</button>
                  <button @click="updateStatus(order, 2)" class="text-red-600 hover:text-red-700 text-sm">取消</button>
                </div>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="total > pageSize" class="flex items-center justify-between px-6 py-4 border-t border-gray-200">
        <span class="text-sm text-gray-500">共 {{ total }} 条</span>
        <div class="flex gap-2">
          <button @click="page--; fetchOrders()" :disabled="page <= 1" class="px-3 py-1 border rounded text-sm disabled:opacity-50">上一页</button>
          <button @click="page++; fetchOrders()" :disabled="page * pageSize >= total" class="px-3 py-1 border rounded text-sm disabled:opacity-50">下一页</button>
        </div>
      </div>
    </div>
  </div>
</template>
