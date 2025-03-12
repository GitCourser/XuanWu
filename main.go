package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"xuanwu/config"
	serve "xuanwu/gin"
	xwlog "xuanwu/log"
	"xuanwu/xuanwu"
)

// 添加Windows命令行参数
var hideWindow = flag.Bool("hide", false, "在Windows平台下隐藏命令提示符窗口")

func init() {
	if !config.IsWindows { //windows上设置时区会报错,不设置也会正常显示,linux日志时间会差8小时
		TIME_LOCATION, err := time.LoadLocation("Asia/Shanghai")
		if err != nil {
			log.Printf("time时区设置失败")
			panic(err)
		}
		time.Local = TIME_LOCATION
	}
}

func main() {
	// 监听系统信号
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	flag.Parse()

	// Windows平台特定逻辑
	if config.IsWindows && *hideWindow {
		hideConsoleWindow()
	}

	//初始化日志文件
	_, Writer := xwlog.LogInit("main.log")
	log.SetOutput(Writer) // 设置默认logger

	// 退出时记录日志
	defer func() {
		log.Println("玄武退出")
	}()

	cfg, err := config.ReadConfigFileToJson()
	if err != nil {
		log.Println("读取配置文件出错")
		return
	}
	fmt.Println("玄武启动，版本：v" + config.Version + "，按 Ctrl+C 退出")

	// 初始化全局配置
	serve.InitGlobalConfig()
	//初始化web服务 传递端口
	go serve.InitApi(cfg, nil)
	//初始化定时任务
	go xuanwu.CronInit(cfg)

	fmt.Println(time.Now())
	log.Println("玄武启动, 版本: v" + config.Version)

	<-sigChan
}
