package routers

import (
	controllers "ai-transform-backend/transform-web-api/controller"
	"github.com/gin-gonic/gin"
)

func InitTransformRouters(g *gin.RouterGroup, controller *controllers.Transform) {
	v1 := g.Group("/v1")
	v1.POST("/translate", controller.Translate)
	v1.GET("/records", controller.GetRecords)
}
