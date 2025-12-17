<template>
  <header class="header">
    <div class="header-content">
      <div class="header-left">
        <slot name="left" />
      </div>
      
      <div class="header-center">
        <slot name="center" />
      </div>
      
      <div class="header-right">
        <slot name="right" />
      </div>
    </div>
  </header>
</template>

<script setup lang="ts">
// Simple header component for flat design
</script>

<style scoped>
.header {
  @apply fixed top-0 left-0 right-0;
  @apply bg-white border-b border-gray-200;
  @apply z-10;
  height: var(--header-height);
}

.header-content {
  @apply h-full px-6;
  @apply flex items-center justify-between;
}

.header-left,
.header-center,
.header-right {
  @apply flex items-center gap-4;
}

.header-center {
  @apply flex-1 justify-center;
}
</style>
