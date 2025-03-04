package xuanwu

import (
	"log"
	"sync"
	"xuanwu/config"
	xwlog "xuanwu/log"
)

var (
	logCleanDays = 7 // 默认7天
	logCleanLock sync.RWMutex
)

func init() {
	// 初始化日志清理天数
	if cfg, err := config.ReadConfigFileToJson(); err == nil {
		if days := cfg.Get("log_clean_days").Int(); days > 0 {
			logCleanDays = int(days)
		}
	}
}

// 系统任务
var SystemTask = []TaskInfo{
	{
		Name:    "定时清理日志",
		Times:   []string{"@daily"},
		WorkDir: "",
		Exec:    "",
		System:  true,
		Enable:  true,
		Func:    cleanLogsTask,
	},
}

// UpdateLogCleanDays 更新日志清理天数
func UpdateLogCleanDays(days int) {
	if days <= 0 {
		return
	}
	logCleanLock.Lock()
	logCleanDays = days
	logCleanLock.Unlock()
}

// cleanLogsTask 清理过期日志任务
func cleanLogsTask() {
	// 记录任务开始
	log.Printf("定时清理日志")
	
	// 使用当前的清理天数
	logCleanLock.RLock()
	days := logCleanDays
	logCleanLock.RUnlock()

	if err := xwlog.CleanLogs(days); err != nil {
		log.Printf("清理日志失败: %v", err)
		return
	}
}
