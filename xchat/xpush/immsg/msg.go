package immsg

import (
	"encoding/json"
	"errors"
	"fmt"
)

// message type
const (
	MHIMMessageText             = iota // 文本
	MHIMMessageImage                   // 图片
	MHIMMessageLocation                // 地理位置
	MHIMMessageVoice                   // 音频
	MHIMMessageVideo                   // 视频
	MHIMMessageFile                    // 文件
	MHIMMessageLuckRedPacket           // 运气红包
	MHIMMessageBombRedPacket           // 炸弹红包
	MHIMMessageWelfareRedPacket        // 福利红包
	MHIMMessageBegRedPacket            // 讨红包
	MHIMMeesageRedPacketNotice         // 红包消息通知
	// redpacket type
	RedPacketLuck    = iota // 运气红包
	RedPacketWelfare        // 福利红包
	RedPacketBeg            // 讨红包
	RedPacketBomb           // 炸弹红包
)

// ParseMsg parse im message.
func ParseMsg(data []byte) (msg string, err error) {
	var m map[string]interface{}
	json.Unmarshal(data, &m)
	fmt.Println(m)
	msgType, ok := m["messageType"]
	if !ok {
		return msg, errors.New("no message type")
	}
	t, ok := msgType.(int)
	if !ok {
		return msg, errors.New("wrong messag type")
	}
	switch t {
	case MHIMMessageText:
		msg, ok = m["text"].(string)
		if !ok {
			err = errors.New("Wrong test message")
		}
	case MHIMMessageImage:
		msg = "[图片]"
	case MHIMMessageVoice:
		msg = "[语音]"
	case MHIMMessageVideo:
		msg = "[视频]"
	case MHIMMessageFile:
		msg = "[文件]"
	case MHIMMessageLuckRedPacket:
		msg = "[运气红包]"
	case MHIMMessageBombRedPacket:
		msg = "[炸弹红包]"
	case MHIMMessageWelfareRedPacket:
		msg = "[福利红包]"
	case MHIMMessageBegRedPacket:
		msg = "[讨红包]"
	case MHIMMeesageRedPacketNotice:
		msg = "[红包]"
	}

	return msg, err
}
