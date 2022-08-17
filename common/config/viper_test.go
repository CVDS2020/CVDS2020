package config

import (
	"fmt"
	"github.com/CVDS2020/CVDS2020/common/assert"
	"github.com/CVDS2020/CVDS2020/common/def"
	"gopkg.in/yaml.v3"
	"os"
	"testing"
)

type DB struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

func (d *DB) PreHandle() PreHandlerConfig {
	if d == nil {
		d = new(DB)
	}
	if d.Username == "" {
		d.Username = "root"
	}
	return d
}

func (d *DB) PostHandle() error {
	fmt.Println("post handle")
	return nil
}

type Client struct {
	Addr string
	Port int
	DBS  []*DB
}

func (c *Client) PreHandle() PreHandlerConfig {
	if c == nil {
		c = new(Client)
	}
	def.SetDefault(&c.Addr, "127.0.0.1")
	def.SetDefault(&c.Port, 80)
	//if c.DBS == nil {
	//	c.DBS = map[string]*DB{
	//		"1": nil,
	//	}
	//}
	if len(c.DBS) == 0 {
		c.DBS = append(c.DBS, nil)
	}
	return c
}

type Http struct {
	Addr   string
	Client Client
}

type Config struct {
	*DB   `yaml:"db"`
	*Http `yaml:"http"`
}

func (c *Config) PreHandle() PreHandlerConfig {
	return c
}

func (c *Config) PostHandle() (PostHandlerConfig, error) {
	return c, nil
}

func TestViper(t *testing.T) {

}

func TestConfig(t *testing.T) {
	config := Config{}
	parser := &Parser{}
	parser.AddConfigFile("C:\\Users\\suy\\Documents\\Language\\Go\\common\\config\\config.yaml", TypeYaml)
	parser.Unmarshal(&config)
	os.Stdout.Write(assert.Must(yaml.Marshal(config)))
}
