<template>
  <div class="ink-scene" aria-hidden="true">
    <svg viewBox="0 0 1440 900" preserveAspectRatio="xMidYMid slice">
      <circle class="ink-sun" cx="1118" cy="188" r="74" />

      <path
        class="ridge ridge-far"
        d="M-80 610 Q 100 430 260 500 T 520 438 T 770 500 T 1020 438 T 1260 500 T 1520 450 L1520 920 L-80 920 Z"
      />
      <path
        class="ridge ridge-mid"
        d="M-80 700 Q 120 560 300 620 T 560 552 T 830 630 T 1100 560 T 1520 640 L1520 920 L-80 920 Z"
      />
      <path
        class="ridge ridge-near"
        d="M-80 820 Q 160 700 360 752 T 680 690 T 980 762 T 1260 706 T 1520 760 L1520 920 L-80 920 Z"
      />

      <g class="ink-birds">
        <path d="M936 308 q 12 -8 24 0 q 12 -8 24 0" />
        <path d="M1010 340 q 9 -6 18 0 q 9 -6 18 0" />
        <path d="M890 368 q 8 -5 16 0 q 8 -5 16 0" />
      </g>

      <path
        class="ink-cloud"
        d="M360 330 q 32 -22 64 0 q 32 -22 64 0 q 32 -22 64 0 M400 330 q 22 18 44 0 q 22 18 44 0"
      />
      <path class="ink-rule" d="M80 780 L1360 780" />
    </svg>
  </div>
</template>
