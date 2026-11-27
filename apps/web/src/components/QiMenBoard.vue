<script setup lang="ts">
import { computed } from 'vue'
import type { QiMenPayload } from '../types'

const props = defineProps<{ payload: QiMenPayload }>()

// 传统九宫方位：巽四、离九、坤二在上；震三、中五、兑七在中；艮八、坎一、乾六在下。
const slots: Record<number, { row: number; col: number }> = {
  4: { row: 1, col: 1 },
  9: { row: 1, col: 2 },
  2: { row: 1, col: 3 },
  3: { row: 2, col: 1 },
  5: { row: 2, col: 2 },
  7: { row: 2, col: 3 },
  8: { row: 3, col: 1 },
  1: { row: 3, col: 2 },
  6: { row: 3, col: 3 }
}

const cells = computed(() =>
  props.payload.palaces.map((palace) => ({
    ...palace,
    ...(slots[palace.index] || { row: 1, col: 1 })
  }))
)

const examRef = computed(() => {
  const jing = props.payload.palaces.find((p) => p.door === '景门')
  const fu = props.payload.palaces.find((p) => p.star === '天辅')
  return [
    jing ? `景门 ${jing.name}` : '',
    fu ? `天辅 ${fu.name}` : ''
  ]
    .filter(Boolean)
    .join(' / ')
})
</script>

<template>
  <div class="qm-board">
    <div class="qm-grid">
      <article
        v-for="cell in cells"
        :key="cell.index"
        class="qm-cell"
        :class="{
          'is-center': cell.index === 5,
          'is-zhifu': cell.is_zhi_fu,
          'is-zhishi': cell.is_zhi_shi,
          'is-kong': cell.kong,
          'is-horse': cell.horse,
          'is-po': cell.men_po,
          'is-edge-right': cell.col === 3,
          'is-edge-bottom': cell.row === 3
        }"
        :style="{ gridRow: cell.row, gridColumn: cell.col }"
      >
        <div class="qm-cell-head">
          <span>{{ cell.name }}</span>
          <em>{{ cell.direction }}</em>
        </div>
        <div class="qm-stars">
          <b>{{ cell.star }}</b>
          <span v-if="cell.door">{{ cell.door }}</span>
          <span v-else class="is-quiet">中宫</span>
        </div>
        <div class="qm-gans">
          <span class="qm-heaven">{{ cell.heaven_gan }}</span>
          <i>天</i>
          <span class="qm-earth">{{ cell.earth_gan }}</span>
          <i>地</i>
        </div>
        <div class="qm-cell-foot">
          <span v-if="cell.god" class="qm-god">{{ cell.god }}</span>
          <span v-if="cell.kong" class="qm-flag is-kong">空</span>
          <span v-if="cell.horse" class="qm-flag is-horse">马</span>
          <span v-if="cell.men_po" class="qm-flag is-po">迫</span>
          <span v-if="cell.is_zhi_fu" class="qm-flag is-fu">符</span>
          <span v-if="cell.is_zhi_shi" class="qm-flag is-shi">使</span>
        </div>
      </article>
    </div>

    <div class="qm-sheets">
      <section class="qm-sheet">
        <header><span>四柱</span><em>{{ payload.solar_term }}{{ payload.yuan }}</em></header>
        <div class="qm-sheet-body">
          <span>年 <b>{{ payload.year_gz }}</b></span>
          <span>月 <b>{{ payload.month_gz }}</b></span>
          <span>日 <b>{{ payload.day_gz }}</b></span>
          <span>时 <b>{{ payload.hour_gz }}</b></span>
        </div>
      </section>
      <section class="qm-sheet">
        <header><span>近事用神</span><em>只给落宫</em></header>
        <div class="qm-sheet-body">
          <span>考试 <b>{{ examRef || '未定' }}</b></span>
          <span>空亡 <b>{{ payload.kong.join('、') }}</b></span>
          <span>马星 <b>{{ payload.horse }}</b></span>
          <span>旬首 <b>{{ payload.xun_shou }}</b></span>
        </div>
      </section>
    </div>
  </div>
</template>

<style scoped>
.qm-board {
  width: 100%;
  min-width: 540px;
  color: #2d2a24;
}

.qm-grid {
  position: relative;
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  grid-template-rows: repeat(3, minmax(132px, auto));
  border: 1px solid #2d2a24;
  background: #f8f3e8;
}

.qm-cell {
  position: relative;
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 20px 10px 10px;
  border-right: 1px solid #c9bfae;
  border-bottom: 1px solid #c9bfae;
  background: #f8f3e8;
}

.qm-cell.is-edge-right {
  border-right: 0;
}

.qm-cell.is-edge-bottom {
  border-bottom: 0;
}

.qm-cell.is-center {
  background: #f1eadb;
  border-left: 1px solid #2d2a24;
  border-right: 1px solid #2d2a24;
}

.qm-cell.is-zhifu {
  box-shadow: inset 0 0 0 1px #9e3327;
  background: #fbf5e8;
}

.qm-cell.is-zhishi {
  box-shadow: inset 0 0 0 1px #746a5a;
}

.qm-cell.is-kong {
  background: #efe6d6;
}

.qm-cell-head {
  position: absolute;
  top: 5px;
  left: 7px;
  right: 7px;
  display: flex;
  justify-content: space-between;
  color: #8a7f6c;
  font-size: 10px;
}

.qm-cell-head em {
  font-style: normal;
  font-size: 9px;
}

.qm-stars {
  display: flex;
  align-items: baseline;
  gap: 7px;
  min-height: 19px;
}

.qm-stars b {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 17px;
  font-weight: 500;
  color: #2d2a24;
}

.qm-stars span {
  color: #9e3327;
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 15px;
}

.qm-stars span.is-quiet {
  color: #8a7f6c;
  font-size: 11px;
  font-family: inherit;
}

.qm-gans {
  display: flex;
  align-items: baseline;
  gap: 4px;
  padding: 2px 0;
}

.qm-heaven {
  color: #9e3327;
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 30px;
  line-height: 1;
}

.qm-earth {
  color: #2d2a24;
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 19px;
  line-height: 1;
}

.qm-gans i {
  color: #8a7f6c;
  font-size: 9px;
  font-style: normal;
}

.qm-cell-foot {
  display: flex;
  flex-wrap: wrap;
  align-items: center;
  gap: 4px;
  margin-top: auto;
  min-height: 16px;
}

.qm-god {
  color: #746a5a;
  font-size: 11px;
}

.qm-flag {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 15px;
  height: 15px;
  border: 1px solid #c9bfae;
  color: #746a5a;
  font-size: 9px;
}

.qm-flag.is-kong {
  border-color: #8a7f6c;
  color: #5d5549;
}

.qm-flag.is-horse {
  border-color: #9e3327;
  color: #9e3327;
}

.qm-flag.is-po {
  border-color: #2d2a24;
  color: #2d2a24;
}

.qm-flag.is-fu {
  border-color: #9e3327;
  background: #9e3327;
  color: #f8f3e8;
}

.qm-flag.is-shi {
  border-color: #2d2a24;
  background: #2d2a24;
  color: #f8f3e8;
}

.qm-sheets {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 10px;
  margin-top: 12px;
}

.qm-sheet {
  border: 1px solid #c9bfae;
  background: #f8f3e8;
}

.qm-sheet header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  padding: 7px 9px;
  border-bottom: 1px solid #c9bfae;
}

.qm-sheet header span {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 15px;
  letter-spacing: 0.14em;
}

.qm-sheet header em {
  color: #8a7f6c;
  font-size: 10px;
  font-style: normal;
}

.qm-sheet-body {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 14px;
  padding: 8px 9px;
  color: #746a5a;
  font-size: 11px;
}

.qm-sheet-body b {
  color: #2d2a24;
  font-weight: 500;
}

@media (max-width: 640px) {
  .qm-board {
    min-width: 470px;
  }

  .qm-grid {
    grid-template-rows: repeat(3, minmax(116px, auto));
  }

  .qm-heaven {
    font-size: 25px;
  }

  .qm-stars b {
    font-size: 14px;
  }

  .qm-stars span {
    font-size: 13px;
  }

  .qm-sheets {
    grid-template-columns: 1fr;
  }
}
</style>
