package mid

import (
	"fmt"
	"sync"
)

// SessionID is uniqe session id.
type SessionID uint64

// Session is a user connection.
type Session struct {
	sync.RWMutex
	ID   SessionID
	User string
	// only push msg which id > PushMsgID.
	pushMsgID uint64
	seq       uint64
	curSeq    uint64
	sending   chan struct{}
}

func (s *Session) String() string {
	return fmt.Sprintf("[%d]->%s", s.ID, s.User)
}

// Sending starts sending.
func (s *Session) Sending(seq uint64) bool {
	return s.doSending(seq, false)
}

func (s *Session) doSending(seq uint64, yield bool) bool {
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
		case s.sending <- struct{}{}:
		default:
		}
	}

	<-s.sending

	return s.doSending(seq, true)
}

/* debug lock.
 */
func (s *Session) Lock() {
	l.Info("%d: LOCK, %d", s.ID, s.curSeq)
	s.RWMutex.Lock()
}

func (s *Session) Unlock() {
	l.Info("%d: UNLOCK, %d", s.ID, s.curSeq)
	s.RWMutex.Unlock()
}

// DoneSending done sending.
func (s *Session) DoneSending(seq uint64) {
	s.Lock()
	if s.curSeq == seq {
		s.curSeq++
	}
	s.Unlock()

	select {
	case s.sending <- struct{}{}:
	default:
	}
}

// GetSetPushID get last push id if not pushed, and set current id and get a push seq id.
func (s *Session) GetSetPushID(id uint64) (uint64, uint64, bool) {
	s.Lock()
	defer s.Unlock()

	if id <= s.pushMsgID {
		return 0, 0, false
	}

	seq := s.seq
	s.seq++

	lastID := s.pushMsgID
	s.pushMsgID = id

	return seq, lastID, true
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
