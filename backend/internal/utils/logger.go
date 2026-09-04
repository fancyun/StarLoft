package utils

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"
)

// 日志类型（文件按类型分开，便于查询）
const (
	LogAccess   = "access"   // HTTP 请求日志
	LogBusiness = "business" // 业务/系统日志（标准 log 输出）
	LogError    = "error"    // 错误日志
	LogAdmin    = "admin"    // 管理后台操作日志
)

// 全局日志器（在 InitLoggers 中初始化）
var (
	AccessLogger *log.Logger // 访问日志：access.log
	ErrorLogger  *log.Logger // 错误日志：error.log
	AdminLogger  *log.Logger // 管理操作日志：admin.log
)

const (
	defaultMaxSize = 100 // 单个日志文件最大 100MB
)

// rotatingFileWriter 按大小自动轮转的文件写入器（仅使用标准库实现）。
// 轮转后的历史备份一律保留，不自动删除（文件名带递增序号）。
type rotatingFileWriter struct {
	mu       sync.Mutex
	dir      string
	filename string
	maxSize  int64

	file *os.File
	size int64
}

// newRotatingFileWriter 创建轮转文件写入器
func newRotatingFileWriter(dir, filename string) (*rotatingFileWriter, error) {
	w := &rotatingFileWriter{
		dir:      dir,
		filename: filename,
		maxSize:  defaultMaxSize * 1024 * 1024,
	}
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	if err := w.open(); err != nil {
		return nil, err
	}
	return w, nil
}

// open 打开（或新建）日志文件并记录当前大小
func (w *rotatingFileWriter) open() error {
	path := filepath.Join(w.dir, w.filename)
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return err
	}
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return err
	}
	w.file = file
	w.size = info.Size()
	return nil
}

// nextBackupIndex 计算下一个备份序号（取现有备份中的最大值 + 1），保证历史备份永不删除、永不覆盖
func (w *rotatingFileWriter) nextBackupIndex() int {
	max := 0
	entries, _ := os.ReadDir(w.dir)
	prefix := w.filename + "."
	for _, e := range entries {
		name := e.Name()
		if len(name) <= len(prefix) || name[:len(prefix)] != prefix || e.IsDir() {
			continue
		}
		var idx int
		if _, err := fmt.Sscanf(name[len(prefix):], "%d", &idx); err == nil && idx > max {
			max = idx
		}
	}
	return max + 1
}

// rotate 将当前文件轮转为带递增序号的历史备份，重新打开新文件（历史备份保留）
func (w *rotatingFileWriter) rotate() {
	if w.file != nil {
		w.file.Close()
		w.file = nil
	}

	// 当前文件改为 filename.<递增序号>，保留历史
	base := filepath.Join(w.dir, w.filename)
	_ = os.Rename(base, filepath.Join(w.dir, fmt.Sprintf("%s.%d", w.filename, w.nextBackupIndex())))

	// 重新打开
	_ = w.open()
}

// Write 实现 io.Writer，写入并触发轮转
func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	if err != nil {
		return n, err
	}
	w.size += int64(n)

	if w.size >= w.maxSize {
		w.rotate()
	}
	return n, nil
}

// InitLoggers 初始化文件日志（写入 logDir，类型分开存放）
// 将标准 log 包输出重定向到 business.log，同时创建 access/error/admin 日志器。
// 通过 LOG_DIR 环境变量可覆盖目录（默认 /app/logs）。
func InitLoggers(logDir string) error {
	if logDir == "" {
		logDir = "/app/logs"
	}
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return fmt.Errorf("创建日志目录失败: %w", err)
	}

	// 业务/系统日志（标准 log 输出重定向）
	businessWriter, err := newRotatingFileWriter(logDir, LogBusiness+".log")
	if err != nil {
		return fmt.Errorf("初始化业务日志失败: %w", err)
	}
	log.SetOutput(businessWriter)
	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	// 访问日志
	accessWriter, err := newRotatingFileWriter(logDir, LogAccess+".log")
	if err != nil {
		return fmt.Errorf("初始化访问日志失败: %w", err)
	}
	AccessLogger = log.New(accessWriter, "", log.LstdFlags)

	// 错误日志
	errorWriter, err := newRotatingFileWriter(logDir, LogError+".log")
	if err != nil {
		return fmt.Errorf("初始化错误日志失败: %w", err)
	}
	ErrorLogger = log.New(errorWriter, "ERROR: ", log.LstdFlags)

	// 管理操作日志
	adminWriter, err := newRotatingFileWriter(logDir, LogAdmin+".log")
	if err != nil {
		return fmt.Errorf("初始化管理日志失败: %w", err)
	}
	AdminLogger = log.New(adminWriter, "", log.LstdFlags)

	log.Printf("文件日志已初始化，目录: %s（business/access/error/admin 分文件记录）", logDir)
	return nil
}
