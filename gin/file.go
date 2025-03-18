package serve

import (
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"time"
	"xuanwu/gin/response"
	"xuanwu/lib/pathutil"

	"github.com/gin-gonic/gin"
)

type FileInfo struct {
	Name      string    `json:"name"`
	Path      string    `json:"path"`
	IsDir     bool      `json:"is_dir"`
	Size      int64     `json:"size"`
	UpdatedAt time.Time `json:"updated_at"`
}

// 验证文件路径是否在DATA_DIR下，返回完整路径，如果不合法返回空字符串
func validatePath(subPath string) string {
	fullPath := pathutil.GetDataPath(subPath)
	if !strings.HasPrefix(fullPath, filepath.Join(pathutil.GetRootDir(), pathutil.DATA_DIR)) {
		return ""
	}
	return filepath.Clean(fullPath)
}

// 获取文件列表
func HandlerFileList(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		path = "."
	}

	fullPath := validatePath(path)
	if fullPath == "" {
		response.ErrMesage(c, "非法路径")
		return
	}

	files, err := os.ReadDir(fullPath)
	if err != nil {
		response.ErrMesage(c, "读取目录失败")
		return
	}

	var fileInfos []FileInfo
	for _, f := range files {
		info, err := f.Info()
		if err != nil {
			continue
		}

		relativePath, err := filepath.Rel(filepath.Join(pathutil.GetRootDir(), pathutil.DATA_DIR), filepath.Join(fullPath, f.Name()))
		if err != nil {
			continue
		}

		fileInfos = append(fileInfos, FileInfo{
			Name:      f.Name(),
			Path:      relativePath,
			IsDir:     f.IsDir(),
			Size:      info.Size(),
			UpdatedAt: info.ModTime(),
		})
	}
	response.OkData(c, fileInfos)
}

// 上传文件
func HandlerFileUpload(c *gin.Context) {
	path := c.PostForm("path")
	if path == "" {
		path = "."
	}

	fullPath := validatePath(path)
	if fullPath == "" {
		response.ErrMesage(c, "非法路径")
		return
	}

	file, err := c.FormFile("file")
	if err != nil {
		response.ErrMesage(c, "获取上传文件失败")
		return
	}

	if err := pathutil.EnsureDir(fullPath); err != nil {
		response.ErrMesage(c, "创建目录失败")
		return
	}

	dst := filepath.Join(fullPath, file.Filename)
	if err := c.SaveUploadedFile(file, dst); err != nil {
		response.ErrMesage(c, "保存文件失败")
		return
	}

	response.OkMesage(c, "上传成功")
}

// 下载文件
func HandlerFileDownload(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.ErrMesage(c, "路径不能为空")
		return
	}

	fullPath := validatePath(path)
	if fullPath == "" {
		response.ErrMesage(c, "非法路径")
		return
	}

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		response.ErrMesage(c, "文件不存在")
		return
	}

	if fileInfo.IsDir() {
		response.ErrMesage(c, "不能下载文件夹")
		return
	}

	file, err := os.Open(fullPath)
	if err != nil {
		response.ErrMesage(c, "打开文件失败")
		return
	}
	defer file.Close()

	fileName := filepath.Base(path)
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+fileName)
	c.Writer.Header().Set("Transfer-Encoding", "chunked")

	chunkSize := 1024 * 1024 // 1MB
	chunks := math.Ceil(float64(fileInfo.Size()) / float64(chunkSize))

	for i := 0; i < int(chunks); i++ {
		data := make([]byte, chunkSize)
		n, err := file.Read(data)
		if err != nil && err != io.EOF {
			c.AbortWithStatus(500)
			return
		}
		c.Data(200, "application/octet-stream", data[:n])
		if err == io.EOF {
			break
		}
	}
}

// 获取文件内容
func HandlerFileContent(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.ErrMesage(c, "路径不能为空")
		return
	}

	fullPath := validatePath(path)
	if fullPath == "" {
		response.ErrMesage(c, "非法路径")
		return
	}

	fileInfo, err := os.Stat(fullPath)
	if err != nil {
		response.ErrMesage(c, "文件不存在")
		return
	}

	if fileInfo.IsDir() {
		response.ErrMesage(c, "不能读取文件夹内容")
		return
	}

	// 限制文件大小为10MB
	if fileInfo.Size() > 10*1024*1024 {
		response.ErrMesage(c, "文件太大")
		return
	}

	content, err := os.ReadFile(fullPath)
	if err != nil {
		response.ErrMesage(c, "读取文件失败")
		return
	}

	response.OkData(c, string(content))
}

// 编辑文件
func HandlerFileEdit(c *gin.Context) {
	var req struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrMesage(c, "无效的请求参数")
		return
	}

	if req.Path == "" {
		response.ErrMesage(c, "路径不能为空")
		return
	}

	fullPath := validatePath(req.Path)
	if fullPath == "" {
		response.ErrMesage(c, "非法路径")
		return
	}

	fileInfo, err := os.Stat(fullPath)
	if err == nil && fileInfo.IsDir() {
		response.ErrMesage(c, "不能编辑文件夹")
		return
	}

	if err := os.WriteFile(fullPath, []byte(req.Content), 0644); err != nil {
		response.ErrMesage(c, "保存文件失败")
		return
	}

	response.OkMesage(c, "保存成功")
}

// 删除文件
func HandlerFileDelete(c *gin.Context) {
	path := c.Query("path")
	if path == "" {
		response.ErrMesage(c, "路径不能为空")
		return
	}

	fullPath := validatePath(path)
	if fullPath == "" {
		response.ErrMesage(c, "非法路径")
		return
	}

	if err := os.RemoveAll(fullPath); err != nil {
		response.ErrMesage(c, "删除失败")
		return
	}

	response.OkMesage(c, "删除成功")
}

// 批量上传文件的结果
type UploadResult struct {
	Name  string `json:"name"`
	Error string `json:"error,omitempty"`
}

// 批量上传文件
func HandlerBatchUpload(c *gin.Context) {
	path := c.PostForm("path")
	if path == "" {
		path = "."
	}

	fullPath := validatePath(path)
	if fullPath == "" {
		response.ErrMesage(c, "非法路径")
		return
	}

	form, err := c.MultipartForm()
	if err != nil {
		response.ErrMesage(c, "获取上传文件失败")
		return
	}

	files := form.File["files[]"]
	if len(files) == 0 {
		response.ErrMesage(c, "没有上传文件")
		return
	}

	if err := pathutil.EnsureDir(fullPath); err != nil {
		response.ErrMesage(c, "创建目录失败")
		return
	}

	var results struct {
		Success []string       `json:"success"`
		Failed  []UploadResult `json:"failed"`
	}

	for _, file := range files {
		dst := filepath.Join(fullPath, file.Filename)
		if err := c.SaveUploadedFile(file, dst); err != nil {
			results.Failed = append(results.Failed, UploadResult{
				Name:  file.Filename,
				Error: "保存失败: " + err.Error(),
			})
		} else {
			results.Success = append(results.Success, file.Filename)
		}
	}

	response.OkData(c, results)
}

// 重命名文件或目录
func HandlerFileRename(c *gin.Context) {
    var req struct {
        Path    string `json:"path"`     // 原路径
        NewPath string `json:"new_path"` // 新路径
    }

    if err := c.ShouldBindJSON(&req); err != nil {
        response.ErrMesage(c, "无效的请求参数")
        return
    }

    if req.Path == "" || req.NewPath == "" {
        response.ErrMesage(c, "路径不能为空")
        return
    }

    // 验证原路径
    oldFullPath := validatePath(req.Path)
    if oldFullPath == "" {
        response.ErrMesage(c, "原路径非法")
        return
    }

    // 验证新路径
    newFullPath := validatePath(req.NewPath)
    if newFullPath == "" {
        response.ErrMesage(c, "新路径非法")
        return
    }

    // 检查原路径是否存在
    if _, err := os.Stat(oldFullPath); err != nil {
        response.ErrMesage(c, "原文件不存在")
        return
    }

    // 检查新路径是否已存在
    if _, err := os.Stat(newFullPath); err == nil {
        response.ErrMesage(c, "目标文件已存在")
        return
    }

    // 执行重命名
    if err := os.Rename(oldFullPath, newFullPath); err != nil {
        response.ErrMesage(c, "重命名失败: "+err.Error())
        return
    }

    response.OkMesage(c, "重命名成功")
}

// 创建文件夹
func HandlerMkdir(c *gin.Context) {
	var req struct {
		Path string `json:"path"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrMesage(c, "无效的请求参数")
		return
	}

	if req.Path == "" {
		response.ErrMesage(c, "路径不能为空")
		return
	}

	fullPath := validatePath(req.Path)
	if fullPath == "" {
		response.ErrMesage(c, "非法路径")
		return
	}

	if err := pathutil.EnsureDir(fullPath); err != nil {
		response.ErrMesage(c, "创建目录失败")
		return
	}

	response.OkMesage(c, "创建成功")
}
