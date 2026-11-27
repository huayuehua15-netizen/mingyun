package store

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

type filePayload struct {
	Version int      `json:"version"`
	Records []Record `json:"records"`
}

// File 在内存索引之上把盘面快照落成一份 JSON 文件。
//
// 这是个人自部署的持久化层，不是多实例数据库：写入时串行拍快照、
// 临时文件 + rename 原子替换。数据量变大后应迁移到 Postgres。
type File struct {
	*Memory
	path    string
	writeMu sync.Mutex
}

// NewFile 打开或创建一份文件存储。文件不存在时返回空存储。
func NewFile(path string) (*File, error) {
	if path == "" {
		return nil, fmt.Errorf("文件存储路径不能为空")
	}
	f := &File{Memory: NewMemory(), path: path}
	if err := f.load(); err != nil {
		return nil, err
	}
	return f, nil
}

func (f *File) load() error {
	data, err := os.ReadFile(f.path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("读取盘面文件 %s: %w", f.path, err)
	}
	if len(data) == 0 {
		return fmt.Errorf("盘面文件 %s 为空", f.path)
	}
	var payload filePayload
	if err := json.Unmarshal(data, &payload); err != nil {
		return fmt.Errorf("解析盘面文件 %s: %w", f.path, err)
	}
	if payload.Version != 0 && payload.Version != 1 {
		return fmt.Errorf("盘面文件 %s 版本 %d 不受支持", f.path, payload.Version)
	}
	for _, rec := range payload.Records {
		if rec.ChartID == "" {
			return fmt.Errorf("盘面文件 %s 含空 chart_id", f.path)
		}
		if rec.CreatedAt.IsZero() {
			rec.CreatedAt = time.Now()
		}
		if err := f.Memory.Save(rec); err != nil {
			return fmt.Errorf("载入盘面 %s: %w", rec.ChartID, err)
		}
	}
	return nil
}

// Save 先写入内存索引，再原子刷新整份文件。
func (f *File) Save(rec Record) error {
	if err := f.Memory.Save(rec); err != nil {
		return err
	}
	if err := f.persist(); err != nil {
		f.Memory.remove(rec.ChartID)
		return fmt.Errorf("盘面落盘失败: %w", err)
	}
	return nil
}

func (f *File) persist() error {
	f.writeMu.Lock()
	defer f.writeMu.Unlock()

	payload := filePayload{Version: 1, Records: f.Memory.snapshot()}
	data, err := json.MarshalIndent(payload, "", "  ")
	if err != nil {
		return err
	}
	dir := filepath.Dir(f.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("创建数据目录 %s: %w", dir, err)
	}
	if err := os.Chmod(dir, 0o700); err != nil {
		return fmt.Errorf("收紧数据目录 %s: %w", dir, err)
	}
	tmp, err := os.CreateTemp(dir, ".charts-*.tmp")
	if err != nil {
		return fmt.Errorf("创建盘面临时文件: %w", err)
	}
	tmpName := tmp.Name()
	defer os.Remove(tmpName)
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("写入盘面临时文件: %w", err)
	}
	if err := tmp.Chmod(0o600); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("设置盘面文件权限: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return fmt.Errorf("关闭盘面临时文件: %w", err)
	}
	if err := os.Rename(tmpName, f.path); err != nil {
		return fmt.Errorf("替换盘面文件: %w", err)
	}
	return nil
}
