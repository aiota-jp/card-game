package main

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"testing"
)

// Only built into the test executable; communicates over stdin, without a server.
func TestUIBridge(t *testing.T) {
	if os.Getenv("CARDLOUNGE_UI_TEST") != "1" {
		t.Skip("optional UI bridge")
	}
	controllers := map[string]any{"PokerApp": NewPokerApp(), "BlackjackApp": NewBlackjackApp(), "BaccaratApp": NewBaccaratApp()}
	dec, enc := json.NewDecoder(os.Stdin), json.NewEncoder(os.Stdout)
	for {
		var req struct {
			Controller, Method string
			Args               []json.RawMessage
		}
		if dec.Decode(&req) != nil {
			return
		}
		target, ok := controllers[req.Controller]
		if !ok {
			enc.Encode(map[string]string{"error": "unknown controller"})
			continue
		}
		m := reflect.ValueOf(target).MethodByName(req.Method)
		if !m.IsValid() || m.Type().NumIn() != len(req.Args) {
			enc.Encode(map[string]string{"error": "invalid method"})
			continue
		}
		args := []reflect.Value{}
		var err error
		for i, raw := range req.Args {
			v := reflect.New(m.Type().In(i))
			if err = json.Unmarshal(raw, v.Interface()); err != nil {
				break
			}
			args = append(args, v.Elem())
		}
		if err != nil {
			enc.Encode(map[string]string{"error": err.Error()})
			continue
		}
		out := m.Call(args)
		if len(out) > 1 && !out[1].IsNil() {
			enc.Encode(map[string]string{"error": fmt.Sprint(out[1].Interface())})
		} else {
			enc.Encode(map[string]any{"result": out[0].Interface()})
		}
	}
}
