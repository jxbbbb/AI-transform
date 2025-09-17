package cos

import (
	"ai-transform-backend/pkg/storage"
	"context"
	"encoding/base64"
	"github.com/tencentyun/cos-go-sdk-v5"
	"io"
	"mime"
	"net/http"
	url2 "net/url"
	"path"
)

type cosStorageFactory struct {
	bucketUrl string
	secretId  string
	secretKey string
	cdnDomain string
}

func NewCosStorageFactory(bucketUrl, secretId, secretKey, cdnDomain string) storage.StorageFactory {
	return &cosStorageFactory{
		bucketUrl: bucketUrl,
		secretId:  secretId,
		secretKey: secretKey,
		cdnDomain: cdnDomain,
	}
}
func (f *cosStorageFactory) CreateStorage() storage.Storage {
	return newCos(f.bucketUrl, f.secretId, f.secretKey, f.cdnDomain)
}

type cosStorage struct {
	bucketUrl string
	cdnDomain string
	client    *cos.Client
}

func newCos(bucketUrl, secretId, secretKey, cdnDomain string) storage.Storage {
	u, _ := url2.Parse(bucketUrl)
	b := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(b, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  secretId,
			SecretKey: secretKey,
		},
	})

	return &cosStorage{
		bucketUrl: bucketUrl,
		client:    client,
		cdnDomain: cdnDomain,
	}
}

func (s *cosStorage) Upload(r io.Reader, md5Digest []byte, dstPath string) (url string, err error) {
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: mime.TypeByExtension(path.Ext(dstPath)),
		},
		ACLHeaderOptions: &cos.ACLHeaderOptions{},
	}
	if len(md5Digest) != 0 {
		opt.ObjectPutHeaderOptions.ContentMD5 = base64.StdEncoding.EncodeToString(md5Digest)
	}
	_, err = s.client.Object.Put(context.Background(), dstPath, r, opt)
	if err != nil {
		return "", err
	}
	url = s.bucketUrl + dstPath
	if s.cdnDomain != "" {
		url = s.cdnDomain + dstPath
	}
	return url, err
}

func (s *cosStorage) UploadFromFile(filePath string, dstPath string) (url string, err error) {
	opt := &cos.ObjectPutOptions{
		ObjectPutHeaderOptions: &cos.ObjectPutHeaderOptions{
			ContentType: mime.TypeByExtension(path.Ext(dstPath)),
		},
		ACLHeaderOptions: &cos.ACLHeaderOptions{},
	}
	_, err = s.client.Object.PutFromFile(context.Background(), dstPath, filePath, opt)
	if err != nil {
		return "", err
	}
	url = s.bucketUrl + dstPath
	if s.cdnDomain != "" {
		url = s.cdnDomain + dstPath
	}
	return url, err
}
func (s *cosStorage) DownloadFile(objectKey string, dstPath string) error {
	opt := &cos.MultiDownloadOptions{
		ThreadPoolSize: 5,
	}
	_, err := s.client.Object.Download(context.Background(), objectKey, dstPath, opt)
	if err != nil {
		return err
	}
	return nil
}
