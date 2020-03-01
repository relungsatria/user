package config

import (
	"fmt"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"log"
	"sync"
)

type Config struct {
	Database map[string]Database
	Redis map[string]Redis
}

type Database struct {
	Master string
	MaxLifeTime int
	MaxIdle int
	MaxOpen int
}

type Redis struct {
	Address string
}

const (
	configFilePath = "./config/file/"

	UserDB = "userdb"
	UserRedis = "userredis"
)

var config *Config
var once sync.Once

func GetConfig() *Config{
	once.Do(func() {
		config = &Config{}
		Load("database.yml", &config.Database)
		Load("redis.yml", &config.Redis)
	})
	log.Println(fmt.Sprintf("%+v", config))
	return config
}

func Load(fileName string, container interface{}) {
	file, err := ioutil.ReadFile(configFilePath+fileName)
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(file, container)
	if err != nil {
		log.Fatal(err)
	}
}