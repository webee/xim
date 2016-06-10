package mid

import (
	"fmt"
	"sync"
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

func (p *PushState) setSending() {
	<-p.sending
}

func (p *PushState) doneSending() {
	p.sending <- struct{}{}
}

func (p *PushState) getTask(id uint64, isSender bool) (task chan []*Message, lastID uint64, valid bool) {
	p.Lock()
	defer p.Unlock()

	if !isSender {
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

func newPushState(s *Session, pushID uint64) *PushState {
	p := &PushState{
		sending:        make(chan struct{}, 1),
		pushing:        make(chan struct{}, 1),
		pushingMutex:   make(chan struct{}, 1),
		s:              s,
		pushMsgID:      pushID,
		taskChan:       make(chan chan []*Message, 64),
		notifyTaskChan: make(chan chan []*NotifyMessage, 32),
	}
	p.sending <- struct{}{}
	p.pushing <- struct{}{}
	p.pushingMutex <- struct{}{}
	return p
}

// Session is a user connection.
type Session struct {
	sync.Mutex
	ID         SessionID
	User       string
	pushStates map[uint64]*PushState
	msgTopic   string
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
	}
}

// GetChatPustState returns chat's push state.
func (s *Session) GetChatPustState(chatID uint64) *PushState {
	s.Lock()
	defer s.Unlock()
	p, ok := s.pushStates[chatID]
	if !ok {
		p = newPushState(s, 0)
		s.pushStates[chatID] = p
	}
	return p
}

// GetPushState get last push id if not pushed, and set current id and get a push task.
func (s *Session) GetPushState(chatID uint64, id uint64) (p *PushState, task chan []*Message, lastID uint64, valid bool) {
	s.Lock()
	p, ok := s.pushStates[chatID]
	if !ok {
		p = newPushState(s, id-1)
		s.pushStates[chatID] = p
	}
	s.Unlock()

	task, lastID, valid = p.getTask(id, false)
	return
}

// GetNotifyPushState get notify task.
func (s *Session) GetNotifyPushState(chatID uint64) (p *PushState, task chan []*NotifyMessage) {
	s.Lock()
	p, ok := s.pushStates[chatID]
	if !ok {
		p = newPushState(s, 0)
		s.pushStates[chatID] = p
	}
	s.Unlock()

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
