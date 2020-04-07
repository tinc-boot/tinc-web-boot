package network

import "io/ioutil"

func ConfigFromFile(name string) (*Config, error) {
	var cfg Config
	data, err := ioutil.ReadFile(name)
	if err != nil {
		return nil, err
	}
	return &cfg, cfg.UnmarshalText(data)
}
