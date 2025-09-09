package serve

import (
	"fmt"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"xuanwu/config"
	"xuanwu/gin/cron"
	"xuanwu/public"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
)

type ApiData struct {
	Cookie    string // 解析出来的username
	Token     string // 未解析的cookie,也就是token
	RootRoute *gin.Engine
	AddApi    map[string]string
	Port      string
}

// 判断字符串是否为UDS路径（包含路径分隔符且不是纯数字）
func isUDSPath(port string) bool {
	// 如果是纯数字，则认为是端口号
	if _, err := strconv.Atoi(port); err == nil {
		return false
	}
	// 如果包含路径分隔符，则认为是UDS路径
	return strings.Contains(port, "/")
}

func InitApi(cfg gjson.Result, addApi map[string]string) {
	ApiData := &ApiData{
		Cookie: "", //刷新token
		Port:   "4165",
	}
	ApiData.AddApi = addApi

	// 端口配置优先级：环境变量 XW_PORT > 配置文件 port > 默认值 4165
	xwPort := os.Getenv("XW_PORT")
	if xwPort != "" {
		ApiData.Port = xwPort
	} else if cfg.Get("port").String() != "" {
		ApiData.Port = cfg.Get("port").String()
	}

	ApiData.Init()
}

func (p *ApiData) Init() {

	gin.SetMode(gin.ReleaseMode) // 关闭gin启动时路由打印
	RootRoute := gin.Default()
	p.RootRoute = RootRoute
	RootRoute.Use(p.CookieHandler()) //全局用户认证

	// 添加缓存中间件
	RootRoute.Use(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 缓存过滤
		shouldNotCache := strings.HasPrefix(path, "/api") ||
			path == "/" ||
			path == "/index.html"

		// 静态资源长期缓存（1年）
		if !shouldNotCache {
			c.Header("Cache-Control", "public, max-age=31536000, immutable")
		}
	})

	routeApi := RootRoute.Group("/api") // api接口总路由

	// 管理接口
	routeAdmin := routeApi.Group("/user")
	routeAdmin.GET("/profile", p.HandlerGetUserProfile)    // 获取用户配置
	routeAdmin.POST("/profile", p.HandlerUpdateUserProfile) // 更新用户配置

	// 登录接口
	routeAuth := routeApi.Group("/auth")
	routeAuth.POST("/login", p.LoginHandle)
	routeAuth.GET("/logout", p.LogoutHandler)
	routeAuth.GET("/check-default", p.CheckDefaultCredentials) // 检查是否为默认用户名密码

	// 定时任务接口
	routeCron := routeApi.Group("/cron")
	/* 任务源 */
	routeCron.GET("/list", cron.HandlerTaskList)    //获取任务列表（包含运行状态）
	routeCron.GET("/delete", cron.HandlerDeleteTask)   //删除源任务
	routeCron.POST("/add", cron.HandlerAddTask)        //添加任务源
	routeCron.POST("/batch-add", cron.HandlerBatchAddTask) //批量添加任务源
	routeCron.POST("/update", cron.HandlerAddTask)     //更新任务（复用添加接口）
	/* 任务控制 */
	routeCron.GET("/enable", cron.HandlerEnableTask)   //启用任务
	routeCron.GET("/disable", cron.HandlerDisableTask) //禁用任务
	routeCron.POST("/execute", cron.HandlerExecuteTask) //立即执行任务

	// 文件管理接口
	routeFile := routeApi.Group("/file")
	routeFile.GET("/list", HandlerFileList)       // 获取文件列表
	routeFile.POST("/upload", HandlerFileUpload)  // 上传文件
	routeFile.POST("/batch-upload", HandlerBatchUpload) // 批量上传文件
	routeFile.POST("/mkdir", HandlerMkdir)       // 创建文件夹
	routeFile.GET("/download", HandlerFileDownload) // 下载文件
	routeFile.GET("/content", HandlerFileContent) // 获取文件内容
	routeFile.POST("/edit", HandlerFileEdit)     // 编辑文件
	routeFile.GET("/delete", HandlerFileDelete)  // 删除文件
	routeFile.POST("/rename", HandlerFileRename) // 重命名文件

	// 静态文件处理
	distFS, err := fs.Sub(public.Public, "dist")
	if err != nil {
		log.Println("加载后台文件失败,web服务停止")
		return
	}

	// 使用 NoRoute 处理所有非 API 请求
	RootRoute.NoRoute(func(c *gin.Context) {
		// 如果是 API 请求，返回 404
		if strings.HasPrefix(c.Request.URL.Path, "/api") {
			c.Status(http.StatusNotFound)
			return
		}

		// 非 API 请求，尝试提供静态文件
		path := c.Request.URL.Path
		if path == "/" {
			path = "index.html"
		}

		content, err := fs.ReadFile(distFS, strings.TrimPrefix(path, "/"))
		if err != nil {
			// 如果文件不存在，返回 index.html
			content, err = fs.ReadFile(distFS, "index.html")
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
		}

		// 设置适当的 Content-Type
		if strings.HasSuffix(path, ".html") {
			c.Header("Content-Type", "text/html")
		} else if strings.HasSuffix(path, ".css") {
			c.Header("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".js") {
			c.Header("Content-Type", "application/javascript")
		}

		c.Data(http.StatusOK, c.ContentType(), content)
	})

	// 判断是否使用 UDS (Unix Domain Socket)
	if isUDSPath(p.Port) && !config.IsWindows {
		// 使用 UDS 监听
		socketPath := p.Port
		// 删除可能存在的旧socket文件
		os.Remove(socketPath)

		listener, err := net.Listen("unix", socketPath)
		if err != nil {
			log.Printf("UDS 监听失败: %v", err)
			return
		}
		defer listener.Close()

		// 设置socket文件权限为0666，确保nginx/caddy等反向代理可以访问
		if err := os.Chmod(socketPath, 0666); err != nil {
			log.Printf("设置UDS权限失败: %v", err)
			return
		}

		fmt.Println("Web UDS：" + socketPath + " (权限: 0666)")
		log.Printf("Web服务启动，UDS监听：%s (权限: 0666)", socketPath)

		server := &http.Server{Handler: RootRoute}
		if err := server.Serve(listener); err != nil {
			log.Printf("UDS 服务启动失败: %v", err)
		}
	} else {
		// 使用端口监听
		fmt.Println("Web 端口：" + p.Port)
		log.Printf("Web服务启动，端口监听：%s", p.Port)
		RootRoute.Run(":" + p.Port)
	}
}
