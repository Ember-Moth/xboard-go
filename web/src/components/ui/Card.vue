<template>
  <div :class="cardClasses">
    <div v-if="$slots.header" class="card-header">
      <slot name="header" />
    </div>
    <div class="card-body">
      <slot />
    </div>
    <div v-if="$slots.footer" class="card-footer">
      <slot name="footer" />
    </div>
  </div>
</template>

<script setup lang="ts">
import { computed } from 'vue'

interface Props {
  variant?: 'default' | 'bordered' | 'elevated'
  padding?: 'none' | 'sm' | 'md' | 'lg'
}

const props = withDefaults(defineProps<Props>(), {
  variant: 'default',
  padding: 'md',
})

const cardClasses = computed(() => {
  const classes = ['card']
  
  // Variant styles
  if (props.variant === 'bordered') {
    classes.push('border-2')
  } else if (props.variant === 'elevated') {
    classes.push('shadow-md')
  }
  
  // Padding styles
  const paddingClasses = {
    none: 'p-0',
    sm: 'p-3',
    md: 'p-4',
    lg: 'p-6',
  }
  classes.push(paddingClasses[props.padding])
  
  return classes.join(' ')
})
</script>

<style scoped>
.card-header {
  @apply pb-3 mb-3 border-b border-gray-200;
}

.card-footer {
  @apply pt-3 mt-3 border-t border-gray-200;
}

.card-body {
  @apply flex-1;
}
</style>
