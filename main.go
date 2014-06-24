package main

import (
	"flag"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/awilliams/linode"
)

var args struct {
	help       bool
	version    bool
	pp         bool
	update     bool
	stdout     bool
	configPath string
	sshPath    string
}

func init() {
	flag.BoolVar(&args.help, "h", false, "Print help/usage")
	flag.BoolVar(&args.version, "v", false, "Print version")
	flag.BoolVar(&args.pp, "pp", false, "Print Linodes in nicely formatted manner")
	flag.BoolVar(&args.update, "update", false, "Update ssh config file with rendered config")
	flag.BoolVar(&args.stdout, "o", true, "Print rendered config")
	flag.StringVar(&args.configPath, "c", "~/.linode-ssh-config.ini", "Path to configuration file")
	flag.StringVar(&args.sshPath, "F", "~/.ssh/config", "Path to SSH config file")
}

var linodeClient *linode.Client
var config *configuration

func main() {
	flag.Parse()
	var err error
	configPath, err := expandPath(args.configPath)
	if err != nil {
		fatal(err)
	}
	config, err = loadConfig(configPath)
	if err != nil {
		fatal(err)
	}
	linodeClient = linode.NewClient(config.APIKey)

	switch true {
	case args.help:
		printHelp()
	case args.version:
		printVersion()
	case args.pp:
		printPrettyLinodes()
	case args.update:
		updateSSHConfig()
	case args.stdout:
		printSSHConfig()
	default:
		printHelp()
	}
}

func printPrettyLinodes() {
	m := linodes()
	displayGroups := make(sort.StringSlice, len(m))
	i := 0
	for displayGroup := range m {
		displayGroups[i] = displayGroup
		i++
	}
	sort.Sort(displayGroups)

	w := new(tabwriter.Writer)
	w.Init(os.Stdout, 1, 1, 2, ' ', 0)
	var ipType string
	for _, displayGroup := range displayGroups {
		nodes := m[displayGroup]
		fmt.Fprintf(w, "%s (%d)\t\t\t\n", displayGroup, len(nodes))
		fmt.Fprintln(w, "Label\tRunning\tRam\tLinodeId")
		fmt.Fprintln(w, "\t\t\t")

		for _, nodeWithIPs := range nodes {
			fmt.Fprintf(w, "%s\t%v\t%d\t%d\n", nodeWithIPs.node.Label, nodeWithIPs.node.IsRunning(), nodeWithIPs.node.RAM, nodeWithIPs.node.ID)
			for _, ip := range nodeWithIPs.ips {
				if ip.IsPublic() {
					ipType = "public"
				} else {
					ipType = "private"
				}
				fmt.Fprintf(w, "%s\t%s\t\t\n", ip.IP, ipType)
			}
			fmt.Fprintln(w, "\t\t\t")
		}
	}
	w.Flush()
}

func printSSHConfig() {
	sshPath, err := expandPath(args.sshPath)
	if err != nil {
		fatal(err)
	}
	cfg := newSSHConfig(sshPath)
	bytes, err := cfg.render()
	if err != nil {
		fatal(err)
	}
	fmt.Print(string(bytes))
}

func updateSSHConfig() {
	sshPath, err := expandPath(args.sshPath)
	if err != nil {
		fatal(err)
	}
	cfg := newSSHConfig(sshPath)
	if err = cfg.update(); err != nil {
		fatal(err)
	}
	fmt.Printf("Updated %s with %d Linodes\n", cfg.path, cfg.count)
}

func printVersion() {
	fmt.Printf("%s v%s\n", appName, appVersion)
}

const usage = "usage: %s <flag>\n\nflags:\n"

func printHelp() {
	fmt.Printf(usage, appName)
	flag.PrintDefaults()
}

type linodeWithIPs struct {
	node linode.Linode
	ips  []linode.LinodeIP
}

// linodes grouped by their DisplayGroup, filtered by the DisplayGroups specified in the config file
func linodes() map[string][]*linodeWithIPs {
	nodes, err := linodeClient.LinodeList()
	if err != nil {
		fatal(err)
	}

	m := make(map[int]*linodeWithIPs, len(nodes))
	ret := make(map[string][]*linodeWithIPs, len(nodes))
	ids := make([]int, 0, len(nodes))
	for _, n := range nodes {
		if config.filterRunning(n.IsRunning()) && config.filterDisplayGroup(n.DisplayGroup) {
			v := &linodeWithIPs{node: n}
			m[n.ID] = v
			ret[n.DisplayGroup] = append(ret[n.DisplayGroup], v)
			ids = append(ids, n.ID)
		}
	}

	ipMap, err := linodeClient.LinodeIPList(ids)
	if err != nil {
		fatal(err)
	}
	for nodeID, ips := range ipMap {
		m[nodeID].ips = ips
	}

	return ret
}

func expandPath(p string) (string, error) {
	if p[:2] == "~/" {
		usr, err := user.Current()
		if err != nil {
			return p, err
		}
		p = strings.Replace(p, "~", usr.HomeDir, 1)
	}
	return filepath.Abs(p)
}

func fileExists(path string) bool {
	exists := true
	if _, err := os.Stat(path); os.IsNotExist(err) {
		exists = false
	}
	return exists
}

func fatal(v interface{}) {
	fmt.Fprintln(os.Stderr, v)
	os.Exit(1)
}
