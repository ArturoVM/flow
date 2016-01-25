package main

import (
	"common"
	"interp"
)

func interpEvent(event interp.Event) {
	switch event.Type {
	case interp.InterpDone:
		evData := event.Data.(map[string]string)
		args := map[string]string{
			"peer": evData["peer"],
			"msg":  "\n" + evData["result"],
		}
		sendNetworkingCommand("send", args)
	case interp.Error:
		evData := event.Data.(map[string]string)
		args := map[string]string{
			"peer": evData["peer"],
			"msg":  "\n" + evData["error"],
		}
		sendNetworkingCommand("send", args)
	}
}

func sendInterpCommand(cmd string, args map[string]string) {
	interp.In() <- common.Command{
		Cmd:  cmd,
		Args: args,
	}
}
