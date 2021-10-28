package confs

import "github.com/BurntSushi/toml"

type BnWsAPIConf struct {
	Scheme    string   `toml:"scheme"`
	BaseURL   string   `toml:"base_url"`
	Endpoints []string `toml:"end_points"`
}

type KafkaServiceConf struct {
	KafkaBrokers []string `toml:"brokers"`
	Topic        string   `toml:"topic"`
}

type ServerConf struct {
	BnAPIConf BnWsAPIConf      `toml:"bn_api_conf"`
	KafkaConf KafkaServiceConf `toml:"kafka_conf"`
}

func (c *ServerConf) LoadFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}

type KafkaConsumerConf struct {
	KafkaBrokers []string `toml:"brokers"`
	Topics       []string `toml:"topics"`
	GroupId      string   `toml:"group_id"`
}

type InfluxDBConf struct {
	Token    string `toml:"token"`
	Bucket   string `toml:"bucket"`
	Org      string `toml:"org"`
	SeverURL string `toml:"server_url"`
}

type BnConsumerConf struct {
	KafkaConf  KafkaConsumerConf `toml:"kafka_conf"`
	InfluxConf InfluxDBConf      `toml:"influx_conf"`
}

func (c *BnConsumerConf) LoadFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}

type TickersSvrConf struct {
	Host       string       `toml:"host"`
	Port       int          `toml:"port"`
	InfluxConf InfluxDBConf `toml:"influx_conf"`
}

func (c *TickersSvrConf) LoadFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}
