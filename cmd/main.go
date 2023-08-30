package main

import (
	"github.com/gin-gonic/gin"

	"github.com/VitalyMyalkin/avito/internal/handlers"
)

func main() {

	newApp := handlers.NewApp()

	// задаем роутер и хендлеры
	router := gin.Default()
	router.POST("/:slug", newApp.AddSegment)
	router.DELETE("/:slug", newApp.RemoveSegment)
	router.PATCH("/:id", newApp.RefreshUserSegments)
	router.GET("/:id", newApp.GetUserSegments)

	// запускаем сервер
	router.Run("localhost:8080")
}
