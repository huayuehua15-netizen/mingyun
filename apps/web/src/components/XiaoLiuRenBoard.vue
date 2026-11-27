<script setup lang="ts">
import { computed } from 'vue'
import type { XiaoLiuRenPayload } from '../types'

const props = defineProps<{ payload: XiaoLiuRenPayload }>()

const stages = computed(() => [
  { label: '月 · 背景', palace: props.payload.month_palace },
  { label: '日 · 事体', palace: props.payload.day_palace },
  { label: '时 · 动向', palace: props.payload.final_palace }
])

function isFinal(index: number): boolean {
  return index === props.payload.final_palace.index
}

function isStage(index: number): boolean {
  return (
    index === props.payload.month_palace.index ||
    index === props.payload.day_palace.index ||
    index === props.payload.final_palace.index
  )
}
</script>

<template>
  <div class="xlr-board">
    <div class="xlr-path">
      <template v-for="(stage, i) in stages" :key="stage.label">
        <article class="xlr-stage" :class="{ 'is-final': i === 2 }">
          <span>{{ stage.label }}</span>
          <b>{{ stage.palace.name }}</b>
          <em>{{ stage.palace.element }} · {{ stage.palace.deity }} · {{ stage.palace.direction }}</em>
        </article>
        <div v-if="i < stages.length - 1" class="xlr-arrow" aria-hidden="true">→</div>
      </template>
    </div>

    <div class="xlr-grid">
      <article
        v-for="palace in payload.six_palaces"
        :key="palace.index"
        class="xlr-cell"
        :class="{ 'is-final': isFinal(palace.index), 'is-stage': isStage(palace.index) }"
      >
        <span class="xlr-index">0{{ palace.index + 1 }}</span>
        <b>{{ palace.name }}</b>
        <em>{{ palace.element }} · {{ palace.deity }} · {{ palace.direction }}</em>
        <small>{{ palace.meaning }}</small>
        <i class="xlr-luck" :class="palace.luck === '吉' ? 'is-good' : 'is-bad'">{{ palace.luck }}</i>
      </article>
    </div>

    <footer class="xlr-foot">
      <span>落宫 <b>{{ payload.final_palace.name }}</b></span>
      <span>月{{ payload.month_palace.name }} · 日{{ payload.day_palace.name }} · 时{{ payload.final_palace.name }}</span>
      <span>近事窗口 1—15 日</span>
    </footer>
  </div>
</template>

<style scoped>
.xlr-board {
  width: 100%;
  min-width: 520px;
  color: #2d2a24;
}

.xlr-path {
  display: flex;
  align-items: stretch;
  gap: 8px;
  margin-bottom: 14px;
}

.xlr-stage {
  flex: 1;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  padding: 10px 8px;
  border: 1px solid #c9bfae;
  background: #f8f3e8;
}

.xlr-stage.is-final {
  border-color: #9e3327;
  background: #fbf6e9;
}

.xlr-stage span {
  color: #8a7f6c;
  font-size: 11px;
  letter-spacing: 0.08em;
}

.xlr-stage b {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 30px;
  font-weight: 500;
  color: #9e3327;
  line-height: 1.05;
}

.xlr-stage em {
  color: #746a5a;
  font-size: 10px;
  font-style: normal;
}

.xlr-arrow {
  align-self: center;
  color: #9e3327;
  font-size: 16px;
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
}

.xlr-grid {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  border: 1px solid #2d2a24;
  background: #f8f3e8;
}

.xlr-cell {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 3px;
  min-height: 112px;
  padding: 12px 8px 10px;
  border-right: 1px solid #c9bfae;
  border-bottom: 1px solid #c9bfae;
  text-align: center;
  background: #f8f3e8;
}

.xlr-cell:nth-child(3n) {
  border-right: 0;
}

.xlr-cell:nth-child(n + 4) {
  border-bottom: 0;
}

.xlr-cell.is-stage {
  background: #f2ebdc;
}

.xlr-cell.is-final {
  outline: 1px solid #9e3327;
  outline-offset: -3px;
  background: #fbf5e8;
}

.xlr-index {
  position: absolute;
  top: 5px;
  left: 7px;
  color: #8a7f6c;
  font-size: 10px;
  letter-spacing: 0.12em;
}

.xlr-cell b {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 25px;
  font-weight: 500;
  color: #2d2a24;
  line-height: 1.1;
}

.xlr-cell.is-final b {
  color: #9e3327;
}

.xlr-cell em {
  color: #746a5a;
  font-size: 10px;
  font-style: normal;
}

.xlr-cell small {
  color: #8a7f6c;
  font-size: 10px;
  line-height: 1.45;
}

.xlr-luck {
  position: absolute;
  top: 6px;
  right: 7px;
  padding: 1px 4px;
  border: 1px solid #c9bfae;
  color: #746a5a;
  font-size: 9px;
  font-style: normal;
}

.xlr-luck.is-good {
  border-color: #9e3327;
  color: #9e3327;
}

.xlr-luck.is-bad {
  border-color: #8a7f6c;
  color: #5d5549;
}

.xlr-foot {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 16px;
  margin-top: 12px;
  color: #746a5a;
  font-size: 11px;
}

.xlr-foot b {
  color: #9e3327;
  font-weight: 500;
}

@media (max-width: 640px) {
  .xlr-board {
    min-width: 460px;
  }

  .xlr-stage b {
    font-size: 24px;
  }

  .xlr-cell b {
    font-size: 21px;
  }
}
</style>
