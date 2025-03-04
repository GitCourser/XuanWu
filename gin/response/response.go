package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 请求失败  http.StatusForbidden 403
func ErrMesage(c *gin.Context, errmsg interface{}) {
	c.JSON(http.StatusOK, gin.H{ //请求失败也返回200状态码,显示msg信息
		"code":    1,
		"message": errmsg,
	})
}

// 请求成功 // 返回数据,不需要提示信息
func OkData(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code": 0, //0表示成功
		"data": data,
	})
}

// 有操作提示信息,无数据
func OkMesage(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0, //0表示成功
		"message": data,
	})
}

// 有操作提示信息,有数据
func OkMesageData(c *gin.Context, msg string, data interface{}) {
	c.JSON(http.StatusOK, gin.H{
		"code":    0, //0表示成功
		"message": msg,
		"data":    data,
	})
}

// 程序未安装
func NotInstallMesage(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"code":    10000, //10000表示程序未安装
		"message": "程序未安装",
	})
}

// 权限验证失败
func AuthMesage(c *gin.Context) {
	c.JSON(http.StatusUnauthorized, gin.H{
		"code":    403,
		"message": "权限认证失败",
	})
}

// http.StatusContinue =100客户端上传数据需要请求服务端同意才可以上传显示的状态码。
// http.StatusOK= 200成功连接访问。
// http.StatusFound = 302页面跳转的状态码。
// http.StatusBadRequest = 400非法请求，服务端无法解析。
// http.StatusUnauthorized = 401 权限受限，未通过。
// http.StatusForbidden = 403禁止访问。
// http.StatusNotFound = 404请求页面不存在。
// http.StatusInternalServerError = 500服务器内部错误。
