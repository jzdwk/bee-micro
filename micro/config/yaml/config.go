package yaml

import (
	"fmt"

	yaml "github.com/asim/go-micro/plugins/config/encoder/yaml/v3"
	"github.com/asim/go-micro/v3/config"
	"github.com/asim/go-micro/v3/config/reader"
	"github.com/asim/go-micro/v3/config/reader/json"
	"github.com/asim/go-micro/v3/config/source/file"
)

type Config struct {
	Server Server
	Consul Consul
}

type Server struct {
	Name    string `json:"name"`
	Address string `json:"address"`
}

type Consul struct {
	Address string `json:"address"`
}

func GetInfo() (info Config) {
	// new yaml encoder
	enc := yaml.NewEncoder()

	// new config
	c, _ := config.NewConfig(
		config.WithReader(
			json.NewReader( // json reader for internal config merge
				reader.WithEncoder(enc),
			),
		),
	)

	// load the config from a file source
	if err := c.Load(file.NewSource(
		file.WithPath("./conf/yaml_config/config.yaml"),
	)); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println("data", c.Map())

	if err := c.Get("info").Scan(&info); err != nil {
		fmt.Println(err)
		return
	}

	return info
}
