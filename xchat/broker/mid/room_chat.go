package mid

import (
	"fmt"
	"sort"
	"sync"
	"xim/xchat/logic/db"
	"xim/xchat/logic/service/types"
)

// TODO: use area to refactor.

// Chat is a chat.
type Chat struct {
	area    uint32
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
	areaLimit uint32
	rooms     map[uint64]*RoomChats
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
	rooms = &Rooms{
		areaLimit: 1000,
		rooms:     make(map[uint64]*RoomChats, 4),
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
func (rc *RoomChats) Add(id SessionID) (area uint32, chatID uint64, err error) {
	rc.Lock()
	defer rc.Unlock()

	chats := []*Chat{}
	for _, chat := range rc.chats {
		chats = append(chats, chat)
	}
	sort.Sort(ByCountDesc(chats))

	areaLimit := int(rooms.areaLimit)
	for _, chat := range chats {
		if _, ok := chat.members[id]; ok {
			// 已经加入
			return chat.area, chat.chatID, nil
		}

		if len(chat.members) < areaLimit {
			chat.members[id] = struct{}{}
			return chat.area, chat.chatID, nil
		}
	}
	// 都满了, 从服务器获取所有会话
	ids := []uint64{}
	for _, chat := range chats {
		ids = append(ids, chat.chatID)
	}

	roomChats := []db.RoomChat{}
	if err := xchatLogic.Call(types.RPCXChatFetchNewRoomChats, &types.FetchNewRoomChatsArgs{
		RoomID:  rc.roomID,
		ChatIDs: ids,
	}, &roomChats); err != nil {
		return 0, 0, err
	}

	var newChat *Chat
	for _, roomChat := range roomChats {
		_, ok := rc.chats[roomChat.ChatID]
		if !ok {
			chat := &Chat{
				area:    roomChat.Area,
				chatID:  roomChat.ChatID,
				members: make(map[SessionID]struct{}, 32),
			}
			rc.chats[roomChat.ChatID] = chat
			if newChat == nil {
				newChat = chat
			}
		}
	}

	if newChat != nil {
		newChat.members[id] = struct{}{}
		return newChat.area, newChat.chatID, nil
	}

	return 0, 0, fmt.Errorf("add to room failed")
}

// Remove remove session from room chat.
func (rc *RoomChats) Remove(chatID uint64, id SessionID) {
	rc.Lock()
	defer rc.Unlock()
	chat, ok := rc.chats[chatID]
	if !ok {
		return
	}

	delete(chat.members, id)
}

// Members returns room chat's members.
func (rc *RoomChats) Members(chatID uint64) (ids []SessionID) {
	rc.RLock()
	defer rc.RUnlock()

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
	defer rc.RUnlock()

	_, ok := rc.chats[chatID]
	return ok
}

// SetAreaLimit set area's limit.
func (rm *Rooms) SetAreaLimit(limit uint32) {
	rm.Lock()
	defer rm.Unlock()

	rm.areaLimit = limit
}

// Enter adds session to room's chat members.
func (rm *Rooms) Enter(roomID uint64, id SessionID) (area uint32, chatID uint64, err error) {
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
