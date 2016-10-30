package mid

import (
	"strconv"
	"sync"
	"time"
	pubtypes "xim/xchat/logic/pub/types"
	"xim/xchat/logic/service/types"

	"github.com/patrickmn/go-cache"
)

var (
	pushStatesCache = cache.New(10*time.Minute, 10*time.Minute)
)

// PushState is chat's push state.
type PushState struct {
	sync.RWMutex
	chatID   uint64
	chatType string
	// only accept msg which id > pendingMsgID.
	pendingMsgID uint64
	pushedMsgID  uint64
	pushing      chan struct{}
	replace      chan struct{}
	msgs         []*pubtypes.ChatMessage
}

// Pending pending push message.
func (p *PushState) Pending(msg *pubtypes.ChatMessage) bool {
	p.Lock()
	if p.pendingMsgID >= msg.ID {
		if msg.ID > p.pushedMsgID {
			// 缓存消息
			p.msgs = append(p.msgs, msg)
		}
		p.Unlock()
		return false
	}
	p.msgs = append(p.msgs, msg)

	p.pendingMsgID = msg.ID
	close(p.replace)
	p.replace = make(chan struct{})
	p.Unlock()

	select {
	case p.pushing <- NT:
		return true
	case <-p.replace:
		return false
	}
}

// FetchMsgs fetch to push messages.
func (p *PushState) FetchMsgs(msgID uint64) []*pubtypes.ChatMessage {
	pushedMsgID := p.pushedMsgID

	if pushedMsgID == 0 {
		// 初始状态, 等待一会儿，防止#n+m比#n先到的情况, 会错过m条消息
		time.Sleep(10 * time.Millisecond)
		// 选择所有缓存消息中最小id的确定pushedMsgID
		pushedMsgID = msgID - 1
		if len(p.msgs) > 1 {
			for _, msg := range p.msgs {
				if msg.ID <= pushedMsgID {
					pushedMsgID = msg.ID - 1
				}
			}
		}
	}

	p.RLock()
	count := 0
	msgCount := int(msgID - pushedMsgID)
	msgs := make([]*pubtypes.ChatMessage, msgCount)
	for _, msg := range p.msgs {
		if msg.ID <= msgID {
			idx := msg.ID - pushedMsgID - 1
			if msgs[idx] == nil {
				msgs[idx] = msg
				count++
			}
		}
	}
	defer p.RUnlock()

	if count < msgCount {
		missedIDs := []uint64{}
		for i, msg := range msgs {
			if msg == nil {
				missedIDs = append(missedIDs, uint64(i)+pushedMsgID+1)
			}
		}

		missedMsgs := []pubtypes.ChatMessage{}
		args := &types.FetchChatMessagesByIDsArgs{
			ChatID:   p.chatID,
			ChatType: p.chatType,
			MsgIDs:   missedIDs,
		}
		if err := xchatLogic.Call(types.RPCXChatFetchChatMessagesByIDs, args, &missedMsgs); err != nil {
			l.Warning("fetch chat[%d] messages by ids error: %+v, %s", p.chatID, args, err.Error())
		}
		for i := range missedMsgs {
			msg := &missedMsgs[i]
			msgs[msg.ID-pushedMsgID-1] = msg
		}
	}

	return msgs
}

// Done done pushing.
func (p *PushState) Done(msgID uint64) {
	p.Lock()
	defer p.Unlock()

	p.pushedMsgID = msgID

	newMsgs := p.msgs[:0]
	for _, msg := range p.msgs {
		if msg.ID > msgID {
			newMsgs = append(newMsgs, msg)
		}
	}
	p.msgs = newMsgs
	<-p.pushing
}

func newPushState(chatID uint64, chatType string) *PushState {
	return &PushState{
		chatID:   chatID,
		chatType: chatType,
		pushing:  make(chan struct{}, 1),
		replace:  make(chan struct{}),
		msgs:     []*pubtypes.ChatMessage{},
	}
}

// GetChatPushState returns chat's push state.
func GetChatPushState(chatID uint64, chatType string) *PushState {
	var p *PushState
	key := strconv.FormatUint(chatID, 36)
	value, ok := pushStatesCache.Get(key)
	if !ok {
		p = newPushState(chatID, chatType)
		if err := pushStatesCache.Add(key, p, cache.DefaultExpiration); err != nil {
			return GetChatPushState(chatID, chatType)
		}
	} else {
		p = value.(*PushState)
	}
	return p
}

// RemoveChatPustState remove chat's push state.
func RemoveChatPustState(chatID uint64) {
	key := strconv.FormatUint(chatID, 36)
	pushStatesCache.Delete(key)
}
