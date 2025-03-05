<div align="center">
<img src="https://github.com/GitCourser/xuanwu-ui/blob/main/public/logo.png?raw=true"><p>

# 玄武
### 跨平台定时任务管理系统
（Docker, Linux, Windows）
</div>

![image](https://github.com/user-attachments/assets/235a964c-133d-45d4-8911-5861f7ad72ff)

## 功能

- 各种命令行工具
- 在线管理文件
- 在线查看任务日志
- 任务日志按期自动清理
- cron支持秒级扩展

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

## 安全相关

- 单用户系统，默认用户名和密码都是 `admin`
- 如果公网能访问此服务，请务必修改用户名和密码
- 除了在系统设置中更改，也可以在启动程序前直接添加配置文件 `data/config.json`

## 配置文件

配置文件在程序数据目录 `data/config.json`，如果手动修改要重启程序  
密码为 `sha256` 加密后的值，可添加 `"port": 12345` 修改默认端口  
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
