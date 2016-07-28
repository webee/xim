package turnpike

import (
	"fmt"
	"time"
)

const (
	defaultAuthTimeout = 2 * time.Minute
)

// A Realm is a WAMP routing and administrative domain.
//
// Clients that have connected to a WAMP router are joined to a realm and all
// message delivery is handled by the realm.
type Realm struct {
	_   string
	URI URI
	Broker
	Dealer
	Authorizer
	Interceptor
	CRAuthenticators map[string]CRAuthenticator
	Authenticators   map[string]Authenticator
	// DefaultAuth      func(details map[string]interface{}) (map[string]interface{}, error)
	AuthTimeout time.Duration
	clients     map[ID]*Session
	localClient
	actor Actor
}

type localClient struct {
	*Client
}

func (r *Realm) getPeerWithSize(s int, rt Router, details map[string]interface{}) (Peer, error) {
	peerA, peerB := localPipeWithSize(s)
	if details == nil {
		details = make(map[string]interface{})
	}
	sess := &Session{Peer: peerA, Id: NewID(), Details: details, kill: make(chan URI, 1)}
	rt.OpenSession(r, sess)
	tlog.Println("Established internal session:", sess)
	return peerB, nil
}

func (r *Realm) getPeer(rt Router, details map[string]interface{}) (Peer, error) {
	return r.getPeerWithSize(100, rt, details)
}

// Close disconnects all clients after sending a goodbye message
func (r *Realm) Close() {
	r.actor.Acts() <- func() {
		for _, client := range r.clients {
			select {
			case client.kill <- ErrSystemShutdown:
			default:
			}
		}
	}

	var (
		nclients int
	)
	for {
		r.actor.SyncAct(func() {
			nclients = len(r.clients)
		})
		if nclients == 0 {
			break
		}
	}

	r.actor.Close()
}

func (r *Realm) init(rt Router) {
	r.clients = make(map[ID]*Session)
	r.actor = NewChannelActor()
	p, _ := r.getPeer(rt, nil)
	r.localClient.Client = NewClient(p)
	if r.Broker == nil {
		r.Broker = NewDefaultBroker()
	}
	if r.Dealer == nil {
		r.Dealer = NewDefaultDealer()
	}
	if r.Authorizer == nil {
		r.Authorizer = NewDefaultAuthorizer()
	}
	if r.Interceptor == nil {
		r.Interceptor = NewDefaultInterceptor()
	}
	if r.AuthTimeout == 0 {
		r.AuthTimeout = defaultAuthTimeout
	}

	r.actor.Start()
}

func (l *localClient) onJoin(details map[string]interface{}) {
	l.Publish("wamp.session.on_join", []interface{}{details}, nil)
}

func (l *localClient) onLeave(session ID) {
	l.Publish("wamp.session.on_leave", []interface{}{session}, nil)
}

// KillSession disconnects a session.
func (r *Realm) KillSession(id ID) {
	r.actor.Acts() <- func() {
		if sess, ok := r.clients[id]; ok {
			select {
			case sess.kill <- ErrAbortAndOut:
			default:
			}
		}
	}
}

func (r *Realm) handleSession(sess *Session) {
	r.actor.SyncAct(func() {
		r.clients[sess.Id] = sess
	})
	r.onJoin(sess.Details)

	defer func() {
		r.actor.Acts() <- func() {
			delete(r.clients, sess.Id)
		}
		r.Broker.RemovePeer(sess.Peer)
		r.Dealer.RemovePeer(sess.Peer)
		r.onLeave(sess.Id)
	}()
	c := sess.Receive()
	// TODO: what happens if the realm is closed?

	for {
		var msg Message
		var open bool
		select {
		case msg, open = <-c:
			if !open {
				tlog.Println("lost session:", sess)
				return
			}
		case reason := <-sess.kill:
			switch reason {
			case ErrAbortAndOut:
				logErr(sess.Send(&Abort{Reason: reason, Details: make(map[string]interface{})}))
			default:
				logErr(sess.Send(&Goodbye{Reason: reason, Details: make(map[string]interface{})}))
			}
			tlog.Printf("kill session %s: %v", sess, reason)
			// TODO: wait for client Goodbye?
			return
		}

		tlog.Printf("[%s] %s: %+v", sess, msg.MessageType(), msg)
		if isAuthz, err := r.Authorizer.Authorize(sess, msg); !isAuthz {
			errMsg := &Error{Type: msg.MessageType()}
			if err != nil {
				errMsg.Error = ErrAuthorizationFailed
				tlog.Printf("[%s] authorization failed: %v", sess, err)
			} else {
				errMsg.Error = ErrNotAuthorized
				tlog.Printf("[%s] %s UNAUTHORIZED", sess, msg.MessageType())
			}
			logErr(sess.Send(errMsg))
			continue
		}

		r.Interceptor.Intercept(sess, &msg)

		switch msg := msg.(type) {
		case *Goodbye:
			logErr(sess.Send(&Goodbye{Reason: ErrGoodbyeAndOut, Details: make(map[string]interface{})}))
			tlog.Printf("[%s] leaving: %v", sess, msg.Reason)
			return

		// Broker messages
		case *Publish:
			r.Broker.Publish(sess.Peer, msg)
		case *Subscribe:
			r.Broker.Subscribe(sess.Peer, msg)
		case *Unsubscribe:
			r.Broker.Unsubscribe(sess.Peer, msg)

		// Dealer messages
		case *Register:
			r.Dealer.Register(sess.Peer, msg)
		case *Unregister:
			r.Dealer.Unregister(sess.Peer, msg)
		case *Call:
			r.Dealer.Call(sess.Peer, msg)
		case *Yield:
			r.Dealer.Yield(sess.Peer, msg)

		// Error messages
		case *Error:
			if msg.Type == INVOCATION {
				// the only type of ERROR message the router should receive
				r.Dealer.Error(sess.Peer, msg)
			} else {
				tlog.Printf("invalid ERROR message received: %v", msg)
			}

		default:
			tlog.Println("Unhandled message:", msg.MessageType())
		}
	}
}

func (r *Realm) handleAuth(client Peer, details map[string]interface{}) (*Welcome, error) {
	msg, err := r.authenticate(details)
	if err != nil {
		return nil, err
	}
	// we should never get anything besides WELCOME and CHALLENGE
	if msg.MessageType() == WELCOME {
		return msg.(*Welcome), nil
	}
	// Challenge response
	challenge := msg.(*Challenge)
	if err := client.Send(challenge); err != nil {
		return nil, err
	}

	msg, err = GetMessageTimeout(client, r.AuthTimeout)
	if err != nil {
		return nil, err
	}
	tlog.Printf("%s: %+v", msg.MessageType(), msg)
	if authenticate, ok := msg.(*Authenticate); !ok {
		return nil, fmt.Errorf("unexpected %s message received", msg.MessageType())
	} else {
		return r.checkResponse(challenge, authenticate)
	}
}

// Authenticate either authenticates a client or returns a challenge message if
// challenge/response authentication is to be used.
func (r Realm) authenticate(details map[string]interface{}) (Message, error) {
	tlog.Println("details:", details)
	if len(r.Authenticators) == 0 && len(r.CRAuthenticators) == 0 {
		return &Welcome{}, nil
	}
	// TODO: this might not always be a []interface{}. Using the JSON unmarshaller it will be,
	// but we may have serializations that preserve more of the original type.
	// For now, the tests just explicitly send a []interface{}
	_authmethods, ok := details["authmethods"].([]interface{})
	if !ok {
		return nil, fmt.Errorf("No authentication supplied")
	}
	authmethods := []string{}
	for _, method := range _authmethods {
		if m, ok := method.(string); ok {
			authmethods = append(authmethods, m)
		} else {
			tlog.Printf("invalid authmethod value: %v", method)
		}
	}
	for _, method := range authmethods {
		if auth, ok := r.CRAuthenticators[method]; ok {
			if challenge, err := auth.Challenge(details); err != nil {
				return nil, err
			} else {
				return &Challenge{AuthMethod: method, Extra: challenge}, nil
			}
		}
		if auth, ok := r.Authenticators[method]; ok {
			if authDetails, err := auth.Authenticate(details); err != nil {
				return nil, err
			} else {
				return &Welcome{Details: addAuthMethod(authDetails, method)}, nil
			}
		}
	}
	// TODO: check default auth (special '*' auth?)
	return nil, fmt.Errorf("could not authenticate with any method")
}

// checkResponse determines whether the response to the challenge is sufficient to gain access to the Realm.
func (r Realm) checkResponse(chal *Challenge, auth *Authenticate) (*Welcome, error) {
	authenticator, ok := r.CRAuthenticators[chal.AuthMethod]
	if !ok {
		return nil, fmt.Errorf("authentication method has been removed")
	}
	if details, err := authenticator.Authenticate(chal.Extra, auth.Signature); err != nil {
		return nil, err
	} else {
		return &Welcome{Details: addAuthMethod(details, chal.AuthMethod)}, nil
	}
}

func addAuthMethod(details map[string]interface{}, method string) map[string]interface{} {
	if details == nil {
		details = make(map[string]interface{})
	}
	details["authmethod"] = method
	return details
}

// r := Realm{
// 	Authenticators: map[string]turnpike.Authenticator{
// 		"wampcra": turnpike.NewCRAAuthenticatorFactoryFactory(mySecret),
// 		"ticket": turnpike.NewTicketAuthenticator(myTicket),
// 		"asdfasdf": myAsdfAuthenticator,
// 	},
// 	BasicAuthenticators: map[string]turnpike.BasicAuthenticator{
// 		"anonymous": nil,
// 	},
// }
