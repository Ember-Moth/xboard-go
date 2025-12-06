<script setup lang="ts">
import { ref, onMounted, watch } from 'vue'
import api from '@/api'

interface User {
  id: number
  email: string
  balance: number
  plan_id: number | null
  transfer_enable: number
  u: number
  d: number
  expired_at: number | null
  banned: boolean
  is_admin: boolean
  is_staff: boolean
  created_at: number
}

const users = ref<User[]>([])
const loading = ref(false)
const total = ref(0)
const page = ref(1)
const pageSize = 20
const search = ref('')
const showModal = ref(false)
const editingUser = ref<User | null>(null)
const editForm = ref({
  email: '',
  balance: 0,
  transfer_enable: 0,
  expired_at: '',
  banned: false,
  is_admin: false,
  password: '',
})

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const formatDate = (ts: number | null) => {
  if (!ts) return '永久'
  return new Date(ts * 1000).toLocaleDateString()
}

const fetchUsers = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/users', {
      params: { page: page.value, page_size: pageSize, search: search.value }
    })
    users.value = res.data.data || []
    total.value = res.data.total || 0
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const openEditModal = (user: User) => {
  editingUser.value = user
  editForm.value = {
    email: user.email,
    balance: user.balance,
    transfer_enable: user.transfer_enable,
    expired_at: user.expired_at ? new Date(user.expired_at * 1000).toISOString().split('T')[0] : '',
    banned: user.banned,
    is_admin: user.is_admin,
    password: '',
  }
  showModal.value = true
}

const saveUser = async () => {
  if (!editingUser.value) return
  try {
    const data: any = {
      email: editForm.value.email,
      balance: editForm.value.balance,
      transfer_enable: editForm.value.transfer_enable,
      banned: editForm.value.banned,
      is_admin: editForm.value.is_admin,
    }
    if (editForm.value.expired_at) {
      data.expired_at = Math.floor(new Date(editForm.value.expired_at).getTime() / 1000)
    }
    if (editForm.value.password) {
      data.password = editForm.value.password
    }
    await api.put(`/api/v2/admin/user/${editingUser.value.id}`, data)
    showModal.value = false
    fetchUsers()
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  }
}

const resetTraffic = async (user: User) => {
  if (!confirm(`确定要重置用户 "${user.email}" 的流量吗？`)) return
  try {
    await api.post(`/api/v2/admin/user/${user.id}/reset_traffic`)
    fetchUsers()
  } catch (e: any) {
    alert(e.response?.data?.error || '重置失败')
  }
}

const deleteUser = async (user: User) => {
  if (!confirm(`确定要删除用户 "${user.email}" 吗？此操作不可恢复！`)) return
  try {
    await api.delete(`/api/v2/admin/user/${user.id}`)
    fetchUsers()
  } catch (e: any) {
    alert(e.response?.data?.error || '删除失败')
  }
}

watch(search, () => {
  page.value = 1
  fetchUsers()
})

onMounted(fetchUsers)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">用户管理</h1>
        <p class="text-gray-500 mt-1">管理系统用户</p>
      </div>
      <div class="flex items-center gap-4">
        <input 
          v-model="search" 
          type="text" 
          placeholder="搜索邮箱..." 
          class="px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
        />
      </div>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

      <div v-else class="overflow-x-auto">
        <table class="w-full">
          <thead class="bg-gray-50 border-b border-gray-200">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">用户</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">余额</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">流量</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">到期</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="user in users" :key="user.id" class="hover:bg-gray-50">
              <td class="px-6 py-4">
                <div class="flex items-center gap-3">
                  <div class="w-8 h-8 rounded-full bg-gradient-to-br from-primary-400 to-primary-600 flex items-center justify-center text-white text-sm font-medium">
                    {{ user.email.charAt(0).toUpperCase() }}
                  </div>
                  <div>
                    <div class="font-medium text-gray-900">{{ user.email }}</div>
                    <div class="text-xs text-gray-500">ID: {{ user.id }}</div>
                  </div>
                </div>
              </td>
              <td class="px-6 py-4 text-sm">¥{{ (user.balance / 100).toFixed(2) }}</td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ formatBytes(user.u + user.d) }} / {{ formatBytes(user.transfer_enable) }}
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">{{ formatDate(user.expired_at) }}</td>
              <td class="px-6 py-4">
                <span v-if="user.banned" class="px-2 py-1 bg-red-100 text-red-600 rounded-full text-xs">封禁</span>
                <span v-else-if="user.is_admin" class="px-2 py-1 bg-purple-100 text-purple-600 rounded-full text-xs">管理员</span>
                <span v-else class="px-2 py-1 bg-green-100 text-green-600 rounded-full text-xs">正常</span>
              </td>
              <td class="px-6 py-4 text-right space-x-2">
                <button @click="openEditModal(user)" class="text-primary-600 hover:text-primary-700 text-sm">编辑</button>
                <button @click="resetTraffic(user)" class="text-yellow-600 hover:text-yellow-700 text-sm">重置流量</button>
                <button @click="deleteUser(user)" class="text-red-600 hover:text-red-700 text-sm">删除</button>
              </td>
            </tr>
          </tbody>
        </table>
      </div>

      <!-- Pagination -->
      <div v-if="total > pageSize" class="flex items-center justify-between px-6 py-4 border-t border-gray-200">
        <span class="text-sm text-gray-500">共 {{ total }} 条</span>
        <div class="flex gap-2">
          <button @click="page--; fetchUsers()" :disabled="page <= 1" class="px-3 py-1 border rounded text-sm disabled:opacity-50">上一页</button>
          <button @click="page++; fetchUsers()" :disabled="page * pageSize >= total" class="px-3 py-1 border rounded text-sm disabled:opacity-50">下一页</button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4">
        <div class="absolute inset-0 bg-black/30" @click="showModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-lg p-6">
          <h3 class="text-lg font-bold mb-4">编辑用户</h3>
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">邮箱</label>
              <input v-model="editForm.email" type="email" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">余额 (分)</label>
                <input v-model.number="editForm.balance" type="number" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">流量 (字节)</label>
                <input v-model.number="editForm.transfer_enable" type="number" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">到期时间</label>
              <input v-model="editForm.expired_at" type="date" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">新密码 (留空不修改)</label>
              <input v-model="editForm.password" type="password" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
            </div>
            <div class="flex items-center gap-4">
              <label class="flex items-center gap-2">
                <input v-model="editForm.banned" type="checkbox" class="rounded" />
                <span class="text-sm">封禁</span>
              </label>
              <label class="flex items-center gap-2">
                <input v-model="editForm.is_admin" type="checkbox" class="rounded" />
                <span class="text-sm">管理员</span>
              </label>
            </div>
          </div>
          <div class="flex gap-3 mt-6">
            <button @click="showModal = false" class="flex-1 px-4 py-2 border border-gray-200 text-gray-600 rounded-xl hover:bg-gray-50">取消</button>
            <button @click="saveUser" class="flex-1 px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
