package networking

import (
	"common"
	"fmt"
	"log"

	"os"

	"github.com/hashicorp/mdns"
)

// EventType define la clase de eventos que se pueden emitir
type EventType int

const (
	// Connection representa cliente conectado
	Connection EventType = iota
	// Disconnection representa cliente desconectado
	Disconnection
	// PeerLookup significa que el nodo está buscando peers
	PeerLookup
	// PeersFound significa que se encontraron peers
	PeersFound
	// Interp representa una solicitud de interpretación
	Interp
	// PeerSelected significa que un peer fue escogido
	PeerSelected
	// UsageRequested significa que otro nodo ha pedido el % de CPU de este
	UsageRequested
	// EvalRequested significa que otro nodo le ha pedido evaluar algo a este
	EvalRequested
	// GotEvalReply notifica que el resultado de la evaluación está listo
	GotEvalReply
	// Error reporta un error
	Error
)

// Event se utiliza para representar un evento emitido
type Event struct {
	Type EventType
	Data interface{}
}

var in chan common.Command
var out chan Event

func init() {
	in = make(chan common.Command)
	out = make(chan Event)
}

// Start inicia el módulo
func Start() <-chan Event {
	go loop(in)
	go setServer()
	return out
}

// In regresa el channel para mandar comandos al módulo
func In() chan<- common.Command {
	return in
}

func loop(input <-chan common.Command) {
	host, err := os.Hostname()
	if err != nil {
		log.Fatal("unable to announce on mDNS: unable to obtain hostname")
	}
	info := []string{"Flow distributed computing peer"}
	service, err := mdns.NewMDNSService(host, "_flow._tcp", "", "", 3569, nil, info)
	if err != nil {
		log.Fatal("unable to create mDNS service")
	}
	server, err := mdns.NewServer(&mdns.Config{Zone: service})
	if err != nil {
		log.Fatal("unable to start mDNS service")
	}
	defer server.Shutdown()

	for c := range input {
		switch c.Cmd {
		case "lookup-peers":
			p := LookupPeers()
			peerTable := <-p
			if len(peerTable) < 1 {
				out <- Event{
					Type: Error,
					Data: "no peers found",
				}
			} else {
				out <- Event{
					Type: PeersFound,
					Data: peerTable,
				}
			}
		case "select-peer":
			p, err := selectPeer()
			if err != nil {
				out <- Event{
					Type: Error,
					Data: fmt.Sprintf("error selecting peer: %s", err),
				}
			} else {
				out <- Event{
					Type: PeerSelected,
					Data: p,
				}
			}
		case "reply":
			ipTable[c.Args["peer"]] <- c.Args["msg"]
			close(ipTable[c.Args["peer"]])
			delete(ipTable, c.Args["peer"])
		case "eval":
			evalInst := "eval::|:|flow-code|:|" + c.Args["code"] + "|:|flow-code|:|"
			SendEval(evalInst)
		default:
		}
	}
}
