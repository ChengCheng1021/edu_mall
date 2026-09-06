package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"mall/adaptor"
	"mall/config"
	"time"

	"github.com/go-redis/redis"
)

type IVerify interface {
	SetCaptchaKey(ctx context.Context, key string, value string, expire time.Duration) error
	GetCaptchaKey(ctx context.Context, key string) (string, error)
	SetCaptchaTicket(ctx context.Context, ticket string, value string, expire time.Duration) error
	GetCaptchaTicket(ctx context.Context, ticket string) (string, error)

	SetVerifyCode(ctx context.Context, mobile string, sceneCode string, value interface{}, expire time.Duration) error
	GetVerifyCode(ctx context.Context, mobile string, sceneCode string) (string, error)
	DelVerifyCode(ctx context.Context, mobile string, sceneCode string)
	CheckVerifyFreLimited(ctx context.Context, mobile string) (bool, error)

	SetAdminUserToken(ctx context.Context, userId int64, token string, tokenData string, expire time.Duration) error
	GetAdminUserToken(ctx context.Context, token string) (string, error)
	CleanAdminUserToken(ctx context.Context, userId int64) error

	SetUserToken(ctx context.Context, userId int64, token string, tokenData string, expire time.Duration) error
	GetUserToken(ctx context.Context, token string) (string, error)
	CleanUserToken(ctx context.Context, userId int64) error

	SetScanTask(ctx context.Context, onceID string, value interface{}, expire time.Duration) error
	GetScanTask(ctx context.Context, onceID string, out interface{}) error
	DelScanTask(ctx context.Context, onceID string) error
	SetScanSceneToken(ctx context.Context, sceneToken string, onceID string, expire time.Duration) error
	GetScanSceneToken(ctx context.Context, sceneToken string) (string, error)
	DelScanSceneToken(ctx context.Context, sceneToken string) error

	IncrPasswordErr(ctx context.Context, mobile string, expire time.Duration) (int64, error)
	DeletePasswordErr(ctx context.Context, mobile string) error
}

type Verify struct {
	redis *redis.Client
}

func NewVerify(adaptor adaptor.IAdaptor) *Verify {
	return &Verify{
		redis: adaptor.GetRedis(),
	}
}

func fmtVerifyCaptchaKey(key string) string {
	return fmt.Sprintf("%s:captcha:%s", config.ServerFullName, key)
}

func fmtVerifyCaptchaTicket(key string) string {
	return fmt.Sprintf("%s:captcha:ticket:%s", config.ServerFullName, key)
}
func (v *Verify) SetCaptchaKey(ctx context.Context, key string, value string, expire time.Duration) error {
	redisKey := fmtVerifyCaptchaKey(key)
	return v.redis.Set(redisKey, value, expire).Err()
}
func (v *Verify) GetCaptchaKey(ctx context.Context, key string) (string, error) {
	redisKey := fmtVerifyCaptchaKey(key)
	get, err := v.redis.Get(redisKey).Result()
	if err != nil {
		return "", err
	}
	v.redis.Del(redisKey)
	return get, nil
}
func (v *Verify) SetCaptchaTicket(ctx context.Context, ticket string, value string, expire time.Duration) error {
	redisKey := fmtVerifyCaptchaTicket(ticket)
	return v.redis.Set(redisKey, value, expire).Err()
}
func (v *Verify) GetCaptchaTicket(ctx context.Context, ticket string) (string, error) {
	redisKey := fmtVerifyCaptchaTicket(ticket)
	get, err := v.redis.Get(redisKey).Result()
	if err != nil {
		return "", err
	}
	v.redis.Del(redisKey)
	return get, nil
}

func fmtVerifyVerifyCode(mobile string, sceneCode string) string {
	return fmt.Sprintf("%s:verify:code:%s:%s", config.ServerFullName, mobile, sceneCode)
}

func (v *Verify) SetVerifyCode(ctx context.Context, mobile string, sceneCode string, value interface{}, expire time.Duration) error {
	redisKey := fmtVerifyVerifyCode(mobile, sceneCode)
	return v.redis.Set(redisKey, value, expire).Err()
}
func (v *Verify) GetVerifyCode(ctx context.Context, mobile string, sceneCode string) (string, error) {
	redisKey := fmtVerifyVerifyCode(mobile, sceneCode)
	return v.redis.Get(redisKey).Result()
}
func (v *Verify) DelVerifyCode(ctx context.Context, mobile string, sceneCode string) {
	redisKey := fmtVerifyVerifyCode(mobile, sceneCode)
	_ = v.redis.Del(redisKey)
}

func fmtVerifyFreLimitedKey(keyValue string) string {
	return fmt.Sprintf("%s:verify:frelimit:%s", config.ServerFullName, keyValue)
}
func (v *Verify) CheckVerifyFreLimited(ctx context.Context, keyValue string) (bool, error) {
	redisKey := fmtVerifyFreLimitedKey(keyValue)
	count, err := v.redis.Incr(redisKey).Result()
	if err != nil {
		return true, err
	}
	if count <= 3 { // 一个变量，一个小时最多三条验证码短信
		return false, nil
	}
	_, err = v.redis.Expire(redisKey, time.Hour).Result()
	if err != nil {
		return true, err
	}
	return false, nil
}

// -------- 管理后台用户token ---------//
func fmtVerifyAdminUserToken(token string) string {
	return fmt.Sprintf("%s:admin:user:token:%s", config.ServerFullName, token)
}

func fmtUserMapTokenAdminUser(userId int64) string {
	return fmt.Sprintf("%s:admin:token:user:%d", config.ServerFullName, userId)
}
func (v *Verify) SetAdminUserToken(ctx context.Context, userID int64, token string, tokenData string, expire time.Duration) error {
	redisKey := fmtVerifyAdminUserToken(token)
	_, err := v.redis.Set(redisKey, tokenData, expire).Result()
	if err != nil {
		return err
	}
	userMapTokenKey := fmtUserMapTokenAdminUser(userID)
	return v.redis.Set(userMapTokenKey, token, expire).Err()
}
func (v *Verify) GetAdminUserToken(ctx context.Context, token string) (string, error) {
	redisKey := fmtVerifyAdminUserToken(token)
	get, err := v.redis.Get(redisKey).Result()
	if err != nil {
		return "", err
	}
	return get, nil
}

func (v *Verify) CleanAdminUserToken(ctx context.Context, userId int64) error {
	userMapTokenKey := fmtUserMapTokenAdminUser(userId)
	token, err := v.redis.Get(userMapTokenKey).Result()
	if err != nil {
		return err
	}
	redisKey := fmtVerifyAdminUserToken(token)
	return v.redis.Del(redisKey, userMapTokenKey).Err()
}

// -------- C端用户token ---------//
func fmtVerifyUserToken(token string) string {
	return fmt.Sprintf("%s:user:token:%s", config.ServerFullName, token)
}

func fmtUserMapTokenUser(userId int64) string {
	return fmt.Sprintf("%s:token:user:%d", config.ServerFullName, userId)
}
func (v *Verify) SetUserToken(ctx context.Context, userID int64, token string, tokenData string, expire time.Duration) error {
	redisKey := fmtVerifyUserToken(token)
	_, err := v.redis.Set(redisKey, tokenData, expire).Result()
	if err != nil {
		return err
	}
	userMapTokenKey := fmtUserMapTokenUser(userID)
	return v.redis.Set(userMapTokenKey, token, expire).Err()
}
func (v *Verify) GetUserToken(ctx context.Context, token string) (string, error) {
	redisKey := fmtVerifyUserToken(token)
	get, err := v.redis.Get(redisKey).Result()
	if err != nil {
		return "", err
	}
	return get, nil
}

func (v *Verify) CleanUserToken(ctx context.Context, userId int64) error {
	userMapTokenKey := fmtUserMapTokenUser(userId)
	token, err := v.redis.Get(userMapTokenKey).Result()
	if err != nil {
		return err
	}
	redisKey := fmtVerifyUserToken(token)
	return v.redis.Del(redisKey, userMapTokenKey).Err()
}

func fmtVerifyScanTask(onceID string) string {
	return fmt.Sprintf("%s:user:wechat:scan:%s", config.ServerFullName, onceID)
}

func fmtVerifyScanSceneToken(sceneToken string) string {
	return fmt.Sprintf("%s:user:wechat:scan:scene:%s", config.ServerFullName, sceneToken)
}

func (v *Verify) SetScanTask(ctx context.Context, onceID string, value interface{}, expire time.Duration) error {
	bs, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return v.redis.Set(fmtVerifyScanTask(onceID), string(bs), expire).Err()
}

func (v *Verify) GetScanTask(ctx context.Context, onceID string, out interface{}) error {
	get, err := v.redis.Get(fmtVerifyScanTask(onceID)).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(get), out)
}

func (v *Verify) DelScanTask(ctx context.Context, onceID string) error {
	return v.redis.Del(fmtVerifyScanTask(onceID)).Err()
}

func (v *Verify) SetScanSceneToken(ctx context.Context, sceneToken string, onceID string, expire time.Duration) error {
	return v.redis.Set(fmtVerifyScanSceneToken(sceneToken), onceID, expire).Err()
}

func (v *Verify) GetScanSceneToken(ctx context.Context, sceneToken string) (string, error) {
	return v.redis.Get(fmtVerifyScanSceneToken(sceneToken)).Result()
}

func (v *Verify) DelScanSceneToken(ctx context.Context, sceneToken string) error {
	return v.redis.Del(fmtVerifyScanSceneToken(sceneToken)).Err()
}

func fmtVerifyPasswordErr(mobile string) string {
	return fmt.Sprintf("%s:admin:user:password:errcount:%s", config.ServerFullName, mobile)
}

func (v *Verify) IncrPasswordErr(ctx context.Context, mobile string, expire time.Duration) (int64, error) {
	redisKey := fmtVerifyPasswordErr(mobile)
	incr, err := v.redis.Incr(redisKey).Result()
	if err != nil {
		return 0, err
	}
	if incr == 1 {
		v.redis.Expire(redisKey, expire)
	}
	return incr, err
}
func (v *Verify) DeletePasswordErr(ctx context.Context, mobile string) error {
	redisKey := fmtVerifyPasswordErr(mobile)
	return v.redis.Del(redisKey).Err()
}
