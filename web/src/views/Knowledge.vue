<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import api from '@/api'

interface KnowledgeItem {
  id: number
  title: string
  body: string
  category: string
  language: string
}

const categories = ref<string[]>([])
const items = ref<KnowledgeItem[]>([])
const selectedCategory = ref('')
const selectedItem = ref<KnowledgeItem | null>(null)
const loading = ref(false)
const searchQuery = ref('')

// 默认教程分类
const defaultCategories = [
  { name: 'iOS', icon: '📱', color: 'from-blue-400 to-blue-600' },
  { name: 'Android', icon: '🤖', color: 'from-green-400 to-green-600' },
  { name: 'Windows', icon: '💻', color: 'from-cyan-400 to-cyan-600' },
  { name: 'macOS', icon: '🍎', color: 'from-gray-400 to-gray-600' },
  { name: 'Linux', icon: '🐧', color: 'from-orange-400 to-orange-600' },
  { name: '常见问题', icon: '❓', color: 'from-purple-400 to-purple-600' },
]

const filteredItems = computed(() => {
  let result = items.value
  if (selectedCategory.value) {
    result = result.filter(item => item.category === selectedCategory.value)
  }
  if (searchQuery.value) {
    const query = searchQuery.value.toLowerCase()
    result = result.filter(item => 
      item.title.toLowerCase().includes(query) || 
      item.body.toLowerCase().includes(query)
    )
  }
  return result
})

const fetchCategories = async () => {
  try {
    const res = await api.get('/api/v1/knowledge/categories')
    categories.value = res.data.data || []
  } catch (e) {
    console.error(e)
  }
}

const fetchItems = async () => {
  loading.value = true
  try {
    const res = await api.get('/api/v1/knowledge', {
      params: { category: selectedCategory.value }
    })
    items.value = res.data.data || []
  } catch (e) {
    console.error(e)
  } finally {
    loading.value = false
  }
}

const selectCategory = (category: string) => {
  selectedCategory.value = category
  selectedItem.value = null
  fetchItems()
}

const selectItem = (item: KnowledgeItem) => {
  selectedItem.value = item
}

const goBack = () => {
  selectedItem.value = null
}

const getCategoryIcon = (cat: string) => {
  const found = defaultCategories.find(c => c.name === cat)
  return found?.icon || '📄'
}

const getCategoryColor = (cat: string) => {
  const found = defaultCategories.find(c => c.name === cat)
  return found?.color || 'from-gray-400 to-gray-600'
}

onMounted(() => {
  fetchCategories()
  fetchItems()
})
</script>

<template>
  <div class="space-y-6 animate-fade-in">
    <!-- Header -->
    <div class="flex items-center justify-between">
      <div>
        <h1 class="text-2xl font-bold text-gray-900">使用教程</h1>
        <p class="text-gray-500 mt-1">各平台客户端使用指南</p>
      </div>
    </div>

    <!-- Search Bar -->
    <div class="bg-white rounded-2xl shadow-sm border border-gray-100 p-4">
      <div class="relative">
        <span class="absolute left-4 top-1/2 -translate-y-1/2 text-gray-400">🔍</span>
        <input 
          v-model="searchQuery"
          type="text" 
          placeholder="搜索教程..." 
          class="w-full pl-12 pr-4 py-3 bg-gray-50 rounded-xl border-0 focus:ring-2 focus:ring-primary-500 focus:bg-white transition-all"
        />
      </div>
    </div>

    <!-- Category Cards -->
    <div v-if="!selectedItem" class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-6 gap-4">
      <button
        v-for="cat in defaultCategories"
        :key="cat.name"
        @click="selectCategory(cat.name)"
        :class="selectedCategory === cat.name ? 'ring-2 ring-primary-500 ring-offset-2' : ''"
        class="bg-white rounded-2xl shadow-sm border border-gray-100 p-4 hover:shadow-md transition-all group"
      >
        <div :class="`bg-gradient-to-br ${cat.color}`" class="w-12 h-12 rounded-xl flex items-center justify-center text-white text-2xl mx-auto mb-3 group-hover:scale-110 transition-transform shadow-lg">
          {{ cat.icon }}
        </div>
        <p class="text-sm font-medium text-gray-700 text-center">{{ cat.name }}</p>
      </button>
    </div>

    <!-- Content Area -->
    <div class="flex gap-6">
      <!-- Sidebar (when not viewing article) -->
      <div v-if="!selectedItem && categories.length > 0" class="w-48 flex-shrink-0 hidden lg:block">
        <div class="bg-white rounded-2xl p-4 shadow-sm border border-gray-100 sticky top-4">
          <h3 class="text-sm font-medium text-gray-500 mb-3">全部分类</h3>
          <ul class="space-y-1">
            <li>
              <button
                @click="selectCategory('')"
                :class="selectedCategory === '' ? 'bg-primary-50 text-primary-600' : 'text-gray-600 hover:bg-gray-50'"
                class="w-full text-left px-3 py-2 rounded-lg text-sm transition flex items-center gap-2"
              >
                <span>📚</span> 全部
              </button>
            </li>
            <li v-for="cat in categories" :key="cat">
              <button
                @click="selectCategory(cat)"
                :class="selectedCategory === cat ? 'bg-primary-50 text-primary-600' : 'text-gray-600 hover:bg-gray-50'"
                class="w-full text-left px-3 py-2 rounded-lg text-sm transition flex items-center gap-2"
              >
                <span>{{ getCategoryIcon(cat) }}</span> {{ cat }}
              </button>
            </li>
          </ul>
        </div>
      </div>

      <!-- Main Content -->
      <div class="flex-1">
        <!-- Article Detail -->
        <div v-if="selectedItem" class="bg-white rounded-2xl shadow-sm border border-gray-100 overflow-hidden">
          <!-- Article Header -->
          <div class="bg-gradient-to-r from-primary-500 to-primary-600 p-6 text-white">
            <button
              @click="goBack"
              class="flex items-center gap-2 text-white/80 hover:text-white mb-4 transition-colors"
            >
              <span>←</span> 返回列表
            </button>
            <h2 class="text-xl font-bold">{{ selectedItem.title }}</h2>
            <div class="flex items-center gap-2 mt-2 text-white/80 text-sm">
              <span class="px-2 py-0.5 bg-white/20 rounded-full">{{ selectedItem.category }}</span>
            </div>
          </div>
          
          <!-- Article Body -->
          <div class="p-6">
            <div class="prose prose-sm max-w-none prose-headings:text-gray-900 prose-p:text-gray-600 prose-a:text-primary-600" v-html="selectedItem.body"></div>
          </div>
        </div>

        <!-- Article List -->
        <div v-else class="space-y-3">
          <div v-if="loading" class="bg-white rounded-2xl p-12 text-center text-gray-500 shadow-sm border border-gray-100">
            <div class="animate-spin w-8 h-8 border-2 border-primary-500 border-t-transparent rounded-full mx-auto mb-3"></div>
            加载中...
          </div>
          
          <div v-else-if="filteredItems.length === 0" class="bg-white rounded-2xl p-12 text-center text-gray-500 shadow-sm border border-gray-100">
            <span class="text-4xl mb-3 block">📭</span>
            <p>暂无相关教程</p>
            <p class="text-sm mt-1">请尝试其他分类或关键词</p>
          </div>

          <div
            v-else
            v-for="item in filteredItems"
            :key="item.id"
            @click="selectItem(item)"
            class="bg-white rounded-2xl p-5 shadow-sm border border-gray-100 cursor-pointer hover:border-primary-200 hover:shadow-md transition-all group"
          >
            <div class="flex items-center gap-4">
              <div :class="`bg-gradient-to-br ${getCategoryColor(item.category)}`" class="w-12 h-12 rounded-xl flex items-center justify-center text-white text-xl flex-shrink-0 shadow-md">
                {{ getCategoryIcon(item.category) }}
              </div>
              <div class="flex-1 min-w-0">
                <h3 class="font-medium text-gray-900 group-hover:text-primary-600 transition-colors">{{ item.title }}</h3>
                <span class="text-xs text-gray-400 mt-1">{{ item.category }}</span>
              </div>
              <span class="text-gray-400 group-hover:translate-x-1 transition-transform">›</span>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>
