// Package entropy 为摇卦提供不可预测的随机源，并生成可复核的凭证。
//
// 引擎本身不取随机（见 docs/02 6.3），随机量必须由调用方传入。这一层就是
// 调用方：用 crypto/rand 摇六爻，再对"结果 + 服务端密钥"做 HMAC，留下一个
// 可以复算的 entropy_proof。它不是密码学刚需，但对"这卦是不是后台编的"
// 这种质疑，是目前唯一能拿出来的答案。
package entropy

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"time"
)

// Toss 一次摇卦记录。
type Toss struct {
	// ID 本次摇卦的标识，12 字节随机数的十六进制。
	ID string `json:"entropy_id"`
	// Backs 六爻自下而上的背数，取值 0—3。
	Backs [6]int `json:"backs"`
	// Proof 对 ID、背数与时刻的 HMAC-SHA256（十六进制）。
	Proof string `json:"entropy_proof"`
	// Source 随机源标识，固定为 crypto/rand。
	Source string `json:"source"`
	// CreatedAt 摇卦时刻（UTC）。
	CreatedAt time.Time `json:"created_at"`
}

// BacksString 返回 "3,2,1,0,3,1" 形式的背数串，用于展示与复核。
func (t Toss) BacksString() string {
	return fmt.Sprintf("%d,%d,%d,%d,%d,%d", t.Backs[0], t.Backs[1], t.Backs[2], t.Backs[3], t.Backs[4], t.Backs[5])
}

// Source 摇卦随机源。
type Source struct {
	secret []byte
	rnd    io.Reader
	now    func() time.Time
}

// NewSource 用服务端密钥构造随机源。密钥至少 16 字节，建议 32 字节。
func NewSource(secret []byte) (*Source, error) {
	if len(secret) < 16 {
		return nil, fmt.Errorf("熵源密钥至少 16 字节，实为 %d", len(secret))
	}
	return &Source{secret: append([]byte(nil), secret...), rnd: rand.Reader, now: time.Now}, nil
}

// NewSourceWithReader 仅供测试注入确定性的 io.Reader，生产代码不要使用。
func NewSourceWithReader(secret []byte, r io.Reader, now func() time.Time) *Source {
	return &Source{secret: append([]byte(nil), secret...), rnd: r, now: now}
}

// Toss 摇一卦：先取 12 字节定 ID，再取 6 字节，每字节的低三位对应三枚铜钱。
func (s *Source) Toss() (Toss, error) {
	idBytes := make([]byte, 12)
	if _, err := io.ReadFull(s.rnd, idBytes); err != nil {
		return Toss{}, fmt.Errorf("读取随机数: %w", err)
	}
	coinBytes := make([]byte, 6)
	if _, err := io.ReadFull(s.rnd, coinBytes); err != nil {
		return Toss{}, fmt.Errorf("读取随机数: %w", err)
	}

	var backs [6]int
	for i, b := range coinBytes {
		n := 0
		if b&1 != 0 {
			n++
		}
		if b&2 != 0 {
			n++
		}
		if b&4 != 0 {
			n++
		}
		backs[i] = n
	}

	t := Toss{
		ID:        hex.EncodeToString(idBytes),
		Backs:     backs,
		Source:    "crypto/rand",
		CreatedAt: s.now().UTC(),
	}
	t.Proof = s.proof(t.ID, t.Backs, t.CreatedAt)
	return t, nil
}

// Verify 复核凭证。任意字段被改动都会失败。
func (s *Source) Verify(t Toss) bool {
	want := s.proof(t.ID, t.Backs, t.CreatedAt)
	return hmac.Equal([]byte(want), []byte(t.Proof))
}

func (s *Source) proof(id string, backs [6]int, at time.Time) string {
	msg := id + "|" + fmt.Sprintf("%d,%d,%d,%d,%d,%d", backs[0], backs[1], backs[2], backs[3], backs[4], backs[5]) +
		"|" + at.UTC().Format(time.RFC3339Nano)
	mac := hmac.New(sha256.New, s.secret)
	mac.Write([]byte(msg))
	return hex.EncodeToString(mac.Sum(nil))
}
