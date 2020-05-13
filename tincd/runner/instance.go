package runner

import (
	"bufio"
	"context"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"tinc-web-boot/utils"
)

var subnetEventPattern = regexp.MustCompile(`(\w+)_SUBNET\s+from\s+([^\s]+)\s+\(([^\s]+)\s+port\s+(\d+)\)\:\s+\d+\s+[\w\d]+\s+([^\s]+)\s+([^#]+)`)

func fromLine(line string) *SubnetEvent {
	match := subnetEventPattern.FindAllStringSubmatch(line, -1)
	if len(match) == 0 {
		return nil
	}
	groups := match[0]
	if len(groups) != 7 {
		return nil
	}
	var event SubnetEvent
	event.Add = groups[1] == "ADD"
	event.Advertising.Node = groups[2]
	event.Advertising.Host = groups[3]
	event.Advertising.Port = groups[4]
	event.Peer.Node = groups[5]
	event.Peer.Subnet = groups[6]
	return &event
}

type SubnetEvent struct {
	Add         bool
	Advertising struct {
		Node string
		Host string
		Port string
	}
	Peer struct {
		Node   string
		Subnet string
	}
}

func RunTinc(global context.Context, tincBin string, dir string) <-chan SubnetEvent {

	ctx, abort := context.WithCancel(global)
	defer abort()
	var events = make(chan SubnetEvent)

	reader, writer := io.Pipe()
	scanner := bufio.NewScanner(reader)

	cmd := exec.CommandContext(ctx, tincBin, "-D", "-d", "-d", "-d", "-d",
		"--pidfile", filepath.Join(dir, "pid.run"),
		"-c", dir)
	cmd.Dir = dir
	cmd.Stderr = os.Stderr
	utils.SetCmdAttrs(cmd)

	cmd.Stdout = writer

	go func() {
		defer writer.Close()
		_ = cmd.Run()
	}()

	go func() {
		defer close(events)
		for scanner.Scan() {
			if event := fromLine(scanner.Text()); event != nil {
				select {
				case events <- *event:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return events
}
