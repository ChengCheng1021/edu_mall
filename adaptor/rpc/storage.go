package rpc

import (
	"context"
	"errors"
	"fmt"
	"mall/adaptor"
	"mall/config"
	"mall/service/do"
	"mall/utils/logger"
	"mall/utils/tools"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
	sts "github.com/tencentyun/qcloud-cos-sts-sdk/go"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type IStorage interface {
	// 获取临时上传密钥
	GetTempSecret(ctx context.Context, req *do.GetTempSecret) (*do.TempSecret, error)
	// 获取文件预览地址
	GetPreviewUrl(ctx context.Context, req *do.GetPreviewUrl) (map[string]string, error)
	// 删除文件--视频文件太大，业务层判断是否已经更新视频，更新则原来的文件
	DeleteFile(ctx context.Context, req *do.DeleteFile) error
}

type Storage struct {
	db   *gorm.DB
	conf *config.Config
}

func NewStorage(adaptor adaptor.IAdaptor) *Storage {

	return &Storage{
		conf: adaptor.GetConfig(),
	}
}

func (s *Storage) GetTempSecret(ctx context.Context, req *do.GetTempSecret) (*do.TempSecret, error) {
	client, err := s.getPreviewClient(ctx)
	if err != nil {
		return nil, err
	}
	path := s.conf.Storage.Bucket.Paths[req.Scene]
	idx := strings.LastIndex(req.FileName, ".")
	postFix := req.FileName[idx+1:]
	fileKey := path + "/" + tools.UUIDHex() + "." + postFix

	stClient := sts.NewClient(s.conf.Storage.SecretID, s.conf.Storage.SecretKey, nil)
	res, err := stClient.GetCredential(s.getCredentialOptions(path))
	if err != nil {
		return nil, err
	}
	var (
		fileUrl    string
		expireTime = time.Hour
	)
	if req.Scene == "intro" {
		expireTime = time.Hour * 24 * 365 * 10
	}
	pre, err := client.Object.GetPresignedURL(
		ctx,
		http.MethodGet,
		fileKey,
		s.conf.Storage.SecretID,
		s.conf.Storage.SecretKey,
		expireTime, nil)
	if err != nil {
		return nil, err
	}
	fileUrl = pre.String()
	//}else {
	//	fileMap, err := s.GetPreviewUrlWithAuth(ctx, &do.GetPreviewUrl{
	//		Keys:        []string{fileKey},
	//		ExpireHours: 2,
	//	})
	//	if err != nil {
	//		return nil, err
	//	}
	//	fileUrl = fileMap[fileKey]
	//}
	return &do.TempSecret{
		Bucket:        s.conf.Storage.Bucket.BucketName,
		Region:        s.conf.Storage.Bucket.Region,
		Key:           fileKey,
		FileUrl:       fileUrl,
		ExpiredTime:   int64(res.ExpiredTime),
		StartTime:     int64(res.StartTime),
		SecretId:      res.Credentials.TmpSecretID,
		SecretKey:     res.Credentials.TmpSecretKey,
		SecurityToken: res.Credentials.SessionToken,
	}, nil
}

func (s *Storage) getCredentialOptions(path string) *sts.CredentialOptions {
	return &sts.CredentialOptions{
		DurationSeconds: int64(time.Hour.Seconds()),
		Region:          s.conf.Storage.Bucket.Region,
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
						fmt.Sprintf("qcs::cos:%s:uid/%s:%s/%s/*",
							s.conf.Storage.Bucket.Region,
							s.conf.Storage.AppID,
							s.conf.Storage.Bucket.BucketName,
							path),
					},
					// 开始构建生效条件 condition
					// 关于 condition 的详细设置规则和COS支持的condition类型可以参考https://cloud.tencent.com/document/product/436/71306
					Condition: map[string]map[string]interface{}{},
				},
			},
		},
	}
}

func (s *Storage) getPreviewClient(ctx context.Context) (*cos.Client, error) {
	u, err := url.Parse(s.conf.Storage.Bucket.CdnDomain)
	if err != nil {
		return nil, err
	}
	baseUrl := &cos.BaseURL{BucketURL: u}
	previewClient := cos.NewClient(baseUrl, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  s.conf.Storage.SecretID,  // 用户的 SecretId，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考 https://cloud.tencent.com/document/product/598/37140
			SecretKey: s.conf.Storage.SecretKey, // 用户的 SecretKey，建议使用子账号密钥，授权遵循最小权限指引，降低使用风险。子账号密钥获取可参考 https://cloud.tencent.com/document/product/598/37140
		},
	})
	return previewClient, nil
}

func (s *Storage) GetPreviewUrlWithAuth(ctx context.Context, req *do.GetPreviewUrl) (map[string]string, error) {
	loc, _ := time.LoadLocation("Asia/Shanghai")
	_, err := url.Parse(s.conf.Storage.Bucket.CdnDomain)
	if err != nil {
		return nil, err
	}
	fileKeyMap := make(map[string]string)
	for _, key := range req.Keys {
		if len(key) == 0 {
			continue
		}
		url, err := tools.GenerateSignedURL(tools.CategoryD, tools.GenerateURLConfig{
			BaseURL:   s.conf.Storage.Bucket.CdnDomain,
			Path:      key,
			Suffix:    "",
			Key:       s.conf.Storage.Bucket.SignKey,
			SignKey:   "sign",
			TimeKey:   "t",
			TTLFormat: 10,
			RandStr:   tools.UUIDHex(),
			UID:       0,
			Location:  loc,
		})
		if err != nil {
			return nil, err
		}
		fileKeyMap[key] = url
	}
	return fileKeyMap, nil
}

func (s *Storage) GetPreviewUrl(ctx context.Context, req *do.GetPreviewUrl) (map[string]string, error) {
	//return s.GetPreviewUrlWithAuth(ctx, req)
	client, err := s.getPreviewClient(ctx)
	if err != nil {
		return nil, err
	}
	fileKeyMap := make(map[string]string)
	for _, fileKey := range req.Keys {
		if fileKey == "" {
			continue
		}
		pre, err := client.Object.GetPresignedURL(
			ctx,
			http.MethodGet,
			fileKey,
			s.conf.Storage.SecretID,
			s.conf.Storage.SecretKey,
			time.Duration(req.ExpireHours)*time.Hour, nil)
		if err != nil {
			return nil, err
		}
		fileKeyMap[fileKey] = pre.String()
	}
	return fileKeyMap, nil
}

func (s *Storage) DeleteFile(ctx context.Context, req *do.DeleteFile) error {
	u, err := url.Parse(s.conf.Storage.Bucket.Domain)
	if err != nil {
		return errors.New("DeleteFiles bucket config Domain error")
	}
	baseUrl := &cos.BaseURL{BucketURL: u}
	client := cos.NewClient(baseUrl, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  s.conf.Storage.SecretID,
			SecretKey: s.conf.Storage.SecretKey,
		},
	})
	for _, srcKey := range req.Keys {
		if srcKey == "" {
			continue
		}
		resp, err := client.Object.Delete(context.Background(), srcKey, nil)
		if err != nil {
			logger.Error("DeleteFile Delete error", zap.Error(err), zap.Any("resp", resp))
		}
	}
	return nil
}
