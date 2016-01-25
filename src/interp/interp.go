package interp

import (
	"bytes"
	"common"
	"fmt"
	"io"
	"os"

	zygo "github.com/glycerine/zygomys/repl"
)

// EventType define la clase de eventos que se pueden emitir
type EventType int

const (
	// InterpDone señaliza que el interprete terminó de evaluar código
	InterpDone EventType = iota
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

var env *zygo.Glisp

func init() {
	env = zygo.NewGlisp()
	in = make(chan common.Command)
	out = make(chan Event)
}

// Start inicia el módulo
func Start() <-chan Event {
	go loop(in)
	return out
}

// In regresa el channel para mandar comandos al módulo
func In() chan<- common.Command {
	return in
}

func loop(input <-chan common.Command) {
	for c := range input {
		switch c.Cmd {
		case "interp":
			interpOut, err := captureOutput()
			if err != nil {
				sendError(c.Args["peer"],
					fmt.Sprintf("unable to capture script output: %s", err.Error()))
				return
			}
			if err := env.LoadString(c.Args["code"]); err != nil {
				<-interpOut
				env.Clear()
				sendError(c.Args["peer"],
					fmt.Sprintf("unable to load code: %s", err.Error()))
			} else if expr, err := env.Run(); err != nil {
				<-interpOut
				env.Clear()
				sendError(c.Args["peer"],
					fmt.Sprintf("unable to evaluate code: %s", err.Error()))
			} else {
				env.Clear()
				o := <-interpOut
				out <- Event{
					Type: InterpDone,
					Data: map[string]string{
						"peer":   c.Args["peer"],
						"result": o + expr.SexpString(),
					},
				}
			}
		}
	}
}

func sendError(peer, err string) {
	out <- Event{
		Type: Error,
		Data: map[string]string{
			"peer":  peer,
			"error": err,
		},
	}
}

func captureOutput() (<-chan string, error) {
	stdOut := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	os.Stdout = w
	interpOut := make(chan string)
	go func() {
		var buf bytes.Buffer
		io.Copy(&buf, r)
		interpOut <- buf.String()
		w.Close()
		os.Stdout = stdOut
	}()
	return interpOut, nil
}
