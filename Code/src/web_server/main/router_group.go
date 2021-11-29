package main

import (
	"github.com/gin-gonic/gin"
	"sync"
)

type orderHandler struct {
	sync.RWMutex
}

func newOrderHandler() *productHandler {
	return nil
}

func (u *orderHandler) Create(c *gin.Context) {

}

func (u *orderHandler) Get(c *gin.Context) {

}

func router_group() {
	router := gin.Default()
	productHandler := newProductHandler()
	orderHandler := newOrderHandler()

	v1 := router.Group("/v1", gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))
	{
		productv1 := v1.Group("/products")
		{
			// 路由匹配
			productv1.POST("", productHandler.Create)
			productv1.GET(":name", productHandler.Get)
		}
		orderv1 := v1.Group("/orders")
		{
			// 路由匹配
			orderv1.POST("", orderHandler.Create)
			orderv1.GET(":name", orderHandler.Get)
		}
	}

	v2 := router.Group("/v2", gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))
	{
		productv2 := v2.Group("/products")
		{
			// 路由匹配
			productv2.POST("", productHandler.Create)
			productv2.GET(":name", productHandler.Get)
		}
	}
}

func router_group_middle() {
	router := gin.New()

	router.Use(gin.Logger(), gin.Recovery()) // 中间件作用于所有的HTTP请求

	v1 := router.Group("/v1").Use(gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))
	v1.POST("/login", Login).Use(gin.BasicAuth(gin.Accounts{"foo": "bar", "colin": "colin404"}))
}

func Login(context *gin.Context) {

}
