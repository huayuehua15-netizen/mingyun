<script setup lang="ts">
import type { ViewName } from '../types'

const items: { key: ViewName; glyph: string; label: string; hint: string }[] = [
  { key: 'home', glyph: '山', label: '山门', hint: '回到入口' },
  { key: 'cast', glyph: '卦', label: '问卦', hint: '案上起卦' },
  { key: 'archive', glyph: '册', label: '卦例', hint: '翻看旧卦' },
  { key: 'codex', glyph: '典', label: '三术', hint: '术数典籍' }
]

defineProps<{ current: ViewName; archiveCount: number }>()
defineEmits<{ (e: 'navigate', view: ViewName): void }>()
</script>

<template>
  <nav class="game-nav" aria-label="主导航">
    <button
      v-for="item in items"
      :key="item.key"
      class="game-nav-item"
      :class="{ 'is-active': current === item.key }"
      :aria-current="current === item.key ? 'page' : undefined"
      :title="item.hint"
      @click="$emit('navigate', item.key)"
    >
      <i>{{ item.glyph }}</i>
      <span>{{ item.label }}</span>
      <em v-if="item.key === 'archive' && archiveCount">{{ archiveCount }}</em>
    </button>
  </nav>
</template>
