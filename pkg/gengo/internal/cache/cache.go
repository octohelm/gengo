// Package cache 为 gengo 提供全局生成缓存，按 actionID 存储时间戳。
//
// 缓存位于 $GOCACHE/gengo 目录中，文件布局为 {id[0:2]}/{id}，
// 文件内容为 8 字节小端 int64 时间戳。
package cache

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Cache 按 actionID 在文件系统上存储生成标记。
type Cache struct {
	dir string
}

// New 使用默认目录（$GOCACHE/gengo）创建缓存实例。
// 若 GOCACHE 未设置，回退到 $UserCacheDir/go-build/gengo。
func New() (*Cache, error) {
	dir := os.Getenv("GOCACHE")
	if dir == "" {
		d, err := os.UserCacheDir()
		if err != nil {
			return nil, fmt.Errorf("cache: 无法确定缓存目录: %w", err)
		}
		dir = filepath.Join(d, "go-build")
	}
	return NewWithDir(filepath.Join(dir, "gengo")), nil
}

// NewWithDir 使用指定目录创建缓存实例。
func NewWithDir(dir string) *Cache {
	return &Cache{dir: dir}
}

// Exists 返回 actionID 对应的缓存条目是否已存在。
func (c *Cache) Exists(actionID string) bool {
	_, err := os.Stat(c.filePath(actionID))
	return err == nil
}

// Mark 为 actionID 写入当前时间戳作为缓存标记。
func (c *Cache) Mark(actionID string) error {
	fp := c.filePath(actionID)

	if err := os.MkdirAll(filepath.Dir(fp), os.ModePerm); err != nil {
		return fmt.Errorf("cache: 创建缓存目录失败: %w", err)
	}

	tmp, err := os.CreateTemp(filepath.Dir(fp), ".tmp-*")
	if err != nil {
		return fmt.Errorf("cache: 创建临时文件失败: %w", err)
	}
	tmpName := tmp.Name()

	buf := make([]byte, 8)
	now := timeNow()
	binary.LittleEndian.PutUint64(buf, uint64(now))

	if _, err := tmp.Write(buf); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("cache: 写入缓存失败: %w", err)
	}
	tmp.Close()

	if err := os.Rename(tmpName, fp); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("cache: 重命名缓存文件失败: %w", err)
	}

	return nil
}

// filePath 返回 actionID 对应的缓存文件路径（{id[0:2]}/{id}）。
func (c *Cache) filePath(actionID string) string {
	return filepath.Join(c.dir, actionID[:2], actionID)
}

// timeNow 返回当前 UnixNano 时间戳，提取便于测试替换。
var timeNow = func() int64 { return time.Now().UnixNano() }
