<h1 align="center">copilot-gpt4-service</h1>

<p align="center">
⚔️ Convert GitHub Copilot to ChatGPT
</p>

<p align="center">
English | <a href="README_CN.md">简体中文</a>
</p>

## Supported APIs

- `GET /`: Home page
- `GET /healthz`: Health check
- `GET /v1/models`: Get model list
- `POST /v1/chat/completions`: Chat API
- `POST /v1/embeddings`
    - for embeddings api  
    Pay attention that this api is **not totally compatible** with OpenAI API.  
    For `input` field, OpenAI API accepts following types:
        - `string`: The string that will be turned into an embedding.
        - `array`: The array of strings that will be turned into an embedding.
        - `array`: The array of integers that will be turned into an embedding.
        - `array`: The array of arrays containing integers that will be turned into an embedding.

        Unfortunately, this service only accepts the first 2 types as well as the array of arrays containing strings.

## How To Use

1. Install and start the copilot-gpt4-service, e.g., after local startup, the API default address is: `http://127.0.0.1:8080`;
2. Get your GitHub account GitHub Copilot Plugin Token (see below for details);
3. Install a third-party client, e.g., [ChatGPT-Next-Web](https://github.com/ChatGPTNextWeb/ChatGPT-Next-Web), and fill in the settings with the API address of copilot-gpt4-service and the GitHub Copilot Plugin Token in the settings, and you can use the GPT-4 model to have a conversation.

## Deployment Methods

### Best Practice Approach

As verified and discussed by the community, the best practice approach is.

1. Local deployment for personal use only (Recommended);
2. Deploy with your own server integration [ChatGPT-Next-Web](https://github.com/ChatGPTNextWeb/ChatGPT-Next-Web), the service is not public;
3. server deployment, public but for personal use (e.g. multi-client scenarios [Chatbox](https://github.com/Bin-Huang/chatbox), [OpenCat APP](https://opencat.app/), [ChatX APP](https://apps.apple.com/us/app/chatx-ai-chat-client/id6446304087)).

### Not Recommended Approaches

1. Providing an interface as a public service

    Making multiple token requests from the same IP address can be flagged as abnormal behavior.

2. Offering public services using the same client web interface (e.g., ChatGPT-Next-Web) with the default API and API Key

    Making too many requests with the same token can be flagged as abnormal behavior.

3. Deploying with serverless providers

    Serverless providers have short service lifecycles and frequently change IP addresses, which can be flagged as abnormal behavior.

4. Other abusive behaviors or profiteering behaviors.

**⚠️ Very important: The above not recommended methods may cause GitHub Copilot to be banned, and it may not be possible to unban after being banned.**

## Clients

To use copilot-gpt4-service, you need to use it with a third-party client. The following clients have been tested and are supported:

- [ChatGPT-Next-Web](https://github.com/ChatGPTNextWeb/ChatGPT-Next-Web) (recommended)
- [Chatbox](https://github.com/Bin-Huang/chatbox): Supports Windows, Mac, and Linux platforms
- [OpenCat APP](https://opencat.app/): Supports iOS and Mac platforms
- [ChatX APP](https://apps.apple.com/us/app/chatx-ai-chat-client/id6446304087): Supports iOS and Mac platforms

## Server

The deployment methods of the copilot-gpt4-service currently include Docker deployment, binary startup, Kubernetes deployment, and source code startup.

### Configuration

You can configure the service using command line parameters or environment variables or the environment variable configuration file `config.env` (you can copy `config.env.example` from the project root directory to `config.env` and modify it). The default service configuration items are as follows:

```yaml
HOST=0.0.0.0 # Service listening address, default is 0.0.0.0.
PORT=8080 # Service listening port, default is 8080.
CACHE=true # Whether to enable persistence, default is true.
CACHE_PATH=db/cache.sqlite3 # Path of persistent cache (effective only when CACHE=true), default is db/cache.sqlite3.
DEBUG=false # Whether to enable debug mode, more logs will be output after enabling, default is false.
LOGGING=true # Whether to enable logging, default is true.
LOG_LEVEL=info # Log level, optional values: panic, fatal, error, warn, info, debug, trace (Note: effective only when LOGGING=true), default is info.
COPILOT_TOKEN=ghp_xxxxxxx # Default GitHub Copilot Token, if this item is set, the Token carried with the request will be ignored. Default is empty.
SUPER_TOKEN=randomtoken,randomtoken2 # Super Token is a user-defined standalone token that can access COPILOT_TOKEN above. This allows you to share the service without exposing your COPILOT_TOKEN. Multiple tokens are separated by commas. Default is empty.
ENABLE_SUPER_TOKEN=false # Whether to enable SUPER_TOKEN, default is false. If false, but COPILOT_TOKEN is not empty, COPILOT_TOKEN will be used without any authentication for all requests.
CORS_PROXY_NEXTCHAT=false # Whether to enable the CORS proxy for NextChat desktop application. It will then be served on the '$HOST:$PORT/cors-proxy-nextchat/' endpoint. Make sure to update it in your application settings
RATE_LIMIT=0 # The number of requests allowed per minute, if 0 there is no limit, default is 0.
```

**Note:** All of the above configuration items can be configured through command line parameters or environment variables. The priority of command line parameters is the highest, the priority of environment variables is second, and the priority of the configuration file is the lowest. The command line parameter name is the lowercase form of the environment variable name, such as `HOST` corresponding to the command line parameter is `host`.

### Docker Deployment

Docker deployment requires the installation of Docker first, and then execute the command.

#### One-click Deployment

Start the service with default configuration parameters, as follows:

```bash
docker run -d \
  --name copilot-gpt4-service \
  --restart always \
  -p 8080:8080 \
  aaamoon/copilot-gpt4-service:latest
```

The startup can also be configured by carrying parameters through environment variables or command lines, such as setting **HOST** through environment variables and setting **LOG_LEVEL** through command line parameters.

```bash
docker run -d \
  --name copilot-gpt4-service \
  -e HOST=0.0.0.0 \
  --restart always \
  -p 8080:8080 \
  aaamoon/copilot-gpt4-service:latest -log_level=debug
```

#### Compose Startup

```bash
# Pull source code
git clone https://github.com/aaamoon/copilot-gpt4-service && cd copilot-gpt4-service
# Compose startup, startup parameter configuration can be done by modifying the docker-compose.yml file
docker compose up -d
```

If you need to update the container, you can re-pull the code and build the image in the source code folder, the command is as follows:

```bash
git pull && docker compose pull && docker compose up -d
```

### Binary Startup

You can pull the source code to compile the binary executable file yourself, or download the binary executable file of the corresponding system architecture from the repository, and then execute the following command to start (note that **copilot-gpt4-service** is the name of the binary executable file, please replace it according to the actual name):

```bash
# Quick start (use default configuration, if there is a config.env file in the folder where the executable file is located, the config.env file configuration is used first)
./copilot-gpt4-service

# Modify the listening port through the command line
./copilot-gpt4-service -port 3000

# View help, see available parameters
./copilot-gpt4-service -h
```

**Note:** Please make sure that executable permissions have been set before running. If there are no executable permissions, you can set them through the `chmod +x copilot-gpt4-service` command.

### Kubernetes Deployment

Supports deployment through Kubernetes, the specific deployment method is as follows:

```bash
helm repo add aaamoon https://charts.kii.la && helm repo update # Source by github pages
helm install copilot-gpt4-service aaamoon/copilot-gpt4-service
```

### Source Code Startup

Pull the code and start directly, suitable for development environment use, please make sure that the local Go environment has been installed.

```bash
git clone https://github.com/aaamoon/copilot-gpt4-service && cd copilot-gpt4-service && go run .
```

## Support HTTPS

<details> <summary> Use Caddy to support HTTPS </summary>

<p>

Using [Caddy](https://caddyserver.com/docs/) can easily provide HTTPS support for port services, automatically manage certificates, worry-free.

Here's an example of using Caddy on a Debian/Ubuntu system, for other systems please refer to the [Caddy official documentation](https://caddyserver.com/docs/).

### Install Caddy

```bash
sudo apt install -y debian-keyring debian-archive-keyring apt-transport-https curl
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/gpg.key' | sudo gpg --dearmor -o /usr/share/keyrings/caddy-stable-archive-keyring.gpg
curl -1sLf 'https://dl.cloudsmith.io/public/caddy/stable/debian.deb.txt' | sudo tee /etc/apt/sources.list.d/caddy-stable.list
sudo apt update
sudo apt install caddy
```

### Configure Caddy

```bash
sudo vi /etc/caddy/Caddyfile
```

If you are going to use the domain name `your.domain.com`, please make sure of the following:

-   Please perform DNS resolution first and resolve your domain name to the server IP address.
-   Open port 80 and port 443, and the port is not occupied by other programs, such as Nginx, Xray, etc.

Then add the following content to Caddyfile:

```bash
your.domain.com {
    reverse_proxy localhost:8080
}
```

### Start Caddy

Execute the following command to start Caddy:

```bash
# Start Caddy
sudo systemctl start caddy

# Set Caddy to start automatically
sudo systemctl enable caddy

# View Caddy running status
sudo systemctl status caddy
```

If the command is executed successfully, you can now access the copilot-gpt4-service service via `https://your.domain.com`.

</p>

</details>

## Installation with ChatGPT-Next-Web

```bash
helm install copilot-gpt4-service aaamoon/copilot-gpt4-service \
  --set chatgpt-next-web.enabled=true \
  --set chatgpt-next-web.config.OPENAI_API_KEY=[ your openai api key ] \   #Token obtained by copilot
  --set chatgpt-next-web.config.CODE=[ backend access code ] \    # Access password for next chatgpt web ui
  --set chatgpt-next-web.service.type=NodePort \
  --set chatgpt-next-web.service.nodePort=30080
```

## Obtaining Copilot Token

Your account needs to have GitHub Copilot service enabled.

There are currently two ways to obtain the GitHub Copilot Plugin Token:

1. Obtain it through the Python script, only the requests library is required (recommended).
2. Obtain it by installing [GitHub Copilot CLI](https://githubnext.com/projects/copilot-cli/) and authorizing (recommended).
3. Authorized access through a third-party interface, not recommended because it is not secure.

### Obtaining Through Python Script

First, install the requests library:

```bash
pip install requests
```

Then, run the following command to obtain the GitHub Copilot Plugin Token:

**For Linux/MacOS Platforms**

```bash
python3 <(curl -fsSL https://raw.githubusercontent.com/aaamoon/copilot-gpt4-service/master/shells/get_copilot_token.py)
```

You can set proxy by setting environment variables or modifying the 3th line of the script.

**For Windows Platform**

Download the Python script and double-click to run it: [get_copilot_token.py](https://raw.githubusercontent.com/aaamoon/copilot-gpt4-service/master/shells/get_copilot_token.py).

You can set proxy by modifying the 3th line of the script.

### Obtaining Through GitHub Copilot CLI

**For Linux/MacOS Platforms**

```bash
# Execute the command and select the corresponding way to get the Token
bash -c "$(curl -fsSL https://raw.githubusercontent.com/aaamoon/copilot-gpt4-service/master/shells/get_copilot_token.sh)"
```

**For Windows Platform**

Download the batch script and double-click to run it: [get_copilot_token.bat](https://raw.githubusercontent.com/aaamoon/copilot-gpt4-service/master/shells/get_copilot_token.bat).

## Frequently Asked Questions

### Model Support

According to the test, the model parameters support GPT-4 and GPT-3.5-turbo, and the actual test will be processed at the default 3.5 when using other models (compared with the return results of the OpenAI API, guess it should be the earliest versions of GPT-4-0314 and GPT-3.5-turbo-0301)

### How To Determine If It's The GPT-4 Model

Why weren't I invited when my parents got married?

-   GPT-3.5 They considered you too young at that time, so they didn't invite you.
-   GPT-4 They got married before you were born.

### Explanation Of HTTP Response Status Codes

-   401: The GitHub Copilot Plugin Token used has expired or is incorrect. Please obtain it again.
-   403: The account used does not have GitHub Copilot enabled.

### Got "Failed to authenticate: You do not have access to GitHub Copilot CLI"

If you're using an account that hasn't used Copilot before (e.g. with a recently activated student pack or pro), you can try to create an [Codespaces](https://github.com/codespaces) then install Copilot extension inside. This may take you to the [Copilot signup](https://github.com/github-copilot/free_signup) page.

Don't trust the "Please join the waitlist" tips because copilot-cli is [now in public beta](https://github.blog/changelog/2023-11-08-github-copilot-in-the-cli-now-in-public-beta/).

## Acknowledgements

### Contributors

<a href="https://github.com/aaamoon/copilot-gpt4-service/graphs/contributors">
  <img src="https://contrib.rocks/image?repo=aaamoon/copilot-gpt4-service&anon=0" />
</a>

## LICENSE

[MIT](https://opensource.org/license/mit/)
