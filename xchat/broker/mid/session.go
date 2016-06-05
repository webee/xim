package mid

import (
	"fmt"
	"sync"
)

// SessionID is uniqe session id.
type SessionID uint64

type pushState struct {
	sync.RWMutex
	// only push msg which id > PushMsgID.
	pushMsgID uint64
	seq       uint64
	curSeq    uint64
	pushing   chan struct{}
	s         *Session
}

// Pushing starts push.
func (s *pushState) Pushing(seq uint64) bool {
	return s.doPushing(seq, false)
}

func (s *pushState) doPushing(seq uint64, yield bool) bool {
	// TODO: use RWLock.
	s.RLock()
	curSeq := s.curSeq
	s.RUnlock()

	if curSeq == seq {
		return true
	}
	if curSeq > seq {
		// impossible!!!
		l.Emergency("current seq: %d, seq: %d", curSeq, seq)
		return false
	}

	if yield {
		select {
		case s.pushing <- struct{}{}:
		default:
		}
	}

	<-s.pushing

	return s.doPushing(seq, true)
}

/* debug lock.
 */
func (s *pushState) Lock() {
	l.Info("%d: LOCK, %d", s.s.ID, s.curSeq)
	s.RWMutex.Lock()
}

func (s *pushState) Unlock() {
	l.Info("%d: UNLOCK, %d", s.s.ID, s.curSeq)
	s.RWMutex.Unlock()
}

// DonePushing done push.
func (s *pushState) DonePushing(seq uint64) {
	s.Lock()
	if s.curSeq == seq {
		s.curSeq++
	}
	s.Unlock()

	select {
	case s.pushing <- struct{}{}:
	default:
	}
}

// Session is a user connection.
type Session struct {
	sync.Mutex
	ID         SessionID
	User       string
	pushStates map[uint64]*pushState
}

func (s *Session) String() string {
	return fmt.Sprintf("[%d]->%s", s.ID, s.User)
}

func newSession(id SessionID, user string) *Session {
	return &Session{
		ID:         id,
		User:       user,
		pushStates: make(map[uint64]*pushState),
	}
}

// GetSetPushID get last push id if not pushed, and set current id and get a push seq id.
func (s *Session) GetSetPushID(chatID uint64, id uint64) (*pushState, uint64, uint64, bool) {
	s.Lock()
	defer s.Unlock()
	p, ok := s.pushStates[chatID]
	if !ok {
		p = &pushState{
			pushing: make(chan struct{}),
			s:       s,
		}
		s.pushStates[chatID] = p
	}

	if id <= p.pushMsgID {
		return nil, 0, 0, false
	}

	seq := p.seq
	p.seq++

	lastID := p.pushMsgID
	p.pushMsgID = id

	return p, seq, lastID, true
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
