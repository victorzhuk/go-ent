---
name: vue-development
description: Vue 3 development with Composition API, Pinia, Vue Router, TypeScript, and best practices
---

# Vue 3 Development

## Composition API
```vue
<script setup lang="ts">
import { ref, computed, onMounted } from 'vue'

const count = ref(0)
const doubled = computed(() => count.value * 2)

function increment() {
  count.value++
}

onMounted(() => {
  console.log('Component mounted')
})
</script>

<template>
  <button @click="increment">{{ count }} ({{ doubled }})</button>
</template>
```

## State Management (Pinia)
```typescript
export const useUserStore = defineStore('user', () => {
  const users = ref<User[]>([])
  const loading = ref(false)

  const activeUsers = computed(() =>
    users.value.filter(u => u.status === 'active')
  )

  async function fetchUsers() {
    loading.value = true
    try {
      users.value = await api.getUsers()
    } finally {
      loading.value = false
    }
  }

  return { users, loading, activeUsers, fetchUsers }
})
```

## Composables (Custom Hooks)
```typescript
export function useAsync<T>(fn: () => Promise<T>) {
  const data = ref<T | null>(null)
  const error = ref<Error | null>(null)
  const loading = ref(false)

  async function execute() {
    loading.value = true
    error.value = null
    try {
      data.value = await fn()
    } catch (e) {
      error.value = e as Error
    } finally {
      loading.value = false
    }
  }

  return { data, error, loading, execute }
}
```

## Best Practices
- Use `<script setup>` for cleaner component code
- Use TypeScript with Vue — defineProps, defineEmits with type literals
- Use Pinia for state management (Vuex successor)
- Use Vue Router with typed route params
- Use `v-memo` for expensive list rendering
- Keep components small — extract composables for reusable logic
- Use `provide/inject` for deep prop drilling avoidance
