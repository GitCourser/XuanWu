package serve

import (
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"strings"
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

func InitApi(cfg gjson.Result, addApi map[string]string) {
	ApiData := &ApiData{
		Cookie: "", //刷新token
		Port:   "4165",
	}
	ApiData.AddApi = addApi
	if cfg.Get("port").String() != "" {
		ApiData.Port = cfg.Get("port").String()
	}
	ApiData.Init()
}

func (p *ApiData) Init() {

	gin.SetMode(gin.ReleaseMode) // 关闭gin启动时路由打印
	RootRoute := gin.Default()
	p.RootRoute = RootRoute
	RootRoute.Use(p.CookieHandler()) //全局用户认证

	routeApi := RootRoute.Group("/api") // api接口总路由

	// 管理接口
	routeAdmin := routeApi.Group("/user")
	routeAdmin.GET("/profile", p.HandlerGetUserProfile)    // 获取用户配置
	routeAdmin.POST("/profile", p.HandlerUpdateUserProfile) // 更新用户配置

	// 登录接口
	routeAuth := routeApi.Group("/auth")
	routeAuth.POST("/login", p.LoginHandle)
	routeAuth.GET("/logout", p.LogoutHandler)

	// 定时任务接口
	routeCron := routeApi.Group("/cron")
	/* 任务源 */
	routeCron.GET("/list", cron.HandlerTaskList)    //获取任务列表（包含运行状态）
	routeCron.GET("/delete", cron.HandlerDeleteTask)   //删除源任务
	routeCron.POST("/add", cron.HandlerAddTask)        //添加任务源
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
			// 回退到index.html时也添加缓存
			c.Header("Cache-Control", "max-age=31536000, public")
			c.Data(http.StatusOK, "text/html", content)
			return
		}

		// 设置适当的 Content-Type
		if strings.HasSuffix(path, ".html") {
			c.Header("Content-Type", "text/html")
		} else if strings.HasSuffix(path, ".css") {
			c.Header("Content-Type", "text/css")
		} else if strings.HasSuffix(path, ".js") {
			c.Header("Content-Type", "application/javascript")
		}

		// 添加缓存头
		c.Header("Cache-Control", "max-age=31536000, public")
		c.Data(http.StatusOK, c.ContentType(), content)
	})

	fmt.Println("Web 端口：" + p.Port)
	RootRoute.Run(":" + p.Port)
}
