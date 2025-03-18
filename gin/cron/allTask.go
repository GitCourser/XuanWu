package cron

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"xuanwu/config"
	r "xuanwu/gin/response"
	"xuanwu/lib/pathutil"
	mycron "xuanwu/xuanwu"

	"github.com/gin-gonic/gin"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

// 放处理过的json字符串
type JsonParams struct {
	data string
}

/* 给json设置值 */
func (j *JsonParams) Set(key string, value interface{}) {
	j.data, _ = sjson.Set(j.data, key, value)
}

/*
添加或更新任务
*/
func HandlerAddTask(c *gin.Context) {
	// 声明map结构
	var jsonData map[string]interface{}
	// 解析请求体到map
	if err := c.BindJSON(&jsonData); err != nil {
		r.ErrMesage(c, "请求参数错误")
		return
	}
	// 获取name参数
	name := jsonData["name"].(string)
	if name == "" {
		r.ErrMesage(c, "任务名称不能为空")
		return
	}

	times := jsonData["times"]
	if times == nil {
		r.ErrMesage(c, "任务类型不能为空")
		return
	}
	workdir := jsonData["workdir"]
	if workdir == nil {
		r.ErrMesage(c, "工作目录不能为空")
		return
	}
	exec := jsonData["exec"]
	if exec == nil {
		r.ErrMesage(c, "执行命令不能为空")
		return
	}
	jp := &JsonParams{data: ""}
	jp.Set("name", name)
	jp.Set("times", jsonData["times"])
	jp.Set("workdir", jsonData["workdir"])
	jp.Set("exec", jsonData["exec"])
	jp.Set("enable", jsonData["enable"])

	cfg, err := config.ReadConfigFileToJson()
	if err != nil {
		log.Println("读取配置文件出错")
		return
	}

	// 检查任务是否已存在
	isUpdate := false
	result := gjson.Get(cfg.Raw, "task.#.name")
	for i, isname := range result.Array() {
		if isname.String() == name {
			isUpdate = true
			// 更新配置文件
			jp := &JsonParams{data: cfg.Raw}
			jp.Set(fmt.Sprintf("task.%v.times", i), times)
			jp.Set(fmt.Sprintf("task.%v.workdir", i), workdir)
			jp.Set(fmt.Sprintf("task.%v.exec", i), exec)
			jp.Set(fmt.Sprintf("task.%v.enable", i), jsonData["enable"])
			configPath := pathutil.GetConfigPath()
			err := config.WriteConfigFile(configPath, []byte(jp.data))
			if err != nil {
				r.ErrMesage(c, "更新失败,配置文件写入失败")
				return
			}
			break
		}
	}

	if !isUpdate {
		// 添加新任务
		var newObj map[string]interface{}
		json.Unmarshal([]byte(jp.data), &newObj)
		value, _ := sjson.Set(cfg.Raw, "task.-1", newObj)
		configPath := pathutil.GetConfigPath()
		err = config.WriteConfigFile(configPath, []byte(value))
		if err != nil {
			r.ErrMesage(c, "添加失败,配置文件写入失败")
			return
		}
	}

	// 如果enable为true，启用任务
	if enable, ok := jsonData["enable"].(bool); ok && enable {
		// 先禁用任务（如果存在）
		for _, e := range mycron.C.Entries() {
			if taskInfo, exists := mycron.TaskData[e.ID]; exists && taskInfo.Name == name {
				mycron.C.Remove(e.ID)
				// 关闭日志文件
				if taskInfo.Writer != nil {
					taskInfo.Writer.Close()
				}
				// 停止写入日志
				taskInfo.Log.SetOutput(io.Discard)
				// 从映射表中删除
				delete(mycron.TaskData, e.ID)
			}
		}

		// 添加到cron
		TaskData := mycron.TaskInfo{
			Name: name,
			Times: func() []string {
				timesArray, ok := times.([]interface{})
				if !ok {
					return []string{}
				}
				var result []string
				for _, t := range timesArray {
					if str, ok := t.(string); ok {
						result = append(result, str)
					}
				}
				return result
			}(),
			WorkDir: workdir.(string),
			Exec:    exec.(string),
			Enable:  true,
		}
		mycron.AddRunFunc(TaskData)
	}

	if isUpdate {
		r.OkMesage(c, "更新成功")
	} else {
		r.OkMesage(c, "添加成功")
	}
}

/* 删除任务源 */
func HandlerDeleteTask(c *gin.Context) {
	name := c.Query("name")
	if name == "" {
		r.ErrMesage(c, "任务名称不能为空")
		return
	}
	cfg, err := config.ReadConfigFileToJson()
	if err != nil {
		log.Println("读取配置文件出错")
		return
	}
	result := gjson.Get(cfg.Raw, "task.#.name")
	for i, isname := range result.Array() {
		if isname.String() == name {
			value, _ := sjson.Delete(cfg.Raw, fmt.Sprintf("task.%v", i))
			configPath := pathutil.GetConfigPath()
			err := config.WriteConfigFile(configPath, []byte(value))
			if err != nil {
				r.ErrMesage(c, "删除失败,配置文件写入失败")
				return
			}
			r.OkMesage(c, "删除成功")
			return
		}
	}
	r.ErrMesage(c, "删除失败,任务不存在")
}
