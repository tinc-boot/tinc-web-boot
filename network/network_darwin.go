package network

import (
	"context"
	"errors"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

func (network *Network) postConfigure(ctx context.Context, config *Config, tincBin string) error {
	return nil
}

func (network *Network) beforeConfigure(config *Config) error {
	tap, err := network.findAvailableTap()
	if err != nil {
		log.Println("found no available TAP devices:", err, "assuming OS will create it dynamically")
	}
	config.Interface = ""
	config.Device = tap
	return nil
}

func (network *Network) findAvailableTap() (string, error) {
	darwinMaxTapDevices, err := getMaximumAvailableTapDevices()
	if err != nil {
		return "", err
	}

	storage := &Storage{Root: filepath.Dir(network.Root)}
	nets, err := storage.List()
	if err != nil {
		return "", err
	}
	var used = make([]bool, darwinMaxTapDevices)
	for _, net := range nets {
		cfg, err := net.Read()
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return "", err
		}
		numStr := regexp.MustCompile(`[0-9]+`).FindString(cfg.Device)
		v, _ := strconv.Atoi(numStr)
		if v < len(used) {
			used[v] = true
		}
	}
	for i := len(used) - 1; i >= 0; i-- {
		if !used[i] {
			return "/dev/tap" + strconv.Itoa(i), nil
		}
	}
	return "", errors.New("all tap devices are used")
}

func getMaximumAvailableTapDevices() (int, error) {
	devices, err := ioutil.ReadDir("/dev")
	if err != nil {
		return 0, err
	}
	var darwinMaxTapDevices int
	for _, device := range devices {
		if strings.HasPrefix(device.Name(), "tap") {
			darwinMaxTapDevices++
		}
	}
	return darwinMaxTapDevices, nil
}
