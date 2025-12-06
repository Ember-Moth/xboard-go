<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

const stats = ref({
  user: { total: 0, active: 0 },
  order: { total: 0, today_count: 0, today_income: 0, month_count: 0, month_income: 0 },
  server: { total: 0 },
  ticket: { pending: 0 },
})

const loading = ref(false)

const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`

const fetchStats = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/stats/overview')
    stats.value = res.data.data
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

onMounted(fetchStats)
</script>

<template>
  <div class="space-y-6">
    <div>
      <h1 class="text-2xl font-bold text-gray-900">管理后台</h1>
      <p class="text-gray-500 mt-1">系统概览</p>
    </div>

    <!-- Stats Grid -->
    <div class="grid grid-cols-2 lg:grid-cols-3 xl:grid-cols-6 gap-4">
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <p class="text-sm text-gray-500">总用户</p>
        <p class="text-2xl font-bold text-gray-900 mt-1">{{ stats.user.total }}</p>
      </div>
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <p class="text-sm text-gray-500">活跃用户</p>
        <p class="text-2xl font-bold text-green-600 mt-1">{{ stats.user.active }}</p>
      </div>
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <p class="text-sm text-gray-500">总订单</p>
        <p class="text-2xl font-bold text-gray-900 mt-1">{{ stats.order.total }}</p>
      </div>
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <p class="text-sm text-gray-500">今日收入</p>
        <p class="text-2xl font-bold text-primary-600 mt-1">{{ formatPrice(stats.order.today_income) }}</p>
      </div>
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <p class="text-sm text-gray-500">待处理工单</p>
        <p class="text-2xl font-bold text-yellow-600 mt-1">{{ stats.ticket.pending }}</p>
      </div>
      <div class="bg-white rounded-xl p-4 shadow-sm">
        <p class="text-sm text-gray-500">节点数量</p>
        <p class="text-2xl font-bold text-blue-600 mt-1">{{ stats.server.total }}</p>
      </div>
    </div>

    <!-- Quick Actions -->
    <div class="grid grid-cols-1 lg:grid-cols-2 gap-6">
      <div class="bg-white rounded-xl p-6 shadow-sm">
        <h2 class="text-lg font-semibold mb-4">快捷操作</h2>
        <div class="grid grid-cols-2 gap-3">
          <RouterLink to="/admin/users" class="flex items-center gap-2 p-3 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors">
            <span>👥</span>
            <span class="text-sm">用户管理</span>
          </RouterLink>
          <RouterLink to="/admin/servers" class="flex items-center gap-2 p-3 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors">
            <span>🌐</span>
            <span class="text-sm">节点管理</span>
          </RouterLink>
          <RouterLink to="/admin/orders" class="flex items-center gap-2 p-3 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors">
            <span>📋</span>
            <span class="text-sm">订单管理</span>
          </RouterLink>
          <RouterLink to="/admin/tickets" class="flex items-center gap-2 p-3 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors">
            <span>💬</span>
            <span class="text-sm">工单管理</span>
          </RouterLink>
          <RouterLink to="/admin/plans" class="flex items-center gap-2 p-3 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors">
            <span>💎</span>
            <span class="text-sm">套餐管理</span>
          </RouterLink>
          <RouterLink to="/admin/coupons" class="flex items-center gap-2 p-3 rounded-lg bg-gray-50 hover:bg-gray-100 transition-colors">
            <span>🎫</span>
            <span class="text-sm">优惠券管理</span>
          </RouterLink>
        </div>
      </div>

      <div class="bg-white rounded-xl p-6 shadow-sm">
        <h2 class="text-lg font-semibold mb-4">本月统计</h2>
        <div class="space-y-3">
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">本月订单</span>
            <span class="font-medium">{{ stats.order.month_count }} 笔</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">本月收入</span>
            <span class="font-medium text-green-600">{{ formatPrice(stats.order.month_income) }}</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">今日订单</span>
            <span class="font-medium">{{ stats.order.today_count }} 笔</span>
          </div>
          <div class="flex justify-between text-sm">
            <span class="text-gray-500">今日收入</span>
            <span class="font-medium text-green-600">{{ formatPrice(stats.order.today_income) }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
