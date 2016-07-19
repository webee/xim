package turnpike

import "sync"

// Broker is the interface implemented by an object that handles routing EVENTS
// from Publishers to Subscribers.
type Broker interface {
	// Publishes a message to all Subscribers.
	Publish(Sender, *Publish)
	// Subscribes to messages on a URI.
	Subscribe(Sender, *Subscribe)
	// Unsubscribes from messages on a URI.
	Unsubscribe(Sender, *Unsubscribe)
	// Remove a subscriber's subscriptions.
	RemovePeer(sub Sender)
}

// A super simple broker that matches URIs to Subscribers.
type defaultBroker struct {
	routes        map[URI]map[ID]Sender
	subscriptions map[ID]URI
	subscribers   map[Sender][]ID
	lock          sync.RWMutex
}

// NewDefaultBroker initializes and returns a simple broker that matches URIs to
// Subscribers.
func NewDefaultBroker() Broker {
	return &defaultBroker{
		routes:        make(map[URI]map[ID]Sender),
		subscriptions: make(map[ID]URI),
		subscribers:   make(map[Sender][]ID),
	}
}

// Publish sends a message to all subscribed clients except for the sender.
//
// If msg.Options["acknowledge"] == true, the publisher receives a Published event
// after the message has been sent to all subscribers.
func (br *defaultBroker) Publish(pub Sender, msg *Publish) {
	pubID := NewID()
	evtTemplate := Event{
		Publication: pubID,
		Arguments:   msg.Arguments,
		ArgumentsKw: msg.ArgumentsKw,
		Details:     make(map[string]interface{}),
	}

	br.lock.RLock()
	for id, sub := range br.routes[msg.Topic] {
		// shallow-copy the template
		event := evtTemplate
		event.Subscription = id
		// don't send event to publisher
		if sub != pub {
			sub.Send(&event)
		}
	}
	br.lock.RUnlock()

	// only send published message if acknowledge is present and set to true
	if doPub, _ := msg.Options["acknowledge"].(bool); doPub {
		pub.Send(&Published{Request: msg.Request, Publication: pubID})
	}
}

// Subscribe subscribes the client to the given topic.
func (br *defaultBroker) Subscribe(sub Sender, msg *Subscribe) {
	id := NewID()

	br.lock.Lock()
	if _, ok := br.routes[msg.Topic]; !ok {
		br.routes[msg.Topic] = make(map[ID]Sender)
	}
	br.routes[msg.Topic][id] = sub
	br.subscriptions[id] = msg.Topic

	// subscribers
	ids, ok := br.subscribers[sub]
	if !ok {
		ids = []ID{}
	}
	ids = append(ids, id)
	br.subscribers[sub] = ids

	br.lock.Unlock()

	sub.Send(&Subscribed{Request: msg.Request, Subscription: id})
}

func (br *defaultBroker) RemovePeer(sub Sender) {
	tlog.Printf("broker remove peer %p", &sub)
	br.lock.Lock()
	defer br.lock.Unlock()

	for _, id := range br.subscribers[sub] {
		br.unsubscribe(sub, id)
	}
}

func (br *defaultBroker) Unsubscribe(sub Sender, msg *Unsubscribe) {
	br.lock.Lock()
	if !br.unsubscribe(sub, msg.Subscription) {
		br.lock.Unlock()
		err := &Error{
			Type:    msg.MessageType(),
			Request: msg.Request,
			Error:   ErrNoSuchSubscription,
		}
		sub.Send(err)
		tlog.Printf("Error unsubscribing: no such subscription %v", msg.Subscription)
		return
	}
	br.lock.Unlock()

	sub.Send(&Unsubscribed{Request: msg.Request})
}

func (br *defaultBroker) unsubscribe(sub Sender, id ID) bool {
	tlog.Printf("broker unsubscribing: %p, %d", &sub, id)
	topic, ok := br.subscriptions[id]
	if !ok {
		return false
	}
	delete(br.subscriptions, id)

	if r, ok := br.routes[topic]; !ok {
		tlog.Printf("Error unsubscribing: unable to find routes for %s topic", topic)
	} else if _, ok := r[id]; !ok {
		tlog.Printf("Error unsubscribing: %s route does not exist for %v subscription", topic, id)
	} else {
		delete(r, id)
		if len(r) == 0 {
			delete(br.routes, topic)
		}
	}

	// subscribers
	ids := br.subscribers[sub][:0]
	for _, id := range br.subscribers[sub] {
		if id != id {
			ids = append(ids, id)
		}
	}
	if len(ids) == 0 {
		delete(br.subscribers, sub)
	} else {
		br.subscribers[sub] = ids
	}

	return true
}
