<script setup lang="ts">
import { computed } from 'vue'
import type { ZiWeiPalace, ZiWeiPayload, ZiWeiStar } from '../types'

const props = defineProps<{ payload: ZiWeiPayload }>()

const slots: Record<string, { row: number; col: number }> = {
  巳: { row: 1, col: 1 },
  午: { row: 1, col: 2 },
  未: { row: 1, col: 3 },
  申: { row: 1, col: 4 },
  酉: { row: 2, col: 4 },
  戌: { row: 3, col: 4 },
  亥: { row: 4, col: 4 },
  子: { row: 4, col: 3 },
  丑: { row: 4, col: 2 },
  寅: { row: 4, col: 1 },
  卯: { row: 3, col: 1 },
  辰: { row: 2, col: 1 }
}

const byBranch = computed(() => {
  const map: Record<string, ZiWeiPalace> = {}
  for (const palace of props.payload.palaces) map[palace.branch] = palace
  return map
})

const ring = computed(() =>
  Object.entries(slots).map(([branch, slot]) => ({
    branch,
    ...slot,
    palace: byBranch.value[branch]
  }))
)

function starLine(stars: ZiWeiStar[] | undefined): string {
  if (!stars?.length) return '—'
  return stars.map((star) => `${star.name}${star.hua ? '化' + star.hua : ''}`).join(' ')
}
</script>

<template>
  <div class="zw-board">
    <div class="zw-ring">
      <article
        v-for="cell in ring"
        :key="cell.branch"
        class="zw-cell"
        :class="{ 'is-ming': cell.palace?.is_ming, 'is-shen': cell.palace?.is_shen }"
        :style="{ gridRow: cell.row, gridColumn: cell.col }"
      >
        <div class="zw-cell-head">
          <span>{{ cell.branch }}</span>
          <em>{{ cell.palace?.stem_branch }}</em>
        </div>
        <h3>{{ cell.palace?.name }}</h3>
        <p class="zw-major">{{ starLine(cell.palace?.major_stars) }}</p>
        <p class="zw-minor">{{ starLine(cell.palace?.minor_stars) }}</p>
        <footer>
          <span v-if="cell.palace?.daxian_range">{{ cell.palace?.daxian_range }}岁</span>
          <i v-if="cell.palace?.is_shen">身</i>
        </footer>
      </article>

      <div class="zw-center">
        <div class="zw-center-kicker">{{ payload.year_gz }}年 · {{ payload.gender === 'male' ? '乾造' : '坤造' }}</div>
        <b>{{ payload.ming_gong }}宫命</b>
        <em>{{ payload.wu_xing_ju }}</em>
        <dl>
          <div><dt>身宫</dt><dd>{{ payload.shen_gong }}宫</dd></div>
          <div><dt>紫微</dt><dd>{{ payload.ziwei_pos }}宫</dd></div>
          <div><dt>天府</dt><dd>{{ payload.tianfu_pos }}宫</dd></div>
          <div><dt>四化</dt><dd>{{ payload.sihua_year.map((s) => s.star + '化' + s.hua).join('、') }}</dd></div>
        </dl>
      </div>
    </div>

    <div class="zw-daxian">
      <header><span>大限</span><em>起运 {{ payload.da_xian[0]?.start_age }} 岁 · 十年一宫</em></header>
      <div class="zw-daxian-track">
        <span v-for="dx in payload.da_xian" :key="dx.index">
          <b>{{ dx.range }}</b>
          <em>{{ dx.name }} · {{ dx.branch }}</em>
        </span>
      </div>
    </div>

    <footer class="zw-foot">
      <span>公历/真太阳时 {{ payload.solar_date }}</span>
      <span>农历 {{ payload.lunar_date }}</span>
      <span>命宫 {{ payload.ming_gong }} · 身宫 {{ payload.shen_gong }}</span>
    </footer>
  </div>
</template>

<style scoped>
.zw-board {
  width: 100%;
  min-width: 560px;
  color: #2d2a24;
}

.zw-ring {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  grid-template-rows: repeat(4, minmax(112px, auto));
  border: 1px solid #2d2a24;
  background: #f8f3e8;
}

.zw-cell {
  display: flex;
  flex-direction: column;
  gap: 3px;
  padding: 8px 9px 7px;
  border-right: 1px solid #c9bfae;
  border-bottom: 1px solid #c9bfae;
  background: #f8f3e8;
}

.zw-cell:nth-child(4n) {
  border-right: 0;
}

.zw-cell.is-ming {
  background: #fbf2df;
  box-shadow: inset 3px 0 0 #9e3327;
}

.zw-cell.is-shen {
  border-bottom-color: #9e3327;
}

.zw-cell-head {
  display: flex;
  justify-content: space-between;
  color: #8a7f6c;
  font-size: 10px;
}

.zw-cell-head span {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 14px;
  color: #2d2a24;
}

.zw-cell h3 {
  margin: 0;
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 16px;
  font-weight: 500;
  letter-spacing: 0.12em;
}

.zw-major {
  min-height: 32px;
  margin: 2px 0 0;
  color: #9e3327;
  font-size: 12px;
  line-height: 1.45;
}

.zw-minor {
  min-height: 18px;
  margin: 0;
  color: #746a5a;
  font-size: 10px;
  line-height: 1.4;
}

.zw-cell footer {
  display: flex;
  justify-content: space-between;
  margin-top: auto;
  color: #8a7f6c;
  font-size: 10px;
}

.zw-cell footer i {
  color: #9e3327;
  font-style: normal;
}

.zw-center {
  grid-area: 2 / 2 / 4 / 4;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 4px;
  padding: 12px;
  border-right: 1px solid #c9bfae;
  border-bottom: 1px solid #c9bfae;
  background: #f1eadb;
  text-align: center;
}

.zw-center-kicker {
  color: #8a7f6c;
  font-size: 11px;
  letter-spacing: 0.08em;
}

.zw-center b {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 26px;
  font-weight: 500;
  letter-spacing: 0.14em;
  color: #2d2a24;
}

.zw-center > em {
  color: #9e3327;
  font-size: 12px;
  font-style: normal;
}

.zw-center dl {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px 12px;
  margin: 8px 0 0;
  text-align: left;
}

.zw-center dl div {
  display: flex;
  gap: 5px;
  font-size: 10px;
}

.zw-center dt {
  color: #8a7f6c;
}

.zw-center dd {
  margin: 0;
  color: #2d2a24;
}

.zw-daxian {
  margin-top: 14px;
  border: 1px solid #c9bfae;
  background: #f8f3e8;
}

.zw-daxian header {
  display: flex;
  justify-content: space-between;
  padding: 8px 10px;
  border-bottom: 1px solid #c9bfae;
}

.zw-daxian header span {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 16px;
  letter-spacing: 0.16em;
}

.zw-daxian header em {
  color: #8a7f6c;
  font-size: 10px;
  font-style: normal;
}

.zw-daxian-track {
  display: flex;
  overflow-x: auto;
  gap: 0;
}

.zw-daxian-track span {
  flex: 0 0 auto;
  min-width: 92px;
  padding: 7px 9px;
  border-right: 1px solid #e3dacb;
  font-size: 11px;
}

.zw-daxian-track b {
  display: block;
  color: #9e3327;
  font-weight: 500;
}

.zw-daxian-track em {
  color: #746a5a;
  font-style: normal;
}

.zw-foot {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 14px;
  margin-top: 10px;
  color: #8a7f6c;
  font-size: 10px;
}

@media (max-width: 640px) {
  .zw-board {
    min-width: 500px;
  }

  .zw-ring {
    grid-template-rows: repeat(4, minmax(96px, auto));
  }

  .zw-cell h3 {
    font-size: 14px;
  }

  .zw-center b {
    font-size: 21px;
  }
}
</style>
