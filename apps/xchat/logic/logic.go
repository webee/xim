package logic

import (
	"log"
	"strings"

	"gopkg.in/jcelliott/turnpike.v2"
)

// Start starts connect to broker and serving.
func Start(config *Config) {
	if config.Debug {
		turnpike.Debug()
	}
	c, err := turnpike.NewWebsocketClient(turnpike.JSON, config.BrokerURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	d, err := c.JoinRealm("xchat", nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("joined xchat:", d)

	if err := c.Subscribe(URIWAMPSessionOnJoin, onJoin); err != nil {
		log.Fatalf("Error subscribing to %s: %s\n", URIWAMPSessionOnJoin, err)
	}

	if err := c.Subscribe(URIWAMPSessionOnLeave, onLeave); err != nil {
		log.Fatalf("Error subscribing to %s: %s\n", URIWAMPSessionOnLeave, err)
	}

	if err := c.Register(URITestToUpper, toUpper, map[string]interface{}{}); err != nil {
		log.Fatalf("Error register %s: %s\n", URITestToUpper, err)
	}
}

func onJoin(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(map[string]interface{})
	sessionID := uint64(details["session"].(float64))
	role := details["authrole"].(string)
	log.Printf("<%s:%d> joined\n", role, sessionID)
}

func onLeave(args []interface{}, kwargs map[string]interface{}) {
	sessionID := uint64(args[0].(float64))
	log.Printf("<%d> left\n", sessionID)
}

func toUpper(args []interface{}, kargs map[string]interface{}, details map[string]interface{}) (result *turnpike.CallResult) {
	caller := uint64(details["caller"].(float64))
	role := details["caller_authrole"].(string)
	log.Printf("<%s:%d> [rpc]%s: %v, %v, %v\n", role, caller, URITestToUpper, args, kargs, details)
	s, ok := args[0].(string)
	if !ok {
		return &turnpike.CallResult{Err: turnpike.URI("xchat.invalid-argument")}
	}
	res := strings.ToUpper(s)
	return &turnpike.CallResult{Args: []interface{}{res}}
}
