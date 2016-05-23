package db

import (
	"log"

	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

var (
	initialMsgStore *MsgStore
)

// InitMsgDB init the msg db.
func InitMsgDB(mangoURL string) {
	session, err := mgo.Dial(mangoURL)
	if err != nil {
		log.Panicln(err)
	}
	initialMsgStore = NewMsgStore("xim", session)
	msgCol := initialMsgStore.c("msg")
	if err := msgCol.Create(&mgo.CollectionInfo{}); err != nil {
		log.Println("create msg collction: ", err)
	}
	if err := msgCol.EnsureIndex(mgo.Index{
		Key:    []string{"channel", "-id"},
		Unique: true,
	}); err != nil {
		log.Println("create msg collction index(channel,-id): ", err)
	}
}

// GetMsgStore returns a msg store instance.
func GetMsgStore() *MsgStore {
	return initialMsgStore.Copy()
}

// MsgCounter is the channel message counter.
type MsgCounter struct {
	Channel string `bson:"_id"`
	Seq     int    `bson:"id"`
}

// ChannelMsg is the channel message.
type ChannelMsg struct {
	Channel string      `bson:"channel"`
	ID      int         `bson:"id"`
	Ts      int64       `bson:"ts"`
	User    string      `bson:"user"`
	Msg     interface{} `bson:"msg"`
}

// MsgStore handles message store.
type MsgStore struct {
	dbName  string
	session *mgo.Session
}

// NewMsgStore creates a new msg store.
func NewMsgStore(db string, session *mgo.Session) *MsgStore {
	return &MsgStore{db, session}
}

// Copy returns a copied message store.
func (s *MsgStore) Copy() *MsgStore {
	return NewMsgStore(s.dbName, s.session.Copy())
}

// Close free resources.
func (s *MsgStore) Close() {
	s.session.Close()
}

func (s *MsgStore) db() *mgo.Database {
	return s.session.DB(s.dbName)
}

func (s *MsgStore) c(c string) *mgo.Collection {
	db := s.db()
	return db.C(c)
}

// LastID gets channel's last message id.
func (s *MsgStore) LastID(channel string) (int, error) {
	msgCol := s.c("msg")
	var channelMsgs []ChannelMsg
	err := msgCol.Find(bson.M{"channel": channel}).Select(bson.M{"id": 1, "_id": 0}).Sort("-id").Limit(1).All(&channelMsgs)
	if err != nil {
		return 0, err
	}

	if len(channelMsgs) < 1 {
		return 0, nil
	}
	return channelMsgs[0].ID, nil
}

// AddChannelMsg add channel msg to db.
func (s *MsgStore) AddChannelMsg(channel string, id int, ts int64, user string, msg interface{}) error {
	channelMsg := ChannelMsg{
		Channel: channel,
		ID:      id,
		Ts:      ts,
		User:    user,
		Msg:     msg,
	}

	msgCol := s.c("msg")

	return msgCol.Insert(&channelMsg)
}
