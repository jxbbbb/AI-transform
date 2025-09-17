package tmt

import (
	machine_translate "ai-transform-backend/pkg/machine-translate"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
	tmt1 "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/tmt/v20180321"
)

type tmtFactory struct {
	secretId  string
	secretKey string
	endpoint  string
	region    string
}

func NewCreateTmtFactory(secretID, secretKey, endpoint, region string) machine_translate.TranslatorFactory {
	return &tmtFactory{
		secretId:  secretID,
		secretKey: secretKey,
		endpoint:  endpoint,
		region:    region,
	}
}

func (f *tmtFactory) CreateTranslator() (machine_translate.Translator, error) {
	credential := common.NewCredential(f.secretId, f.secretKey)
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = f.endpoint
	client, err := tmt1.NewClient(credential, f.region, cpf)
	if err != nil {
		return nil, err
	}
	return &tmt{
		client: client,
	}, nil
}

type tmt struct {
	client *tmt1.Client
}

// 腾讯云限制 默认 5次/秒,单次请求的文本长度总和需要低于6000字符，低于6000字符，低于6000字符
func (t *tmt) TextTranslateBatch(textList []string, sourceLanguage, targetLanguage string) ([]*string, error) {
	request := tmt1.NewTextTranslateBatchRequest()
	request.Source = common.StringPtr(sourceLanguage)
	request.Target = common.StringPtr(targetLanguage)
	request.ProjectId = common.Int64Ptr(0)
	request.SourceTextList = common.StringPtrs(textList)
	response, err := t.client.TextTranslateBatch(request)
	if err != nil {
		return nil, err
	}
	return response.Response.TargetTextList, err
}
