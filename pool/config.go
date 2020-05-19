package pool

import (
	"encoding/json"
	"errors"
	"os"
)

type Config struct {
	AutoStart StringSet `json:"auto_start,omitempty"`

	_filename string
}

func (cfg *Config) Filename() string { return cfg._filename }

func (cfg *Config) Save() error {
	return cfg.SaveAs(cfg._filename)
}

func (cfg *Config) SaveAs(filename string) error {
	if filename == "" {
		return errors.New("file name not specified")
	}
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	err = enc.Encode(cfg)
	if err != nil {
		return err
	}
	cfg._filename = filename
	return nil
}

func (cfg *Config) Load() error {
	return cfg.LoadFrom(cfg._filename)
}

func (cfg *Config) LoadFrom(filename string) error {
	if filename == "" {
		return errors.New("file name not specified")
	}
	f, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	err = json.NewDecoder(f).Decode(cfg)
	if err != nil {
		return err
	}

	cfg._filename = filename
	return nil
}

type StringSet map[string]bool

func (s *StringSet) MarshalJSON() ([]byte, error) {
	var keys = make([]string, 0, len(*s))
	for k := range *s {
		keys = append(keys, k)
	}
	return json.Marshal(keys)
}

func (s *StringSet) UnmarshalJSON(bytes []byte) error {
	var keys []string
	err := json.Unmarshal(bytes, &keys)
	if err != nil {
		return err
	}
	for _, k := range keys {
		(*s)[k] = true
	}
	return nil
}

func (s *StringSet) Has(key string) bool { return (*s)[key] }

func (s *StringSet) Set(key string) {
	(*s)[key] = true
}

func (s *StringSet) Del(key string) {
	delete(*s, key)
}
