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
	// only push msg which id > PushMsgID.
	pushMsgID      uint64
	pushing        chan struct{}
	s              *Session
	taskChan       chan chan []*Message
	notifyTaskChan chan chan []*NotifyMessage
}

func newPushState(s *Session, pushID uint64) *PushState {
	return &PushState{
		pushing:        make(chan struct{}, 2),
		s:              s,
		pushMsgID:      pushID,
		taskChan:       make(chan chan []*Message, 64),
		notifyTaskChan: make(chan chan []*NotifyMessage, 32),
	}
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

// GetPushState get last push id if not pushed, and set current id and get a push task.
func (s *Session) GetPushState(chatID uint64, id uint64) (p *PushState, task chan []*Message, lastID uint64, valid bool) {
	s.Lock()
	p, ok := s.pushStates[chatID]
	if !ok {
		p = newPushState(s, id-1)
		p.pushing <- struct{}{}
		s.pushStates[chatID] = p
	}
	s.Unlock()

	p.Lock()
	defer p.Unlock()
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

// GetNotifyPushState get notify task.
func (s *Session) GetNotifyPushState(chatID uint64) (p *PushState, task chan []*NotifyMessage) {
	s.Lock()
	p, ok := s.pushStates[chatID]
	if !ok {
		p = newPushState(s, 0)
		p.pushing <- struct{}{}
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
