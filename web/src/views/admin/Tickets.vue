<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

interface TicketMessage {
  id: number
  user_id: number
  message: string
  is_admin: boolean
  created_at: number
}

interface Ticket {
  id: number
  user_id: number
  user_email: string
  subject: string
  level: number
  status: number
  reply_status: number
  created_at: number
  updated_at: number
  messages?: TicketMessage[]
}

const tickets = ref<Ticket[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 20
const statusFilter = ref<string>('')

const showDetail = ref(false)
const currentTicket = ref<Ticket | null>(null)
const replyMessage = ref('')
const replying = ref(false)

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '开启', class: 'bg-green-100 text-green-600' },
  1: { text: '已关闭', class: 'bg-gray-100 text-gray-600' },
}

const replyStatusMap: Record<number, { text: string; class: string }> = {
  0: { text: '待回复', class: 'bg-yellow-100 text-yellow-600' },
  1: { text: '已回复', class: 'bg-blue-100 text-blue-600' },
}

const levelMap: Record<number, { text: string; class: string }> = {
  0: { text: '低', class: 'text-gray-500' },
  1: { text: '中', class: 'text-yellow-600' },
  2: { text: '高', class: 'text-red-600' },
}

const formatDate = (ts: number) => new Date(ts * 1000).toLocaleString()

const fetchTickets = async () => {
  loading.value = true
  try {
    const params: any = { page: page.value, page_size: pageSize }
    if (statusFilter.value !== '') {
      params.status = statusFilter.value
    }
    const res = await api.get('/api/v2/admin/tickets', { params })
    tickets.value = res.data.data || []
    total.value = res.data.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const openDetail = async (ticket: Ticket) => {
  try {
    const res = await api.get(`/api/v2/admin/ticket/${ticket.id}`)
    currentTicket.value = res.data.data
    showDetail.value = true
    replyMessage.value = ''
  } catch (e: any) {
    alert(e.response?.data?.error || '获取详情失败')
  }
}

const sendReply = async () => {
  if (!currentTicket.value || !replyMessage.value.trim()) return
  replying.value = true
  try {
    await api.post(`/api/v2/admin/ticket/${currentTicket.value.id}/reply`, {
      message: replyMessage.value
    })
    replyMessage.value = ''
    // 刷新详情
    const res = await api.get(`/api/v2/admin/ticket/${currentTicket.value.id}`)
    currentTicket.value = res.data.data
    fetchTickets()
  } catch (e: any) {
    alert(e.response?.data?.error || '回复失败')
  } finally {
    replying.value = false
  }
}

const closeTicket = async () => {
  if (!currentTicket.value) return
  if (!confirm('确定要关闭此工单吗？')) return
  try {
    await api.post(`/api/v2/admin/ticket/${currentTicket.value.id}/close`)
    showDetail.value = false
    fetchTickets()
  } catch (e: any) {
    alert(e.response?.data?.error || '关闭失败')
  }
}

onMounted(fetchTickets)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">工单管理</h1>
        <p class="text-gray-500 mt-1">处理用户工单</p>
      </div>
      <div class="flex items-center gap-4">
        <select v-model="statusFilter" @change="page = 1; fetchTickets()" class="px-4 py-2 border border-gray-200 rounded-xl">
          <option value="">全部状态</option>
          <option value="0">开启</option>
          <option value="1">已关闭</option>
        </select>
      </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

      <div v-else-if="tickets.length === 0" class="text-center py-12 text-gray-500">暂无工单</div>

      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50 border-b border-gray-200">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">主题</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">优先级</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">更新时间</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="ticket in tickets" :key="ticket.id" class="hover:bg-gray-50">
              <td class="px-6 py-4">
                <div class="font-medium text-gray-900">{{ ticket.subject }}</div>
                <div class="text-xs text-gray-500">#{{ ticket.id }}</div>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ ticket.user_email || `用户 ${ticket.user_id}` }}
              </td>
              <td class="px-6 py-4">
                <span :class="['text-sm font-medium', levelMap[ticket.level]?.class]">
                  {{ levelMap[ticket.level]?.text }}
                </span>
              </td>
              <td class="px-6 py-4">
                <div class="flex items-center gap-2">
                  <span :class="['px-2 py-1 rounded-full text-xs', statusMap[ticket.status]?.class]">
                    {{ statusMap[ticket.status]?.text }}
                  </span>
                  <span :class="['px-2 py-1 rounded-full text-xs', replyStatusMap[ticket.reply_status]?.class]">
                    {{ replyStatusMap[ticket.reply_status]?.text }}
                  </span>
                </div>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">{{ formatDate(ticket.updated_at) }}</td>
              <td class="px-6 py-4 text-right">
                <button @click="openDetail(ticket)" class="text-primary-600 hover:text-primary-700 text-sm">查看</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="total > pageSize" class="flex items-center justify-between px-6 py-4 border-t border-gray-200">
        <span class="text-sm text-gray-500">共 {{ total }} 条</span>
        <div class="flex gap-2">
          <button @click="page--; fetchTickets()" :disabled="page <= 1" class="px-3 py-1 border rounded text-sm disabled:opacity-50">上一页</button>
          <button @click="page++; fetchTickets()" :disabled="page * pageSize >= total" class="px-3 py-1 border rounded text-sm disabled:opacity-50">下一页</button>
        </div>
      </div>
    </div>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="showDetail && currentTicket" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showDetail = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl max-h-[90vh] flex flex-col">
          <div class="p-6 border-b border-gray-200">
            <div class="flex items-center justify-between">
              <div>
                <h3 class="text-lg font-bold">{{ currentTicket.subject }}</h3>
                <p class="text-sm text-gray-500">{{ currentTicket.user_email }} · #{{ currentTicket.id }}</p>
              </div>
              <div class="flex items-center gap-2">
                <span :class="['px-2 py-1 rounded-full text-xs', statusMap[currentTicket.status]?.class]">
                  {{ statusMap[currentTicket.status]?.text }}
                </span>
                <button v-if="currentTicket.status === 0" @click="closeTicket" class="text-red-600 hover:text-red-700 text-sm">关闭工单</button>
              </div>
            </div>
          </div>

          <div class="flex-1 overflow-y-auto p-6 space-y-4">
            <div v-for="msg in currentTicket.messages" :key="msg.id" :class="['p-4 rounded-xl', msg.is_admin ? 'bg-primary-50 ml-8' : 'bg-gray-50 mr-8']">
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium" :class="msg.is_admin ? 'text-primary-600' : 'text-gray-600'">
                  {{ msg.is_admin ? '管理员' : '用户' }}
                </span>
                <span class="text-xs text-gray-400">{{ formatDate(msg.created_at) }}</span>
              </div>
              <p class="text-sm text-gray-700 whitespace-pre-wrap">{{ msg.message }}</p>
            </div>
          </div>

          <div v-if="currentTicket.status === 0" class="p-6 border-t border-gray-200">
            <div class="flex gap-3">
              <textarea v-model="replyMessage" rows="2" placeholder="输入回复内容..." class="flex-1 px-4 py-2 border border-gray-200 rounded-xl resize-none"></textarea>
              <button @click="sendReply" :disabled="replying || !replyMessage.trim()" class="px-6 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600 disabled:opacity-50">
                {{ replying ? '发送中...' : '发送' }}
              </button>
            </div>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
