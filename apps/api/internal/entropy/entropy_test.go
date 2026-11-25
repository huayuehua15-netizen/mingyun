package entropy

import (
	"bytes"
	"testing"
	"time"
)

func fixedReader() *bytes.Reader {
	b := make([]byte, 64)
	for i := range b {
		b[i] = byte(i)
	}
	return bytes.NewReader(b)
}

func newTestSource() *Source {
	at := time.Date(2026, 11, 25, 10, 0, 0, 0, time.UTC)
	return NewSourceWithReader(bytes.Repeat([]byte{0x2a}, 32), fixedReader(), func() time.Time { return at })
}

func TestToss(t *testing.T) {
	src := newTestSource()
	toss, err := src.Toss()
	if err != nil {
		t.Fatalf("摇卦失败: %v", err)
	}
	if len(toss.ID) != 24 {
		t.Fatalf("entropy_id 长度应为 24，实为 %d", len(toss.ID))
	}
	if toss.Source != "crypto/rand" {
		t.Fatalf("随机源标识错误: %s", toss.Source)
	}
	for i, b := range toss.Backs {
		if b < 0 || b > 3 {
			t.Fatalf("第%d爻背数越界: %d", i+1, b)
		}
	}
	if got := toss.BacksString(); got == "" {
		t.Fatal("背数串不应为空")
	}

	// 同一个 reader 再摇一次，结果应当不同（ID 与背数都消耗了随机流）。
	second, err := src.Toss()
	if err != nil {
		t.Fatalf("第二次摇卦失败: %v", err)
	}
	if second.ID == toss.ID {
		t.Fatal("两次摇卦不应得到相同 ID")
	}
}

func TestVerify(t *testing.T) {
	src := newTestSource()
	toss, err := src.Toss()
	if err != nil {
		t.Fatalf("摇卦失败: %v", err)
	}
	if !src.Verify(toss) {
		t.Fatal("原始凭证应通过复核")
	}

	tampered := toss
	tampered.Backs[0] = (tampered.Backs[0] + 1) % 4
	if src.Verify(tampered) {
		t.Fatal("改动背数后复核应失败")
	}

	tampered = toss
	tampered.CreatedAt = toss.CreatedAt.Add(time.Second)
	if src.Verify(tampered) {
		t.Fatal("改动时刻后复核应失败")
	}

	other, err := NewSource(bytes.Repeat([]byte{0x11}, 32))
	if err != nil {
		t.Fatal(err)
	}
	if other.Verify(toss) {
		t.Fatal("换密钥后复核应失败")
	}
}

func TestNewSourceRejectsShortSecret(t *testing.T) {
	if _, err := NewSource([]byte("short")); err == nil {
		t.Fatal("过短的密钥应当报错")
	}
}
