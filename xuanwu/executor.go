package xuanwu

import (
	"bufio"
	"bytes"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
	"xuanwu/config"
	"xuanwu/lib/pathutil"
	xwlog "xuanwu/log"

	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/transform"
)

// 从env.ini文件加载环境变量
func loadEnvFromIni() error {
    envPath := pathutil.GetEnvPath()

    // 检查文件是否存在
    if _, err := os.Stat(envPath); os.IsNotExist(err) {
        return nil // 文件不存在，直接返回
    }

    // 打开文件
    file, err := os.Open(envPath)
    if err != nil {
        return err
    }
    defer file.Close()

    // 创建scanner
    scanner := bufio.NewScanner(file)

    // 逐行读取
    for scanner.Scan() {
        line := strings.TrimSpace(scanner.Text())

        // 跳过空行和注释
        if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
            continue
        }

        // 分割键值对
        parts := strings.SplitN(line, "=", 2)
        if len(parts) != 2 {
            continue // 跳过不符合格式的行
        }

        key := strings.TrimSpace(parts[0])
        value := strings.TrimSpace(parts[1])

        // 设置环境变量
        if key != "" {
            os.Setenv(key, value)
        }
    }

    return scanner.Err()
}

// 处理工作目录路径
func HandleWorkDir(workDir string) string {
	// 如果工作目录为空,则返回data目录
	if workDir == "" {
		return pathutil.GetDataPath("")
	}

	// Windows系统检查盘符
	if config.IsWindows {
		if len(workDir) >= 2 && workDir[1] == ':' {
			return workDir
		}
	} else {
		// Linux/Unix系统检查根目录
		if strings.HasPrefix(workDir, "/") {
			return workDir
		}
	}

	// 相对路径处理
	return pathutil.GetDataPath(workDir)
}

// 检查字符串是否为有效的UTF-8
func isValidUTF8(s string) bool {
	// 简单检查：如果包含常见的UTF-8中文字符或者没有乱码特征，认为是UTF-8
	for _, r := range s {
		if r == '\uFFFD' { // UTF-8解码失败的替换字符
			return false
		}
	}
	return true
}

// Windows环境下的智能编码检测和转换
func detectAndConvertEncoding(data []byte) string {
	// 先尝试UTF-8解码
	if utf8Text := string(data); isValidUTF8(utf8Text) {
		return utf8Text
	}

	// 如果UTF-8解码失败，尝试GBK转UTF-8
	reader := transform.NewReader(bytes.NewReader(data), simplifiedchinese.GBK.NewDecoder())
	var buf bytes.Buffer
	if _, err := io.Copy(&buf, reader); err == nil {
		return buf.String()
	}

	// 如果都失败，返回原始字符串
	return string(data)
}

// 创建一个支持编码转换的Scanner
func newEncodingScanner(reader io.Reader) *bufio.Scanner {
	// Linux/Unix系统统一使用UTF-8，无需转换
	if !config.IsWindows {
		return bufio.NewScanner(reader)
	}

	// Windows环境下使用智能编码检测
	scanner := bufio.NewScanner(reader)
	scanner.Split(func(data []byte, atEOF bool) (advance int, token []byte, err error) {
		// 使用默认的分割函数
		advance, token, err = bufio.ScanLines(data, atEOF)
		if token != nil {
			// 对每行进行智能编码转换
			convertedText := detectAndConvertEncoding(token)
			token = []byte(convertedText)
		}
		return
	})
	return scanner
}

// 执行任务命令
func ExecTask(command string, workDir string, logger *log.Logger) error {
	// 记录开始时间
	startTime := time.Now()

	// 如果logger实现了我们的接口，设置开始时间
	if tw, ok := logger.Writer().(xwlog.TaskLogWriter); ok {
		tw.SetStartTime(startTime)
	}

	// 处理工作目录
	workDir = HandleWorkDir(workDir)

	// 加载环境变量
	if err := loadEnvFromIni(); err != nil {
		logger.Printf("加载环境变量失败: %v\n", err)
	}

	// 创建命令
	var cmd *exec.Cmd
	if config.IsWindows {
		cmd = exec.Command("cmd", "/c", command)
	} else {
		cmd = exec.Command("sh", "-c", command)
	}

	// 设置工作目录
	if workDir != "" {
		cmd.Dir = workDir
	}

	// 获取输出管道
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return err
	}

	// 使用WaitGroup等待所有输出读取完成
	var wg sync.WaitGroup

	// 开始执行命令
	if err := cmd.Start(); err != nil {
		return err
	}

	// 异步读取标准输出
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := newEncodingScanner(stdout)
		for scanner.Scan() {
			logger.Println(scanner.Text())
		}
	}()

	// 异步读取标准错误
	wg.Add(1)
	go func() {
		defer wg.Done()
		scanner := newEncodingScanner(stderr)
		for scanner.Scan() {
			logger.Println(scanner.Text())
		}
	}()

	// 等待命令执行完成
	err = cmd.Wait()

	// 等待所有输出读取完成
	wg.Wait()

	// 计算并输出执行用时
	duration := time.Since(startTime)
	logger.Printf("\n任务完成，用时: %v\n", duration)

	return err
}
