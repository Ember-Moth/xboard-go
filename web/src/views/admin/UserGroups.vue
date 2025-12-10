<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'

interface UserGroup {
  id: number
  name: string
  description: string
  server_ids: number[]
  plan_ids: number[]
  servers: Array<{ id: number; name: string; type: string }>
  plans: Array<{ id: number; name: string }>
  default_transfer_enable: number
  default_speed_limit: number | null
  default_device_limit: number | null
  sort: number
}

interface Server {
  id: number
  name: string
  type: string
  host: string
}

interface Plan {
  id: number
  name: string
  transfer_enable: number
}

const groups = ref<UserGroup[]>([])
const servers = ref<Server[]>([])
const plans = ref<Plan[]>([])
const loading = ref(false)
const showModal = ref(false)
const editingGroup = ref<UserGroup | null>(null)

const editForm = ref({
  name: '',
  description: '',
  server_ids: [] as number[],
  plan_ids: [] as number[],
  default_transfer_enable: 0,
  default_speed_limit: null as number | null,
  default_device_limit: null as number | null,
  sort: 0,
})

const formatBytes = (bytes: number) => {
  if (bytes === 0) return '0 B'
  const k = 1024
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(k))
  return parseFloat((bytes / Math.pow(k, i)).toFixed(2)) + ' ' + sizes[i]
}

const fetchGroups = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/user-groups')
    groups.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const fetchServers = async () => {
  try {
    const res = await api.get('/api/v2/admin/servers')
    servers.value = res.data.data || []
  } catch (e) {
    console.error(e)
  }
}

const fetchPlans = async () => {
  try {
    const res = await api.get('/api/v2/admin/plans')
    plans.value = res.data.data || []
  } catch (e) {
    console.error(e)
  }
}

const openCreateModal = () => {
  editingGroup.value = null
  editForm.value = {
    name: '',
    description: '',
    server_ids: [],
    plan_ids: [],
    default_transfer_enable: 107374182400, // 100GB
    default_speed_limit: null,
    default_device_limit: null,
    sort: 0,
  }
  showModal.value = true
}

const openEditModal = (group: UserGroup) => {
  editingGroup.value = group
  editForm.value = {
    name: group.name,
    description: group.description,
    server_ids: [...group.server_ids],
    plan_ids: [...group.plan_ids],
    default_transfer_enable: group.default_transfer_enable,
    default_speed_limit: group.default_speed_limit,
    default_device_limit: group.default_device_limit,
    sort: group.sort,
  }
  showModal.value = true
}

const saveGroup = async () => {
  try {
    if (editingGroup.value) {
      await api.put(`/api/v2/admin/user-group/${editingGroup.value.id}`, editForm.value)
    } else {
      await api.post('/api/v2/admin/user-group', editForm.value)
    }
    showModal.value = false
    fetchGroups()
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  }
}

const deleteGroup = async (group: UserGroup) => {
  if (!confirm(`确定要删除用户组 "${group.name}" 吗？`)) return
  try {
    await api.delete(`/api/v2/admin/user-group/${group.id}`)
    fetchGroups()
  } catch (e: any) {
    alert(e.response?.data?.error || '删除失败')
  }
}

const toggleServer = (serverId: number) => {
  const index = editForm.value.server_ids.indexOf(serverId)
  if (index > -1) {
    editForm.value.server_ids.splice(index, 1)
  } else {
    editForm.value.server_ids.push(serverId)
  }
}

const togglePlan = (planId: number) => {
  const index = editForm.value.plan_ids.indexOf(planId)
  if (index > -1) {
    editForm.value.plan_ids.splice(index, 1)
  } else {
    editForm.value.plan_ids.push(planId)
  }
}

onMounted(() => {
  fetchGroups()
  fetchServers()
  fetchPlans()
})
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">用户组管理</h1>
        <p class="text-gray-500 mt-1">管理用户组，控制节点和套餐访问权限</p>
      </div>
      <button @click="openCreateModal" class="px-4 py-2 bg-indigo-500 text-white rounded-xl hover:bg-indigo-600 transition-colors flex items-center gap-2">
        <svg class="w-5 h-5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 4v16m8-8H4"/>
        </svg>
        创建用户组
      </button>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

    <div v-else class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
      <div v-for="group in groups" :key="group.id" class="bg-white rounded-2xl p-6 shadow-sm hover:shadow-md transition-shadow">
        <div class="flex items-start justify-between mb-4">
          <div class="flex-1">
            <h3 class="text-lg font-semibold text-gray-900">{{ group.name }}</h3>
            <p class="text-sm text-gray-500 mt-1">{{ group.description || '暂无描述' }}</p>
          </div>
          <span class="px-2 py-1 bg-gray-100 text-gray-600 rounded-lg text-xs">排序: {{ group.sort }}</span>
        </div>

        <div class="space-y-3 mb-4">
          <div class="flex items-center gap-2 text-sm">
            <svg class="w-4 h-4 text-cyan-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 12h14M5 12a2 2 0 01-2-2V6a2 2 0 012-2h14a2 2 0 012 2v4a2 2 0 01-2 2M5 12a2 2 0 00-2 2v4a2 2 0 002 2h14a2 2 0 002-2v-4a2 2 0 00-2-2m-2-4h.01M17 16h.01"/>
            </svg>
            <span class="text-gray-600">{{ group.servers?.length || 0 }} 个节点</span>
          </div>
          <div class="flex items-center gap-2 text-sm">
            <svg class="w-4 h-4 text-indigo-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 3v4M3 5h4M6 17v4m-2-2h4m5-16l2.286 6.857L21 12l-5.714 2.143L13 21l-2.286-6.857L5 12l5.714-2.143L13 3z"/>
            </svg>
            <span class="text-gray-600">{{ group.plans?.length || 0 }} 个套餐</span>
          </div>
          <div class="flex items-center gap-2 text-sm">
            <svg class="w-4 h-4 text-green-500" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M7 16a4 4 0 01-.88-7.903A5 5 0 1115.9 6L16 6a5 5 0 011 9.9M15 13l-3-3m0 0l-3 3m3-3v12"/>
            </svg>
            <span class="text-gray-600">默认流量: {{ formatBytes(group.default_transfer_enable) }}</span>
          </div>
        </div>

        <div class="flex gap-2">
          <button @click="openEditModal(group)" class="flex-1 px-3 py-2 bg-gray-100 text-gray-700 rounded-xl hover:bg-gray-200 transition-colors text-sm">
            编辑
          </button>
          <button @click="deleteGroup(group)" class="px-3 py-2 bg-red-50 text-red-600 rounded-xl hover:bg-red-100 transition-colors text-sm">
            删除
          </button>
        </div>
      </div>
    </div>

    <!-- Edit Modal -->
    <Teleport to="body">
      <div v-if="showModal" class="fixed inset-0 z-50 flex items-center justify-center p-4 overflow-y-auto">
        <div class="absolute inset-0 bg-black/30" @click="showModal = false"></div>
        <div class="relative bg-white rounded-2xl shadow-xl w-full max-w-4xl p-6 my-8">
          <h3 class="text-xl font-bold mb-6">{{ editingGroup ? '编辑用户组' : '创建用户组' }}</h3>
          
          <div class="space-y-6">
            <!-- 基本信息 -->
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">组名称 *</label>
                <input v-model="editForm.name" type="text" placeholder="如：VIP用户" class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">排序</label>
                <input v-model.number="editForm.sort" type="number" class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
              </div>
            </div>

            <div>
              <label class="block text-sm font-medium text-gray-700 mb-2">描述</label>
              <textarea v-model="editForm.description" rows="2" placeholder="用户组描述..." class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent"></textarea>
            </div>

            <!-- 默认配置 -->
            <div class="grid grid-cols-3 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">默认流量 (GB)</label>
                <input v-model.number="editForm.default_transfer_enable" type="number" step="1073741824" class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
                <p class="text-xs text-gray-500 mt-1">{{ formatBytes(editForm.default_transfer_enable) }}</p>
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">速度限制 (Mbps)</label>
                <input v-model.number="editForm.default_speed_limit" type="number" placeholder="不限制" class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
              </div>
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-2">设备数限制</label>
                <input v-model.number="editForm.default_device_limit" type="number" placeholder="不限制" class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-indigo-500 focus:border-transparent" />
              </div>
            </div>

            <!-- 节点选择 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-3">可访问的节点</label>
              <div class="grid grid-cols-2 gap-2 max-h-48 overflow-y-auto p-3 bg-gray-50 rounded-xl">
                <label v-for="server in servers" :key="server.id" class="flex items-center gap-2 p-2 hover:bg-white rounded-lg cursor-pointer transition-colors">
                  <input type="checkbox" :checked="editForm.server_ids.includes(server.id)" @change="toggleServer(server.id)" class="rounded text-indigo-500" />
                  <span class="text-sm">{{ server.name }}</span>
                  <span class="text-xs text-gray-500">({{ server.type }})</span>
                </label>
              </div>
              <p class="text-xs text-gray-500 mt-2">已选择 {{ editForm.server_ids.length }} 个节点</p>
            </div>

            <!-- 套餐选择 -->
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-3">可购买的套餐</label>
              <div class="grid grid-cols-2 gap-2 max-h-48 overflow-y-auto p-3 bg-gray-50 rounded-xl">
                <label v-for="plan in plans" :key="plan.id" class="flex items-center gap-2 p-2 hover:bg-white rounded-lg cursor-pointer transition-colors">
                  <input type="checkbox" :checked="editForm.plan_ids.includes(plan.id)" @change="togglePlan(plan.id)" class="rounded text-indigo-500" />
                  <span class="text-sm">{{ plan.name }}</span>
                </label>
              </div>
              <p class="text-xs text-gray-500 mt-2">已选择 {{ editForm.plan_ids.length }} 个套餐</p>
            </div>
          </div>

          <div class="flex gap-3 mt-6">
            <button @click="showModal = false" class="flex-1 px-4 py-2.5 border border-gray-200 text-gray-600 rounded-xl hover:bg-gray-50 transition-colors">取消</button>
            <button @click="saveGroup" class="flex-1 px-4 py-2.5 bg-indigo-500 text-white rounded-xl hover:bg-indigo-600 transition-colors">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
