package main

import (
	"fmt"
	"log"
	"os"
	"os/user"
	"path"

	"code.google.com/p/gcfg"
	"github.com/awilliams/linode-ssh-config/api"
	"github.com/mgutz/ansi"
)

const CONFIG_NAME = ".linode-ssh-config.ini"

type Configuration struct {
	ApiKey       string   `gcfg:"api-key"`
	DisplayGroup []string `gcfg:"display-group"`
	User         string   `gcfg:"user"`
	IdentityFile string   `gcfg:"identity-file"`
}

func (self Configuration) ContainsDisplayGroup(g string) bool {
	if len(self.DisplayGroup) == 0 {
		return true
	}
	for _, configDisplayGroup := range self.DisplayGroup {
		if configDisplayGroup == g {
			return true
		}
	}
	return false
}

func loadConfig() (*Configuration, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	configPath := path.Join(usr.HomeDir, CONFIG_NAME)
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("No config file found at: %s", configPath)
	}

	var iniConfig struct {
		Linode Configuration
	}
	err = gcfg.ReadFileInto(&iniConfig, configPath)
	if err != nil {
		return nil, err
	}

	return &iniConfig.Linode, nil
}

func prettyPrintLinodes(l api.Linodes) {
	for displayGroup, linodes := range l {
		fmt.Printf("%s\t[%d]\n\n", ansi.Color(displayGroup, "green"), len(linodes))
		for _, linode := range linodes {
			labelColor := "magenta"
			if linode.Status != 1 {
				labelColor = "blue"
			}
			fmt.Printf(" * %-25s\tRunning=%v, Ram=%d, LinodeId=%d\n", ansi.Color(linode.Label, labelColor), linode.Status == 1, linode.Ram, linode.Id)
			for _, ip := range linode.Ips {
				var ipType string
				if ip.Public == 1 {
					ipType = "Public"
				} else {
					ipType = "Private"
				}
				fmt.Printf("   %-15s\t%s\n", ip.Ip, ipType)
			}
			fmt.Println("")
		}
	}
}

func sshConfigPrintLinodes(config Configuration, l api.Linodes) {
	sshConfig, err := NewSSHConfig(config, l)
	if err != nil {
		log.Fatal(err)
	}
	bytes, err := sshConfig.render()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Print(string(bytes))
}

func sshConfigUpdate(config Configuration, l api.Linodes) {
	sshConfig, err := NewSSHConfig(config, l)
	if err != nil {
		log.Fatal(err)
	}
	err = sshConfig.update()
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Updated %s with %d Linodes\n", sshConfig.Path, l.Size())
}

func linodes(config *Configuration) api.Linodes {
	linodes, err := api.FetchLinodesWithIps(config.ApiKey)
	if err != nil {
		log.Fatal(err)
	}
	return linodes
}

func printHelp() {
	help := `Generate your .ssh/config file with aliases to your Linodes
  
  Usage:        
  (no args)     Prints generated ssh config to stdout
  --pp          Nicely formatted list of linodes
  --update      Writes generated ssh config to ~/.ssh/config
  --help        Print this message
 `
	fmt.Println(help)
}

func main() {
	config, err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}

	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "--pp":
			prettyPrintLinodes(linodes(config))
		case "--update":
			sshConfigUpdate(*config, linodes(config))
		case "help":
			printHelp()
		case "-h":
			printHelp()
		case "--help":
			printHelp()
		default:
			printHelp()
			log.Fatal("Unrecognized argument")
		}
	} else {
		sshConfigPrintLinodes(*config, linodes(config))
	}
}
