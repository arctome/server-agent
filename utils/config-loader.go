package Utils

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type SAConfig struct {
	Basic   SAConfigBasic    `yaml:"basic"`
	Feature SAConfigService  `yaml:"feature"`
	Servers []SAConfigServer `yaml:"servers"`
}

type SAConfigBasic struct {
	Mode      string `yaml:"mode"`
	Port      string `yaml:"port"`
	Pass      string `yaml:"pass"`
	Salt      string `yaml:"salt"`
	EnableSSL bool   `yaml:"enable_ssl"`
}

type SAConfigService struct {
	Prefork      bool `yaml:"prefork"`
	SinglePageUI bool `yaml:"spa_ui"`
}

type SAConfigServer struct {
	Name      string `yaml:"name"`
	Host      string `yaml:"host"`
	Port      string `yaml:"port"`
	Pass      string `yaml:"pass"`
	UseProxy  bool   `yaml:"use_proxy"`
	EnableSSL bool   `yaml:"enable_ssl"`
	EnableSSH bool   `yaml:"enable_ssh"`
	SSHUser   string `yaml:"ssh_user"`
	SSHPort   string `yaml:"ssh_port"`
	Location  string `yaml:"location"`
}

func LoadConf(file string, cnf interface{}) error {
	yamlFile, err := ioutil.ReadFile(file)
	if err == nil {
		err = yaml.Unmarshal(yamlFile, cnf)
	}
	return err
}
