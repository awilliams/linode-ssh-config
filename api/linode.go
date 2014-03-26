package api

import (
	"sort"
	"strconv"
)

func FetchLinodesWithIps(apiKey string) (Linodes, error) {
	api, err := NewApiRequest(apiKey)
	if err != nil {
		return nil, err
	}

	linodes, err := FetchLinodeList(*api)
	if err != nil {
		return nil, err
	}

	linodeIps, err := FetchLinodeIpList(*api, linodes.Ids())
	if err != nil {
		return nil, err
	}

	// associate ips with linodes
	for _, linodeDisplayGroup := range linodes {
		for _, linode := range linodeDisplayGroup {
			if ips, ok := linodeIps[linode.Id]; ok {
				sortLinodeIps(ips)
				linode.Ips = ips
			}
		}
		sortLinodes(linodeDisplayGroup)
	}

	return linodes, nil
}

// map of Linodes by their display group
type Linodes map[string][]*Linode

func (self Linodes) Ids() []int {
	ids := []int{}
	for _, linodeDisplayGroup := range self {
		for _, linode := range linodeDisplayGroup {
			ids = append(ids, linode.Id)
		}
	}
	return ids
}

type Linode struct {
	Id           int    `json:"LINODEID"`
	Status       int    `json:"STATUS"`
	Label        string `json:"LABEL"`
	DisplayGroup string `json:"LPM_DISPLAYGROUP"`
	Ram          int    `json:"TOTALRAM"`
	Ips          []*LinodeIp
}

func (self *Linode) PublicIp() string {
	var ip string
	for _, linodeIp := range self.Ips {
		if linodeIp.Public == 1 {
			ip = linodeIp.Ip
			break
		}
	}
	return ip
}

func (self *Linode) PrivateIp() string {
	var ip string
	for _, linodeIp := range self.Ips {
		if linodeIp.Public == 0 {
			ip = linodeIp.Ip
			break
		}
	}
	return ip
}

func (self *Linode) IsRunning() bool {
	return self.Status == 1
}

func FetchLinodeList(api apiRequest) (Linodes, error) {
	api.AddAction("linode.list")

	var jsonData struct {
		Linodes []Linode `json:"DATA,omitempty"`
	}
	err := api.GetJson(&jsonData)
	if err != nil {
		return nil, err
	}

	linodes := make(Linodes)
	for _, linode := range jsonData.Linodes {
		l := linode
		linodes[linode.DisplayGroup] = append(linodes[linode.DisplayGroup], &l)
	}

	return linodes, nil
}

// map of LinodeIps by their Linode.Id
type LinodeIps map[int][]*LinodeIp

type LinodeIp struct {
	LinodeId int    `json:"LINODEID"`
	Ip       string `json:"IPADDRESS"`
	Public   int    `json:"ISPUBLIC"`
}

func FetchLinodeIpList(api apiRequest, linodeIds []int) (LinodeIps, error) {
	apiMethod := "linode.ip.list"
	// one batch request for all linode Ids
	for _, linodeId := range linodeIds {
		action := api.AddAction(apiMethod)
		action.Set("LinodeID", strconv.Itoa(linodeId))
	}

	var jsonData []struct {
		LinodeIps []LinodeIp `json:"DATA"`
	}
	err := api.GetJson(&jsonData)
	if err != nil {
		return nil, err
	}

	linodeIps := make(LinodeIps)
	for _, ipList := range jsonData {
		for _, linodeIp := range ipList.LinodeIps {
			i := linodeIp
			linodeIps[linodeIp.LinodeId] = append(linodeIps[linodeIp.LinodeId], &i)
		}
	}

	return linodeIps, nil
}

// Sort functions

type sortedLinodeIps []*LinodeIp

func (self sortedLinodeIps) Len() int {
	return len(self)
}
func (self sortedLinodeIps) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}

// Public first
func (self sortedLinodeIps) Less(i, j int) bool {
	return self[i].Public > self[j].Public
}
func sortLinodeIps(ips []*LinodeIp) {
	sort.Sort(sortedLinodeIps(ips))
}

type sortedLinodes []*Linode

func (self sortedLinodes) Len() int {
	return len(self)
}
func (self sortedLinodes) Swap(i, j int) {
	self[i], self[j] = self[j], self[i]
}
func (self sortedLinodes) Less(i, j int) bool {
	return self[i].Label < self[j].Label
}
func sortLinodes(linodes []*Linode) {
	sort.Sort(sortedLinodes(linodes))
}
