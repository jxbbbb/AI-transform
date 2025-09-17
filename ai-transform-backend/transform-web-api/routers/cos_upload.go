package routers

import (
	"ai-transform-backend/transform-web-api/controller"
	"github.com/gin-gonic/gin"
)

func InitCosUploadRouters(g *gin.RouterGroup, controller *controllers.CosUpload) {
	v1 := g.Group("/v1")
	v1.GET("/cos/presigned/url", controller.GetPresignedURL)
}
