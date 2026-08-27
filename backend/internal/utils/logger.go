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
	defaultMaxSize   = 100 // 单个日志文件最大 100MB
	defaultMaxBackup = 10  // 最多保留 10 个备份
)

// rotatingFileWriter 按大小自动轮转的文件写入器（仅使用标准库实现）
type rotatingFileWriter struct {
	mu        sync.Mutex
	dir       string
	filename  string
	maxSize   int64
	maxBackup int

	file *os.File
	size int64
}

// newRotatingFileWriter 创建轮转文件写入器
func newRotatingFileWriter(dir, filename string) (*rotatingFileWriter, error) {
	w := &rotatingFileWriter{
		dir:       dir,
		filename:  filename,
		maxSize:   defaultMaxSize * 1024 * 1024,
		maxBackup: defaultMaxBackup,
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

// rotate 将当前文件轮转为带序号的文件
func (w *rotatingFileWriter) rotate() {
	if w.file != nil {
		w.file.Close()
		w.file = nil
	}

	// 删除最旧的备份，为轮转腾出空间
	base := filepath.Join(w.dir, w.filename)
	oldest := filepath.Join(w.dir, fmt.Sprintf("%s.%d", w.filename, w.maxBackup))
	_ = os.Remove(oldest)

	// 从新到旧依次改名：filename.N-1 -> filename.N
	for i := w.maxBackup - 1; i >= 1; i-- {
		from := filepath.Join(w.dir, fmt.Sprintf("%s.%d", w.filename, i))
		to := filepath.Join(w.dir, fmt.Sprintf("%s.%d", w.filename, i+1))
		if _, err := os.Stat(from); err == nil {
			_ = os.Rename(from, to)
		}
	}

	// 当前文件改名为 filename.1
	_ = os.Rename(base, filepath.Join(w.dir, w.filename+".1"))

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
