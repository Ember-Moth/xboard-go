<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import api from '@/api'

interface Plan {
  id: number
  name: string
  transfer_enable: number
  speed_limit: number | null
  prices: Record<string, number>
  content: string | null
}

const router = useRouter()
const plans = ref<Plan[]>([])
const loading = ref(false)
const selectedPlan = ref<Plan | null>(null)
const selectedPeriod = ref('')
const showModal = ref(false)
const ordering = ref(false)
const couponCode = ref('')
const couponChecking = ref(false)
const couponDiscount = ref(0)
const couponError = ref('')

const periods = [
  { key: 'monthly', name: '月付', months: 1 },
  { key: 'quarterly', name: '季付', months: 3 },
  { key: 'half_yearly', name: '半年付', months: 6 },
  { key: 'yearly', name: '年付', months: 12 },
  { key: 'two_yearly', name: '两年付', months: 24 },
  { key: 'three_yearly', name: '三年付', months: 36 },
  { key: 'onetime', name: '一次性', months: -1 },
]

const formatBytes = (gb: number) => gb >= 1024 ? `${(gb / 1024).toFixed(0)} TB` : `${gb} GB`
const formatPrice = (cents: number) => `¥${(cents / 100).toFixed(2)}`
const getAvailablePeriods = (plan: Plan) => periods.filter(p => plan.prices && plan.prices[p.key] > 0)
const getLowestPrice = (plan: Plan) => {
  const available = getAvailablePeriods(plan)
  return available.length === 0 ? 0 : Math.min(...available.map(p => plan.prices[p.key]))
}

const openOrderModal = (plan: Plan) => {
  selectedPlan.value = plan
  const available = getAvailablePeriods(plan)
  selectedPeriod.value = available[0]?.key || ''
  couponCode.value = ''
  couponDiscount.value = 0
  couponError.value = ''
  showModal.value = true
}

const checkCoupon = async () => {
  if (!couponCode.value.trim() || !selectedPlan.value || !selectedPeriod.value) return
  couponChecking.value = true
  couponError.value = ''
  couponDiscount.value = 0
  try {
    const res = await api.post('/api/v1/user/coupon/check', { code: couponCode.value, plan_id: selectedPlan.value.id, period: selectedPeriod.value })
    couponDiscount.value = res.data.data.discount
  } catch (e: any) {
    couponError.value = e.response?.data?.error || '优惠券无效'
  } finally {
    couponChecking.value = false
  }
}

const getFinalPrice = () => {
  if (!selectedPlan.value || !selectedPeriod.value) return 0
  return Math.max(0, (selectedPlan.value.prices[selectedPeriod.value] || 0) - couponDiscount.value)
}

const createOrder = async () => {
  if (!selectedPlan.value || !selectedPeriod.value) return
  ordering.value = true
  try {
    await api.post('/api/v1/user/order/create', { plan_id: selectedPlan.value.id, period: selectedPeriod.value, coupon_code: couponCode.value || undefined })
    showModal.value = false
    router.push('/orders')
  } catch (e: any) {
    alert(e.response?.data?.error || '创建订单失败')
  } finally {
    ordering.value = false
  }
}

const fetchPlans = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v1/guest/plans')
    plans.value = res.data.data || []
  } catch (e) {} finally { loading.value = false }
}

onMounted(fetchPlans)
</script>

<template>
  <div class="space-y-4 pb-20 lg:pb-0">
    <!-- Header -->
    <div class="bg-gradient-to-r from-indigo-500 to-purple-600 rounded-2xl p-5 text-white">
      <h1 class="text-lg font-semibold">选择套餐</h1>
      <p class="text-white/80 text-sm mt-1">选择适合您的订阅计划</p>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-400">
      <div class="animate-spin w-6 h-6 border-2 border-indigo-500 border-t-transparent rounded-full mx-auto mb-2"/>
      加载中...
    </div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4">
      <div v-for="plan in plans" :key="plan.id" class="bg-white rounded-2xl shadow-sm p-5 hover:shadow-md transition-shadow">
        <div class="text-center mb-4">
          <h3 class="text-lg font-bold text-gray-900">{{ plan.name }}</h3>
          <div class="mt-2">
            <span class="text-2xl font-bold text-indigo-600">{{ formatPrice(getLowestPrice(plan)) }}</span>
            <span class="text-gray-400 text-sm">/起</span>
          </div>
        </div>
        <div class="space-y-2 mb-4 text-sm">
          <div class="flex items-center gap-2 text-gray-600">
            <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>
            {{ formatBytes(plan.transfer_enable) }} 流量
          </div>
          <div class="flex items-center gap-2 text-gray-600">
            <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>
            {{ plan.speed_limit ? `${plan.speed_limit} Mbps` : '不限速' }}
          </div>
          <div class="flex items-center gap-2 text-gray-600">
            <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"/></svg>
            全部节点
          </div>
        </div>
        <button @click="openOrderModal(plan)" class="w-full py-2.5 bg-indigo-500 text-white rounded-xl font-medium hover:bg-indigo-600 transition-colors active:scale-[0.98]">
          立即订购
        </button>
      </div>
    </div>

    <!-- Order Modal -->
    <Teleport to="body">
      <Transition name="modal">
        <div v-if="showModal" class="fixed inset-0 z-50 flex items-end lg:items-center justify-center p-4" @click.self="showModal = false">
          <div class="fixed inset-0 bg-black/50" @click="showModal = false"/>
          <div class="relative bg-white rounded-t-2xl lg:rounded-2xl w-full max-w-md p-5 max-h-[85vh] overflow-y-auto">
            <div class="flex items-center justify-between mb-4">
              <h3 class="text-lg font-bold">确认订单</h3>
              <button @click="showModal = false" class="p-1 rounded-lg hover:bg-gray-100">
                <svg class="w-5 h-5 text-gray-500" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12"/></svg>
              </button>
            </div>
            
            <p class="text-gray-600 text-sm mb-4">套餐：<span class="font-medium text-gray-900">{{ selectedPlan?.name }}</span></p>

            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">选择周期</label>
              <div class="grid grid-cols-2 gap-2">
                <button v-for="period in getAvailablePeriods(selectedPlan!)" :key="period.key" @click="selectedPeriod = period.key; couponDiscount = 0; couponError = ''" :class="['p-3 rounded-xl border-2 transition-all text-left', selectedPeriod === period.key ? 'border-indigo-500 bg-indigo-50' : 'border-gray-200 hover:border-indigo-300']">
                  <div class="font-medium text-sm">{{ period.name }}</div>
                  <div class="text-xs text-gray-500">{{ formatPrice(selectedPlan!.prices[period.key]) }}</div>
                </button>
              </div>
            </div>

            <div class="mb-4">
              <label class="block text-sm font-medium text-gray-700 mb-2">优惠券</label>
              <div class="flex gap-2">
                <input v-model="couponCode" type="text" placeholder="输入优惠券码" class="flex-1 px-3 py-2 border border-gray-200 rounded-xl text-sm"/>
                <button @click="checkCoupon" :disabled="couponChecking || !couponCode.trim()" class="px-3 py-2 bg-gray-100 text-gray-700 rounded-xl text-sm hover:bg-gray-200 disabled:opacity-50">
                  {{ couponChecking ? '...' : '验证' }}
                </button>
              </div>
              <p v-if="couponError" class="text-red-500 text-xs mt-1">{{ couponError }}</p>
              <p v-if="couponDiscount > 0" class="text-green-500 text-xs mt-1">优惠 {{ formatPrice(couponDiscount) }}</p>
            </div>

            <div class="mb-4 p-3 bg-gray-50 rounded-xl text-sm">
              <div class="flex justify-between text-gray-600 mb-1">
                <span>套餐价格</span>
                <span>{{ formatPrice(selectedPlan!.prices[selectedPeriod] || 0) }}</span>
              </div>
              <div v-if="couponDiscount > 0" class="flex justify-between text-green-600 mb-1">
                <span>优惠</span>
                <span>-{{ formatPrice(couponDiscount) }}</span>
              </div>
              <div class="flex justify-between font-bold text-base border-t border-gray-200 pt-2 mt-2">
                <span>应付</span>
                <span class="text-indigo-600">{{ formatPrice(getFinalPrice()) }}</span>
              </div>
            </div>

            <div class="flex gap-3">
              <button @click="showModal = false" class="flex-1 py-2.5 bg-gray-100 text-gray-700 rounded-xl font-medium hover:bg-gray-200">取消</button>
              <button @click="createOrder" :disabled="ordering" class="flex-1 py-2.5 bg-indigo-500 text-white rounded-xl font-medium hover:bg-indigo-600 disabled:opacity-50">
                {{ ordering ? '创建中...' : '确认订购' }}
              </button>
            </div>
          </div>
        </div>
      </Transition>
    </Teleport>
  </div>
</template>

<style scoped>
.modal-enter-active, .modal-leave-active { transition: all 0.3s ease; }
.modal-enter-from, .modal-leave-to { opacity: 0; }
.modal-enter-from > div:last-child, .modal-leave-to > div:last-child { transform: translateY(100%); }
@media (min-width: 1024px) {
  .modal-enter-from > div:last-child, .modal-leave-to > div:last-child { transform: scale(0.95); }
}
</style>
