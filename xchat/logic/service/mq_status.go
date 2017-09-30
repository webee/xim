package service

import "time"

type userStatus struct {
	User   string    `json:"user"`
	Status string    `json:"status"`
	Ts     time.Time `json:"ts"`
}

func publishUserStatus(user string, status string, ts time.Time) {
	/* FIXME: 暂时停止
	m := userStatus{
		User:   user,
		Status: status,
		Ts:     ts,
	}

	b, err := json.Marshal(&m)
	if err != nil {
		l.Warning("json encoding error: %s", err.Error())
		return
	}

	mq.Publish(mq.XChatUserStatuses, string(b))
	*/
}
