package main

import "os"

type Config struct {
	Port           string
	MongoURI       string
	MongoDB        string
	FrontendOrigin string
}

func LoadConfig() Config {
	return Config{
		Port:           getEnv("PORT", "8080"),
		MongoURI:       getEnv("MONGODB_URI", "mongodb://localhost:27017"),
		MongoDB:        getEnv("MONGODB_DB", "mediasequencer"),
		FrontendOrigin: getEnv("FRONTEND_ORIGIN", "*"),
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
