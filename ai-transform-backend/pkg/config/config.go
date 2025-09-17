package config

import (
	"github.com/spf13/viper"
	"log"
)

type Config struct {
	Http struct {
		IP   string
		Port int
		Mode string
	}
	Mysql struct {
		DSN         string
		MaxLifeTime int
		MaxOpenConn int
		MaxIdleConn int
	}
	Log struct {
		Level   string
		LogPath string `mapstructure:"logPath"`
	} `mapstructure:"log"`
	Cos struct {
		SecretId  string
		SecretKey string
		Region    string
		BucketUrl string
		Bucket    string
		CDNDomain string
		AppID     string
	}
	Asr struct {
		SecretId  string
		SecretKey string
		Region    string
		Endpoint  string
		Modals    []string
	}
	Tmt struct {
		SecretID  string
		SecretKey string
		Endpoint  string
		Region    string
	}
	DependOn struct {
		GPT      []string
		ReferWav struct {
			Address string
		}
		User struct {
			Address string
		}
	}
	ExternalKafka struct {
		User          string
		Pwd           string
		SaslMechanism string
		Address       []string
	}
	Kafka struct {
		User          string
		Pwd           string
		SaslMechanism string
		Address       []string
	}
}

var conf *Config

func InitConfig(filePath string, typ ...string) {
	v := viper.New()
	v.SetConfigFile(filePath)
	if len(typ) > 0 {
		v.SetConfigType(typ[0])
	}
	err := v.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	conf = &Config{}
	err = v.Unmarshal(conf)
	if err != nil {
		log.Fatal(err)
	}

}

func GetConfig() *Config {
	return conf
}
