package mid

import (
	"fmt"
	"sync"
	"time"
)

// constants.
const (
	InitialClientInfo = "*"
)

// const variables.
var (
	NT = struct{}{}
)

// SessionID is uniqe session id.
type SessionID uint64

// TaskChan is chat's msg push task channels.
type TaskChan struct {
	tasks          chan chan []*Message
	statelessTasks chan chan []StatelessMsg
	pushing        chan struct{}
	pushingMutex   chan struct{}
}

func newTaskChan() *TaskChan {
	t := &TaskChan{
		tasks:          make(chan chan []*Message, 64),
		statelessTasks: make(chan chan []StatelessMsg, 32),
		pushing:        make(chan struct{}, 1),
		pushingMutex:   make(chan struct{}, 1),
	}
	t.pushing <- NT
	t.pushingMutex <- NT
	return t
}

// NewTask append a new message push task.
func (t *TaskChan) NewTask() (task chan []*Message) {
	task = make(chan []*Message, 1)
	t.tasks <- task
	return
}

// NewStatelessTask append a new stateless message push task.
func (t *TaskChan) NewStatelessTask() (task chan []StatelessMsg) {
	task = make(chan []StatelessMsg, 1)
	t.statelessTasks <- task
	return
}

// Session is a user connection.
type Session struct {
	sync.Mutex
	ID         SessionID
	User       string
	created    time.Time
	taskChan   *TaskChan
	msgTopic   string
	clientInfo string
	// roomID->chatID
	roomsLock sync.RWMutex
	rooms     map[uint64]uint64
}

func (s *Session) String() string {
	return fmt.Sprintf("[%d]->%s", s.ID, s.User)
}

func newSession(id SessionID, User string) *Session {
	return &Session{
		ID:         id,
		User:       User,
		created:    time.Now(),
		taskChan:   newTaskChan(),
		msgTopic:   fmt.Sprintf(URIXChatUserMsg, id),
		clientInfo: InitialClientInfo,
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
	}
	return
}

var (
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
	us[s.ID] = NT
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

// GetSlowSessions get sessions who pub client info late.
func GetSlowSessions() []SessionID {
	sessLock.RLock()
	defer sessLock.RUnlock()
	t := time.Now()
	sesses := []SessionID{}
	for id, sess := range sessions {
		if t.Sub(sess.created).Seconds() > 30 && sess.clientInfo == InitialClientInfo {
			sesses = append(sesses, id)
		}
	}
	return sesses
}
