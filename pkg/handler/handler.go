package handler

import (
	"time"

	"github.com/IBM/sarama"
	"github.com/MsSabo/todo-app/pkg/ads"
	"github.com/MsSabo/todo-app/pkg/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	services *service.Service
	producer *sarama.SyncProducer
}

func wrapper(f func(c *gin.Context)) func(*gin.Context) {
	return func(t *gin.Context) {
		start := time.Now()

		defer func() {
			ads.ObserveRequest(time.Since(start), t.Writer.Status())
		}()

		f(t)
	}
}

func NewHandler(services *service.Service, producer *sarama.SyncProducer) *Handler {
	return &Handler{services: services, producer: producer}
}

func (h *Handler) InitRoutes() *gin.Engine {
	router := gin.New()

	auth := router.Group("/auth")
	{
		auth.POST("/sign-up", h.signUp)
		auth.POST("/sign-in", h.signIn)
	}

	api := router.Group("/api", h.userIdentity)
	{
		lists := api.Group("/lists")
		{
			lists.POST("/", wrapper(h.createList))
			lists.GET("/", wrapper(h.getAllLists))
			lists.GET("/:id", wrapper(h.getListById))
			lists.PUT("/:id", wrapper(h.updateList))
			lists.DELETE("/:id", wrapper(h.deleteList))

			items := lists.Group(":id/items")
			{
				items.POST("/", wrapper(h.createItem))
				items.GET("/", wrapper(h.getAllItem))
			}
		}
		items := api.Group("/items")
		{
			items.GET("/:item_id", wrapper(h.getItemById))
			items.PUT("/:item_id", wrapper(h.updateItem))
			items.DELETE("/:item_id", wrapper(h.deleteItem))
		}
	}

	return router
}
