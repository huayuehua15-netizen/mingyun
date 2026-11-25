<script setup lang="ts">
import { computed } from 'vue'
import type { LiuYaoPayload } from '../types'

const props = withDefaults(
  defineProps<{
    payload: LiuYaoPayload
    revealed?: number
  }>(),
  { revealed: 6 }
)

const VB_W = 1200
const VB_H = 470
const ROW_TOP = 420
const STEP = 64

const POS_X = 40
const LIUSHEN_X = 86
const BADGE_X = 150
const LINE_X = 196
const LINE_W = 400
const YIN_W = 190
const YIN_GAP = 20
const MOVING_X = 612
const LIUQIN_X = 642
const NAJIA_X = 704
const MARKS_X = 790
const CHANGE_X = 980
const CHANGE_W = 190
const CHANGE_YIN_W = 91
const CHANGE_YIN_GAP = 8

function yOf(pos: number): number {
  return ROW_TOP - (pos - 1) * STEP
}

function changedIsYin(kind: string): boolean {
  return kind.includes('阴')
}

const rows = computed(() =>
  [...props.payload.lines]
    .sort((a, b) => a.position - b.position)
    .map((line) => ({
      ...line,
      y: yOf(line.position),
      yin: line.yin_yang === '阴',
      visible: line.position <= props.revealed
    }))
)

const hasChange = computed(() => !!props.payload.changed_gua)

const fuShenAt = computed(() => {
  const map = new Map<number, string>()
  for (const f of props.payload.fu_shen ?? []) {
    map.set(f.under_position, `${f.liu_qin} ${f.gan}${f.zhi}${f.wu_xing}`)
  }
  return map
})
</script>

<template>
  <svg
    class="board-svg"
    :viewBox="`0 0 ${VB_W} ${VB_H}`"
    role="img"
    :aria-label="`六爻盘：${payload.gua_name}`"
  >
    <text class="col-title" :x="LINE_X + LINE_W / 2" y="36" text-anchor="middle">本卦</text>
    <text v-if="hasChange" class="col-title" :x="CHANGE_X + CHANGE_W / 2" y="36" text-anchor="middle">变卦</text>
    <line v-if="hasChange" x1="946" y1="56" x2="946" :y2="VB_H - 16" class="col-divider" />

    <g v-for="row in rows" :key="row.position">
      <!-- 爻位 -->
      <text class="pos-num" :x="POS_X" :y="row.y + 5">{{ ['初', '二', '三', '四', '五', '上'][row.position - 1] }}</text>

      <!-- 世应 -->
      <template v-if="row.visible && (row.shi || row.ying)">
        <rect
          class="badge"
          :x="BADGE_X"
          :y="row.y - 14"
          width="28"
          height="28"
          fill="var(--paper-2)"
          :stroke="row.shi ? 'var(--cinnabar)' : 'var(--stone)'"
        />
        <text
          class="badge-text"
          :x="BADGE_X + 14"
          :y="row.y + 5"
          text-anchor="middle"
          :fill="row.shi ? 'var(--cinnabar)' : 'var(--stone)'"
        >
          {{ row.shi ? '世' : '应' }}
        </text>
      </template>

      <!-- 六神 -->
      <text v-if="row.visible" class="liushen" :x="LIUSHEN_X" :y="row.y + 5">{{ row.liu_shen }}</text>

      <!-- 本卦爻线 -->
      <template v-if="row.visible">
        <template v-if="!row.yin">
          <rect
            :x="LINE_X"
            :y="row.y - 4"
            :width="LINE_W"
            height="8"
            :fill="row.moving ? 'var(--cinnabar)' : 'var(--ink)'"
          />
        </template>
        <template v-else>
          <rect
            :x="LINE_X"
            :y="row.y - 4"
            :width="YIN_W"
            height="8"
            :fill="row.moving ? 'var(--cinnabar)' : 'var(--ink)'"
          />
          <rect
            :x="LINE_X + YIN_W + YIN_GAP"
            :y="row.y - 4"
            :width="YIN_W"
            height="8"
            :fill="row.moving ? 'var(--cinnabar)' : 'var(--ink)'"
          />
        </template>
        <text v-if="row.moving" class="moving-mark" :x="MOVING_X" :y="row.y + 7">
          {{ row.kind === '老阳' ? '○' : '×' }}
        </text>
      </template>
      <line v-else :x1="LINE_X" :y1="row.y" :x2="LINE_X + LINE_W" :y2="row.y" class="ghost-line" />

      <!-- 六亲纳甲 -->
      <template v-if="row.visible">
        <text class="liuqin" :x="LIUQIN_X" :y="row.y + 6">{{ row.liu_qin }}</text>
        <text class="najia" :x="NAJIA_X" :y="row.y + 6">{{ row.gan }}{{ row.zhi }}{{ row.wu_xing }}</text>
        <text v-if="row.marks && row.marks.length" class="marks" :x="MARKS_X" :y="row.y + 5">
          {{ row.marks.join('·') }}
        </text>
        <text v-if="fuShenAt.get(row.position)" class="fushen" :x="LIUQIN_X" :y="row.y + 22">
          伏 {{ fuShenAt.get(row.position) }}
        </text>
      </template>

      <!-- 变卦 -->
      <template v-if="hasChange">
        <template v-if="row.visible">
          <template v-if="row.change">
            <template v-if="!changedIsYin(row.change.kind)">
              <rect :x="CHANGE_X" :y="row.y - 4" :width="CHANGE_W" height="8" fill="var(--ink)" />
            </template>
            <template v-else>
              <rect :x="CHANGE_X" :y="row.y - 4" :width="CHANGE_YIN_W" height="8" fill="var(--ink)" />
              <rect
                :x="CHANGE_X + CHANGE_YIN_W + CHANGE_YIN_GAP"
                :y="row.y - 4"
                :width="CHANGE_YIN_W"
                height="8"
                fill="var(--ink)"
              />
            </template>
            <text class="change-label" :x="CHANGE_X" :y="row.y + 20">
              {{ row.change.liu_qin }} {{ row.change.gan }}{{ row.change.zhi }}{{ row.change.wu_xing }}
            </text>
          </template>
          <template v-else>
            <template v-if="!row.yin">
              <rect :x="CHANGE_X" :y="row.y - 2.5" :width="CHANGE_W" height="5" class="static-line" />
            </template>
            <template v-else>
              <rect :x="CHANGE_X" :y="row.y - 2.5" :width="CHANGE_YIN_W" height="5" class="static-line" />
              <rect
                :x="CHANGE_X + CHANGE_YIN_W + CHANGE_YIN_GAP"
                :y="row.y - 2.5"
                :width="CHANGE_YIN_W"
                height="5"
                class="static-line"
              />
            </template>
          </template>
        </template>
      </template>
    </g>
  </svg>
</template>

<style scoped>
.board-svg text {
  font-family: "PingFang SC", "Microsoft YaHei", "Noto Sans SC", sans-serif;
  font-size: 15px;
  fill: var(--ink-2);
}

.col-title {
  font-family: "Songti SC", "STSong", "SimSun", serif !important;
  font-size: 12.5px;
  fill: var(--grey) !important;
  letter-spacing: 0.22em;
}

.pos-num {
  font-family: "Songti SC", "STSong", "SimSun", serif !important;
  font-size: 15px !important;
  fill: var(--ink) !important;
  letter-spacing: 0.1em;
}

.badge-text {
  font-family: "Songti SC", "STSong", "SimSun", serif !important;
  font-size: 16px !important;
}

.liushen {
  fill: var(--ink-2) !important;
  font-size: 14px !important;
}

.liuqin {
  font-family: "Songti SC", "STSong", "SimSun", serif !important;
  font-size: 16px !important;
  fill: var(--ink) !important;
}

.najia {
  font-family: "Songti SC", "STSong", "SimSun", serif !important;
  font-size: 16px !important;
  fill: var(--ink) !important;
  font-variant-numeric: tabular-nums;
}

.marks {
  fill: var(--cinnabar) !important;
  font-size: 11.5px !important;
}

.fushen {
  fill: var(--grey) !important;
  font-size: 11px !important;
}

.moving-mark {
  fill: var(--cinnabar) !important;
  font-size: 19px !important;
}

.static-line {
  fill: #c9bda8;
}

.change-label {
  fill: var(--ink-2) !important;
  font-size: 11.5px !important;
}

.col-divider {
  stroke: var(--line);
  stroke-width: 1;
}

.ghost-line {
  stroke: var(--line);
  stroke-width: 1;
  stroke-dasharray: 2 4;
}
</style>
