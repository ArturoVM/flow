package main

import (
	"common"
	"log"
	"networking"
)

func formatPeersFound(peers []string) string {
	peerMsg := ""
	for i := range peers {
		peerMsg += "\t" + peers[i] + "\n"
	}
	return peerMsg
}

func netEvent(event networking.Event) {
	switch event.Type {
	case networking.PeersFound:
		peers, ok := event.Data.([]string)
		if !ok {
			log.Fatalf("datos incorrectos para evento 'peers-found'")
		}
		p := formatPeersFound(peers)
		uiPrint("[networking-module]:\n\n" + p)
	case networking.PeerSelected:
		peer, ok := event.Data.(string)
		peerMsg := ""
		if !ok {
			log.Fatalf("datos incorrectos para evento 'peer-selected'")
		} else {
			peerMsg = "selected peer " + peer
		}
		uiPrint(peerMsg)
	case networking.UsageRequested:
		args := map[string]string{"peer": event.Data.(string)}
		sendUsageCommand("get-usage", args)
	case networking.EvalRequested:
		args := event.Data.(map[string]string)
		sendInterpCommand("interp", args)
	case networking.GotEvalReply:
		uiPrint(event.Data.(string))
	case networking.Error:
		uiPrint("[networking-module]:\n\n\t" + event.Data.(string))
	}
}

func sendNetworkingCommand(cmd string, args map[string]string) {
	networking.In() <- common.Command{
		Cmd:  cmd,
		Args: args,
	}
}
