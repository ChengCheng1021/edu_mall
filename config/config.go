package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/gogf/gf/util/gconv"
	"github.com/spf13/viper"
	_ "github.com/spf13/viper/remote"
	"gopkg.in/yaml.v3"
)

const (
	ServerName     = "mall"
	ServerFullName = "edu.mall"
)

var (
	etcdKey         string
	etcdEnv         string
	etcdAddr        string
	localConfigPath string
	GlobalConfig    Config
)

type Config struct {
	Server  Server            `yaml:"server"`
	Mysql   Mysql             `yaml:"mysql"`
	Redis   Redis             `yaml:"redis"`
	BizConf BizConf           `yaml:"biz_conf"`
	AppConf map[int32]AppConf `yaml:"app_conf"`
	Storage Storage           `yaml:"storage"`
	AliPay  AliPay            `yaml:"ali_pay"`
}

type Server struct {
	HttpPort    int    `yaml:"http_port"`
	Env         string `yaml:"env"`
	EnablePprof bool   `yaml:"enable_pprof"`
	LogLevel    string `yaml:"log_level"`
}

type Mysql struct {
	Dialect  string `yaml:"dialect"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Database string `yaml:"database"`
	Charset  string `yaml:"charset"`
	ShowSql  bool   `yaml:"show_sql"`
	MaxOpen  int    `yaml:"max_open"`
	MaxIdle  int    `yaml:"max_idle"`
}

type Storage struct {
	SecretID  string `yaml:"secret_id"`
	SecretKey string `yaml:"secret_key"`
	AppID     string `yaml:"app_id"`
	Bucket    Bucket `yaml:"bucket"`
}

type Bucket struct {
	Region     string            `yaml:"region"`
	BucketName string            `yaml:"bucket_name"`
	Domain     string            `yaml:"domain"`
	CdnDomain  string            `yaml:"cdn_domain"`
	SignKey    string            `yaml:"sign_key"`
	Paths      map[string]string `yaml:"paths"`
}

type AliPay struct {
	AppID            string `yaml:"app_id"`
	ApiKey           string `yaml:"api_key"`
	ReturnURL        string `yaml:"return_url"`
	Sandbox          bool   `yaml:"sandbox"`    // 是否生产环境
	NotifyURL        string `yaml:"notify_url"` // 1>回调，2>定时查询（每10秒钟）
	PublicKey        string `yaml:"public_key"`
	AppPublicCert    string `yaml:"app_public_cert"`
	AlipayPublicCert string `yaml:"alipay_public_cert"`
	AlipayRootCert   string `yaml:"alipay_root_cert"`
}

func (m *Mysql) GetDsn() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=Local",
		m.User, m.Password, m.Host, m.Port, m.Database, m.Charset)
}

type AppConf struct {
	AppType   string `yaml:"app_type"`
	AppName   string `yaml:"app_name"`
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

type BizConf struct {
	LarkGroupID  string `yaml:"lark_group_id"`
	MobileSecret string `yaml:"mobile_secret"`
	BizSecret    string `yaml:"biz_secret"`
}

type Redis struct {
	Addr    string `yaml:"addr"`
	PWD     string `yaml:"password"`
	DBIndex int    `yaml:"db_index"`
	MaxIdle int    `yaml:"max_idle"`
	MaxOpen int    `yaml:"max_open"`
}

type WechatPay struct {
	AppID        string `yaml:"app_id"`
	MchID        string `yaml:"mch_id"`
	ApiKey       string `yaml:"api_key"`
	CertSerialNo string `yaml:"cert_serial_no"`
	CallbackUrl  string `yaml:"callback_url"` // 1>回调，2>定时查询（每10秒钟）
	IsProd       bool   `yaml:"is_prod"`      // 是否生产环境
}

type AliSms struct {
	Endpoint     string `yaml:"endpoint"`
	AccessKey    string `yaml:"access_key"`
	AccessSecret string `yaml:"access_secret"`
}

type WechatApp struct {
	AppName   string `yaml:"app_name"`
	AppID     string `yaml:"app_id"`
	AppSecret string `yaml:"app_secret"`
}

func init() {
	flag.StringVar(&localConfigPath, "c", ServerName+"_local.yml", "default config path")
	flag.StringVar(&etcdAddr, "r", os.Getenv("ETCD_ADDR"), "default etcd address")
	flag.StringVar(&etcdEnv, "k", os.Getenv("ETCD_ENV"), "etcd env")
}

func InitConfig() *Config {
	var (
		err      error
		tempConf = &Config{}
		vipConf  = viper.New()
	)
	vipConf.SetConfigType("yaml")
	flag.Parse()
	etcdKey = fmt.Sprintf("/configs/%s/%s/system", ServerFullName, etcdEnv)
	// etcd地址存在，优先使用etcd的配置
	if etcdAddr != "" {
		tempConf, err = getFromRemoteAndWatchUpdate(vipConf)
		if err != nil {
			panic(err)
		}
		return tempConf
	}

	// 从本地获取
	tempConf, err = getFromLocal()
	if err != nil {
		panic(err)
	}
	return tempConf
}

func getFromRemoteAndWatchUpdate(v *viper.Viper) (*Config, error) {
	tempConf := Config{}
	if err := v.AddRemoteProvider("etcd3", etcdAddr, etcdKey); err != nil {
		return nil, err
	}
	if err := v.ReadRemoteConfig(); err != nil {
		return nil, err
	}

	// 反序列化到结构体
	err := v.Unmarshal(&tempConf, func(config *mapstructure.DecoderConfig) {
		config.TagName = "yaml"
	})

	if err != nil {
		return nil, err
	}

	go func() {
		for {
			time.Sleep(time.Minute * 1)
			if err := v.WatchRemoteConfig(); err == nil {
				_ = v.Unmarshal(&GlobalConfig)
				fmt.Println(">>> etcd config hot-reloaded: ", gconv.String(GlobalConfig))
			}
		}
	}()
	return &tempConf, nil
}

func getFromLocal() (*Config, error) {
	tempConf := Config{}
	if _, err := os.Stat(localConfigPath); err == nil {
		content, err := os.ReadFile(localConfigPath)
		if err != nil {
			return nil, err
		}
		err = yaml.Unmarshal(content, &tempConf)
		if err != nil {
			return nil, err
		}
		return &tempConf, nil
	}
	return nil, fmt.Errorf("local config file not found ,file_name: %s", localConfigPath)
}
