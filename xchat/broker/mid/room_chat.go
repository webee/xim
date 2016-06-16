package mid

import (
	"fmt"
	"sort"
	"sync"
	"xim/xchat/logic/service/types"
)

// Chat is a chat.
type Chat struct {
	chatID  uint64
	members map[SessionID]struct{}
}

// RoomChats is room chats.
type RoomChats struct {
	sync.RWMutex
	roomID uint64
	chats  map[uint64]*Chat
}

// Rooms is the room chats.
type Rooms struct {
	sync.RWMutex
	rooms map[uint64]*RoomChats
}

// ByCountDesc implements sort.Interface for []*Chat by members count desc.
type ByCountDesc []*Chat

func (a ByCountDesc) Len() int      { return len(a) }
func (a ByCountDesc) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a ByCountDesc) Less(i, j int) bool {
	if len(a[i].members) > len(a[j].members) {
		return true
	}
	if len(a[i].members) == len(a[j].members) {
		return a[i].chatID < a[j].chatID
	}
	return false
}

var (
	areaLimit = int(1000)
	rooms     = &Rooms{
		rooms: make(map[uint64]*RoomChats, 4),
	}
)

// NewRoomChats creates a room chats.
func NewRoomChats(roomID uint64) *RoomChats {
	return &RoomChats{
		roomID: roomID,
		chats:  make(map[uint64]*Chat, 8),
	}
}

// Add add session to room chat.
func (rc *RoomChats) Add(id SessionID) (chatID uint64, err error) {
	rc.Lock()
	defer rc.Unlock()

	chats := []*Chat{}
	for _, chat := range rc.chats {
		chats = append(chats, chat)
	}
	sort.Sort(ByCountDesc(chats))

	for _, chat := range chats {
		if len(chat.members) < areaLimit {
			chat.members[id] = struct{}{}
			return chat.chatID, nil
		}
	}
	// 都满了, 从服务器获取所有会话
	ids := []uint64{}
	for _, chat := range chats {
		ids = append(ids, chat.chatID)
	}

	chatIDs := []uint64{}
	if err := xchatLogic.Call(types.RPCXChatFetchNewRoomChatIDs, &types.FetchNewRoomChatIDs{
		RoomID:  rc.roomID,
		ChatIDs: ids,
	}, &chatIDs); err != nil {
		return 0, err
	}

	var newChat *Chat
	for _, chatID := range chatIDs {
		_, ok := rc.chats[chatID]
		if !ok {
			chat := &Chat{
				chatID:  chatID,
				members: make(map[SessionID]struct{}, 32),
			}
			rc.chats[chatID] = chat
			if newChat == nil {
				newChat = chat
			}
		}
	}

	if newChat != nil {
		newChat.members[id] = struct{}{}
		return newChat.chatID, nil
	}

	return 0, fmt.Errorf("add to room failed")
}

// Remove remove session from room chat.
func (rc *RoomChats) Remove(chatID uint64, id SessionID) {
	rc.Lock()
	rc.Unlock()
	chat, ok := rc.chats[chatID]
	if !ok {
		return
	}

	delete(chat.members, id)
}

// Members returns room chat's members.
func (rc *RoomChats) Members(chatID uint64) (ids []SessionID) {
	rc.RLock()
	rc.RUnlock()

	chat, ok := rc.chats[chatID]
	if !ok {
		return
	}

	for id := range chat.members {
		ids = append(ids, id)
	}

	return
}

// HasChat checks if room has chat.
func (rc *RoomChats) HasChat(chatID uint64) bool {
	rc.RLock()
	rc.RUnlock()

	_, ok := rc.chats[chatID]
	return ok
}

// Enter adds session to room's chat members.
func (rm *Rooms) Enter(roomID uint64, id SessionID) (chatID uint64, err error) {
	rm.Lock()
	defer rm.Unlock()

	room, ok := rm.rooms[roomID]
	if !ok {
		room = NewRoomChats(roomID)
		rm.rooms[roomID] = room
	}
	return room.Add(id)
}

// Exit removes session from room chat's members.
func (rm *Rooms) Exit(roomID, chatID uint64, id SessionID) {
	rm.RLock()
	defer rm.RUnlock()

	room, ok := rm.rooms[roomID]
	if !ok {
		return
	}
	room.Remove(chatID, id)
}

// Members returns room chat's members.
func (rm *Rooms) Members(roomID, chatID uint64) (ids []SessionID) {
	rm.RLock()
	defer rm.RUnlock()
	room, ok := rm.rooms[roomID]
	if !ok {
		return
	}

	return room.Members(chatID)
}

// ChatMembers returns room chat's members.
func (rm *Rooms) ChatMembers(chatID uint64) (ids []SessionID) {
	rm.RLock()
	defer rm.RUnlock()
	for _, rc := range rm.rooms {
		if rc.HasChat(chatID) {
			return rc.Members(chatID)
		}
	}
	return
}
