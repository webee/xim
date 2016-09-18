package mid

import (
	"errors"
	"time"
	"xim/xchat/logic/service/types"

	"gopkg.in/webee/turnpike.v2"
)

// ping
func ping(s *Session, args []interface{}, kwargs map[string]interface{}) (rargs []interface{}, rkwargs map[string]interface{}, rerr APIError) {
	l.Info("[rpc]%s: %v, %+v\n", URIXChatPing, args, kwargs)

	// TODO: 添加多种探测功能，rpc, 获取状态等
	method := args[0].(string)

	switch method {
	case "info":
		info := args[1].(string)
		res, err := fetchInfo(info, args[2:])
		if err != nil {
			rerr = newDefaultAPIError(err.Error())
			return
		}
		return []interface{}{true, res}, nil, nil
	case "cmd":
		cmd := args[1].(string)
		err := doCmd(cmd, args[2:])
		if err != nil {
			rerr = newDefaultAPIError(err.Error())
			return
		}
		return []interface{}{true}, nil, nil
	}

	payloadSize := 0
	if len(args) > 0 {
		payloadSize = int(args[1].(float64))
	}

	if payloadSize < 0 {
		payloadSize = 0
	} else if payloadSize > 1024*1024 {
		payloadSize = 1024 * 1024
	}

	payload := []byte{}
	for i := 0; i < payloadSize; i++ {
		payload = append(payload, 0x31)
	}

	sleep := int64(args[2].(float64))

	if method == "rpc" {
		var content string
		if err := xchatLogic.Call(types.RPCXChatPing, &types.PingArgs{
			Sleep:   sleep,
			Payload: string(payload),
		}, &content); err != nil {
			l.Warning("%s error: %s", types.RPCXChatPing, err)
			rerr = newDefaultAPIError(err.Error())
			return
		}

		rargs = []interface{}{true, s.ID, content}
		return
	} else if method == "net" {
		time.Sleep(time.Duration(sleep) * time.Millisecond)
		rargs = []interface{}{true, s.ID, string(payload)}
		return
	}
	rerr = newDefaultAPIError("invalid method")
	return
}

func fetchInfo(info string, args []interface{}) (interface{}, error) {
	switch info {
	case "session_ids":
		user := args[0].(string)
		sessions := GetUserSessions(user)
		ids := []SessionID{}
		for _, sess := range sessions {
			ids = append(ids, sess.ID)
		}

		return ids, nil
	case "client_infos":
		user := args[0].(string)
		sessions := GetUserSessions(user)
		clientInfos := []string{}
		for _, sess := range sessions {
			clientInfos = append(clientInfos, sess.clientInfo)
		}

		return clientInfos, nil
	}
	return nil, errors.New("invalid info request")
}

func doCmd(cmd string, args []interface{}) error {
	switch cmd {
	case "send_user_sys_req":
		//user := args[0].(string)
		// TODO: system request user messages.
		return nil
	case "debug_on":
		turnpike.Debug()
		return nil
	case "debug_off":
		turnpike.DebugOff()
		return nil
	}
	return errors.New("invalid cmd request")
}
