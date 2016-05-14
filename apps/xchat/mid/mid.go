package mid

import (
	"log"
	"strings"
	"xim/apps/xchat/router"

	"gopkg.in/jcelliott/turnpike.v2"
)

// Mid is the mid of router and xim.
type Mid struct {
	xchat *turnpike.Client
	xim   *XIMClient
}

var mid *Mid

// Setup initialze mid.
func Setup(config *Config, xchatRouter *router.XChatRouter) {
	xchat, err := xchatRouter.GetLocalClient("xchat", nil)
	if err != nil {
		log.Fatalln("create xchat error:", err)
	}

	mid = &Mid{
		xchat: xchat,
		xim:   NewXIMClient(config),
	}
	mid.Start()
}

// Start starts the mid.
func (m *Mid) Start() {
	xchat := m.xchat
	if err := xchat.Subscribe(URIWAMPSessionOnJoin, m.onJoin); err != nil {
		log.Fatalf("Error subscribing to %s: %s\n", URIWAMPSessionOnJoin, err)
	}

	if err := xchat.Subscribe(URIWAMPSessionOnLeave, m.onLeave); err != nil {
		log.Fatalf("Error subscribing to %s: %s\n", URIWAMPSessionOnLeave, err)
	}

	if err := xchat.BasicRegister(URITestToUpper, toUpper); err != nil {
		log.Fatalf("Error register %s: %s\n", URITestToUpper, err)
	}

	if err := xchat.BasicRegister(URITestAdd, add); err != nil {
		log.Fatalf("Error register %s: %s\n", URITestAdd, err)
	}
}

func routerInitSetup(config *Config, xchat *turnpike.Client, ximClient *XIMClient) {
	if config.Debug {
		turnpike.Debug()
	}
	d, err := xchat.JoinRealm("xchat", map[string]interface{}{"role": "xchat"})
	if err != nil {
		log.Fatal(err)
	}
	log.Println("joined xchat:", d)
}

// 处理用户连接注册
func (m *Mid) onJoin(args []interface{}, kwargs map[string]interface{}) {
	details := args[0].(map[string]interface{})
	role := details["role"].(string)
	// register this user.
	if role == "user" {
		user := details["user"].(string)
		sessionID := uint64(details["session"].(turnpike.ID))
		log.Printf("<%s:%s:%d> joined\n", role, user, sessionID)
	}
}

// 处理用户断开注销
func (m *Mid) onLeave(args []interface{}, kwargs map[string]interface{}) {
	sessionID := uint64(args[0].(turnpike.ID))
	log.Printf("<%d> left\n", sessionID)
	// unregister this user.
}

func sendMsg(args []interface{}, kwargs map[string]interface{}, details map[string]interface{}) (result *turnpike.CallResult) {
	/*
		details := kargs["details"].(map[string]interface{})
		user := details["user"].(string)
		sessionID := uint64(details["session"].(float64))
		role := details["role"].(string)

		chatID := int64(args[0].(float64))
		msg := args[1]
	*/
	return &turnpike.CallResult{}
}

func toUpper(args []interface{}, kargs map[string]interface{}) (result *turnpike.CallResult) {
	details := kargs["details"].(map[string]interface{})
	user := details["user"].(string)
	sessionID := uint64(details["session"].(turnpike.ID))
	role := details["role"].(string)
	log.Printf("<%s:%s:%d> [rpc]%s: %v, %v, %v\n", role, user, sessionID, URITestToUpper, args, kargs, details)
	s, ok := args[0].(string)
	if !ok {
		return &turnpike.CallResult{Err: turnpike.URI(URITestToUpper)}
	}
	res := strings.ToUpper(s)
	return &turnpike.CallResult{Args: []interface{}{res}}
}

func add(args []interface{}, kargs map[string]interface{}) (result *turnpike.CallResult) {
	details := kargs["details"].(map[string]interface{})
	user := details["user"].(string)
	sessionID := uint64(details["session"].(turnpike.ID))
	role := details["role"].(string)
	log.Printf("<%s:%s:%d> [rpc]%s: %v, %v, %v\n", role, user, sessionID, URITestAdd, args, kargs, details)
	a, ok1 := args[0].(float64)
	b, ok2 := args[1].(float64)
	if !(ok1 && ok2) {
		return &turnpike.CallResult{Err: turnpike.URI(URITestToUpper)}
	}
	res := a + b
	return &turnpike.CallResult{Args: []interface{}{res}}
}
