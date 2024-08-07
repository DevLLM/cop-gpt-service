package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port              int
	Cache             bool
	CachePath         string
	Host              string
	Debug             bool
	Logging           bool
	LogLevel          string
	CopilotToken      string
	CORSProxyNextChat bool
	RateLimit         int
	EnableSuperToken  bool
	SuperToken        string
}

var ConfigInstance *Config = &Config{}

// Default Settings
const (
	DefaultPort              = 8080
	DefaultCache             = true
	DefaultCachePath         = "db/cache.sqlite3"
	DefaultHost              = "0.0.0.0"
	DefaultDebug             = false
	DefaultLogging           = true
	DefaultLogLevel          = "info"
	DefaultCORSProxyNextChat = false
	DefaultRateLimit         = 0
	DefaultCopilotToken      = ""
	DefaultEnableSuperToken  = false
	DefaultSuperToken        = ""
)

func init() {
	// if exists config.env, load it
	if _, err := os.Stat("config.env"); err == nil {
		err := godotenv.Load("config.env")
		if err != nil {
			fmt.Println("Error loading config.env file")
		}
	}

	flag.StringVar(&ConfigInstance.Host, "host", getEnvOrDefault("HOST", DefaultHost), "Service listen address.")
	flag.IntVar(&ConfigInstance.Port, "port", getEnvOrDefaultInt("PORT", DefaultPort), "Service listen port.")
	flag.StringVar(&ConfigInstance.CachePath, "cache_path", getEnvOrDefault("CACHE_PATH", DefaultCachePath), "Path to the persistent cache.")
	flag.StringVar(&ConfigInstance.LogLevel, "log_level", getEnvOrDefault("LOG_LEVEL", DefaultLogLevel), "Log level, optional values: panic, fatal, error, warn, info, debug, trace (note: valid only when log_level is true).")
	flag.StringVar(&ConfigInstance.CopilotToken, "copilot_token", getEnvOrDefault("COPILOT_TOKEN", DefaultCopilotToken), "Default Github Copilot Token, if this is set, the Token carried in the request will be ignored. Default is empty.")
	flag.BoolVar(&ConfigInstance.EnableSuperToken, "enable_super_token", getEnvOrDefaultBool("ENABLE_SUPER_TOKEN", DefaultEnableSuperToken), "Enable standalone super token.")
	flag.StringVar(&ConfigInstance.SuperToken, "super_token", getEnvOrDefault("SUPER_TOKEN", DefaultSuperToken), "Value of super token; use ',' to separate multiple tokens.")
	flag.BoolVar(&ConfigInstance.Cache, "cache", getEnvOrDefaultBool("CACHE", DefaultCache), "Whether persistence is enabled or not.")
	flag.BoolVar(&ConfigInstance.Debug, "debug", getEnvOrDefaultBool("DEBUG", DefaultDebug), "Enable debug mode, if enabled, more logs will be output.")
	flag.BoolVar(&ConfigInstance.Logging, "logging", getEnvOrDefaultBool("LOGGING", DefaultLogging), "Enable logging.")
	flag.IntVar(&ConfigInstance.RateLimit, "rate_limit", getEnvOrDefaultInt("RATE_LIMIT", DefaultRateLimit), "Limit the number of requests per minute. 0 means no limit.")
	flag.BoolVar(&ConfigInstance.CORSProxyNextChat, "cors_proxy_nextchat", getEnvOrDefaultBool("CORS_PROXY_NEXTCHAT", DefaultCORSProxyNextChat), "Enable CORS proxy for NextChat.")

	flag.Parse()
}

func getEnvOrDefault(key string, defaultValue string) string {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	return value
}

func getEnvOrDefaultBool(key string, defaultValue bool) bool {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	s, err := strconv.ParseBool(value)
	if err != nil {
		fmt.Println("\033[31mError parsing boolean value for key:", key, "\033[0m")
		return defaultValue
	}
	return s
}

func getEnvOrDefaultInt(key string, defaultValue int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return defaultValue
	}
	s, err := strconv.Atoi(value)
	if err != nil {
		fmt.Println("\033[31mError parsing integer value for key:", key, "\033[0m")
		return defaultValue
	}
	return s
}
