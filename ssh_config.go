package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/user"
	"path"
	"text/template"

	"github.com/awilliams/linode-ssh-config/api"
)

type SSHConfig struct {
	linodes api.Linodes
	path    string
	config  Configuration
}

const SSH_CONFIG_PATH = ".ssh/config"

func NewSSHConfig(config Configuration, linodes api.Linodes) (*SSHConfig, error) {
	usr, err := user.Current()
	if err != nil {
		return nil, err
	}
	path := path.Join(usr.HomeDir, SSH_CONFIG_PATH)

	return &SSHConfig{linodes: linodes, path: path, config: config}, nil
}

// write to the rendered config to disk, making a backup if possible
func (self *SSHConfig) update() error {
	if fileExists(self.path) {
		err := copyFile(self.path, self.path+".linode-ssh-config.bak")
		if err != nil {
			return err
		}
	}
	contents, err := self.render()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(self.path, contents, 0644)
}

// combine the user's config and the generated config
func (self *SSHConfig) render() ([]byte, error) {
	users, err := self.usersConfig()
	if err != nil {
		return nil, err
	}

	generated, err := self.generatedConfig()
	if err != nil {
		return nil, err
	}

	return append(users, generated...), nil
}

var START_TOKEN []byte = []byte("##### START GENERATED LINODE-SSH-CONFIG #####")
var END_TOKEN []byte = []byte("##### END GENERATED LINODE-SSH-CONFIG #####")

// read the user's .ssh/config file, and strip out any previously generated config
func (self *SSHConfig) usersConfig() ([]byte, error) {
	if !fileExists(self.path) {
		return []byte{}, nil
	}

	file, err := os.Open(self.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	insideConfigBlock := false

	strippedBuf := new(bytes.Buffer)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		if !insideConfigBlock && bytes.Equal(scanner.Bytes(), START_TOKEN) {
			insideConfigBlock = true
		}

		if !insideConfigBlock {
			_, err := strippedBuf.Write(append(scanner.Bytes(), '\n'))
			if err != nil {
				return nil, err
			}
		}

		if insideConfigBlock && bytes.Equal(scanner.Bytes(), END_TOKEN) {
			insideConfigBlock = false
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return strippedBuf.Bytes(), nil
}

type sshEntry struct {
	Host    string
	Id      int
	KeyVals map[string]string
}

const entryTemplate = `Host {{ .Host }}{{ range $k, $v := .KeyVals }}
        {{ $k }} {{ $v }}{{ end }}
        
`

// create the generated config section
func (self *SSHConfig) generatedConfig() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(append(START_TOKEN, []byte{'\n', '\n'}...))

	template := template.Must(template.New("entry").Parse(entryTemplate))
	for displayGroup, displayGroupLinodes := range self.linodes {
		if !self.config.ContainsDisplayGroup(displayGroup) {
			continue
		}
		buf.WriteString(fmt.Sprintf("## %s\n\n", displayGroup))
		for _, linode := range displayGroupLinodes {
			if !linode.IsRunning() {
				continue
			}
			entry := sshEntry{Host: linode.Label, Id: linode.Id}
			keyvals := make(map[string]string)
			keyvals["#"] = fmt.Sprintf("%s | Linode ID %d | %dm Ram", linode.DisplayGroup, linode.Id, linode.Ram)
			keyvals["Hostname"] = linode.PublicIp()
			if self.config.User != "" {
				keyvals["User"] = self.config.User
			}
			if self.config.IdentityFile != "" {
				keyvals["IdentityFile"] = self.config.IdentityFile
			}
			entry.KeyVals = keyvals
			if err := template.Execute(buf, entry); err != nil {
				return nil, err
			}
		}
	}

	buf.Write(append(END_TOKEN, '\n'))
	return buf.Bytes(), nil
}

func fileExists(path string) bool {
	exists := true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exists = false
	}
	return exists
}

func copyFile(in, out string) error {
	i, err := os.Open(in)
	if err != nil {
		return err
	}
	defer i.Close()

	o, err := os.Create(out)
	if err != nil {
		return err
	}
	defer o.Close()

	_, err = io.Copy(o, i)
	if err != nil {
		return err
	}

	return o.Sync()
}
