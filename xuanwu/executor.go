package xuanwu

import (
	"bufio"
	"bytes"
	"io"
	"log"
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

// 将GBK编码的文本转换为UTF8
func convertGBKToUTF8(text string) string {
	// 如果不是Windows系统,直接返回
	if !config.IsWindows {
		return text
	}

	// 创建GBK到UTF8的转换器
	reader := transform.NewReader(strings.NewReader(text), simplifiedchinese.GBK.NewDecoder())
	var buf bytes.Buffer
	_, err := io.Copy(&buf, reader)
	if err != nil {
		return text // 如果转换失败,返回原文
	}
	return buf.String()
}

// 创建一个支持编码转换的Scanner
func newEncodingScanner(reader io.Reader) *bufio.Scanner {
	if config.IsWindows {
		// Windows环境下,创建一个GBK到UTF8的转换器
		utf8Reader := transform.NewReader(reader, simplifiedchinese.GBK.NewDecoder())
		return bufio.NewScanner(utf8Reader)
	}
	return bufio.NewScanner(reader)
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
