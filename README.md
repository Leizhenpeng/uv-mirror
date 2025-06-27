# GitHub Proxy

一个用 Golang 实现的 GitHub 文件加速代理服务。

## 功能特性

- 支持 GitHub release、archive 以及项目文件的加速下载
- 支持 git clone 加速  
- 支持所有 GitHub 相关域名代理
- 提供美观的 Web 界面

## 使用方法

### 直接运行

```bash
# 克隆项目
git clone https://github.com/your-username/gh-proxy-go.git
cd gh-proxy-go

# 运行
go run main.go
```

### 编译运行

```bash
# 编译
go build -o gh-proxy main.go

# 运行
./gh-proxy
```

## 使用示例

### Git Clone 加速

```bash
# 原地址
git clone https://github.com/user/repo.git

# 代理地址
git clone https://your-proxy-domain/https://github.com/user/repo.git
```

### 文件下载加速

```bash
# 原地址
https://github.com/user/repo/releases/download/v1.0.0/file.zip

# 代理地址  
https://your-proxy-domain/https://github.com/user/repo/releases/download/v1.0.0/file.zip
```

## 环境变量

| 变量名 | 默认值 | 说明 |
|--------|--------|------|
| `PORT` | `8080` | 服务监听端口 |

## 许可证

MIT License 