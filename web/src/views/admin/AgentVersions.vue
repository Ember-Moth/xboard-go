<template>
  <div class="p-6">
    <div class="mb-6 flex justify-between items-center">
      <h1 class="text-2xl font-bold">Agent 版本管理</h1>
      <button @click="showCreateDialog = true" class="btn btn-primary">
        添加版本
      </button>
    </div>

    <!-- 版本列表 -->
    <div class="bg-white rounded-lg shadow">
      <table class="min-w-full">
        <thead class="bg-gray-50">
          <tr>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">版本号</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">下载地址</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">文件大小</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">更新策略</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">创建时间</th>
            <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">操作</th>
          </tr>
        </thead>
        <tbody class="divide-y divide-gray-200">
          <tr v-for="version in versions" :key="version.id">
            <td class="px-6 py-4 whitespace-nowrap">
              <span class="font-mono">{{ version.version }}</span>
            </td>
            <td class="px-6 py-4">
              <a :href="version.download_url" target="_blank" class="text-blue-600 hover:underline truncate block max-w-xs">
                {{ version.download_url }}
              </a>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              {{ formatFileSize(version.file_size) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span :class="version.strategy === 'auto' ? 'badge-success' : 'badge-warning'">
                {{ version.strategy === 'auto' ? '自动' : '手动' }}
              </span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap">
              <span v-if="version.is_latest" class="badge-primary">最新版本</span>
              <span v-else class="badge-secondary">历史版本</span>
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
              {{ formatDate(version.created_at) }}
            </td>
            <td class="px-6 py-4 whitespace-nowrap text-sm">
              <button v-if="!version.is_latest" @click="setLatest(version.id)" class="text-blue-600 hover:text-blue-900 mr-3">
                设为最新
              </button>
              <button @click="editVersion(version)" class="text-indigo-600 hover:text-indigo-900 mr-3">
                编辑
              </button>
              <button v-if="!version.is_latest" @click="deleteVersion(version.id)" class="text-red-600 hover:text-red-900">
                删除
              </button>
            </td>
          </tr>
        </tbody>
      </table>
    </div>

    <!-- 更新日志 -->
    <div class="mt-8">
      <h2 class="text-xl font-bold mb-4">更新日志</h2>
      <div class="bg-white rounded-lg shadow">
        <table class="min-w-full">
          <thead class="bg-gray-50">
            <tr>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">主机</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">版本变更</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">状态</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">错误信息</th>
              <th class="px-6 py-3 text-left text-xs font-medium text-gray-500 uppercase">时间</th>
            </tr>
          </thead>
          <tbody class="divide-y divide-gray-200">
            <tr v-for="log in updateLogs" :key="log.id">
              <td class="px-6 py-4 whitespace-nowrap">主机 #{{ log.host_id }}</td>
              <td class="px-6 py-4 whitespace-nowrap font-mono">
                {{ log.from_version }} → {{ log.to_version }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap">
                <span v-if="log.status === 'success'" class="badge-success">成功</span>
                <span v-else-if="log.status === 'failed'" class="badge-danger">失败</span>
                <span v-else class="badge-warning">回滚</span>
              </td>
              <td class="px-6 py-4 text-sm text-gray-500">
                {{ log.error_message || '-' }}
              </td>
              <td class="px-6 py-4 whitespace-nowrap text-sm text-gray-500">
                {{ formatDate(log.created_at) }}
              </td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <!-- 创建/编辑对话框 -->
    <div v-if="showCreateDialog" class="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div class="bg-white rounded-lg p-6 w-full max-w-2xl">
        <h2 class="text-xl font-bold mb-4">{{ editingVersion ? '编辑版本' : '添加版本' }}</h2>
        <form @submit.prevent="saveVersion">
          <div class="space-y-4">
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">版本号</label>
              <input v-model="formData.version" type="text" required class="input" placeholder="v1.0.0">
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">下载地址</label>
              <input v-model="formData.download_url" type="url" required class="input">
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">SHA256 哈希</label>
              <input v-model="formData.sha256" type="text" required class="input" maxlength="64">
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">文件大小（字节）</label>
              <input v-model.number="formData.file_size" type="number" required class="input">
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">更新策略</label>
              <select v-model="formData.strategy" class="input">
                <option value="auto">自动更新</option>
                <option value="manual">手动更新</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium text-gray-700 mb-1">发布说明</label>
              <textarea v-model="formData.release_notes" rows="4" class="input"></textarea>
            </div>
          </div>
          <div class="mt-6 flex justify-end space-x-3">
            <button type="button" @click="closeDialog" class="btn btn-secondary">取消</button>
            <button type="submit" class="btn btn-primary">保存</button>
          </div>
        </form>
      </div>
    </div>
  </div>
</template>

<script setup>
import { ref, onMounted } from 'vue'
import { api } from '@/api'

const versions = ref([])
const updateLogs = ref([])
const showCreateDialog = ref(false)
const editingVersion = ref(null)
const formData = ref({
  version: '',
  download_url: '',
  sha256: '',
  file_size: 0,
  strategy: 'manual',
  release_notes: ''
})

const loadVersions = async () => {
  try {
    const res = await api.get('/admin/agent/versions')
    versions.value = res.data.items || []
  } catch (error) {
    console.error('Failed to load versions:', error)
  }
}

const loadUpdateLogs = async () => {
  try {
    const res = await api.get('/admin/agent/update_logs')
    updateLogs.value = res.data.items || []
  } catch (error) {
    console.error('Failed to load update logs:', error)
  }
}

const saveVersion = async () => {
  try {
    if (editingVersion.value) {
      await api.put(`/admin/agent/version/${editingVersion.value.id}`, formData.value)
    } else {
      await api.post('/admin/agent/version', formData.value)
    }
    closeDialog()
    loadVersions()
  } catch (error) {
    console.error('Failed to save version:', error)
    alert('保存失败：' + (error.response?.data?.error || error.message))
  }
}

const editVersion = (version) => {
  editingVersion.value = version
  formData.value = { ...version }
  showCreateDialog.value = true
}

const setLatest = async (id) => {
  if (!confirm('确定要将此版本设为最新版本吗？')) return
  try {
    await api.post(`/admin/agent/version/${id}/set_latest`)
    loadVersions()
  } catch (error) {
    console.error('Failed to set latest:', error)
    alert('操作失败：' + (error.response?.data?.error || error.message))
  }
}

const deleteVersion = async (id) => {
  if (!confirm('确定要删除此版本吗？')) return
  try {
    await api.delete(`/admin/agent/version/${id}`)
    loadVersions()
  } catch (error) {
    console.error('Failed to delete version:', error)
    alert('删除失败：' + (error.response?.data?.error || error.message))
  }
}

const closeDialog = () => {
  showCreateDialog.value = false
  editingVersion.value = null
  formData.value = {
    version: '',
    download_url: '',
    sha256: '',
    file_size: 0,
    strategy: 'manual',
    release_notes: ''
  }
}

const formatFileSize = (bytes) => {
  if (bytes < 1024) return bytes + ' B'
  if (bytes < 1024 * 1024) return (bytes / 1024).toFixed(2) + ' KB'
  return (bytes / 1024 / 1024).toFixed(2) + ' MB'
}

const formatDate = (date) => {
  return new Date(date).toLocaleString('zh-CN')
}

onMounted(() => {
  loadVersions()
  loadUpdateLogs()
})
</script>

<style scoped>
.input {
  @apply w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500;
}

.btn {
  @apply px-4 py-2 rounded-md font-medium transition-colors;
}

.btn-primary {
  @apply bg-blue-600 text-white hover:bg-blue-700;
}

.btn-secondary {
  @apply bg-gray-200 text-gray-700 hover:bg-gray-300;
}

.badge-primary {
  @apply inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-blue-100 text-blue-800;
}

.badge-secondary {
  @apply inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-gray-100 text-gray-800;
}

.badge-success {
  @apply inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-green-100 text-green-800;
}

.badge-warning {
  @apply inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-yellow-100 text-yellow-800;
}

.badge-danger {
  @apply inline-flex px-2 py-1 text-xs font-semibold rounded-full bg-red-100 text-red-800;
}
</style>
