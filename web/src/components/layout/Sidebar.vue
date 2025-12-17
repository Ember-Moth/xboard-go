<template>
  <aside :class="sidebarClasses">
    <div class="sidebar-header">
      <div class="flex items-center gap-3 px-4 py-4">
        <div class="text-xl font-bold">dashGO</div>
      </div>
    </div>
    
    <nav class="sidebar-nav">
      <slot />
    </nav>
  </aside>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  collapsed?: boolean
}

const props = withDefaults(defineProps<Props>(), {
  collapsed: false,
})

const sidebarClasses = computed(() => {
  const classes = [
    'sidebar',
    'fixed left-0 top-0 h-full',
    'bg-white border-r border-gray-200',
    'transition-all duration-200',
  ]
  
  if (props.collapsed) {
    classes.push('w-16')
  } else {
    classes.push('w-60')
  }
  
  return classes.join(' ')
})
</script>

<style scoped>
.sidebar-header {
  @apply border-b border-gray-200;
}

.sidebar-nav {
  @apply py-4;
}
</style>
