package config

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"xuanwu/lib/pathutil"

	"github.com/tidwall/gjson"
)

var (
	IsWindows = runtime.GOOS == "windows"
	Version   = "0.0.0"
)

// 将config文件读取到json字符串
func ReadConfigFileToJson() (gjson.Result, error) {
	configPath := pathutil.GetDataPath("config.json")
	jsonByte, err := os.ReadFile(configPath)
	if err != nil {
		fmt.Println("配置文件读取失败")
		/* 配置文件不存在,创建json文件 */
		str := fmt.Sprintf(`{
			"name": "xuanwu",
			"username":"admin",
			"password":"8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918",
			"cookie_expire_days": 30,
			"log_clean_days": 7,
			"task": []
		  }`)
		err := WriteConfigFile(configPath, []byte(str))
		if err != nil {
			log.Println("配置文件创建失败")
			return gjson.Parse(""), err
		}
		log.Println("配置文件创建成功")
		return gjson.Parse(str), nil
	}

	return gjson.Parse(string(jsonByte)), nil
}

// 写入json到config文件
func WriteConfigFile(filePath string, data []byte) error {
	if err := pathutil.EnsureFile(filePath); err != nil {
		fmt.Println("config文件创建失败")
		return err
	}

	// 解析JSON以验证格式
	var prettyJSON bytes.Buffer
	if err := json.Indent(&prettyJSON, data, "", "    "); err != nil {
		fmt.Println("JSON格式化失败")
		return err
	}

	f, err := os.OpenFile(filePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		fmt.Println("config文件打开失败")
		return err
	}
	defer f.Close()

	_, err = f.Write(prettyJSON.Bytes())
	if err != nil {
		fmt.Println("config文件写入失败")
		return err
	}

	return nil
}
