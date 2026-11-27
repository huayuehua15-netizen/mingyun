<script setup lang="ts">
import { computed } from 'vue'
import type { DaLiuRenPayload, DaLiuRenTianPan } from '../types'

const props = defineProps<{ payload: DaLiuRenPayload }>()

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

const ring = computed(() =>
  props.payload.tian_pan.map((cell: DaLiuRenTianPan) => ({
    ...cell,
    ...(slots[cell.earth] || { row: 1, col: 1 })
  }))
)
</script>

<template>
  <div class="dlr-board">
    <div class="dlr-ring">
      <article
        v-for="cell in ring"
        :key="cell.earth"
        class="dlr-cell"
        :class="{ 'has-sha': (cell.shen_sha || []).length }"
        :style="{ gridRow: cell.row, gridColumn: cell.col }"
      >
        <span class="dlr-earth">{{ cell.earth }}</span>
        <b class="dlr-heaven">{{ cell.heaven }}</b>
        <em class="dlr-jiang">{{ cell.tian_jiang }}</em>
        <span class="dlr-qin">{{ cell.liu_qin }}</span>
        <small v-if="cell.shen_sha?.length" class="dlr-sha">{{ cell.shen_sha?.join('·') }}</small>
      </article>

      <div class="dlr-center">
        <span class="dlr-center-kicker">{{ payload.month_general.zhi }}将 · {{ payload.month_general.name }}</span>
        <b>{{ payload.ke_ti }}</b>
        <em>{{ payload.gate }}课 · {{ payload.hour_zhi }}时</em>
        <div class="dlr-center-chuan">
          <span v-for="c in payload.san_chuan" :key="c.index">{{ c.name }}<strong>{{ c.branch }}</strong></span>
        </div>
      </div>
    </div>

    <div class="dlr-sheets">
      <section class="dlr-sheet">
        <header><span>四课</span><em>{{ payload.day_gan }}日 {{ payload.day_zhi }}支</em></header>
        <ol>
          <li v-for="k in payload.si_ke" :key="k.index">
            <span>{{ k.index }}</span>
            <b>{{ k.upper }}</b>
            <em>上 {{ k.lower }}</em>
            <small>{{ k.relation }}</small>
          </li>
        </ol>
      </section>

      <section class="dlr-sheet">
        <header><span>三传</span><em>{{ payload.gate }} · {{ payload.ke_ti }}</em></header>
        <ol class="dlr-chuan">
          <li v-for="c in payload.san_chuan" :key="c.index">
            <span>{{ c.name }}</span>
            <b>{{ c.branch }}</b>
            <em>{{ c.palace }}宫 · {{ c.tian_jiang }} · {{ c.liu_qin }}</em>
          </li>
        </ol>
      </section>
    </div>

    <footer class="dlr-foot">
      <span>旬空 {{ payload.xun_kong.join('、') }}</span>
      <span v-for="s in payload.shen_sha" :key="s.name">{{ s.value }}</span>
    </footer>
  </div>
</template>

<style scoped>
.dlr-board {
  width: 100%;
  min-width: 520px;
  color: #2d2a24;
}

.dlr-ring {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  grid-template-rows: repeat(4, minmax(78px, auto));
  border: 1px solid #2d2a24;
  background: #f8f3e8;
}

.dlr-cell {
  position: relative;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 2px;
  padding: 8px 6px 10px;
  border-right: 1px solid #c9bfae;
  border-bottom: 1px solid #c9bfae;
  background: #f8f3e8;
}

.dlr-cell:nth-child(4n) {
  border-right: 0;
}

.dlr-earth {
  position: absolute;
  top: 5px;
  left: 7px;
  color: #8a7f6c;
  font-size: 12px;
  letter-spacing: 0.12em;
}

.dlr-heaven {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 28px;
  line-height: 1;
  color: #9e3327;
}

.dlr-jiang {
  color: #2d2a24;
  font-size: 12px;
  font-style: normal;
}

.dlr-qin {
  color: #746a5a;
  font-size: 11px;
}

.dlr-sha {
  color: #9e3327;
  font-size: 10px;
  letter-spacing: 0.04em;
}

.dlr-cell.has-sha {
  background: #fbf6e9;
}

.dlr-center {
  grid-area: 2 / 2 / 4 / 4;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  gap: 6px;
  padding: 12px;
  border-right: 1px solid #c9bfae;
  border-bottom: 1px solid #c9bfae;
  background: #f1eadb;
  text-align: center;
}

.dlr-center-kicker {
  color: #8a7f6c;
  font-size: 12px;
  letter-spacing: 0.08em;
}

.dlr-center b {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 28px;
  letter-spacing: 0.14em;
  color: #2d2a24;
}

.dlr-center em {
  color: #9e3327;
  font-size: 12px;
  font-style: normal;
}

.dlr-center-chuan {
  display: flex;
  gap: 12px;
  margin-top: 4px;
  color: #746a5a;
  font-size: 11px;
}

.dlr-center-chuan strong {
  margin-left: 3px;
  color: #2d2a24;
  font-size: 16px;
  font-weight: 500;
}

.dlr-sheets {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 14px;
  margin-top: 14px;
}

.dlr-sheet {
  border: 1px solid #c9bfae;
  background: #f8f3e8;
}

.dlr-sheet header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 8px;
  padding: 8px 10px;
  border-bottom: 1px solid #c9bfae;
}

.dlr-sheet header span {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 16px;
  letter-spacing: 0.16em;
}

.dlr-sheet header em {
  color: #8a7f6c;
  font-size: 10px;
  font-style: normal;
}

.dlr-sheet ol {
  margin: 0;
  padding: 0;
  list-style: none;
}

.dlr-sheet li {
  display: grid;
  grid-template-columns: 18px 28px 1fr auto;
  align-items: center;
  gap: 6px;
  padding: 7px 10px;
  border-bottom: 1px solid #e3dacb;
  font-size: 12px;
}

.dlr-sheet li:last-child {
  border-bottom: 0;
}

.dlr-sheet li span {
  color: #8a7f6c;
}

.dlr-sheet li b {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 20px;
  font-weight: 500;
  color: #9e3327;
}

.dlr-sheet li em {
  color: #746a5a;
  font-style: normal;
}

.dlr-sheet li small {
  color: #8a7f6c;
  font-size: 10px;
}

.dlr-chuan li {
  grid-template-columns: 38px 28px 1fr;
}

.dlr-chuan li span {
  color: #9e3327;
}

.dlr-foot {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 14px;
  margin-top: 12px;
  color: #746a5a;
  font-size: 11px;
}

@media (max-width: 640px) {
  .dlr-board {
    min-width: 460px;
  }

  .dlr-ring {
    grid-template-rows: repeat(4, minmax(66px, auto));
  }

  .dlr-heaven {
    font-size: 23px;
  }

  .dlr-center b {
    font-size: 22px;
  }

  .dlr-sheets {
    grid-template-columns: 1fr;
  }
}
</style>
