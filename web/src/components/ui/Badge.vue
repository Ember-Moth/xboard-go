<template>
  <span :class="badgeClasses">
    <slot />
  </span>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'success' | 'warning' | 'error' | 'info' | 'default'
  size?: 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  size: 'md',
})

const badgeClasses = computed(() => {
  const classes = ['badge']
  
  // Variant styles
  if (props.variant !== 'default') {
    classes.push(`badge-${props.variant}`)
  } else {
    classes.push('bg-gray-100 text-gray-700 border border-gray-200')
  }
  
  // Size styles
  const sizeClasses = {
    sm: 'px-2 py-0.5 text-xs',
    md: 'px-2.5 py-1 text-xs',
    lg: 'px-3 py-1.5 text-sm',
  }
  classes.push(sizeClasses[props.size])
  
  return classes.join(' ')
})
</script>

<style scoped>
/* Badge styles are defined in style.css */
</style>
