package serve

import (
	"fmt"
	"strings"
	"time"
	r "xuanwu/gin/response"
	"xuanwu/lib"

	"github.com/gin-gonic/gin"
)

// ClearUserToken 清除用户token
func (p *ApiData) ClearUserToken(c *gin.Context) {
	// 获取当前token
	cookie, err := c.Cookie("cookie")
	if err == nil {
		// 将token加入黑名单
		lib.GetTokenBlacklist().AddToBlacklist(cookie)
	}
	
	// 清除cookie
	c.SetCookie("cookie", "", -1, "/", "", false, false)
}

func (p *ApiData) CookieHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		if strings.HasPrefix(c.Request.RequestURI, "/api") {
			cookie, err := c.Cookie("cookie")
			//cookie不存在,用户认证失败
			if c.FullPath() != "/api/auth/login" {
				if err != nil {
					//如果cookie为空,就获取Authorization
					if _, ok := c.Request.Header["Authorization"]; ok {
						// 存在
						cookie = c.Request.Header["Authorization"][0]
					} else {
						//除过login其他都要鉴权
						r.AuthMesage(c)
						c.Abort()
						return
					}
				}
				
				// 检查token是否在黑名单中
				if lib.GetTokenBlacklist().IsBlacklisted(cookie) {
					r.AuthMesage(c)
					c.Abort()
					return
				}
				
				//解密
				username, err := lib.DecryptByAes(cookie)
				if err != nil {
					r.AuthMesage(c)
					c.Abort()
					return
				}
				p.Cookie = string(username)
				p.Token = cookie
			}
		}
		// after request  请求前处理
		c.Next()
	}
}

// 用户登录方法
func (p *ApiData) LoginHandle(c *gin.Context) {
	//定义匿名结构体，字段与json字段对应
	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	//绑定json和结构体
	if err := c.BindJSON(&req); err != nil {
		r.ErrMesage(c, "请求参数错误")
		return
	}
	res := GetUserInfo()
	if res.Username != req.Username { //没有查到用户数据
		r.ErrMesage(c, "用户名错误")
		return
	}
	//密码在设置时候就加密存储,传过来的参数是sha256,直接比较
	if res.Password != req.Password {
		r.ErrMesage(c, "密码错误")
		return
	}
	
	// 生成token时加入时间戳确保唯一性
	tokenStr := fmt.Sprintf("%s_%d", req.Username, time.Now().Unix())
	//加密
	str, _ := lib.EncryptByAes([]byte(tokenStr))
	
	// 使用全局配置的Cookie过期时间
	expireSeconds := GetCookieExpireDays() * 24 * 60 * 60
	
	//设置cookie
	c.SetCookie("cookie", str, expireSeconds, "/", "", false, false)
	
	r.OkMesageData(c, "登录成功", gin.H{
		"token":  str,
		"maxAge": expireSeconds,
	})
}

// 退出登录方法
func (p *ApiData) LogoutHandler(c *gin.Context) {
	p.ClearUserToken(c)
	r.OkMesage(c, "退出登录成功")
}
