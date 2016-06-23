package mid

import (
	"fmt"
	"sync"
	pubtypes "xim/xchat/logic/pub/types"
)

// SessionID is uniqe session id.
type SessionID uint64

// PushState is chat's push state.
type PushState struct {
	sync.Mutex
	sending chan struct{}
	// only push msg which id > PushMsgID.
	pushMsgID      uint64
	pushing        chan struct{}
	pushingMutex   chan struct{}
	s              *Session
	taskChan       chan chan []*Message
	notifyTaskChan chan chan []*NotifyMessage
}

func newPushState(s *Session) *PushState {
	p := &PushState{
		sending:        make(chan struct{}, 1),
		pushing:        make(chan struct{}, 1),
		pushingMutex:   make(chan struct{}, 1),
		s:              s,
		pushMsgID:      0,
		taskChan:       make(chan chan []*Message, 64),
		notifyTaskChan: make(chan chan []*NotifyMessage, 32),
	}
	p.sending <- struct{}{}
	p.pushing <- struct{}{}
	p.pushingMutex <- struct{}{}
	return p
}

func (p *PushState) setSending() {
	<-p.sending
}

func (p *PushState) doneSending() {
	p.sending <- struct{}{}
}

func (p *PushState) getTask(id uint64, isSender bool) (task chan []*Message, lastID uint64, valid bool) {
	p.Lock()
	defer p.Unlock()

	if isSender {
		select {
		case <-p.sending:
			p.sending <- struct{}{}
		default:
			// in sending.
			return
		}
	}

	if id <= p.pushMsgID {
		return
	}
	if p.pushMsgID == 0 {
		p.pushMsgID = id - 1
	}
	valid = true

	task = make(chan []*Message, 1)
	p.taskChan <- task

	lastID = p.pushMsgID
	p.pushMsgID = id
	return
}

// Session is a user connection.
type Session struct {
	sync.Mutex
	ID         SessionID
	User       string
	pushStates map[uint64]*PushState
	msgTopic   string
	clientInfo string
	// roomID->chatID
	roomsLock sync.RWMutex
	rooms     map[uint64]uint64
}

func (s *Session) String() string {
	return fmt.Sprintf("[%d]->%s", s.ID, s.User)
}

func newSession(id SessionID, user string) *Session {
	return &Session{
		ID:         id,
		User:       user,
		pushStates: make(map[uint64]*PushState),
		msgTopic:   fmt.Sprintf(URIXChatUserMsg, id),
		rooms:      make(map[uint64]uint64),
	}
}

// SetClientInfo set session's client info.
func (s *Session) SetClientInfo(info string) {
	s.Lock()
	defer s.Unlock()
	s.clientInfo = info
}

// GetClientInfo returns session's client info.
func (s *Session) GetClientInfo() string {
	s.Lock()
	defer s.Unlock()
	return s.clientInfo
}

// EnterRoom enter to room.
func (s *Session) EnterRoom(roomID uint64) (chatID uint64, err error) {
	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()

	chatID, ok := s.rooms[roomID]
	if ok {
		// 已经加入
		return chatID, nil
	}

	chatID, err = rooms.Enter(roomID, s.ID)
	if err != nil {
		return
	}
	s.rooms[roomID] = chatID
	return
}

// ExitRoom exit from room.
func (s *Session) ExitRoom(roomID, chatID uint64) {
	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()

	cid, ok := s.rooms[roomID]
	if ok && cid == chatID {
		rooms.Exit(roomID, chatID, s.ID)
		delete(s.rooms, roomID)
		s.RemoveChatPustState(chatID)
	}
	return
}

// ExitAllRooms exit from all rooms.
func (s *Session) ExitAllRooms() {
	s.roomsLock.Lock()
	defer s.roomsLock.Unlock()

	for roomID, chatID := range s.rooms {
		rooms.Exit(roomID, chatID, s.ID)
		delete(s.rooms, roomID)
		s.RemoveChatPustState(chatID)
	}
	return
}

// GetChatPustState returns chat's push state.
func (s *Session) GetChatPustState(chatID uint64) *PushState {
	s.Lock()
	defer s.Unlock()
	p, ok := s.pushStates[chatID]
	if !ok {
		p = newPushState(s)
		s.pushStates[chatID] = p
	}
	return p
}

// RemoveChatPustState returns chat's push state.
func (s *Session) RemoveChatPustState(chatID uint64) {
	s.Lock()
	defer s.Unlock()

	delete(s.pushStates, chatID)
}

// GetPushState get last push id if not pushed, and set current id and get a push task.
func (s *Session) GetPushState(msg *pubtypes.ChatMessage) (p *PushState, task chan []*Message, lastID uint64, valid bool) {
	chatID := msg.ChatID
	id := msg.ID
	p = s.GetChatPustState(chatID)

	task, lastID, valid = p.getTask(id, msg.User == s.User)
	return
}

// GetNotifyPushState get notify task.
func (s *Session) GetNotifyPushState(chatID uint64) (p *PushState, task chan []*NotifyMessage) {
	p = s.GetChatPustState(chatID)

	p.Lock()
	defer p.Unlock()

	task = make(chan []*NotifyMessage, 1)
	p.notifyTaskChan <- task

	return
}

var (
	t            = struct{}{}
	sessLock     = sync.RWMutex{}
	sessions     = make(map[SessionID]*Session)
	userSessions = make(map[string]map[SessionID]struct{})
)

// AddSession register the session.
func AddSession(s *Session) {
	sessLock.Lock()
	defer sessLock.Unlock()

	sessions[s.ID] = s
	us, ok := userSessions[s.User]
	if !ok {
		us = make(map[SessionID]struct{})
		userSessions[s.User] = us
	}
	us[s.ID] = t
}

// RemoveSession unregister the session.
func RemoveSession(id SessionID) (s *Session) {
	sessLock.Lock()
	defer sessLock.Unlock()

	s, ok := sessions[id]
	if !ok {
		return
	}
	delete(sessions, id)

	us := userSessions[s.User]
	delete(us, id)
	if len(us) == 0 {
		delete(userSessions, s.User)
	}

	return s
}

// GetSession return the session.
func GetSession(id SessionID) (s *Session, ok bool) {
	sessLock.RLock()
	defer sessLock.RUnlock()

	s, ok = sessions[id]
	return
}

// GetUserSessions return the user's sessions.
func GetUserSessions(user string) []*Session {
	sessLock.RLock()
	defer sessLock.RUnlock()

	ss := []*Session{}
	for id := range userSessions[user] {
		if s, ok := sessions[id]; ok {
			ss = append(ss, s)
		}
	}
	return ss
}

// GetOnlineSessionUsers get online session users.
func GetOnlineSessionUsers() map[uint64]string {
	sessLock.RLock()
	defer sessLock.RUnlock()
	users := map[uint64]string{}
	for user, sesses := range userSessions {
		for i := range sesses {
			users[uint64(i)] = user
		}
	}
	return users
}
