# 使用示例

## 快速开始

```bash
# 启动服务
go run main.go

# 或者编译后运行
go build -o gh-proxy main.go
./gh-proxy
```

服务启动后访问 http://localhost:8080

## 使用示例

### 1. 文件下载加速

原始链接：
```
https://github.com/golang/go/archive/master.zip
```

加速链接：
```
http://localhost:8080/https://github.com/golang/go/archive/master.zip
```

### 2. Raw 文件加速

原始链接：
```
https://raw.githubusercontent.com/golang/go/master/README.md
```

加速链接：
```
http://localhost:8080/https://raw.githubusercontent.com/golang/go/master/README.md
```

### 3. Git Clone 加速

原始命令：
```bash
git clone https://github.com/golang/go.git
```

加速命令：
```bash
git clone http://localhost:8080/https://github.com/golang/go.git
```

### 4. Release 文件下载

原始链接：
```
https://github.com/golang/go/releases/download/go1.21.0/go1.21.0.linux-amd64.tar.gz
```

加速链接：
```
http://localhost:8080/https://github.com/golang/go/releases/download/go1.21.0/go1.21.0.linux-amd64.tar.gz
```

## 配置选项

通过环境变量配置：

```bash
# 自定义端口
PORT=3000 go run main.go

# 启用 jsdelivr CDN 加速
JSDELIVR=true go run main.go

# 启用白名单，只允许特定仓库
WHITELIST="golang/go,microsoft/vscode,facebook/react" go run main.go

# 组合配置
PORT=3000 JSDELIVR=true WHITELIST="golang/go,microsoft/vscode" go run main.go
```

### 白名单功能

白名单功能可以限制只有特定的 GitHub 仓库能够被代理访问，这对于控制服务使用范围很有用：

```bash
# 只允许 Go 相关仓库
WHITELIST="golang/go,golang/tools,golang/example" go run main.go

# 只允许前端框架
WHITELIST="facebook/react,vuejs/vue,angular/angular" go run main.go

# 支持大小写不敏感匹配
WHITELIST="Microsoft/VSCode,GOLANG/GO" go run main.go
```

**白名单格式说明：**
- 格式：`用户名/仓库名`
- 多个仓库用逗号分隔
- 支持大小写不敏感匹配
- 如果不设置 `WHITELIST`，则允许所有仓库访问

## 功能验证

测试服务是否正常：
```bash
# 测试首页
curl http://localhost:8080/

# 测试代理功能
curl -L "http://localhost:8080/https://raw.githubusercontent.com/golang/go/master/README.md"

# 测试 robots.txt
curl http://localhost:8080/robots.txt
```

## 支持的 GitHub 域名

- github.com
- api.github.com  
- raw.githubusercontent.com
- gist.githubusercontent.com
- github.githubassets.com
- avatars.githubusercontent.com
- camo.githubusercontent.com
- user-images.githubusercontent.com

所有这些域名的请求都会被正确代理。 