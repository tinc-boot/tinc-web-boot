package network

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type builder interface {
	Build() (text []byte, err error)
}

func (cfg *Config) Build() (text []byte, err error) {
	var params = map[string][]interface{}{
		"Name":       {cfg.Name},
		"Port":       {fmt.Sprint(cfg.Port)},
		"Interface":  {cfg.Interface},
		"Mode":       {cfg.Mode},
		"AutoStart":  {cfg.AutoStart},
		"DeviceType": {cfg.DeviceType},
		"Device":     {cfg.Device},
		"IP":         {cfg.IP},
		"Mask":       {cfg.Mask},
		"Broadcast":  {cfg.Broadcast},
	}
	for _, con := range cfg.ConnectTo {
		params["ConnectTo"] = append(params["ConnectTo"], con)
	}
	return makeContent(params, "")
}

func (cfg *Config) Parse(text []byte) error {
	params, _ := parseContent(string(text))
	cfg.ConnectTo = params["ConnectTo"]
	cfg.Interface = params.First("Interface", "")
	cfg.Name = params.First("Name", "")
	cfg.Port = params.FirstUint16("Port")
	cfg.Mode = params.First("Mode", "router")
	cfg.AutoStart = params.FirstBool("AutoStart")
	cfg.DeviceType = params.First("DeviceType", "")
	cfg.Device = params.First("Device", "")
	cfg.IP = params.First("IP", "")
	cfg.Mask = params.FirstInt("Mask")
	cfg.Broadcast = params.First("Broadcast", "")
	return nil
}

func (a *Address) Build() (text []byte, err error) {
	out := a.Host
	if a.Port != 0 {
		out += " " + strconv.FormatUint(uint64(a.Port), 10)
	}
	return []byte(out), nil
}

func (a *Address) Parse(text []byte) error {
	vals := strings.Split(string(text), " ")
	a.Host = vals[0]
	if len(vals) == 1 {
		return nil
	}
	p, err := strconv.ParseUint(vals[1], 10, 16)
	if err != nil {
		return err
	}
	a.Port = uint16(p)
	return nil
}

func (n *Node) Build() (text []byte, err error) {
	params := map[string][]interface{}{
		"Name":    {n.Name},
		"Subnet":  {n.Subnet},
		"Port":    {fmt.Sprint(n.Port)},
		"Version": {fmt.Sprint(n.Version)},
	}
	for i := range n.Address {
		params["Address"] = append(params["Address"], &n.Address[i])
	}

	return makeContent(params, n.PublicKey)
}

func (n *Node) Parse(data []byte) error {
	params, tail := parseContent(string(data))
	n.Name = first(params["Name"], "")
	n.Subnet = first(params["Subnet"], "")
	n.PublicKey = tail
	n.Address = nil
	for _, addr := range params["Address"] {
		var a Address
		err := a.Parse([]byte(addr))
		if err != nil {
			return err
		}
		n.Address = append(n.Address, a)
	}
	n.Port = params.FirstUint16("Port")
	n.Version = params.FirstInt("Version")
	return nil
}

func first(vals []string, def string) string {
	if len(vals) > 0 {
		return vals[0]
	}
	return def
}

type multiMap map[string][]string

func (mm *multiMap) First(name, def string) string {
	if mm == nil {
		return def
	}
	v := (*mm)[name]
	if len(v) == 0 {
		return def
	}
	return v[0]
}

func (mm *multiMap) FirstUint16(name string) uint16 {
	v := mm.First(name, "0")
	x, _ := strconv.ParseUint(v, 10, 16)
	return uint16(x)
}

func (mm *multiMap) FirstInt(name string) int {
	v := mm.First(name, "0")
	x, _ := strconv.Atoi(v)
	return x
}

func (mm *multiMap) FirstBool(name string) bool {
	t := strings.ToLower(mm.First(name, "false"))
	return t == "true" || t == "yes" || t == "on"
}

func parseContent(content string) (params multiMap, tail string) {
	var offset = 0
	params = make(map[string][]string)
	for i, r := range []rune(content) {
		if r != '\n' {
			continue
		}
		line := strings.TrimSpace(content[offset:i])

		if len(line) == 0 || line[0] == '#' {
			continue
		}
		kv := strings.SplitN(line, "=", 2)
		if len(kv) != 2 {
			break
		}
		offset = i + 1
		key := strings.TrimSpace(kv[0])
		value := strings.TrimSpace(kv[1])
		params[key] = append(params[key], value)
	}
	tail = content[offset:]
	return
}

func makeContent(params map[string][]interface{}, tail string) ([]byte, error) {
	out := &strings.Builder{}
	var keys = make([]string, len(params))
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		v := params[k]
		for _, item := range v {
			if item == nil {
				continue
			}
			var val []byte
			if coder, ok := item.(builder); ok {
				x, err := coder.Build()
				if err != nil {
					return nil, err
				}
				val = x
			} else {
				val = []byte(fmt.Sprint(item))
			}
			if len(val) == 0 {
				continue
			}
			out.WriteString(k)
			out.WriteString(" = ")
			out.Write(val)
			out.WriteRune('\n')
		}
	}
	out.WriteString(tail)
	return []byte(out.String()), nil
}
