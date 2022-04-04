package json

import (
	"fmt"

	"github.com/asim/go-micro/v3/config"
	"github.com/asim/go-micro/v3/config/source/file"
)

// json格式的config demo
func main() {
	// load the config from a file source
	if err := config.Load(file.NewSource(
		file.WithPath("./app.conf"),
	)); err != nil {
		fmt.Println(err)
		return
	}

	// define our own host type
	type Info struct {
		Appname    string `json:"appname"`
		ServerPort int    `json:"serverPort"`
	}

	var info Info

	// read a database host
	if err := config.Get("info").Scan(&info); err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(info.Appname, info.ServerPort)
}
