<div align="center">
<img src="https://github.com/user-attachments/assets/cacd1e5f-12d9-4389-973c-089ddb2a01eb"><p>

# 玄武
### 跨平台定时任务管理系统
（Docker, Linux, Windows）
</div>

![Image](https://github.com/user-attachments/assets/a8b3193e-0962-4be8-b7c4-452ba267cb77)
![Image](https://github.com/user-attachments/assets/1042f34a-1b19-46d9-bd4a-002428e8f409)

## 功能

- 各种命令行工具
- 在线管理文件
- 在线查看任务日志
- 任务日志按期自动清理
- cron支持秒级扩展

## 安全相关

#### 如果在公网使用记得第一时间修改用户名和密码

## 版本

### docker

镜像中包含 `python 3.11` 和 `nodejs 20` 环境

```sh
docker pull dkcourser/xuanwu
```
建一个目录用于保存数据，挂载路径 `/app/data`，默认端口：4165
```sh
docker run -d \
  -p 4165:4165 \
  -v $PWD/xuanwu:/app/data \
  --name xuanwu \
  dkcourser/xuanwu
```

### 二进制程序

下载：[Releases](https://github.com/GitCourser/xuanwu/releases)  
提供 Linux（amd64，arm64），Windows（amd64）  
脚本环境需要自己安装，[python](https://www.python.org/downloads/windows)，[nodejs](https://nodejs.org/zh-cn/download)，或其他脚本

#### Windows 特殊说明
- 可用 `-hide` 参数隐藏命令窗口（在快捷方式中添加）
- 不支持在软件中设置环境变量

## 配置文件

配置文件在程序数据目录 data/config.json，可手动修改，修改后要重启程序。  
密码为 sha256 加密后的值，可添加 `"port": 12345` 修改默认端口  
示例：
```json
{
    "name": "xuanwu",
    "username": "admin",
    "password": "8c6976e5b5410415bde908bd4dee15dfb167a9c873fc4bb8a81f6f2ab448a918",
    "cookie_expire_days": 30,
    "log_clean_days": 7,
    "task": [
        {
            "enable": true,
            "exec": "dir",
            "name": "test_task_1740128994",
            "times": [
                "0 */1 * * * *"
            ],
            "workdir": ""
        }
    ]
}
```

## 自编译

[前端UI](https://github.com/GitCourser/xuanwu-ui) 构建后将 `dist` 放入后端项目的 `public` 中，也可直接下载构建好的 [Releases](https://github.com/GitCourser/xuanwu-ui/releases)  
后端用 `go 1.24` 编译  
