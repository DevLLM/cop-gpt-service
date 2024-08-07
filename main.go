package main

import (
	"io"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"golang.org/x/time/rate"

	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"copilot-gpt4-service/cache"
	"copilot-gpt4-service/config"
	"copilot-gpt4-service/log"
	"copilot-gpt4-service/tools"
	"copilot-gpt4-service/utils"
)

// Handle the Cross-Origin Resource Sharing (CORS) for requests.
func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "*")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "*")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(200)
			return
		}

		c.Next()
	}
}

// Represent the JSON data structure for the request body.
type CompletionsJsonData struct {
	Messages    interface{} `json:"messages"`
	Model       string      `json:"model"`
	Temperature float64     `json:"temperature"`
	TopP        float64     `json:"top_p"`
	N           int64       `json:"n"`
	Stream      bool        `json:"stream"`
}

type EmbeddingsJsonData struct {
	Input interface{} `json:"input"`
	Model string      `json:"model"`
}

type Message struct {
	Role    *string `json:"role,omitempty"`
	Content string  `json:"content"`
}

type Choice struct {
	Delta         *Message `json:"delta,omitempty"`
	Message       *Message `json:"message,omitempty"`
	Index         int      `json:"index"`
	Finish_reason *string  `json:"finish_reason"`
}

type Usage struct {
	Prompt_tokens     int `json:"prompt_tokens"`
	Completion_tokens int `json:"completion_tokens"`
	Total_tokens      int `json:"total_tokens"`
}

type Data struct {
	Choices []Choice `json:"choices,omitempty"`
	Created int      `json:"created,omitempty"`
	ID      string   `json:"id,omitempty"`
	Object  string   `json:"object,omitempty"`
	Model   string   `json:"model,omitempty"`
	Usage   *Usage   `json:"usage,omitempty"`
}

type EmbeddingData struct {
	Object    string        `json:"object"`
	Index     int           `json:"index"`
	Embedding []interface{} `json:"embedding"`
}
type EmbeddingUsage struct {
	Prompt_tokens int `json:"prompt_tokens"`
	Total_tokens  int `json:"total_tokens"`
}
type Embedding struct {
	Object string          `json:"object"`
	Data   []EmbeddingData `json:"data"`
	Model  string          `json:"model"`
	Usage  EmbeddingUsage  `json:"usage"`
}

// Create request headers to mock Github Copilot Chat requests.
func createHeaders(apptoken string, stream bool) map[string]string {
	item, ok := cache.CacheInstance.Get(apptoken)
	if !ok {
		return nil
	}

	contentType := "application/json; charset=utf-8"
	if stream {
		contentType = "text/event-stream; charset=utf-8"
	}

	return map[string]string{
		"Authorization":         "Bearer " + item.C_token,
		"X-Request-Id":          uuid.NewString(),
		"Vscode-Sessionid":      item.Vscode_sessionid,
		"Vscode-Machineid":      item.Vscode_machineid,
		"Editor-Version":        "vscode/1.83.1",
		"Editor-Plugin-Version": "copilot-chat/0.8.0",
		"Openai-Organization":   "github-copilot",
		"Openai-Intent":         "conversation-panel",
		"Content-Type":          contentType,
		"User-Agent":            "GitHubCopilotChat/0.8.0",
		"Accept":                "*/*",
		"Accept-Encoding":       "gzip,deflate,br",
		"Connection":            "close",
	}
}

func respondWithError(c *gin.Context, httpStatusCode int, errorMessage string) {
	c.JSON(
		httpStatusCode,
		gin.H{
			"error": errorMessage,
			"code":  httpStatusCode,
		},
	)
	c.Abort()
}

func chatCompletions(c *gin.Context) {
	url := "https://api.githubcopilot.com/chat/completions"

	// Get app token from request header
	appToken, ok := utils.GetAuthorization(c)
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_, statusCode, errorInfo := utils.GetAuthorizationFromToken(appToken)
	if len(errorInfo) != 0 {
		respondWithError(c, statusCode, errorInfo)
		return
	}

	jsonBody := &CompletionsJsonData{
		Messages: []map[string]string{
			{"role": "system",
				"content": "\nYou are ChatGPT, a large language model trained by OpenAI.\nKnowledge cutoff: 2021-09\nCurrent model: gpt-4\n"},
		},
		Model:       "gpt-4",
		Temperature: 0.5,
		TopP:        1,
		N:           1,
		Stream:      false,
	}
	_ = c.BindJSON(&jsonBody)

	jsonData, err := json.Marshal(jsonBody)
	if err != nil {
		log.ZLog.Log.Error().Msgf("Error when marshalling the JSON data: %s", err.Error())
		respondWithError(c, http.StatusInternalServerError, "Error when marshalling the JSON data.")
		return
	}

	headers := createHeaders(appToken, jsonBody.Stream)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		error_msg := fmt.Sprintf("Encountering an error when sending the request: %s", err.Error())
		log.ZLog.Log.Err(err).Msg(error_msg)
		respondWithError(c, http.StatusInternalServerError, error_msg)
		return
	}

	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		error_msg := fmt.Sprintf("Encountering an error when receiving the github copilot response: %s", resp.Status)
		log.ZLog.Log.Error().Msg(error_msg)
		respondWithError(c, resp.StatusCode, error_msg)
		return
	}
	// Set the headers for the response
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	if jsonBody.Stream {
		c.Header("Content-Type", "text/event-stream; charset=utf-8")
	} else {
		c.Header("Content-Type", "application/json; charset=utf-8")
	}
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	// Scan the response body line by line
	scanner := bufio.NewScanner(resp.Body)
	for scanner.Scan() {
		line := scanner.Bytes()

		var object string
		if jsonBody.Stream {
			object = "chat.completion.chunk"
		} else {
			object = "chat.completion"
		}

		if len(line) > 0 && !bytes.Contains(line, []byte("data: [DONE]")) {
			tmp := strings.TrimPrefix(string(line), "data: ")
			data := &Data{}
			if err := json.Unmarshal([]byte(tmp), &data); err != nil {
				fmt.Println(err)
			}
			if len(data.Choices) == 0 {
				continue
			}
			if data.Object == "" {
				data.Object = object
			}
			if data.Model == "" {
				data.Model = jsonBody.Model
			}
			if data.Created == 0 {
				data.Created = int(time.Now().Unix())
			}

			newLine, err := json.Marshal(data)
			if err != nil {
				fmt.Println(err)
			}
			if jsonBody.Stream {
				line = []byte(fmt.Sprintf("data: %s", string(newLine)))
			} else {
				line = newLine
			}
		}

		c.Writer.Write(line)
		c.Writer.Write([]byte("\n")) // Add newline to the end of each line
		c.Writer.Flush()
	}
	if err := scanner.Err(); err != nil {
		c.AbortWithError(http.StatusBadGateway, err)
		return
	}

}

func embeddings(c *gin.Context) {
	url := "https://api.githubcopilot.com/embeddings"

	// Get app token from request header
	appToken, ok := utils.GetAuthorization(c)
	if !ok {
		respondWithError(c, http.StatusUnauthorized, "Unauthorized")
		return
	}

	_, statusCode, errorInfo := utils.GetAuthorizationFromToken(appToken)
	if len(errorInfo) != 0 {
		respondWithError(c, statusCode, errorInfo)
		return
	}

	jsonBody := &EmbeddingsJsonData{
		Input: "",
		Model: "text-embedding-ada-002",
	}
	_ = c.BindJSON(&jsonBody)
	// check if the input is empty, if so, return an error
	if jsonBody.Input == "" {
		respondWithError(c, http.StatusBadRequest, "Input cannot be empty.")
		return
	}
	// check if the input is a list, if not, wrap it in a list
	if _, ok := jsonBody.Input.([]interface{}); !ok {
		jsonBody.Input = []interface{}{jsonBody.Input}
	}

	jsonData, err := json.Marshal(jsonBody)
	if err != nil {
		log.ZLog.Log.Error().Msgf("Error when marshalling the JSON data: %s", err.Error())
		respondWithError(c, http.StatusInternalServerError, "Error when marshalling the JSON data.")
		return
	}

	headers := createHeaders(appToken, false)

	req, _ := http.NewRequest("POST", url, bytes.NewReader(jsonData))
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Encountering an error when sending the request.")
	} else {
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return
		} else {
			// Set the headers for the response
			c.Writer.Header().Set("Transfer-Encoding", "chunked")
			c.Writer.Header().Set("X-Accel-Buffering", "no")
			c.Header("Content-Type", "application/json; charset=utf-8")
			c.Header("Cache-Control", "no-cache")
			c.Header("Connection", "keep-alive")

			// Scan the response body line by line (embedding won't be stream=true)
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := scanner.Bytes()

				if len(line) > 0 {
					data := &Embedding{}
					if err := json.Unmarshal([]byte(line), &data); err != nil {
						fmt.Println(err)
					}
					if len(data.Data) == 0 {
						continue
					}
					if data.Object == "" {
						data.Object = "list"
					}
					for i := 0; i < len(data.Data); i++ {
						if data.Data[i].Object == "" {
							data.Data[i].Object = "embedding"
						}
					}
					if data.Model == "" {
						data.Model = jsonBody.Model
					}

					newLine, err := json.Marshal(data)
					if err != nil {
						fmt.Println(err)
					}
					line = newLine
				}

				c.Writer.Write(line)
				c.Writer.Write([]byte("\n")) // Add newline to the end of each line
				c.Writer.Flush()
			}
		}
	}
}

func createMockModel(modelId string) gin.H {
	return gin.H{
		"id":       modelId,
		"object":   "model",
		"created":  1677610602,
		"owned_by": "openai",
		"permission": []gin.H{
			{
				"id":                   "modelperm-" + tools.GenHexStr(12),
				"object":               "model_permission",
				"created":              1677610602,
				"allow_create_engine":  false,
				"allow_sampling":       true,
				"allow_logprobs":       true,
				"allow_search_indices": false,
				"allow_view":           true,
				"allow_fine_tuning":    false,
				"organization":         "*",
				"group":                nil,
				"is_blocking":          false,
			},
		},
		"root":   modelId,
		"parent": nil,
	}
}

func createMockModelsResponse(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"object": "list",
		"data": []gin.H{
			createMockModel("gpt-3.5-turbo"),
			createMockModel("gpt-4"),
		},
	})
}

// corsProxyNextChat endpoint handler for proxying requests from nextChat desktop app
func corsProxyNextChat(c *gin.Context) {
	sp := strings.Split(c.Param("path"), "/")
	if len(sp) < 3 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid proxy path"})
		return
	}

	proto := sp[1]
	host := sp[2]
	path := strings.Join(sp[3:], "/")
	if path != "" {
		path = "/" + path
	}
	url, err := url.Parse(fmt.Sprintf("%s://%s%s", proto, host, path))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	method := c.Request.Header.Get("Method")
	c.Request.Header.Del("Method")
	if method == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Method header is not set"})
		return
	}
	proxyReq, err := http.NewRequest(method, url.String(), c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	proxyReq.Header = c.Request.Header

	resp, err := http.DefaultClient.Do(proxyReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	c.Status(resp.StatusCode)

	_, err = io.Copy(c.Writer, resp.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

func LoggerHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		t := time.Now()

		log.ZLog.Log.Info().Msgf("Request Info:\nMethod: %s\nHost: %s\nURL: %s",
			c.Request.Method, c.Request.Host, c.Request.URL)
		log.ZLog.Log.Debug().Msgf("Request Header:\n%v", c.Request.Header)

		c.Next()

		latency := time.Since(t)
		log.ZLog.Log.Info().Msgf("Response Time: %s\nStatus: %d",
			latency.String(), c.Writer.Status())
		log.ZLog.Log.Debug().Msgf("Response Header:\n%v", c.Writer.Header())
	}
}

func RateLimiterHandler(reqsPerMin int) gin.HandlerFunc {
	var limiter *rate.Limiter
	if reqsPerMin > 0 {
		limiter = rate.NewLimiter(rate.Every(time.Minute), reqsPerMin)
	} else {
		limiter = rate.NewLimiter(rate.Inf, 0)
	}

	return func(c *gin.Context) {
		if !limiter.Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"message": "too many requests",
			})
			return
		}
		c.Next()
	}
}

func startupOutput() {
	tools.PrintStructFieldsAndValues(config.ConfigInstance, "Copilot-GPT4-Service startup configuration:")
	fmt.Println("Service is running at:")
	fmt.Printf(" - %-20s: %s\n", "Local", tools.Colorize(tools.ColorGreen, fmt.Sprintf("http://%s:%d", config.ConfigInstance.Host, config.ConfigInstance.Port)))

	if config.ConfigInstance.Host == "0.0.0.0" {
		ipv4s, err := tools.GetIPv4NetworkIPs()
		if ipv4s != nil && err == nil {
			for _, ip := range ipv4s {
				fmt.Printf(" - %-20s: %s\n", "Network", tools.Colorize(tools.ColorGreen, fmt.Sprintf("http://%s:%d", ip, config.ConfigInstance.Port)))
			}
		}
	}

	fmt.Println("(Press CTRL+C to quit)")

	if config.ConfigInstance.CORSProxyNextChat {
		fmt.Println()
		fmt.Println(tools.Colorize(tools.ColorYellow, "WARNING: CORS_PROXY_NEXTCHAT is enabled. This is a potential security risk if your service is not private."))
	}

	if config.ConfigInstance.CopilotToken != "" && !config.ConfigInstance.EnableSuperToken {
		fmt.Println()
		fmt.Println(tools.Colorize(tools.ColorYellow, "WARNING: COPILOT_TOKEN is set, but ENABLE_SUPER_TOKEN is not enabled. This is a potential security risk if your service is not private."))
	}

	fmt.Println()
	fmt.Println(tools.Colorize(tools.ColorRed, "Warning: Please do not make this service public, for personal use only, otherwise the account or Copilot will be banned."))
	fmt.Println(tools.Colorize(tools.ColorRed, "警告：请不要将此服务公开，仅供个人使用，否则账户或 Copilot 将被封禁。"))
	fmt.Println()
}

func startupCheck() {
	if config.ConfigInstance.Port < 1 || config.ConfigInstance.Port > 65535 {
		fmt.Println(tools.Colorize(tools.ColorRed, fmt.Sprintf("Invalid port %d, use default port %d instead.", config.ConfigInstance.Port, config.DefaultPort)))
		config.ConfigInstance.Port = config.DefaultPort
	}
	if config.ConfigInstance.EnableSuperToken && config.ConfigInstance.SuperToken == "" {
		fmt.Println(tools.Colorize(tools.ColorRed, "You enabled super token but didn't set the super token, please set the super token in the configuration file."))
	}

	if !tools.FilExists("./robots.txt") {
		fmt.Println(tools.Colorize(tools.ColorYellow, "robots.txt not found, creating it..."))
		tools.WriteToFile("./robots.txt", "User-agent: *\nDisallow: /\n", 0644)
	}
}

func main() {
	if config.ConfigInstance.Debug {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()
	router.Use(CORSMiddleware())
	router.Use(LoggerHandler())

	router.StaticFile("/robots.txt", "./robots.txt")

	router.POST("/v1/chat/completions", RateLimiterHandler(config.ConfigInstance.RateLimit), chatCompletions)
	router.POST("/v1/embeddings", RateLimiterHandler(config.ConfigInstance.RateLimit), embeddings)
	router.GET("/v1/models", createMockModelsResponse)
	router.GET("/healthz", func(context *gin.Context) {
		context.JSON(200, gin.H{
			"message": "ok",
		})
	})
	router.GET("/", func(c *gin.Context) {
		c.Data(http.StatusOK, "text/html; charset=utf-8", []byte(`
		<div style="color:red;padding:0 20px;display:grid;align-items:center;justify-content:center;height:98vh;overflow:hidden;font-size:20px;line-height:30px;text-align:center;"><b>Very important: please do not make this service public, for personal use only, otherwise the account or Copilot will be banned.<br>非常重要：请不要将此服务公开，仅供个人使用，否则账户或 Copilot 将被封禁。</b></div>`))
	})
	router.NoRoute(func(c *gin.Context) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": fmt.Sprintf("Invalid URL (%s %s)", c.Request.Method, c.Request.URL.Path),
		})
	})

	if config.ConfigInstance.CORSProxyNextChat {
		router.Any("/cors-proxy-nextchat/*path", corsProxyNextChat)
	}

	startupCheck()
	startupOutput()

	err := router.Run(fmt.Sprintf("%s:%d", config.ConfigInstance.Host, config.ConfigInstance.Port))
	if err != nil {
		log.ZLog.Log.Error().Msgf("Failed to run server: %v", err)
	}
}
