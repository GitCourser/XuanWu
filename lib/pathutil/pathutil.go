package pathutil

import (
	"os"
	"path/filepath"
)

const (
	DATA_DIR   = "data"
	LOG_DIR    = "logs"
	CONFIG_DIR = "config"
)

var (
	executablePath string
	rootDir        string
)

func init() {
	var err error
	executablePath, err = os.Executable()
	if err != nil {
		panic("无法获取可执行文件路径: " + err.Error())
	}
	rootDir = filepath.Dir(executablePath)
}

// GetExecutablePath 获取可执行文件路径
func GetExecutablePath() string {
	return executablePath
}

// GetRootDir 获取根目录
func GetRootDir() string {
	return rootDir
}

// GetDataPath 获取数据目录下的路径
func GetDataPath(subPath string) string {
	return filepath.Join(rootDir, DATA_DIR, subPath)
}

// GetLogPath 获取日志文件路径
func GetLogPath(filename string) string {
	return filepath.Join(rootDir, DATA_DIR, LOG_DIR, filename)
}

// GetConfigPath 获取配置文件路径
func GetConfigPath(filename string) string {
	return filepath.Join(rootDir, DATA_DIR, CONFIG_DIR, filename)
}

// EnsureDir 确保目录存在
func EnsureDir(path string) error {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return os.MkdirAll(path, 0755)
	}
	return nil
}

// EnsureFile 确保文件存在
func EnsureFile(path string) error {
	dir := filepath.Dir(path)
	if err := EnsureDir(dir); err != nil {
		return err
	}
	
	if _, err := os.Stat(path); os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return err
		}
		defer file.Close()
	}
	return nil
}

// IsFileExist 检查文件是否存在
func IsFileExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil || !os.IsNotExist(err)
} 