package controllers

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type record struct {
	ID                 int64  `json:"id"`
	ProjectName        string `json:"project_name"`
	OriginalLanguage   string `json:"original_language"`
	TranslatedLanguage string `json:"translated_language"`
	OriginalVideoUrl   string `json:"original_video_url"`
	TranslatedVideoUrl string `json:"translated_video_url"`
	ExpirationAt       int64  `json:"expiration_at"`
	CreateAt           int64  `json:"create_at"`
}

func (c *Transform) GetRecords(ctx *gin.Context) {
	userID, _ := ctx.Get("User.ID")
	recordsData := c.data.NewTransformRecordsData()
	list, err := recordsData.GetByUserID(userID.(int64))
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	records := make([]record, len(list))
	for index, l := range list {
		records[index].ID = l.ID
		records[index].ProjectName = l.ProjectName
		records[index].OriginalLanguage = l.OriginalLanguage
		records[index].TranslatedLanguage = l.TranslatedLanguage
		records[index].OriginalVideoUrl = l.OriginalVideoUrl
		records[index].TranslatedVideoUrl = l.TranslatedVideoUrl
		records[index].ExpirationAt = l.ExpirationAt
		records[index].CreateAt = l.CreateAt
	}
	ctx.JSON(http.StatusOK, records)
	return
}
