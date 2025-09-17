package middleware

import (
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/log"
	"ai-transform-backend/pkg/zerror"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"strings"
)

func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		token := strings.TrimPrefix(c.Request.Header.Get("Authorization"), "Bearer ")
		if token == "" {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		user, err := checkAuth(token)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			log.Error(err)
			return
		}
		if user == nil {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		fmt.Printf("%+v\n", user)
		c.Set("User.ID", user.ID)
		c.Set("User.Name", user.Name)
		c.Set("User.AvatarUrl", user.AvatarUrl)
	}
}

type userInfo struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	AvatarUrl string `json:"avatar_url"`
}

var client = &http.Client{}

func checkAuth(token string) (*userInfo, error) {
	conf := config.GetConfig()
	path := "/api/v1/login/check/auth"
	url := fmt.Sprintf("%s%s?access_token=%s", conf.DependOn.User.Address, path, token)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	res, err := client.Do(req)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode == 401 {
		return nil, nil
	}
	if res.StatusCode == 500 {
		err = zerror.NewByMsg("服务器内部错误")
		log.Error(err)
		return nil, err
	}
	body, err := io.ReadAll(res.Body)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	user := &userInfo{}
	err = json.Unmarshal(body, user)
	if err != nil {
		log.Error(err)
		return nil, err
	}
	return user, nil
}
