package config

import "os"

func GetPort() string {
	p := os.Getenv("PORT")
	if p == "" {
		p = "8080"
	}
	return p
}

func GetJWTSecret() string {
	if s := os.Getenv("JWT_SECRET"); s != "" {
		return s
	}
	return "dev-secret"
}
