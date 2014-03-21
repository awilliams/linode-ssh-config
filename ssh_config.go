package main

import (
  "bufio"
  "bytes"
  "fmt"
  "github.com/awilliams/linode-ssh-config/api"
  "os"
  "os/user"
  "path"
  "text/template"
)

type SSHConfig struct {
  linodes api.Linodes
  file    *os.File
  config  Configuration
}

const SSH_CONFIG_PATH = ".ssh/config"

func NewSSHConfig(config Configuration, linodes api.Linodes) (*SSHConfig, error) {
  usr, err := user.Current()
  if err != nil {
    return nil, err
  }
  configPath := path.Join(usr.HomeDir, SSH_CONFIG_PATH)
  
  exists := true
  if _, err := os.Stat(configPath); os.IsNotExist(err) {
    exists = false
  }

  var file *os.File
  if exists {
    file, err = os.Open(configPath)
    if err != nil {
      return nil, err
    }
  }

  return &SSHConfig{linodes: linodes, file: file, config: config}, nil
}

// combine the user's config and the generated config
func (self *SSHConfig) render() ([]byte, error){
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
  if self.file == nil {
    return []byte{}, nil
  }
  insideConfigBlock := false
  
  strippedBuf := new(bytes.Buffer)
  scanner := bufio.NewScanner(self.file)
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