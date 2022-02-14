package SAServers

import (
	"log"
	Utils "server-agent/utils"
)

func LoadServerList() []Utils.SAConfigServer {
	cfg := &Utils.SAConfig{}
	if err := Utils.LoadConf("conf.yml", cfg); err != nil {
		log.Panicln(err)
	}

	return cfg.Servers
}
