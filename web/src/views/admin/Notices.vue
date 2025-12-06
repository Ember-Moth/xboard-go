<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

interface Notice {
  id: number
  title: string
  content: string
  img_url: string
  show: boolean
  sort: number
  created_at: number
}

const notices = ref<Notice[]>([])
const loading = ref(false)
const showModal = ref(false)
const editingNotice = ref<Partial<Notice> | null>(null)

const formatDate = (ts: number) => new Date(ts * 1000).toLocaleDateString()

const fetchNotices = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/notices')
    notices.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const openCreateModal = () => {
  editingNotice.value = {
    title: '',
    content: '',
    img_url: '',
    show: true,
    sort: 0,
  }
  showModal.value = true
}

const openEditModal = (notice: Notice) => {
  editingNotice.value = { ...notice }
  showModal.value = true
}

const saveNotice = async () => {
  if (!editingNotice.value) return
  try {
    if (editingNotice.value.id) {
      await api.put(`/api/v2/admin/notice/${editingNotice.value.id}`, editingNotice.value)
    } else {
      await api.post('/api/v2/admin/notice', editingNotice.value)
    }
    showModal.value = false
    fetchNotices()
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  }
}

const deleteNotice = async (notice: Notice) => {
  if (!confirm(`确定要删除公告 "${notice.title}" 吗？`)) return
  try {
    await api.delete(`/api/v2/admin/notice/${notice.id}`)
    fetchNotices()
  } catch (e: any) {
    alert(e.response?.data?.error || '删除失败')
  }
}

onMounted(fetchNotices)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">公告管理</h1>
        <p class="text-gray-500 mt-1">管理系统公告</p>
      </div>
      <button @click="openCreateModal" class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600 transition">
        添加公告
      </button>
    </div>

    <div class="bg-white rounded-xl shadow-sm overflow-hidden">
      <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

      <div v-else-if="notices.length === 0" class="text-center py-12 text-gray-500">暂无公告</div>

      <table v-else class="w-full">
        <thead class="bg-gray-50 border-b border-gray-200">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">标题</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">排序</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">创建时间</th>
            <th class="px-6 py-3 text-right text-xs font-medium text-gray-500 uppercase">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="notice in notices" :key="notice.id" class="hover:bg-gray-50">
            <td class="px-6 py-4 font-medium text-gray-900">{{ notice.title }}</td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ notice.sort }}</td>
            <td class="px-6 py-4">
              <span :class="['px-2 py-1 rounded-full text-xs', notice.show ? 'bg-green-100 text-green-600' : 'bg-gray-100 text-gray-600']">
                {{ notice.show ? '显示' : '隐藏' }}
              </span>
            </td>
            <td class="px-6 py-4 text-sm text-gray-500">{{ formatDate(notice.created_at) }}</td>
            <td class="px-6 py-4 text-right space-x-2">
              <button @click="openEditModal(notice)" class="text-primary-600 hover:text-primary-700 text-sm">编辑</button>
              <button @click="deleteNotice(notice)" class="text-red-600 hover:text-red-700 text-sm">删除</button>
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
          <h3 class="text-lg font-bold mb-4">{{ editingNotice?.id ? '编辑公告' : '添加公告' }}</h3>
          
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">标题</label>
              <input v-model="editingNotice!.title" type="text" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">内容</label>
              <textarea v-model="editingNotice!.content" rows="5" class="w-full px-4 py-2 border border-gray-200 rounded-xl"></textarea>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">图片 URL</label>
              <input v-model="editingNotice!.img_url" type="text" placeholder="可选" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
            </div>
            <div class="grid grid-cols-2 gap-4">
              <div>
                <label class="block text-sm font-medium text-gray-700 mb-1">排序</label>
                <input v-model.number="editingNotice!.sort" type="number" class="w-full px-4 py-2 border border-gray-200 rounded-xl" />
              </div>
              <div class="flex items-center">
                <label class="flex items-center gap-2 mt-6">
                  <input v-model="editingNotice!.show" type="checkbox" class="rounded" />
                  <span class="text-sm">显示</span>
                </label>
              </div>
            </div>
          </div>

          <div class="flex gap-3 mt-6">
            <button @click="showModal = false" class="flex-1 px-4 py-2 border border-gray-200 text-gray-600 rounded-xl hover:bg-gray-50">取消</button>
            <button @click="saveNotice" class="flex-1 px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600">保存</button>
          </div>
        </div>
      </div>
    </Teleport>
  </div>
</template>
