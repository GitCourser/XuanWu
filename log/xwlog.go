package xwlog

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sync"
	"time"
	"xuanwu/lib/pathutil"
)

// LogConfig 日志配置
type LogConfig struct {
	TaskLogFormat bool // true: 任务日志格式(只在第一行显示时间), false: 标准日志格式
}

// TaskLogWriter 导出的接口
type TaskLogWriter interface {
	io.WriteCloser
	SetStartTime(time.Time)
}

// taskLogWriter 实现
type taskLogWriter struct {
	file      *os.File
	lastTime  time.Time // 记录上次写入时间
	startTime time.Time // 记录任务开始时间
	mu        sync.Mutex // 保护并发访问
}

// SetStartTime 设置任务开始时间
func (w *taskLogWriter) SetStartTime(t time.Time) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.startTime = t
	w.lastTime = time.Time{} // 重置lastTime以确保写入新的时间头
}

func (w *taskLogWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 只在第一次写入时输出时间戳
	if w.lastTime.IsZero() {
		// 写入一个空行和时间头
		timeHeader := "\n" + w.startTime.Format("2006-01-02 15:04:05") + "\n\n"
		if _, err := w.file.WriteString(timeHeader); err != nil {
			return 0, err
		}
		w.lastTime = w.startTime
	}

	return w.file.Write(p)
}

// Close 实现io.Closer接口
func (w *taskLogWriter) Close() error {
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

func LogInit(name string) (*log.Logger, io.WriteCloser) {
	return LogInitWithConfig(name, &LogConfig{TaskLogFormat: false})
}

func LogInitWithConfig(name string, config *LogConfig) (*log.Logger, io.WriteCloser) {
	if name == "" { //没有名称时候,返回空日志
		return log.New(os.Stdout, "", 0), nil
	}

	logPath := pathutil.GetLogPath(name)

	// 确保日志目录存在
	if err := pathutil.EnsureDir(pathutil.GetDataPath(pathutil.LOG_DIR)); err != nil {
		log.Fatalf("创建日志目录失败: %v", err)
	}

	// 确保日志文件存在
	if err := pathutil.EnsureFile(logPath); err != nil {
		log.Printf("创建日志文件失败: %v", err)
		return nil, nil
	}

	file, err := os.OpenFile(logPath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatal(err)
	}

	var writer io.WriteCloser = file
	var flags = log.LstdFlags

	if config.TaskLogFormat && name != "main.log" {
		// 对于任务日志，使用自定义writer
		writer = &taskLogWriter{
			file:      file,
			startTime: time.Now(), // 初始化时记录开始时间
		}
		flags = 0 // 不需要标准日志前缀
	}

	logger := log.New(writer, "", flags)
	return logger, writer
}

// CleanLogs 清理过期日志内容
func CleanLogs(cleanDays int) error {
	if cleanDays <= 0 {
		return fmt.Errorf("清理天数必须大于0")
	}

	log.Printf("开始清理日志，清理天数: %d", cleanDays)
	logDir := pathutil.GetDataPath(pathutil.LOG_DIR)
	files, err := ioutil.ReadDir(logDir)
	if err != nil {
		return fmt.Errorf("读取日志目录失败: %v", err)
	}

	cutoffTime := time.Now().AddDate(0, 0, -cleanDays)
	log.Printf("清理截止时间: %v", cutoffTime.Format("2006-01-02 15:04:05"))

	// 匹配日期格式的正则表达式
	// dateRegex := regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}`)
	// 匹配分隔符的正则表达式（日期行）
	splitRegex := regexp.MustCompile(`(?m)^(\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2})`)

	for _, file := range files {
		// 跳过main.log
		if file.Name() == "main.log" {
			continue
		}

		// log.Printf("处理文件: %s", file.Name())
		filePath := filepath.Join(logDir, file.Name())
		content, err := ioutil.ReadFile(filePath)
		if err != nil {
			log.Printf("读取文件失败[%s]: %v", file.Name(), err)
			continue
		}

		// 查找所有日期位置
		matches := splitRegex.FindAllIndex(content, -1)
		if len(matches) == 0 {
			log.Printf("文件[%s]中未找到符合格式的日期", file.Name())
			continue
		}

		var newContent []byte
		var hasExpired bool
		lastEnd := 0

		for i, match := range matches {
			start := match[0]
			end := match[1]
			timeStr := string(content[start:end])

			// 解析时间
			logTime, err := time.Parse("2006-01-02 15:04:05", timeStr)
			if err != nil {
				log.Printf("解析时间失败[%s]: %v, timeStr: %s", file.Name(), err, timeStr)
				continue
			}

			// 确定当前块的结束位置
			blockEnd := len(content)
			if i < len(matches)-1 {
				blockEnd = matches[i+1][0]
			}

			// 检查是否过期
			if logTime.Before(cutoffTime) {
				// log.Printf("发现过期日志块[%s]: %v", file.Name(), timeStr)
				hasExpired = true
				lastEnd = blockEnd
				continue
			}

			// 保留这个块
			if lastEnd < start {
				newContent = append(newContent, content[lastEnd:start]...)
			}
			newContent = append(newContent, content[start:blockEnd]...)
			lastEnd = blockEnd
		}

		// 如果有过期内容，写入新内容
		if hasExpired {
			// log.Printf("准备写入新内容到文件[%s], 新内容长度: %d", file.Name(), len(newContent))
			// 写入文件
			if err := ioutil.WriteFile(filePath, newContent, 0644); err != nil {
				log.Printf("写入文件失败[%s]: %v", file.Name(), err)
				continue
			}
			log.Printf("已清理过期日志内容: %s", file.Name())
		// } else {
		// 	log.Printf("文件[%s]中没有过期内容", file.Name())
		}
	}

	return nil
}
