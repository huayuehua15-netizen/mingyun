<script setup lang="ts">
defineEmits<{
  (e: 'cast', method: 'liuyao' | 'daliuren' | 'ziwei' | 'xiaoliuren' | 'meihua' | 'qimen'): void
  (e: 'home'): void
}>()
</script>

<template>
  <section class="codex-view">
    <span class="view-watermark" aria-hidden="true">典</span>
    <header class="scene-bar scene-bar--page">
      <button class="scene-back" @click="$emit('home')">← 山门</button>
      <div class="scene-heading">
        <span>总卷</span>
        <h1>六术</h1>
      </div>
      <div class="codex-mark">典</div>
    </header>

    <p class="codex-intro">
      命云是一间可以走进去的术数书房。六术各有自己的盘与断法：六爻问事，大六壬起课，紫微斗数看命；小六壬、梅花、奇门专门看近期一件具体事。
    </p>

    <div class="codex-grid">
      <article class="codex-card is-open">
        <div class="codex-no">壹</div>
        <span class="codex-state">已开放</span>
        <h2>六爻</h2>
        <p>以三枚铜钱起卦，装纳甲、六亲、六神与世应。盘面由 Go 引擎复算，AI 只负责把事实讲成人话。</p>
        <dl>
          <div><dt>起卦</dt><dd>铜钱六摇 / 手动录入</dd></div>
          <div><dt>断卦</dt><dd>用神 · 旺衰 · 动变 · 应期</dd></div>
          <div><dt>输出</dt><dd>四段式解读，可离线兜底</dd></div>
        </dl>
        <button class="codex-enter" @click="$emit('cast', 'liuyao')">进入卷一 <i>→</i></button>
      </article>

      <article class="codex-card is-open">
        <div class="codex-no">贰</div>
        <span class="codex-state">已开放</span>
        <h2>大六壬</h2>
        <p>以时课起传，重四课三传与十二天将。它不靠铜钱，靠的是起课那一刻的时间与月将。</p>
        <dl>
          <div><dt>起课</dt><dd>时间 · 月将 · 天地盘</dd></div>
          <div><dt>断课</dt><dd>三传 · 四课 · 天将</dd></div>
          <div><dt>输出</dt><dd>四段式解读，可离线兜底</dd></div>
        </dl>
        <button class="codex-enter" @click="$emit('cast', 'daliuren')">进入卷二 <i>→</i></button>
      </article>

      <article class="codex-card is-open">
        <div class="codex-no">叁</div>
        <span class="codex-state">已开放</span>
        <h2>紫微斗数</h2>
        <p>以出生时刻安十二宫，排星曜与四化。它更像一张命盘，而不是一次问答。</p>
        <dl>
          <div><dt>起盘</dt><dd>生辰 · 性别 · 十二宫</dd></div>
          <div><dt>读盘</dt><dd>星曜 · 四化 · 大限</dd></div>
          <div><dt>输出</dt><dd>四段式解读，可同盘追问</dd></div>
        </dl>
        <button class="codex-enter" @click="$emit('cast', 'ziwei')">进入卷三 <i>→</i></button>
      </article>
      <article class="codex-card is-open">
        <div class="codex-no">肆</div>
        <span class="codex-state">已开放</span>
        <h2>小六壬</h2>
        <p>以农历月、日、时落大安、留连、速喜、赤口、小吉、空亡六宫。最适合近期一件可验证的事。</p>
        <dl>
          <div><dt>起课</dt><dd>月日时 · 三宫落点</dd></div>
          <div><dt>断课</dt><dd>六宫属性 · 近期走向</dd></div>
          <div><dt>输出</dt><dd>四段式解读，可同盘追问</dd></div>
        </dl>
        <button class="codex-enter" @click="$emit('cast', 'xiaoliuren')">进入卷四 <i>→</i></button>
      </article>

      <article class="codex-card is-open">
        <div class="codex-no">伍</div>
        <span class="codex-state">已开放</span>
        <h2>梅花易数</h2>
        <p>时间起卦或报数起卦，看本卦、互卦、变卦与体用生克。盘面轻，适合快速判断一件事的走势。</p>
        <dl>
          <div><dt>起卦</dt><dd>时间 / 报数 · 动爻</dd></div>
          <div><dt>断卦</dt><dd>本互变 · 体用生克</dd></div>
          <div><dt>输出</dt><dd>四段式解读，可同盘追问</dd></div>
        </dl>
        <button class="codex-enter" @click="$emit('cast', 'meihua')">进入卷五 <i>→</i></button>
      </article>

      <article class="codex-card is-open">
        <div class="codex-no">陆</div>
        <span class="codex-state">已开放</span>
        <h2>奇门遁甲</h2>
        <p>时家转盘奇门，按节气与符头定局，布九宫三奇六仪、值符值使、九星八门八神。当前只做基础盘，不含格局。</p>
        <dl>
          <div><dt>起局</dt><dd>节气 · 三元 · 九宫</dd></div>
          <div><dt>断局</dt><dd>值符值使 · 用神宫 · 空亡马星</dd></div>
          <div><dt>输出</dt><dd>四段式解读，可同盘追问</dd></div>
        </dl>
        <button class="codex-enter" @click="$emit('cast', 'qimen')">进入卷六 <i>→</i></button>
      </article>
    </div>

    <div class="codex-note">
      <span>当前版本</span>
      <p>六术链路已打通：确定性排盘、盘面 facts、AI SSE 解读、离线摘要、同盘追问与本机留档。小六壬只做月日时起课；梅花只做时间与两组报数；奇门只做拆补法基础转盘，不含超神接气、十干克应与格局。未达专业金标准的部分已在文档标出。</p>
    </div>
  </section>
</template>
