package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"text/template"
)

func newSSHConfig(p string) *sshConfig {
	return &sshConfig{path: p}
}

type sshConfig struct {
	path  string
	count int // number of linodes. valid until next call to generatedConfig
}

// return the ssh config as a byte slice with the users' previous config and generated config
func (c *sshConfig) render() ([]byte, error) {
	users, err := c.usersConfig()
	if err != nil {
		return nil, err
	}

	generated, err := c.generatedConfig()
	if err != nil {
		return nil, err
	}

	return append(users, generated...), nil
}

// write to the rendered config to disk, making a backup if possible
func (c *sshConfig) update() error {
	if fileExists(c.path) {
		err := copyFile(c.path, c.path+".linode-ssh-config.bak")
		if err != nil {
			return err
		}
	}
	contents, err := c.render()
	if err != nil {
		return err
	}
	return ioutil.WriteFile(c.path, contents, 0644)
}

var startToken = []byte("##### START GENERATED LINODE-SSH-CONFIG #####")
var endToken = []byte("##### END GENERATED LINODE-SSH-CONFIG #####")

// read the user's ssh config file, and strip out any previously generated config
func (c *sshConfig) usersConfig() ([]byte, error) {
	if !fileExists(c.path) {
		return []byte{}, nil
	}

	f, err := os.Open(c.path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	insideConfigBlock := false

	strippedBuf := new(bytes.Buffer)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		if !insideConfigBlock && bytes.Equal(scanner.Bytes(), startToken) {
			insideConfigBlock = true
		}

		if !insideConfigBlock {
			_, err := strippedBuf.Write(append(scanner.Bytes(), '\n'))
			if err != nil {
				return nil, err
			}
		}

		if insideConfigBlock && bytes.Equal(scanner.Bytes(), endToken) {
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
	KeyVals map[string]string
}

const entryTemplate = `Host {{ .Host }}{{ range $k, $v := .KeyVals }}
        {{ $k }} {{ $v }}{{ end }}

`

// create the generated config section
func (c *sshConfig) generatedConfig() ([]byte, error) {
	buf := new(bytes.Buffer)
	buf.Write(startToken)
	buf.WriteRune('\n')
	buf.WriteRune('\n')

	c.count = 0
	tpl := template.Must(template.New("entry").Parse(entryTemplate))
	for displayGroup, nodes := range linodes() {
		if displayGroup != "" {
			buf.WriteString(fmt.Sprintf("## %s\n\n", displayGroup))
		}
		for _, n := range nodes {
			var hostname string
			for _, ip := range n.ips {
				if ip.IsPublic() {
					hostname = ip.IP
					break
				}
			}
			if hostname == "" {
				// no public ip?
				continue
			}
			entry := sshEntry{Host: n.node.Label}
			keyvals := make(map[string]string)
			keyvals["#"] = fmt.Sprintf("Linode ID %d", n.node.ID)
			keyvals["Hostname"] = hostname
			if config.User != "" {
				keyvals["User"] = config.User
			}
			if config.IdentityFile != "" {
				keyvals["IdentityFile"] = config.IdentityFile
			}
			entry.KeyVals = keyvals
			if err := tpl.Execute(buf, entry); err != nil {
				return nil, err
			}
			c.count++
		}
	}
	buf.Write(endToken)
	buf.WriteRune('\n')

	return buf.Bytes(), nil
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
