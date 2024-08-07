package utils

import (
	"copilot-gpt4-service/cache"
	"copilot-gpt4-service/config"
	"copilot-gpt4-service/log"

	"encoding/json"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type Authorization struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}

var superTokenMap = make(map[string]bool)

func init() {
	if config.ConfigInstance.SuperToken != "" && config.ConfigInstance.EnableSuperToken {
		for _, token := range strings.Split(config.ConfigInstance.SuperToken, ",") {
			superTokenMap[token] = true
		}
	}
}

// Set the Authorization in the cache.
func setAuthorizationToCache(copilotToken string, authorization Authorization) {
	new := cache.Authorization{
		C_token:   authorization.Token,
		ExpiresAt: authorization.ExpiresAt,
	}
	cache.CacheInstance.Set(copilotToken, new)
}

// Obtain the Authorization from the cache.
func getAuthorizationFromCache(copilotToken string) *Authorization {
	extraTime := rand.Intn(600) + 300
	if authorization, ok := cache.CacheInstance.Get(copilotToken); ok {
		if authorization.ExpiresAt > time.Now().Unix()+int64(extraTime) {
			return &Authorization{Token: authorization.C_token, ExpiresAt: authorization.ExpiresAt}
		}
	}
	return &Authorization{}
}

// When obtaining the Authorization, first attempt to retrieve it from the cache. If it is not available in the cache, retrieve it through an HTTP request and then set it in the cache.
func GetAuthorizationFromToken(copilotToken string) (string, int, string) {
	authorization := getAuthorizationFromCache(copilotToken)
	if authorization == nil || authorization.Token == "" {
		getAuthorizationUrl := "https://api.github.com/copilot_internal/v2/token"
		client := &http.Client{}
		req, err := http.NewRequest("GET", getAuthorizationUrl, nil)
		if err != nil {
			log.ZLog.Log.Error().Err(err).Msg("Failed to create request: " + getAuthorizationUrl)
			return "", http.StatusInternalServerError, err.Error()
		}
		req.Header.Set("Authorization", "token "+copilotToken)
		response, err := client.Do(req)
		if err != nil {
			log.ZLog.Log.Error().Err(err).Msg("Get GithubCopilot Authorization Token Failed, Request: " + getAuthorizationUrl)
			return "", http.StatusInternalServerError, err.Error()
		}
		defer response.Body.Close()

		body, err := io.ReadAll(response.Body)
		if err != nil {
			log.ZLog.Log.Error().Err(err).Msg("Failed to read getAuthorization response body")
			return "", http.StatusInternalServerError, err.Error()
		}
		if response.StatusCode != 200 {
			log.ZLog.Log.Error().Msgf("Get GithubCopilot Authorization Token Failed, StatusCode: %d, Body: %s", response.StatusCode, string(body))
			return "", response.StatusCode, string(body)
		}

		newAuthorization := &Authorization{}
		if err = json.Unmarshal(body, &newAuthorization); err != nil {
			log.ZLog.Log.Error().Err(err).Msg("Get GithubCopilot Authorization Token Failed, Json Unmarshal Failed")
			return "", http.StatusInternalServerError, err.Error()
		}
		if newAuthorization.Token == "" {
			msg := "Get GithubCopilot Authorization Token Failed, Token is empty"
			log.ZLog.Log.Error().Msg(msg)
			return "", http.StatusInternalServerError, msg
		}

		logMessage := "Token: " + newAuthorization.Token + ", ExpiresAt: " + time.Unix(newAuthorization.ExpiresAt, 0).Format("2006-01-02 15:04:05")
		log.ZLog.Log.Debug().Msg("Get GithubCopilot Authorization Token Success, " + logMessage)
		authorization.Token = newAuthorization.Token
		log.ZLog.Log.Debug().Msg("Successfully obtained Github Copilot Authorization Token, will set in cache. " + logMessage)
		setAuthorizationToCache(copilotToken, *newAuthorization)
	}
	return authorization.Token, http.StatusOK, ""
}

// Retrieve the GitHub Copilot Plugin Token from the request header.
func GetAuthorization(c *gin.Context) (string, bool) {
	copilotToken := strings.TrimPrefix(c.GetHeader("Authorization"), "Bearer ")
	if config.ConfigInstance.CopilotToken != "" &&
		((config.ConfigInstance.EnableSuperToken && superTokenMap[copilotToken]) ||
			!config.ConfigInstance.EnableSuperToken) {
		return config.ConfigInstance.CopilotToken, true
	}

	return copilotToken, copilotToken != ""
}
