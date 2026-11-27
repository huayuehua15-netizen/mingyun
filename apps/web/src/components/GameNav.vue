<script setup lang="ts">
import type { ViewName } from '../types'

const items: { key: ViewName; glyph: string; label: string; hint: string }[] = [
  { key: 'home', glyph: '山', label: '山门', hint: '回到入口' },
  { key: 'cast', glyph: '盘', label: '起盘', hint: '案上起卦起课起局' },
  { key: 'archive', glyph: '册', label: '卦历', hint: '六术起盘与解读记录' },
  { key: 'codex', glyph: '典', label: '六术', hint: '术数总览' },
  { key: 'settings', glyph: '机', label: '模型', hint: '模型接入与运行状态' }
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
