<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

interface Coupon {
  id: number
  code: string
  name: string
  type: number
  value: number
  limit_use: number | null
  limit_use_with_user: number | null
  limit_plan_ids: number[] | null
  limit_period: string[] | null
  started_at: number
  ended_at: number
  created_at: number
}

const coupons = ref<Coupon[]>([])
const loading = ref(false)
const showModal = ref(false)
const editingCoupon = ref<Partial<Coupon> | null>(null)

const typeMap: Record<number, string> = {
  1: '金额',
  2: '比例',
}

const formatDate = (ts: number) => new Date(ts * 1000).toLocaleDateString()
const formatValue = (coupon: Coupon) => {
  if (coupon.type === 1) return `¥${(coupon.value / 100).toFixed(2)}`
  return `${coupon.value}%`
}

const fetchCoupons = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/coupons')
    coupons.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const openCreateModal = () => {
  const now = Math.floor(Date.now() / 1000)
  editingCoupon.value = {
    code: '',
    name: '',
    type: 1,
    value: 0,
    limit_use: null,
    limit_use_with_user: null,
    started_at: now,
    ended_at: now + 30 * 86400,
  }
  showModal.value = true
}

const openEditModal = (coupon: Coupon) => {
  editingCoupon.value = { ...coupon }
  showModal.value = true
}

const generateCode = () => {
  const chars = 'ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789'
  let code = ''
  for (let i = 0; i < 8; i++) {
    code += chars.charAt(Math.floor(Math.random() * chars.length))
  }
  editingCoupon.value!.code = code
}

const saveCoupon = async () => {
  if (!editingCoupon.value) return
  try {
    if (editingCoupon.value.id) {
      await api.put(`/api/v2/admin/coupon/${editingCoupon.value.id}`, editingCoupon.value)
    } else {
      await api.post('/api/v2/admin/coupon', editingCoupon.value)
    }
    showModal.value = false
    fetchCoupons()
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  }
}

const deleteCoupon = async (coupon: Coupon) => {
  if (!confirm(`确定要删除优惠券 "${coupon.code}" 吗？`)) return
  try {
    await api.delete(`/api/v2/admin/coupon/${coupon.id}`)
    fetchCoupons()
  } catch (e: any) {
    alert(e.response?.data?.error || '删除失败')
  }
}

onMounted(fetchCoupons)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">优惠券管理</h1>
        <p class="text-gray-500 mt-1">管理优惠券</p>
      </div>
      <button @click="openCreateModal" class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600 transition">
        添加优惠券
      </button>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

      <div v-else-if="coupons.length === 0" class="text-center py-12 text-gray-500">暂无优惠券</div>

      <table v-else class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">优惠码</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">名称</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">类型</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">优惠</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">使用限制</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">有效期</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="coupon in coupons" :key="coupon.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 font-mono font-medium text-gray-900">{{ coupon.code }}</td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ coupon.name }}</td>
            <td class="px-6 py-4 text-sm">
              <span :class="['px-2 py-1 rounded-full text-xs', coupon.type === 1 ? 'bg-green-100 text-green-600' : 'bg-blue-100 text-blue-600']">
                {{ typeMap[coupon.type] }}
              </span>
            </td>
            <td class="px-6 py-4 text-sm font-medium text-primary-600">{{ formatValue(coupon) }}</td>
            <td class="px-6 py-4 text-sm text-gray-500">
              {{ coupon.limit_use ? `总 ${coupon.limit_use} 次` : '不限' }}
              {{ coupon.limit_use_with_user ? ` / 每人 ${coupon.limit_use_with_user} 次` : '' }}
            </td>
            <td class="px-6 py-4 text-sm text-gray-500">
              {{ formatDate(coupon.started_at) }} - {{ formatDate(coupon.ended_at) }}
            </td>
            <td class="px-6 py-4 text-right space-x-2">
              <button @click="openEditModal(coupon)" class="text-primary-600 hover:text-primary-700 text-sm">编辑</button>
              <button @click="deleteCoupon(coupon)" class="text-red-600 hover:text-red-700 text-sm">删除</button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg p-6">
          <h3 class="text-lg font-bold mb-4">{{ editingCoupon?.id ? '编辑优惠券' : '添加优惠券' }}</h3>
          
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">优惠码</label>
              <div class="flex gap-2">
                <input v-model="editingCoupon!.code" type="text" class="flex-1 px-4 py-2 border border-gray-200 rounded-xl" />
                <button @click="generateCode" type="button" class="px-4 py-2 border border-gray-200 rounded-xl hover:bg-gray-50">生成</button>
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">名称</label>
              <input v-model="editingCoupon!.name" type="text" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">类型</label>
                <select v-model.number="editingCoupon!.type" class="w-full px-4 py-2 border border-gray-200 rounded-xl">
                  <option :value="1">金额 (分)</option>
                  <option :value="2">比例 (%)</option>
                </select>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">优惠值</label>
                <input v-model.number="editingCoupon!.value" type="number" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">总使用次数</label>
                <input v-model.number="editingCoupon!.limit_use" type="number" placeholder="不限" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">每人使用次数</label>
                <input v-model.number="editingCoupon!.limit_use_with_user" type="number" placeholder="不限" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">开始时间</label>
                <input :value="new Date(editingCoupon!.started_at! * 1000).toISOString().split('T')[0]" @input="editingCoupon!.started_at = Math.floor(new Date(($event.target as HTMLInputElement).value).getTime() / 1000)" type="date" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">结束时间</label>
                <input :value="new Date(editingCoupon!.ended_at! * 1000).toISOString().split('T')[0]" @input="editingCoupon!.ended_at = Math.floor(new Date(($event.target as HTMLInputElement).value).getTime() / 1000)" type="date" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
            </div>
          </div>

          <div class="flex gap-3 mt-6">
            <button @click="showModal = false" class="flex-1 px-4 py-2 border border-gray-200 text-gray-600 rounded-xl hover:bg-gray-50">取消</button>
            <button @click="saveCoupon" class="flex-1 px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
