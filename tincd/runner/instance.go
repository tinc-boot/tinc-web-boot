package runner

import (
	"bufio"
	"context"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"regexp"
	"tinc-web-boot/utils"
)

var (
	addSubnetPattern = regexp.MustCompile(`ADD_SUBNET\s+from\s+([^\s]+)\s+\(([^\s]+)\s+port\s+(\d+)\)\:\s+\d+\s+[\w\d]+\s+([^\s]+)\s+([^#]+)`)
	delSubnetPattern = regexp.MustCompile(`DEL_SUBNET\s+[^:]+:\s+\d+\s+[\w\d]+\s+([^\s]+)\s+([^#]+)`)
)

//Sending DEL_SUBNET to everyone (BROADCAST): 11 3f17d1ce hubreddecnet_PEN005 6e:6a:5e:26:39:d2#10
func fromLine(line string) *SubnetEvent {
	if match := addSubnetPattern.FindAllStringSubmatch(line, -1); len(match) > 0 {
		groups := match[0]
		if len(groups) != 6 {
			return nil
		}
		var event SubnetEvent
		event.Add = true
		event.Advertising.Node = groups[1]
		event.Advertising.Host = groups[2]
		event.Advertising.Port = groups[3]
		event.Peer.Node = groups[4]
		event.Peer.Subnet = groups[5]
		return &event
	} else if match := delSubnetPattern.FindAllStringSubmatch(line, -1); len(match) > 0 {
		groups := match[0]
		if len(groups) != 3 {
			return nil
		}
		var event SubnetEvent
		event.Add = false
		event.Peer.Node = groups[1]
		event.Peer.Subnet = groups[2]
		return &event
	}
	return nil
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

	var events = make(chan SubnetEvent)

	reader, writer := io.Pipe()
	scanner := bufio.NewScanner(reader)

	cmd := exec.CommandContext(ctx, tincBin, "-D", "-d", "-d", "-d", "-d",
		"--pidfile", filepath.Join(dir, "pid.run"),
		"-c", dir)
	cmd.Dir = dir
	cmd.Stderr = writer
	utils.SetCmdAttrs(cmd)
	cmd.Stdout = writer

	go func() {
		defer writer.Close()
		defer abort()
		err := cmd.Run()
		if err != nil {
			log.Println("run tincd:", err)
		}
	}()

	go func() {
		defer close(events)
		defer abort()
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
