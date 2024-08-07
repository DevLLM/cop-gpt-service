<h1 align="center">copilot-gpt4-service</h1>

<p align="center">
⚔️ 将 GitHub Copilot 转换为 ChatGPT
</p>

<p align="center">
简体中文 | <a href="README.md">English</a>
</p>

## 支持的 API

- `GET /`: 首页
- `GET /healthz`: 健康检查
- `GET /v1/models`: 获取模型列表
- `POST /v1/chat/completions`: 对话 API
- `POST /v1/embeddings`: 获取文本向量 API
  - 请注意，此 API 与 OpenAI API 不完全兼容。
    对于 `input` 字段，OpenAI API 接受以下类型：
      - string: 将转换为 embedding 的字符串。
      - array: 将转换为 embedding 的字符串数组。
      - array: 将转换为 embedding 的整数数组。
      - array: 将转换为 embedding 的包含整数的数组的数组。
  
    不幸的是，此服务仅接受前两种类型以及包含字符串的数组的数组。


## 如何使用

1. 安装并启动 copilot-gpt4-service 服务，如本地启动后，API 默认地址为：`http://127.0.0.1:8080`;
2. 获取你的 GitHub 账号 GitHub Copilot Plugin Token（详见下文）；
3. 安装第三方客户端，如：[ChatGPT-Next-Web](https://github.com/ChatGPTNextWeb/ChatGPT-Next-Web)，在设置中填入 copilot-gpt4-service 的 API 地址和 GitHub Copilot Plugin Token，即可使用 GPT-4 模型进行对话。

## 部署方式

### 最佳实践方式

经社区验证和讨论，最佳实践方式为：

1. 本地部署，仅个人使用（推荐）；
2. 自用服务器集成 [ChatGPT-Next-Web](https://github.com/ChatGPTNextWeb/ChatGPT-Next-Web) 部署，服务不公开；
3. 服务器部署，公开但个人使用 (例如多客户端使用场景 [Chatbox](https://github.com/Bin-Huang/chatbox), [OpenCat APP](https://opencat.app/), [ChatX APP](https://apps.apple.com/us/app/chatx-ai-chat-client/id6446304087))。

### 不建议方式

1. 以公共服务的方式提供接口

    多个 Token 在同一个 IP 地址进行请求，容易被判定为异常行为

2. 同客户端 Web(例如 ChatGPT-Next-Web) 以默认 API 以及 API Key 的方式提供公共服务

    同一个 Token 请求频率过高，容易被判定为异常行为

3. Serverless 类型的提供商进行部署

    服务生命周期短，更换 IP 地址频繁，容易被判定为异常行为

4. 其他滥用行为或牟利等行为。

### ⚠️ 非常重要

**非常重要：以上不建议的方式，均可能会导致 GitHub Copilot 被封禁，且封禁后可能无法解封。**

**非常重要：以上不建议的方式，均可能会导致 GitHub Copilot 被封禁，且封禁后可能无法解封。**

**非常重要：以上不建议的方式，均可能会导致 GitHub Copilot 被封禁，且封禁后可能无法解封。**

## 客户端

使用 **copilot-gpt4-service** 服务，需要使用第三方客户端，目前已验证支持以下客户端：

- [ChatGPT-Next-Web](https://github.com/ChatGPTNextWeb/ChatGPT-Next-Web) (推荐)。
- [Chatbox](https://github.com/Bin-Huang/chatbox)：支持 Windows, Mac, Linux 平台。
- [OpenCat APP](https://opencat.app/)：支持 iOS、Mac 平台。
- [ChatX APP](https://apps.apple.com/us/app/chatx-ai-chat-client/id6446304087) ：支持 iOS、Mac 平台。

## 服务端

copilot-gpt4-service 服务的部署方式目前包含 Docker 部署、二进制启动、Kubernetes 部署、源码启动实现。

### 服务配置

可使用命令行参数或环境变量或环境变量配置文件 `config.env` 配置服务（可复制项目根目录 `config.env.example` 为 `config.env` 修改），默认服务配置项如下：

```yaml
HOST=0.0.0.0 # 服务监听地址，默认为 0.0.0.0。
PORT=8080 # 服务监听端口，默认为 8080。
CACHE=true # 是否启用持久化，默认为 true。
CACHE_PATH=db/cache.sqlite3 # 持久化缓存的路径（仅当 CACHE=true 时有效），默认为 db/cache.sqlite3。
DEBUG=false # 是否启用调试模式，启用后会输出更多日志，默认为 false。
LOGGING=true # 是否启用日志，默认为 true。
LOG_LEVEL=info # 日志级别，可选值：panic、fatal、error、warn、info、debug、trace（注意：仅当 LOGGING=true 时有效），默认为 info。
COPILOT_TOKEN=ghp_xxxxxxx # 默认的 GitHub Copilot Token，如果设置此项，则请求时携带的 Token 将被忽略。默认为空。
SUPER_TOKEN=randomtoken,randomtoken2 # Super Token 是用户自定义的 Token，用于对请求进行鉴权，若鉴权成功则会使用上方的 COPILOT_TOKEN 处理请求。多个 Token 以英文逗号分隔。默认为空。设置该项可以帮助用户在不泄漏 COPILOT_TOKEN 的情况下分享服务给他人使用。
ENABLE_SUPER_TOKEN=false # 是否启用 Super Token 鉴权，默认为 false。如果未启用但 COPILOT_TOKEN 不为空，则所有请求都会在不鉴权的情况下使用 COPILOT_TOKEN 处理。
CORS_PROXY_NEXTCHAT=false # 启用后，可以通过路由 /cors-proxy-nextchat/ 上为 NextChat 提供代理服务。配置 NextChat 云同步时，如本地部署方式则设置代理地址为：http://localhost:8080/cors-proxy-nextchat/
RATE_LIMIT=0 # 每分钟允许的请求数，如果为 0 则没有限制，默认为 0。
```

**注意：** 以上配置项均可通过命令行参数或环境变量进行配置，命令行参数优先级最高，环境变量优先级次之，配置文件优先级最低。命令行参数名称为为环境变量名称的小写形式，如 `HOST` 对应的命令行参数为 `host`。

### Docker 部署

Docker 部署需要先安装 Docker，然后执行相应命令。

#### 一键部署方式

使用默认配置参数启动服务，如下：

```bash
docker run -d \
  --name copilot-gpt4-service \
  --restart always \
  -p 8080:8080 \
  aaamoon/copilot-gpt4-service:latest
```

启动也可通过环境变量或命令行携带参数配置，如下启动通过环境变量设置 **HOST**、通过命令行参数设置 **LOG_LEVEL**。

```bash
docker run -d \
  --name copilot-gpt4-service \
  -e HOST=0.0.0.0 \
  --restart always \
  -p 8080:8080 \
  aaamoon/copilot-gpt4-service:latest -log_level=debug
```

#### Compose 启动

```bash
# 拉取源代码
git clone https://github.com/aaamoon/copilot-gpt4-service && cd copilot-gpt4-service
# Compose 启动，可通过修改 docker-compose.yml 文件进行启动参数配置
docker compose up -d
```

如需更新容器，可在源代码文件夹重新拉取代码及构建镜像，命令如下：

```bash
git pull && docker compose pull && docker compose up -d
```

### 二进制启动

可以拉取源码自行编译二进制执行文件或从官方下载对应系统架构的二进制执行文件，然后执行以下命令启动（注意 **copilot-gpt4-service** 为二进制执行文件名称，请根据实际名称进行替换）：

```bash
# 快速启动（使用默认配置，如果执行文件所在文件夹有 config.env 文件，则优先使用 config.env 文件配置）
./copilot-gpt4-service

# 通过命令行修改监听端口
./copilot-gpt4-service -port 3000

# 查看帮助，查看可用参数
./copilot-gpt4-service -h
```

**注意：** 运行前请确保已经设置可执行权限，如没有可执行权限，可通过 `chmod +x copilot-gpt4-service` 命令设置。

### Kubernetes 部署

支持通过 Kubernetes 部署，具体部署方式如下：

```bash
helm repo add aaamoon https://charts.kii.la && helm repo update # 源由 github pages 提供
helm install copilot-gpt4-service aaamoon/copilot-gpt4-service
```

### 源码启动

拉取代码，直接启动，适合开发环境使用，请确保本地已经安装 Go 环境。

```bash
git clone https://github.com/aaamoon/copilot-gpt4-service && cd copilot-gpt4-service && go run .
```

## 支持 HTTPS

<details> <summary> 使用 Caddy 支持 HTTPS </summary>

<p>

[Caddy](https://caddyserver.com/docs/) 可以很方便地为端口服务提供 HTTPS 支持，自动管理证书，省心省力。

以下是一个 Debian/Ubuntu 系统上使用 Caddy 的示例，其他系统请参考 [Caddy 官方文档](https://caddyserver.com/docs/)。

### 安装 Caddy

```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

### 配置 Caddy

```bash
sudo vi /etc/caddy/Caddyfile
```

假如你准备使用的域名为 `your.domain.com`，请确保以下条件：

-   请先进行 DNS 解析，将你的域名解析到服务器 IP 地址。
-   开放 80 端口和 443 端口，并且端口没有被其他程序占用，如 Nginx、Xray 等。

然后在 Caddyfile 中添加以下内容：

```bash
your.domain.com {
    reverse_proxy localhost:8080
}
```

### 启动 Caddy

执行以下命令启动 Caddy：

```bash
# 启动 Caddy
sudo systemctl start caddy

# 设置 Caddy 开机自启
sudo systemctl enable caddy

# 查看 Caddy 运行状态
sudo systemctl status caddy
```

如果一切顺利，那此时就可以通过 `https://your.domain.com` 访问 copilot-gpt4-service 服务了。

</p>

</details>

## 与 ChatGPT-Next-Web 一起安装

```bash
helm install copilot-gpt4-service aaamoon/copilot-gpt4-service \
  --set chatgpt-next-web.enabled=true \
  --set chatgpt-next-web.config.OPENAI_API_KEY=[ your openai api key ] \ # copilot 获取的 token
  --set chatgpt-next-web.config.CODE=[ backend access code ] \    # next gpt web ui 的访问密码
  --set chatgpt-next-web.service.type=NodePort \
  --set chatgpt-next-web.service.nodePort=30080
```

## 获取 Copilot Token

首先，你的账号需要开通 GitHub Copilot 服务

获取 GitHub Copilot Plugin Token 的方式目前有两种方式：

1. 通过 Python 脚本获取，只需要 requests 库（推荐）。
2. 通过安装 [GitHub Copilot CLI](https://githubnext.com/projects/copilot-cli/) 授权获取（推荐）。
3. 通过第三方接口授权获取，不推荐，因为不安全。

### 通过 Python 脚本获取

首先确保安装了 Python 3.7+，然后安装 requests 库：

```bash
pip install requests
```

然后执行

**Linux/MacOS 平台获取**

```bash
python3 <(curl -fsSL https://raw.githubusercontent.com/aaamoon/copilot-gpt4-service/master/shells/get_copilot_token.py)
```

可通过设置环境变量或修改脚本第 3 行的字典设置代理。

**Windows 平台获取**

下载脚本，双击运行即可：[get_copilot_token.py](https://raw.githubusercontent.com/aaamoon/copilot-gpt4-service/master/shells/get_copilot_token.py)。

可修改脚本第 3 行的字典设置代理。

### 通过 GitHub Copilot CLI 授权获取

**Linux/MacOS 平台获取**

```bash
# 执行命令，选择对应方式获取Token
bash -c "$(curl -fsSL https://raw.githubusercontent.com/aaamoon/copilot-gpt4-service/master/shells/get_copilot_token.sh)"
```

**Windows 平台获取**

下载批处理脚本，双击运行即可：[get_copilot_token.bat](https://raw.githubusercontent.com/aaamoon/copilot-gpt4-service/master/shells/get_copilot_token.bat)。

## 常见问题

### 模型支持情况

据测试：模型参数支持 GPT-4 和 GPT-3.5-turbo，实测使用其他模型均会以默认的 3.5 处理（对比 OpenAI API 的返回结果，猜测应该是最早的版本 GPT-4-0314 和 GPT-3.5-turbo-0301）

### 如何判断是不是 GPT-4 模型

鲁迅为什么暴打周树人？

-   GPT-3.5 会一本正经的胡说八道
-   GPT-4 表示鲁迅和周树人是同一个人

我爸妈结婚时为什么没有邀请我？

-   GPT-3.5 他们当时认为你还太小，所以没有邀请你。
-   GPT-4 他们结婚时你还没出生。

### HTTP 响应状态码解析说明

-   401: 使用的 GitHub Copilot Plugin Token 过期了或者错误，请重新获取
-   403: 使用的账号没有开通 GitHub Copilot

## 鸣谢

### 贡献者

<a href="https://github.com/aaamoon/copilot-gpt4-service/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=aaamoon/copilot-gpt4-service&anon=0" />
</a>

## 开源协议

[MIT](https://opensource.org/license/mit/)
