<script setup lang="ts">
import { computed } from 'vue'
import type { MeiHuaHexagram, MeiHuaPayload } from '../types'

const props = defineProps<{ payload: MeiHuaPayload }>()

interface HexColumn {
  key: 'ben' | 'hu' | 'bian'
  label: string
  note: string
  hex: MeiHuaHexagram
}

const columns = computed<HexColumn[]>(() => [
  { key: 'ben', label: '本卦', note: '当下', hex: props.payload.ben_gua },
  { key: 'hu', label: '互卦', note: '过程', hex: props.payload.hu_gua },
  { key: 'bian', label: '变卦', note: '收束', hex: props.payload.bian_gua }
])

function displayLines(hex: MeiHuaHexagram, key: HexColumn['key']) {
  return [...hex.lines].reverse().map((yang, index) => ({
    yang,
    position: 6 - index,
    moving: key === 'ben' && 6 - index === props.payload.moving_line
  }))
}
</script>

<template>
  <div class="mh-board">
    <div class="mh-hexes">
      <article v-for="col in columns" :key="col.key" class="mh-hex" :class="{ 'is-main': col.key === 'ben' }">
        <header>
          <span>{{ col.label }}</span>
          <em>{{ col.note }}</em>
        </header>
        <div class="mh-name">{{ col.hex.name }}</div>
        <div class="mh-lines">
          <div
            v-for="line in displayLines(col.hex, col.key)"
            :key="line.position"
            class="mh-line-row"
            :class="{ 'is-moving': line.moving }"
          >
            <span class="mh-pos">{{ line.position }}</span>
            <div class="mh-line" :class="line.yang ? 'is-yang' : 'is-yin'">
              <i></i>
              <i v-if="!line.yang"></i>
            </div>
            <b v-if="line.moving" class="mh-moving">动</b>
          </div>
        </div>
        <footer>
          <span>上{{ col.hex.upper }}</span>
          <span>下{{ col.hex.lower }}</span>
        </footer>
      </article>
    </div>

    <div class="mh-tiyong">
      <div class="mh-role">
        <span>体</span>
        <b>{{ payload.ti_gua.name }}</b>
        <em>{{ payload.ti_gua.element }}</em>
      </div>
      <div class="mh-relation">{{ payload.relation }}</div>
      <div class="mh-role is-yong">
        <span>用</span>
        <b>{{ payload.yong_gua.name }}</b>
        <em>{{ payload.yong_gua.element }}</em>
      </div>
    </div>

    <footer class="mh-foot">
      <span>第 {{ payload.moving_line }} 爻动</span>
      <span>{{ payload.changed_relation }}</span>
      <span v-if="payload.mode === 'numbers'">报数 {{ (payload.numbers || []).join('、') }}</span>
      <span v-else>{{ payload.lunar_date }} · {{ payload.hour_zhi }}时</span>
    </footer>
  </div>
</template>

<style scoped>
.mh-board {
  width: 100%;
  min-width: 520px;
  color: #2d2a24;
}

.mh-hexes {
  display: grid;
  grid-template-columns: repeat(3, minmax(0, 1fr));
  gap: 10px;
}

.mh-hex {
  border: 1px solid #c9bfae;
  background: #f8f3e8;
  padding: 10px 10px 8px;
}

.mh-hex.is-main {
  border-color: #2d2a24;
  background: #fbf6e9;
}

.mh-hex header {
  display: flex;
  align-items: baseline;
  justify-content: space-between;
  gap: 6px;
  border-bottom: 1px solid #e3dacb;
  padding-bottom: 5px;
}

.mh-hex header span {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 15px;
  letter-spacing: 0.12em;
  color: #9e3327;
}

.mh-hex header em {
  color: #8a7f6c;
  font-size: 10px;
  font-style: normal;
}

.mh-name {
  min-height: 33px;
  padding: 7px 0 5px;
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 20px;
  line-height: 1.2;
  text-align: center;
  letter-spacing: 0.06em;
}

.mh-lines {
  display: flex;
  flex-direction: column;
  gap: 5px;
  padding: 3px 0;
}

.mh-line-row {
  display: grid;
  grid-template-columns: 13px 1fr 18px;
  align-items: center;
  gap: 3px;
  min-height: 17px;
}

.mh-pos {
  color: #8a7f6c;
  font-size: 9px;
  text-align: right;
}

.mh-line {
  display: flex;
  align-items: center;
  gap: 7px;
  height: 11px;
}

.mh-line i {
  display: block;
  height: 3px;
  flex: 1;
  background: #2d2a24;
}

.mh-line.is-yin i {
  flex: 0 0 calc(50% - 4px);
}

.mh-line-row.is-moving .mh-line i {
  background: #9e3327;
}

.mh-moving {
  color: #9e3327;
  font-size: 9px;
  font-weight: 500;
  text-align: center;
}

.mh-hex footer {
  display: flex;
  justify-content: space-between;
  gap: 4px;
  margin-top: 4px;
  padding-top: 5px;
  border-top: 1px solid #e3dacb;
  color: #8a7f6c;
  font-size: 10px;
}

.mh-tiyong {
  display: grid;
  grid-template-columns: 1fr auto 1fr;
  align-items: center;
  gap: 10px;
  margin-top: 12px;
  padding: 9px 12px;
  border: 1px solid #c9bfae;
  background: #f1eadb;
}

.mh-role {
  display: flex;
  align-items: baseline;
  gap: 6px;
}

.mh-role span {
  color: #8a7f6c;
  font-size: 11px;
}

.mh-role b {
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 22px;
  font-weight: 500;
  color: #9e3327;
}

.mh-role em {
  color: #746a5a;
  font-size: 10px;
  font-style: normal;
}

.mh-role.is-yong {
  justify-content: flex-end;
}

.mh-role.is-yong b {
  color: #2d2a24;
}

.mh-relation {
  color: #2d2a24;
  font-family: 'Songti SC', 'STSong', 'SimSun', serif;
  font-size: 15px;
  letter-spacing: 0.1em;
  white-space: nowrap;
}

.mh-foot {
  display: flex;
  flex-wrap: wrap;
  gap: 6px 16px;
  margin-top: 10px;
  color: #746a5a;
  font-size: 11px;
}

@media (max-width: 640px) {
  .mh-board {
    min-width: 480px;
  }

  .mh-name {
    font-size: 17px;
  }

  .mh-role b {
    font-size: 18px;
  }

  .mh-relation {
    font-size: 12px;
  }
}
</style>
