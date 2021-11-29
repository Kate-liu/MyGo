package auth

import (
	"github.com/gin-gonic/gin"
	"time"
)

func main() {
	router := gin.New()

	// 认证
	router.Use(gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))

	// RequestID
	router.Use(requestid.New(requestid.Config{
		Generator: func() string {
			return "test"
		},
	}))

	// 跨域
	// CORS for https://foo.com and https://github.com origins, allowing:
	// - PUT and PATCH methods
	// - Origin header
	// - Credentials share
	// - Preflight requests cached for 12 hours
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"https://foo.com"},
		AllowMethods:     []string{"PUT", "PATCH"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))
}
