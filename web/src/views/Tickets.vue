<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'
import dayjs from 'dayjs'

interface TicketMessage {
  id: number
  message: string
  is_me: boolean
  created_at: number
}

interface Ticket {
  id: number
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
const showCreateModal = ref(false)
const showDetailModal = ref(false)
const creating = ref(false)
const replying = ref(false)
const closing = ref(false)
const currentTicket = ref<Ticket | null>(null)
const replyMessage = ref('')

const newTicket = ref({
  subject: '',
  message: '',
  level: 1,
})

const statusMap: Record<number, { text: string; class: string }> = {
  0: { text: '开启', class: 'badge-success' },
  1: { text: '已关闭', class: 'badge-danger' },
}

const replyStatusMap: Record<number, { text: string; class: string }> = {
  0: { text: '待回复', class: 'badge-warning' },
  1: { text: '已回复', class: 'badge-info' },
}

const levelMap: Record<number, string> = {
  0: '低',
  1: '中',
  2: '高',
}

const formatDate = (ts: number) => dayjs.unix(ts).format('YYYY-MM-DD HH:mm')

const createTicket = async () => {
  if (!newTicket.value.subject || !newTicket.value.message) {
    alert('请填写主题和内容')
    return
  }

  creating.value = true
  try {
    await api.post('/api/v1/user/ticket/create', newTicket.value)
    showCreateModal.value = false
    newTicket.value = { subject: '', message: '', level: 1 }
    fetchTickets()
  } catch (e: any) {
    alert(e.response?.data?.error || '创建失败')
  } finally {
    creating.value = false
  }
}

const fetchTickets = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v1/user/tickets')
    tickets.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const openTicketDetail = async (ticket: Ticket) => {
  try {
    const res = await api.get(`/api/v1/user/ticket/${ticket.id}`)
    currentTicket.value = res.data.data
    showDetailModal.value = true
    replyMessage.value = ''
  } catch (e: any) {
    alert(e.response?.data?.error || '获取工单详情失败')
  }
}

const replyTicket = async () => {
  if (!currentTicket.value || !replyMessage.value.trim()) return
  
  replying.value = true
  try {
    await api.post(`/api/v1/user/ticket/${currentTicket.value.id}/reply`, {
      message: replyMessage.value
    })
    replyMessage.value = ''
    // 刷新详情
    const res = await api.get(`/api/v1/user/ticket/${currentTicket.value.id}`)
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
  
  closing.value = true
  try {
    await api.post(`/api/v1/user/ticket/${currentTicket.value.id}/close`)
    showDetailModal.value = false
    fetchTickets()
  } catch (e: any) {
    alert(e.response?.data?.error || '关闭失败')
  } finally {
    closing.value = false
  }
}

onMounted(fetchTickets)
</script>

<template>
  <div class="space-y-6 animate-fade-in">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">工单系统</h1>
        <p class="text-gray-500 mt-1">有问题？提交工单获取帮助</p>
      </div>
      <button @click="showCreateModal = true" class="btn btn-primary">
        新建工单
      </button>
    </div>

    <div class="card">
      <div v-if="loading" class="text-center py-12 text-gray-500">
        加载中...
      </div>

      <div v-else-if="tickets.length === 0" class="text-center py-12">
        <span class="text-5xl mb-4 block">💬</span>
        <p class="text-gray-500">暂无工单</p>
        <button @click="showCreateModal = true" class="btn btn-primary mt-4">
          新建工单
        </button>
      </div>

      <div v-else class="space-y-4">
        <div
          v-for="ticket in tickets"
          :key="ticket.id"
          @click="openTicketDetail(ticket)"
          class="p-4 rounded-xl bg-surface-50 hover:bg-surface-100 transition-colors cursor-pointer"
        >
          <div class="flex items-start justify-between">
            <div class="space-y-2">
              <h3 class="font-medium text-gray-900">{{ ticket.subject }}</h3>
              <div class="flex items-center gap-2">
                <span :class="['badge', statusMap[ticket.status]?.class]">
                  {{ statusMap[ticket.status]?.text }}
                </span>
                <span :class="['badge', replyStatusMap[ticket.reply_status]?.class]">
                  {{ replyStatusMap[ticket.reply_status]?.text }}
                </span>
                <span class="text-sm text-gray-500">优先级: {{ levelMap[ticket.level] }}</span>
              </div>
            </div>
            <div class="text-right text-sm text-gray-500">
              <p>{{ formatDate(ticket.updated_at) }}</p>
              <p class="text-primary-600 mt-1">点击查看详情 →</p>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- Create Modal -->
    <Teleport to="body">
      <div v-if="showCreateModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showCreateModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg p-6 animate-scale-in">
          <h3 class="text-xl font-bold mb-4">新建工单</h3>
          
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">主题</label>
              <input
                v-model="newTicket.subject"
                type="text"
                placeholder="简要描述您的问题"
                class="input"
              />
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">优先级</label>
              <select v-model="newTicket.level" class="input">
                <option :value="0">低</option>
                <option :value="1">中</option>
                <option :value="2">高</option>
              </select>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">内容</label>
              <textarea
                v-model="newTicket.message"
                rows="5"
                placeholder="详细描述您遇到的问题..."
                class="input resize-none"
              ></textarea>
            </div>
          </div>

          <div class="flex gap-3 mt-6">
            <button @click="showCreateModal = false" class="flex-1 btn btn-secondary">
              取消
            </button>
            <button @click="createTicket" :disabled="creating" class="flex-1 btn btn-primary">
              {{ creating ? '提交中...' : '提交' }}
            </button>
          </div>
        </div>
      </div>
    </Teleport>

    <!-- Detail Modal -->
    <Teleport to="body">
      <div v-if="showDetailModal && currentTicket" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30 backdrop-blur-sm" @click="showDetailModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-2xl p-6 animate-scale-in max-h-[90vh] overflow-hidden flex flex-col">
          <div class="flex items-center justify-between mb-4">
            <h3 class="text-xl font-bold">{{ currentTicket.subject }}</h3>
            <div class="flex items-center gap-2">
              <span :class="['badge', statusMap[currentTicket.status]?.class]">
                {{ statusMap[currentTicket.status]?.text }}
              </span>
              <button @click="showDetailModal = false" class="text-gray-400 hover:text-gray-600">✕</button>
            </div>
          </div>

          <!-- Messages -->
          <div class="flex-1 overflow-y-auto space-y-4 mb-4 min-h-[200px] max-h-[400px]">
            <div
              v-for="msg in currentTicket.messages"
              :key="msg.id"
              :class="['p-4 rounded-xl', msg.is_me ? 'bg-primary-50 ml-8' : 'bg-gray-100 mr-8']"
            >
              <div class="flex items-center justify-between mb-2">
                <span class="text-sm font-medium" :class="msg.is_me ? 'text-primary-600' : 'text-gray-600'">
                  {{ msg.is_me ? '我' : '客服' }}
                </span>
                <span class="text-xs text-gray-400">{{ formatDate(msg.created_at) }}</span>
              </div>
              <p class="text-gray-800 whitespace-pre-wrap">{{ msg.message }}</p>
            </div>
          </div>

          <!-- Reply Form -->
          <div v-if="currentTicket.status === 0" class="border-t pt-4">
            <textarea
              v-model="replyMessage"
              rows="3"
              placeholder="输入回复内容..."
              class="input resize-none mb-3"
            ></textarea>
            <div class="flex gap-3">
              <button @click="closeTicket" :disabled="closing" class="btn btn-secondary text-red-500">
                {{ closing ? '关闭中...' : '关闭工单' }}
              </button>
              <button @click="replyTicket" :disabled="replying || !replyMessage.trim()" class="flex-1 btn btn-primary">
                {{ replying ? '发送中...' : '发送回复' }}
              </button>
            </div>
          </div>
          <div v-else class="border-t pt-4 text-center text-gray-500">
            此工单已关闭
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
