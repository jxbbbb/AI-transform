package controllers

import (
	"ai-transform-backend/pkg/config"
	"ai-transform-backend/pkg/constants"
	"ai-transform-backend/pkg/log"
	"ai-transform-backend/pkg/utils"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
)

type CosUpload struct {
	conf *config.Config
	log  log.ILogger
}

func NewCosUpload(conf *config.Config, log log.ILogger) *CosUpload {
	return &CosUpload{
		conf: conf,
		log:  log,
	}
}

func (c *CosUpload) getTmpSecret(userID int64) (*sts.CredentialResult, error) {
	client := sts.NewClient(c.conf.Cos.SecretId, c.conf.Cos.SecretKey, nil)
	opt := &sts.CredentialOptions{
		DurationSeconds: int64(time.Hour.Seconds()),
		Region:          c.conf.Cos.Region,
		Policy: &sts.CredentialPolicy{
			Statement: []sts.CredentialPolicyStatement{
				{
					// 密钥的权限列表。简单上传和分片需要以下的权限，其他权限列表请看 https://cloud.tencent.com/document/product/436/31923
					Action: []string{
						// 简单上传
						"name/cos:PostObject",
						"name/cos:PutObject",
						// 分片上传
						"name/cos:InitiateMultipartUpload",
						"name/cos:ListMultipartUploads",
						"name/cos:ListParts",
						"name/cos:UploadPart",
						"name/cos:CompleteMultipartUpload",
					},
					Effect: "allow",
					Resource: []string{
						// 这里改成允许的路径前缀，可以根据自己网站的用户登录态判断允许上传的具体路径，例子： a.jpg 或者 a/* 或者 * (使用通配符*存在重大安全风险, 请谨慎评估使用)
						// 存储桶的命名格式为 BucketName-APPID，此处填写的 bucket 必须为此格式
						fmt.Sprintf("qcs::cos:%s:uid/", c.conf.Cos.Region) + c.conf.Cos.AppID + ":" + c.conf.Cos.Bucket + fmt.Sprintf("%s/%d/*", constants.COS_INPUT, userID),
					},
				},
			},
		},
	}
	return client.GetCredential(opt)
}

type URLToken struct {
	SessionToken string `url:"x-cos-security-token,omitempty" header:"-"`
}

func (c *CosUpload) GetPresignedURL(ctx *gin.Context) {
	userID, ok := ctx.Get("User.ID")
	fmt.Println(userID)
	if !ok {
		c.log.Error("鉴权失败")
		ctx.JSON(http.StatusUnauthorized, gin.H{})
		return
	}
	filename := ctx.Query("filename")
	ext := path.Ext(filename)
	filename = fmt.Sprintf("%x_%d%s", utils.MD5([]byte(filename)), time.Now().Unix(), ext)

	credential, err := c.getTmpSecret(userID.(int64))
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	tak := credential.Credentials.TmpSecretID
	tsk := credential.Credentials.TmpSecretKey
	token := &URLToken{SessionToken: credential.Credentials.SessionToken}

	u, _ := url.Parse(c.conf.Cos.BucketUrl)
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{})

	filename = fmt.Sprintf("%s/%d/%d_%s", constants.COS_INPUT, userID, userID, filename)
	filename = strings.Trim(filename, "/")

	opt := &cos.PresignedURLOptions{
		Query:  &url.Values{},
		Header: &http.Header{},
	}
	opt.Query.Add("x-cos-security-token", token.SessionToken)
	presignedURL, err := client.Object.GetPresignedURL(context.Background(), http.MethodPut, filename, tak, tsk, time.Minute*5, opt)
	if err != nil {
		c.log.Error(err)
		ctx.JSON(http.StatusInternalServerError, gin.H{})
		return
	}
	fileUrl := fmt.Sprintf("%s%s", c.conf.Cos.CDNDomain, presignedURL.Path)
	ctx.JSON(http.StatusOK, gin.H{
		"presigned_url": presignedURL.String(),
		"file_url":      fileUrl,
	})
}
