<script setup lang="ts">
import { ref, onMounted } from 'vue'
import api from '@/api'

const settings = ref<Record<string, string>>({})
const loading = ref(false)
const saving = ref(false)
const activeTab = ref('basic')

const tabs = [
  { key: 'basic', name: '基础设置', icon: '⚙️' },
  { key: 'mail', name: '邮件设置', icon: '📧' },
  { key: 'telegram', name: 'Telegram', icon: '📱' },
  { key: 'subscribe', name: '订阅设置', icon: '🔗' },
  { key: 'register', name: '注册设置', icon: '📝' },
]

const settingGroups: Record<string, Array<{ key: string; label: string; type: string; placeholder?: string }>> = {
  basic: [
    { key: 'app_name', label: '站点名称', type: 'text', placeholder: 'XBoard' },
    { key: 'app_url', label: '站点地址', type: 'text', placeholder: 'https://example.com' },
    { key: 'app_description', label: '站点描述', type: 'textarea' },
    { key: 'currency', label: '货币单位', type: 'text', placeholder: 'CNY' },
    { key: 'currency_symbol', label: '货币符号', type: 'text', placeholder: '¥' },
  ],
  mail: [
    { key: 'mail_host', label: 'SMTP 服务器', type: 'text', placeholder: 'smtp.example.com' },
    { key: 'mail_port', label: 'SMTP 端口', type: 'text', placeholder: '587' },
    { key: 'mail_username', label: 'SMTP 用户名', type: 'text' },
    { key: 'mail_password', label: 'SMTP 密码', type: 'password' },
    { key: 'mail_encryption', label: '加密方式', type: 'select', placeholder: 'tls' },
    { key: 'mail_from_address', label: '发件人地址', type: 'text' },
    { key: 'mail_from_name', label: '发件人名称', type: 'text' },
  ],
  telegram: [
    { key: 'telegram_bot_enable', label: '启用 Telegram Bot', type: 'checkbox' },
    { key: 'telegram_bot_token', label: 'Bot Token', type: 'password' },
    { key: 'telegram_discuss_link', label: '讨论群链接', type: 'text' },
  ],
  subscribe: [
    { key: 'subscribe_url', label: '订阅地址', type: 'text', placeholder: '留空使用站点地址' },
    { key: 'subscribe_path', label: '订阅路径', type: 'text', placeholder: '/api/v1/client/subscribe' },
    { key: 'subscribe_single_mode', label: '单节点模式', type: 'checkbox' },
  ],
  register: [
    { key: 'register_enable', label: '开放注册', type: 'checkbox' },
    { key: 'email_verify', label: '邮箱验证', type: 'checkbox' },
    { key: 'invite_force', label: '强制邀请', type: 'checkbox' },
    { key: 'invite_commission', label: '邀请佣金比例 (%)', type: 'number' },
    { key: 'try_out_plan_id', label: '试用套餐 ID', type: 'number' },
    { key: 'try_out_hour', label: '试用时长 (小时)', type: 'number' },
  ],
}

const fetchSettings = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v2/admin/settings')
    settings.value = res.data.data || {}
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const saveSettings = async () => {
  saving.value = true
  try {
    await api.post('/api/v2/admin/settings', settings.value)
    alert('保存成功')
  } catch (e: any) {
    alert(e.response?.data?.error || '保存失败')
  } finally {
    saving.value = false
  }
}

onMounted(fetchSettings)
</script>

<template>
  <div class="space-y-6">
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">系统设置</h1>
        <p class="text-gray-500 mt-1">配置系统参数</p>
      </div>
      <button @click="saveSettings" :disabled="saving" class="px-4 py-2 bg-primary-500 text-white rounded-xl hover:bg-primary-600 disabled:opacity-50">
        {{ saving ? '保存中...' : '保存设置' }}
      </button>
    </div>

    <div v-if="loading" class="text-center py-12 text-gray-500">加载中...</div>

    <div v-else class="flex gap-6">
      <!-- Tabs -->
      <div class="w-48 flex-shrink-0">
        <div class="bg-white rounded-xl shadow-sm p-2 space-y-1">
          <button
            v-for="tab in tabs"
            :key="tab.key"
            @click="activeTab = tab.key"
            :class="[
              'w-full flex items-center gap-2 px-4 py-3 rounded-lg text-sm transition-colors',
              activeTab === tab.key ? 'bg-primary-50 text-primary-600' : 'text-gray-600 hover:bg-gray-50'
            ]"
          >
            <span>{{ tab.icon }}</span>
            <span>{{ tab.name }}</span>
          </button>
        </div>
      </div>

      <!-- Content -->
      <div class="flex-1 bg-white rounded-xl shadow-sm p-6">
        <div class="space-y-4">
          <div v-for="item in settingGroups[activeTab]" :key="item.key">
            <label class="block text-sm font-medium text-gray-700 mb-1">{{ item.label }}</label>
            
            <input
              v-if="item.type === 'text' || item.type === 'password' || item.type === 'number'"
              v-model="settings[item.key]"
              :type="item.type"
              :placeholder="item.placeholder"
              class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            />
            
            <textarea
              v-else-if="item.type === 'textarea'"
              v-model="settings[item.key]"
              rows="3"
              :placeholder="item.placeholder"
              class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            ></textarea>
            
            <select
              v-else-if="item.type === 'select'"
              v-model="settings[item.key]"
              class="w-full px-4 py-2 border border-gray-200 rounded-xl focus:ring-2 focus:ring-primary-500 focus:border-transparent"
            >
              <option value="ssl">SSL</option>
              <option value="tls">TLS</option>
              <option value="">无</option>
            </select>
            
            <label v-else-if="item.type === 'checkbox'" class="flex items-center gap-2">
              <input
                v-model="settings[item.key]"
                type="checkbox"
                :true-value="'1'"
                :false-value="'0'"
                class="rounded"
              />
              <span class="text-sm text-gray-600">启用</span>
            </label>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
