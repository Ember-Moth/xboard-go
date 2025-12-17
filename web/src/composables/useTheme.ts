import { ref, computed, watch } from 'vue'

export type ThemeMode = 'light' | 'dark'

const THEME_STORAGE_KEY = 'dashgo-theme'

// Theme state
const currentTheme = ref<ThemeMode>('light')

// Initialize theme from localStorage or system preference
const initTheme = () => {
  const stored = localStorage.getItem(THEME_STORAGE_KEY) as ThemeMode | null
  
  if (stored) {
    currentTheme.value = stored
  } else {
    // For now, we only support light theme (monochrome black & white)
    currentTheme.value = 'light'
  }
  
  applyTheme(currentTheme.value)
}

// Apply theme to document
const applyTheme = (theme: ThemeMode) => {
  document.documentElement.setAttribute('data-theme', theme)
  
  // Update meta theme-color for mobile browsers
  const metaThemeColor = document.querySelector('meta[name="theme-color"]')
  if (metaThemeColor) {
    metaThemeColor.setAttribute('content', theme === 'light' ? '#FFFFFF' : '#000000')
  }
}

// Watch for theme changes
watch(currentTheme, (newTheme) => {
  localStorage.setItem(THEME_STORAGE_KEY, newTheme)
  applyTheme(newTheme)
})

export function useTheme() {
  // Initialize on first use
  if (!document.documentElement.hasAttribute('data-theme')) {
    initTheme()
  }
  
  const theme = computed(() => currentTheme.value)
  
  const setTheme = (newTheme: ThemeMode) => {
    currentTheme.value = newTheme
  }
  
  const toggleTheme = () => {
    currentTheme.value = currentTheme.value === 'light' ? 'dark' : 'light'
  }
  
  const isLight = computed(() => currentTheme.value === 'light')
  const isDark = computed(() => currentTheme.value === 'dark')
  
  return {
    theme,
    setTheme,
    toggleTheme,
    isLight,
    isDark,
  }
}
