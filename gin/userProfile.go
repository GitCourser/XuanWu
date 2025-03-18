package serve

import (
	"encoding/json"
	"log"
	"xuanwu/config"
	r "xuanwu/gin/response"
	"xuanwu/lib"
	"xuanwu/lib/pathutil"
	"xuanwu/xuanwu"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/sjson"
)

// UserProfile 用户配置结构体
type UserProfile struct {
	Username         string `json:"username,omitempty"`          // 用户名
	Password         string `json:"password,omitempty"`          // 密码(SHA256)
	OldPassword      string `json:"old_password,omitempty"`      // 旧密码(SHA256)
	CookieExpireDays int    `json:"cookie_expire_days,omitempty"` // Cookie过期天数
	LogCleanDays     int    `json:"log_clean_days,omitempty"`     // 日志清理天数
}

// UserInfo 用户基本信息
type UserInfo struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// 全局配置缓存
var (
	globalCookieExpireDays = 30 // 默认30天
	globalLogCleanDays     = 7  // 默认7天
)

// InitGlobalConfig 初始化全局配置
func InitGlobalConfig() {
	cfg, err := config.ReadConfigFileToJson()
	if err != nil {
		return
	}

	if days := cfg.Get("cookie_expire_days").Int(); days > 0 {
		globalCookieExpireDays = int(days)
	}

	if days := cfg.Get("log_clean_days").Int(); days > 0 {
		globalLogCleanDays = int(days)
	}
}

// GetCookieExpireDays 获取当前Cookie过期天数
func GetCookieExpireDays() int {
	return globalCookieExpireDays
}

// GetLogCleanDays 获取当前日志清理天数
func GetLogCleanDays() int {
	return globalLogCleanDays
}

/* 从json文件中获取用户信息 公共方法 */
func GetUserInfo() UserInfo {
	cfg, err := config.ReadConfigFileToJson()
	if err != nil {
		log.Println("读取配置文件出错")
		return UserInfo{}
	}
	userInfo := UserInfo{
		Username: cfg.Get("username").String(),
		Password: cfg.Get("password").String(),
	}
	if userInfo.Username == "" { //没有查到用户数据
		return UserInfo{}
	}
	return userInfo
}

// HandlerGetUserProfile 获取用户配置
func (p *ApiData) HandlerGetUserProfile(c *gin.Context) {
	cfg, err := config.ReadConfigFileToJson()
	if err != nil {
		r.ErrMesage(c, "读取配置文件失败")
		return
	}

	profile := UserProfile{
		Username:         cfg.Get("username").String(),
		CookieExpireDays: int(cfg.Get("cookie_expire_days").Int()),
		LogCleanDays:     int(cfg.Get("log_clean_days").Int()),
	}

	if profile.Username == "" {
		r.ErrMesage(c, "获取用户信息失败")
		return
	}

	r.OkData(c, gin.H{
		"profile": profile,
		"version": config.Version,
		"is_windows": config.IsWindows,
	})
}

// HandlerUpdateUserProfile 更新用户配置
func (p *ApiData) HandlerUpdateUserProfile(c *gin.Context) {
	// 先解析为map检查参数名
	var rawMap map[string]interface{}
	if err := c.ShouldBindJSON(&rawMap); err != nil {
		r.ErrMesage(c, "请求参数错误")
		return
	}

	// 有效的参数名集合
	validParams := map[string]bool{
		"username":           true,
		"password":          true,
		"old_password":      true,
		"cookie_expire_days": true,
		"log_clean_days":    true,
	}

	// 检查参数名
	for key := range rawMap {
		if !validParams[key] {
			log.Printf("警告: 检测到无效的参数名: %s", key)
			r.ErrMesage(c, "请求包含无效的参数名")
			return
		}
	}

	// 解析为结构体
	var req UserProfile
	jsonData, _ := json.Marshal(rawMap)
	if err := json.Unmarshal(jsonData, &req); err != nil {
		r.ErrMesage(c, "请求参数错误")
		return
	}

	cfg, err := config.ReadConfigFileToJson()
	if err != nil {
		r.ErrMesage(c, "读取配置文件失败")
		return
	}

	jsonStr := cfg.Raw
	needResetToken := false

	// 更新用户名
	if req.Username != "" {
		// 验证用户名加解密
		encryptedUsername, _ := lib.EncryptByAes([]byte(req.Username))
		decryptedUsername, err := lib.DecryptByAes(encryptedUsername)
		if err != nil || string(decryptedUsername) != req.Username {
			r.ErrMesage(c, "用户名格式错误")
			return
		}
		jsonStr, _ = sjson.Set(jsonStr, "username", req.Username)
		needResetToken = true
	}

	// 更新密码
	if req.Password != "" && req.OldPassword != "" {
		currentPass := cfg.Get("password").String()
		if currentPass != req.OldPassword {
			r.ErrMesage(c, "旧密码错误")
			return
		}

		if req.Password == req.OldPassword {
			r.ErrMesage(c, "新密码不能与旧密码相同")
			return
		}

		jsonStr, _ = sjson.Set(jsonStr, "password", req.Password)
		needResetToken = true
	} else if req.Password != "" {
		r.ErrMesage(c, "请提供旧密码")
		return
	}

	// 更新Cookie过期天数
	if req.CookieExpireDays > 0 {
		jsonStr, _ = sjson.Set(jsonStr, "cookie_expire_days", req.CookieExpireDays)
		globalCookieExpireDays = req.CookieExpireDays
	}

	// 更新日志清理天数
	if req.LogCleanDays > 0 {
		jsonStr, _ = sjson.Set(jsonStr, "log_clean_days", req.LogCleanDays)
		globalLogCleanDays = req.LogCleanDays
		// 更新系统任务中的清理天数
		xuanwu.UpdateLogCleanDays(req.LogCleanDays)
	}

	// 写入配置文件
	configPath := pathutil.GetConfigPath()
	if err := config.WriteConfigFile(configPath, []byte(jsonStr)); err != nil {
		r.ErrMesage(c, "配置文件写入失败")
		return
	}

	// 如果修改了用户名或密码，强制用户重新登录
	if needResetToken {
		p.ClearUserToken(c)
	}

	r.OkMesage(c, "更新成功")
} 