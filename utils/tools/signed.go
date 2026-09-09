package tools

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type URLSignCategory string

const (
	CategoryA URLSignCategory = "A"
	CategoryB URLSignCategory = "B"
	CategoryC URLSignCategory = "C"
	CategoryD URLSignCategory = "D"
)

type GenerateURLConfig struct {
	BaseURL string // 例如：http://www.test.com
	Path    string // 例如：/1.txt
	Suffix  string // 例如：?a=1&b=2 或 a=1&b=2

	Key string // 鉴权密钥

	// 可选：格式必须是 20250714103456
	// 不传则使用当前时间
	Timestamp string

	SignKey string // 默认 sign
	TimeKey string // 默认 t

	// Type A 使用
	RandStr string // 默认 123abc
	UID     int    // 默认 0

	// Type D 使用
	// 10 或 16，默认 10
	TTLFormat int

	// 可选：不传则使用 time.Local
	Location *time.Location
}

func GenerateSignedURL(category URLSignCategory, cfg GenerateURLConfig) (string, error) {
	if cfg.BaseURL == "" {
		return "", fmt.Errorf("BaseURL cannot be empty")
	}
	if cfg.Path == "" {
		return "", fmt.Errorf("Path cannot be empty")
	}
	if cfg.Key == "" {
		return "", fmt.Errorf("Key cannot be empty")
	}

	baseURL := strings.TrimRight(cfg.BaseURL, "/")
	path := normalizePath(cfg.Path)
	suffix := normalizeSuffix(cfg.Suffix)

	signKey := cfg.SignKey
	if signKey == "" {
		signKey = "sign"
	}

	timeKey := cfg.TimeKey
	if timeKey == "" {
		timeKey = "t"
	}

	randStr := cfg.RandStr
	if randStr == "" {
		randStr = "123abc"
	}

	ttlFormat := cfg.TTLFormat
	if ttlFormat == 0 {
		ttlFormat = 10
	}

	loc := cfg.Location
	if loc == nil {
		loc = time.Local
	}

	nowUnix, err := resolveTimestamp(cfg.Timestamp, loc)
	if err != nil {
		return "", err
	}

	switch category {
	case CategoryA:
		ts := strconv.FormatInt(nowUnix, 10)

		raw := fmt.Sprintf("%s-%s-%s-%d-%s", path, ts, randStr, cfg.UID, cfg.Key)
		sign := md5Hex(raw)

		requestURL := fmt.Sprintf(
			"%s%s?%s=%s-%s-%d-%s",
			baseURL,
			path,
			signKey,
			ts,
			randStr,
			cfg.UID,
			sign,
		)

		return requestURL, nil

	case CategoryB:
		ts := time.Unix(nowUnix, 0).In(loc).Format("200601021504")

		raw := fmt.Sprintf("%s%s%s", cfg.Key, ts, path)
		sign := md5Hex(raw)

		requestURL := fmt.Sprintf(
			"%s/%s/%s%s%s",
			baseURL,
			ts,
			sign,
			path,
			suffix,
		)

		return requestURL, nil

	case CategoryC:
		ts := strconv.FormatInt(nowUnix, 16)

		raw := fmt.Sprintf("%s%s%s", cfg.Key, path, ts)
		sign := md5Hex(raw)

		requestURL := fmt.Sprintf(
			"%s/%s/%s%s%s",
			baseURL,
			sign,
			ts,
			path,
			suffix,
		)

		return requestURL, nil

	case CategoryD:
		var ts string
		if ttlFormat == 16 {
			ts = strconv.FormatInt(nowUnix, 16)
		} else {
			ts = strconv.FormatInt(nowUnix, 10)
		}

		raw := fmt.Sprintf("%s%s%s", cfg.Key, path, ts)
		sign := md5Hex(raw)

		requestURL := fmt.Sprintf(
			"%s%s?%s=%s&%s=%s",
			baseURL,
			path,
			signKey,
			sign,
			timeKey,
			ts,
		)

		return requestURL, nil

	default:
		return "", fmt.Errorf("unsupported category: %s", category)
	}
}

func resolveTimestamp(ts string, loc *time.Location) (int64, error) {
	if ts == "" {
		return time.Now().In(loc).Unix(), nil
	}

	t, err := time.ParseInLocation("20060102150405", ts, loc)
	if err != nil {
		return 0, fmt.Errorf("invalid timestamp format, expected yyyyMMddHHmmss, got %s: %w", ts, err)
	}

	return t.Unix(), nil
}

func normalizePath(path string) string {
	if path == "" {
		return path
	}
	if !strings.HasPrefix(path, "/") {
		return "/" + path
	}
	return path
}

func normalizeSuffix(suffix string) string {
	if suffix == "" {
		return ""
	}
	if strings.HasPrefix(suffix, "?") || strings.HasPrefix(suffix, "&") {
		return suffix
	}
	return "?" + suffix
}

func md5Hex(s string) string {
	hash := md5.Sum([]byte(s))
	return hex.EncodeToString(hash[:])
}
