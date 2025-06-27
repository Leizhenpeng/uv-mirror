# GitHub Proxy - Golang 版本

这是一个用 Golang 实现的 GitHub 文件加速代理服务，功能与 [hunshcn/gh-proxy](https://github.com/hunshcn/gh-proxy) 完全一致。

## 功能特性

- ✅ 支持 GitHub release、archive 以及项目文件的加速下载
- ✅ 支持 git clone 加速
- ✅ 支持所有 GitHub 相关域名代理
- ✅ 提供美观的 Web 界面
- ✅ 支持 jsdelivr CDN 加速（可选）
- ✅ 支持文件大小限制
- ✅ 完整的 CORS 支持
- ✅ 支持环境变量配置
- ✅ 支持仓库白名单功能

## 支持的链接类型

- 分支源码：`https://github.com/user/repo/archive/master.zip`
- Release 源码：`https://github.com/user/repo/archive/v1.0.0.tar.gz`
- Release 文件：`https://github.com/user/repo/releases/download/v1.0.0/file.zip`
- 分支文件：`https://github.com/user/repo/blob/master/README.md`
- Commit 文件：`https://github.com/user/repo/blob/commit_hash/filename`
- Raw 文件：`https://raw.githubusercontent.com/user/repo/master/file.txt`
- Gist 文件：`https://gist.githubusercontent.com/user/id/raw/file.py`

## 使用方法

### 1. 直接运行

```bash
# 克隆项目
git clone https://github.com/your-username/gh-proxy-go.git
cd gh-proxy-go

# 运行
go run main.go
```

### 2. 编译运行

```bash
# 编译
go build -o gh-proxy main.go

# 运行
./gh-proxy
```

### 3. 使用 Docker

```bash
# 构建镜像
docker build -t gh-proxy .

# 运行容器
docker run -p 8080:8080 gh-proxy
```

## 环境变量配置

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PORT` | `8080` | 服务监听端口 |
| `JSDELIVR` | `false` | 是否启用 jsdelivr CDN 加速 |
| `CNPMJS` | `false` | 是否启用 cnpmjs 加速 |
| `PREFIX` | `/` | URL 前缀 |
| `ASSET_URL` | `` | 静态资源 URL |
| `WHITELIST` | `` | 白名单仓库列表（逗号分隔，格式：user/repo） |

### 示例

```bash
# 启用 jsdelivr 加速，端口 3000
PORT=3000 JSDELIVR=true go run main.go

# 启用白名单，只允许特定仓库
WHITELIST="golang/go,microsoft/vscode,facebook/react" go run main.go

# 组合配置
PORT=3000 JSDELIVR=true WHITELIST="golang/go,microsoft/vscode" go run main.go
```

## Git Clone 加速

将原来的 GitHub 地址中的 `github.com` 替换为你的代理服务地址：

```bash
# 原地址
git clone https://github.com/user/repo.git

# 代理地址
git clone https://your-proxy-domain/https://github.com/user/repo.git
```

## 文件下载加速

### 方式 1：Web 界面

1. 访问代理服务首页
2. 在输入框中粘贴 GitHub 文件链接
3. 点击"加速下载"或"复制链接"

### 方式 2：直接拼接

在 GitHub 链接前加上你的代理服务地址：

```
# 原地址
https://github.com/user/repo/releases/download/v1.0.0/file.zip

# 代理地址  
https://your-proxy-domain/https://github.com/user/repo/releases/download/v1.0.0/file.zip
```

## 部署

### 传统部署

```bash
# 编译
go build -o gh-proxy main.go

# 后台运行
nohup ./gh-proxy > gh-proxy.log 2>&1 &
```

### Docker 部署

创建 `Dockerfile`：

```dockerfile
FROM golang:1.21-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go build -o gh-proxy main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/

COPY --from=builder /app/gh-proxy .

EXPOSE 8080
CMD ["./gh-proxy"]
```

### 使用 systemd

创建 `/etc/systemd/system/gh-proxy.service`：

```ini
[Unit]
Description=GitHub Proxy Service
After=network.target

[Service]
Type=simple
User=nobody
ExecStart=/path/to/gh-proxy
Restart=always

[Install]
WantedBy=multi-user.target
```

启动服务：

```bash
sudo systemctl enable gh-proxy
sudo systemctl start gh-proxy
```

## API 说明

### 获取代理文件

```
GET /{target_url}
```

示例：
```
GET /https://github.com/user/repo/archive/master.zip
```

### robots.txt

```
GET /robots.txt
```

返回禁止搜索引擎爬取的 robots.txt。

## 注意事项

1. **仅支持 GitHub 相关域名**：出于安全考虑，只代理 GitHub 相关的域名
2. **文件大小限制**：默认限制文件大小为 2GB，超过限制会重定向到原地址
3. **白名单功能**：可通过 `WHITELIST` 环境变量限制只允许特定仓库访问，格式为 `user/repo`，多个仓库用逗号分隔
4. **遵守使用条款**：请遵守 GitHub 的使用条款，不要滥用此服务
5. **生产环境**：生产环境建议添加适当的缓存和限流机制

## 与原版对比

| 功能 | hunshcn/gh-proxy | 此 Golang 版本 |
|------|------------------|----------------|
| 基础代理功能 | ✅ | ✅ |
| Web 界面 | ✅ | ✅ |
| Git Clone 支持 | ✅ | ✅ |
| jsdelivr 加速 | ✅ | ✅ |
| 文件大小限制 | ✅ | ✅ |
| Docker 支持 | ✅ | ✅ |
| 运行时环境 | Node.js/Python | Golang |
| 内存占用 | 较高 | 较低 |
| 性能 | 中等 | 高 |
| 部署复杂度 | 中等 | 简单 |

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！ 