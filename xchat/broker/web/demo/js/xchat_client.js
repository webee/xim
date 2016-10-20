import wamp from 'wamp.js';
import KJUR from 'jsrsasign';


export function wampDebug(debug) {
	if (debug) {
		wamp.debugOn();
	} else {
		wamp.debugOff();
	}
}

export class XChatClient {
	constructor(config) {
		this.user = config.user;
		this.sToken = config.sToken;
		this.wsuri = config.wsuri;
		this.key = config.key;
		this.debug_log = config.debug_log;

		this.onready = config.onready || (()=> {
			});
		this.onclose = config.onclose || (()=> {
			});
		this.onerror = config.onerror || (()=> {
			});
		this.onstatechange = config.onstatechange || (()=> {
			});
		this.state = "";

		this.session = null;
		this.msg_subscribers = [];
		if (config.onmsg) {
			this.subscribeMsg(config.onmsg);
		}

		//this.connection = new autobahn.Connection({
		this.connection = new wamp.Connection({
			url: this.wsuri,
			realm: "xchat",
			authmethods: ["xjwt"],
			authid: this._on_challenge(null, "jwt", null),
			//onchallenge: ::this._on_challenge,
		});

		this.connection.onopen = ::this._on_open;
		this.connection.onclose = ::this._on_close;

		// open wamp connection.
		this.connection.open();
	}

	subscribeMsg(fn, kind, domain) {
		this.msg_subscribers.push({
			kind: kind,
			domain: domain,
			fn: fn
		});
	}

	sendMsg(chat_id, msg, domain, kwargs) {
		if (domain) {
			return this.session.call('xchat.user.msg.send', [chat_id, msg, domain], kwargs);
		}
		return this.session.call('xchat.user.msg.send', [chat_id, msg], kwargs);
	}

	sendNotify(chat_id, msg, domain, kwargs) {
		if (domain) {
			return this.session.call('xchat.user.notify.send', [chat_id, msg, domain], kwargs);
		}
		return this.session.call('xchat.user.notify.send', [chat_id, msg], kwargs);
	}

	pubNotify(chat_id, msg, domain) {
		if (domain) {
			this.session.publish('xchat.user.notify.pub', [chat_id, msg, domain]);
			return
		}
		this.session.publish('xchat.user.notify.pub', [chat_id, msg], kwargs);
	}

	sendUserNotify(user, msg, domain, kwargs) {
		if (domain) {
			return this.session.call('xchat.user.usernotify.send', [user, msg, domain], kwargs);
		}
		return this.session.call('xchat.user.usernotify.send', [user, msg], kwargs);
	}

	pubUserNotify(user, msg, domain, kwargs) {
		if (domain) {
			this.session.publish('xchat.user.usernotify.pub', [user, msg, domain], kwargs);
			return
		}
		this.session.publish('xchat.user.usernotify.pub', [user, msg], kwargs);
	}

	call(method, args, kwargs) {
		return this.session.call(method, args, kwargs);
	}

	publish(topic, args, kwargs) {
		this.session.publish(topic, args, kwargs);
	}

	_on_challenge(session, method, extra) {
		this.debug_log("on_challenge>", method, extra);
		if (method === "jwt") {
			if (!!this.sToken) {
				return this.sToken;
			}
			// Header
			var oHeader = { alg: 'HS256', typ: 'JWT' };
			// Payload
			var oPayload = {};
			var tEnd = KJUR.jws.IntDate.get('now + 1day');
			oPayload.exp = tEnd;
			oPayload.user = this.user.user;
			oPayload.ns = this.user.ns;

			// Sign JWT, password=616161
			var sHeader = JSON.stringify(oHeader);
			var sPayload = JSON.stringify(oPayload);

			var token = KJUR.jws.JWS.sign("HS256", sHeader, sPayload, this.key);

			if (!!this.user.ns) {
				return this.user.ns + ':' + token;
			}
			return token;
		} else {
			throw `don't know how to authenticate using '${method}'`;
		}
	}

	_change_state(state) {
		this.state = state;
		this.onstatechange(this.state);
	}

	_on_open(session, details) {
		this._change_state("opened");
		this.session = session;
		this.session.subscribe(`xchat.user.${this.session.id}.msg`, ::this._on_msg).then(
			::this._on_msg_sub,
			this.onerror
		);

		// publish client info.
		this.session.publish('xchat.user.info.pub', [""]);
	}

	_on_close(reason, details) {
		this._change_state(`closed.${reason}`);

		this.session = null;
		if (reason !== "lost") {
			// 其它的会重连
			this.onclose();
		}
	}

	_on_msg(args, kwargs) {
		var kind = args[0];
		var msgs = args[1];

		msgs.forEach(msg=> {
				this.msg_subscribers.forEach(s=> {
						if (s.kind && s.kind !== kind) {
							return
						}
						if (s.domain && s.domain !== msg.domain) {
							return
						}
						s.fn(kind, msg);
					}
				);
			}
		);
	}

	_on_msg_sub(sub) {
		// ready.
		this.onready(this);
		this._change_state("ready");
	}
}
