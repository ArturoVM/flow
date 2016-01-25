package main

import (
	"common"
	"fmt"
	"usage"
)

func usageEvent(event usage.Event) {
	switch event.Type {
	case usage.SelfUsageReport:
		uiPrint(fmt.Sprintf("[usage-module]:\n\n\t%f", event.Data.(float64)))
	case usage.UsageReport:
		evData := event.Data.(map[string]string)
		args := map[string]string{
			"peer": evData["peer"],
			"msg":  evData["usage"],
		}
		sendNetworkingCommand("reply", args)
	case usage.Error:
		evData := event.Data.(map[string]string)
		args := map[string]string{
			"peer": evData["peer"],
			"msg":  "[remote:usage-module]:\n\n\t" + evData["error"],
		}
		sendNetworkingCommand("reply", args)
	}
}

func sendUsageCommand(cmd string, args map[string]string) {
	usage.In() <- common.Command{
		Cmd:  cmd,
		Args: args,
	}
}
