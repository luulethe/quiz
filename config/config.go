package config

import (
	"io/ioutil"
	"os"

	"github.com/luulethe/quiz/quiz_lib/db"
	"gopkg.in/yaml.v2"
)

// Configuration defines the config
type Configuration struct {
	Debug           bool            `yaml:"debug"` // false: info level, true: debug level
	ProfileAddr     string          `yaml:"pprof"`
	Listen          string          `yaml:"listen"`
	MySQL           []db.Config     `yaml:"mysql"`
	SentryDNS       string          `yaml:"sentry_dns"`
	GeoIPServerAddr string          `yaml:"geoip_server_addr"`
	QuizKafka       *ConsumerConfig `yaml:"quiz_kafka"`
}

type ConsumerConfig struct {
	Brokers       string `yaml:"brokers"`
	Version       string `yaml:"version"`
	ConsumerGroup string `yaml:"consumer_group"`
}

func (c *Configuration) LoadFromFile(confPath string) error {
	configContent, err := ioutil.ReadFile(confPath)
	configContent = []byte(os.ExpandEnv(string(configContent)))

	if err != nil {
		return err
	}
	return yaml.Unmarshal(configContent, c)
}
