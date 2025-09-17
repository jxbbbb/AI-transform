package tasr

import (
	"ai-transform-backend/pkg/asr"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/errors"
	"github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/common/profile"
)
import tasr "github.com/tencentcloud/tencentcloud-sdk-go/tencentcloud/asr/v20190614"

type asrFactory struct {
	secretId  string
	secretKey string
	endpoint  string
	region    string
}

func NewCreateAsrFactory(secretId, secretKey, endpoint, region string) asr.AsrFactory {
	return &asrFactory{
		secretId:  secretId,
		secretKey: secretKey,
		endpoint:  endpoint,
		region:    region,
	}
}
func (f *asrFactory) CreateAsr() (asr.Asr, error) {
	credential := common.NewCredential(
		f.secretId,
		f.secretKey,
	)
	// 实例化一个client选项，可选的，没有特殊需求可以跳过
	cpf := profile.NewClientProfile()
	cpf.HttpProfile.Endpoint = f.endpoint
	// 实例化要请求产品的client对象,clientProfile是可选的
	client, err := tasr.NewClient(credential, f.region, cpf)
	if err != nil {
		return nil, err
	}
	return &tAsr{
		client: client,
	}, nil

}

type tAsr struct {
	client *tasr.Client
}

// TODO 腾讯云限制 默认20次/秒
func (a *tAsr) Asr(url string) (taskID uint64, err error) {

	// 实例化一个请求对象,每个接口都会对应一个request对象
	request := tasr.NewCreateRecTaskRequest()
	request.EngineModelType = common.StringPtr("16k_zh")
	request.ChannelNum = common.Uint64Ptr(1)
	request.ResTextFormat = common.Uint64Ptr(0)
	request.SourceType = common.Uint64Ptr(0)
	request.Url = common.StringPtr(url)
	request.SpeakerDiarization = common.Int64Ptr(0)
	request.ConvertNumMode = common.Int64Ptr(1)

	// 返回的resp是一个CreateRecTaskResponse的实例，与请求对象对应
	response, err := a.client.CreateRecTask(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return 0, err
	}
	if err != nil {
		return 0, err
	}
	return *response.Response.Data.TaskId, nil
}

// TODO 腾讯云限制 默认50次/秒
func (a *tAsr) GetAsrResult(taskID uint64) (result string, status asr.AsrStatus, err error) {
	request := tasr.NewDescribeTaskStatusRequest()
	request.TaskId = common.Uint64Ptr(taskID)
	response, err := a.client.DescribeTaskStatus(request)
	if _, ok := err.(*errors.TencentCloudSDKError); ok {
		return "", convertAsrStatus(*response.Response.Data.Status), err
	}
	if err != nil {
		return "", asr.FAILED, err
	}
	return *response.Response.Data.Result, convertAsrStatus(*response.Response.Data.Status), nil
}
func convertAsrStatus(status int64) asr.AsrStatus {
	switch status {
	case 0:
		return asr.WAITING
	case 3:
		return asr.FAILED
	case 1:
		return asr.DOING
	case 2:
		return asr.SUCCESS
	default:
		return asr.FAILED
	}
}
