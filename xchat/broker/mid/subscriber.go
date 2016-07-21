package mid

import (
	"gopkg.in/webee/turnpike.v2"
)

// Subscriber is a simple subscriber.
type Subscriber func(args []interface{}, kwargs map[string]interface{})

// SessionSubscriber is user session subscriber.
type SessionSubscriber func(s *Session, args []interface{}, kwargs map[string]interface{})

func (s Subscriber) subscribeTo(client *turnpike.Client, topic string) error {
	return client.Subscribe(topic, subTopic(topic, s, false))
}

func (s Subscriber) xsubscribeTo(client *turnpike.Client, topic string) error {
	return client.Subscribe(topic, subTopic(topic, s, true))
}

func (s SessionSubscriber) subscribeTo(client *turnpike.Client, topic string) error {
	return client.Subscribe(topic, subTopic(topic, s.subscriber(), false))
}

func (s SessionSubscriber) xsubscribeTo(client *turnpike.Client, topic string) error {
	return client.Subscribe(topic, subTopic(topic, s.subscriber(), true))
}

func (s SessionSubscriber) subscriber() Subscriber {
	return func(args []interface{}, kwargs map[string]interface{}) {
		sess := getSessionFromDetails(kwargs["details"], false)
		if sess == nil {
			return
		}
		s(sess, args, kwargs)
	}
}

func subTopic(topic string, subscriber Subscriber, logInfo bool) turnpike.EventHandler {
	return func(args []interface{}, kwargs map[string]interface{}) {
		defer func() {
			if r := recover(); r != nil {
				l.Warning("[sub]%s: handle error, %s", topic, r)
			}
		}()
		if logInfo {
			l.Info("[sub]%s: %v, %+v\n", URIXChatPubUserInfo, args, kwargs)
		} else {
			l.Debug("[sub]%s: %v, %+v\n", URIXChatPubUserInfo, args, kwargs)
		}
		subscriber(args, kwargs)
	}
}
