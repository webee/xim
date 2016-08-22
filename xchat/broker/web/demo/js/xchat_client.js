import autobahn from 'autobahn';
import KJUR from 'jsrsasign';


export function autobahn_debug(debug) {
  autobahn.AUTOBAHN_DEBUG = debug;
}

export class XChatClient {
  constructor(config) {
    this.user = config.user;
    this.sToken = config.sToken;
    this.wsuri = config.wsuri;
    this.key = config.key;
    this.debug_log = config.debug_log;

    this.onready = config.onready || (()=> {});

    this.session = null;
    this.msg_subscribers = [];
    if (config.onmsg) {
      this.subscribeMsg(config.onmsg);
    }

    this.connection = new autobahn.Connection({
      url: this.wsuri,
      realm: "xchat",
      authmethods: ["xjwt"],
      authid: this.on_challenge(null, "jwt", null),
      //onchallenge: ::this.on_challenge,
    });

    this.connection.onopen = ::this.on_open;
    this.connection.onclose = ::this.on_close;

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

  sendMsg(chat_id, msg, domain) {
    if (domain) {
      return this.session.call('xchat.user.msg.send', [chat_id, msg, domain]);
    }
    return this.session.call('xchat.user.msg.send', [chat_id, msg]);
  }

  sendNotify(chat_id, msg, domain) {
    if (domain) {
      return this.session.call('xchat.user.notify.send', [chat_id, msg, domain]);
    }
    return this.session.call('xchat.user.notify.send', [chat_id, msg]);
  }

  pubNotify(chat_id, msg, domain) {
    if (domain) {
      this.session.publish('xchat.user.notify.pub', [chat_id, msg, domain]);
      return
    }
    this.session.publish('xchat.user.notify.pub', [chat_id, msg]);
  }

  sendUserNotify(user, msg, domain) {
    if (domain) {
      return this.session.call('xchat.user.usernotify.send', [user, msg, domain]);
    }
    return this.session.call('xchat.user.usernotify.send', [user, msg]);
  }

  pubUserNotify(user, msg, domain) {
    if (domain) {
      this.session.publish('xchat.user.usernotify.pub', [user, msg, domain]);
      return
    }
    this.session.publish('xchat.user.usernotify.pub', [user, msg]);
  }

  call(method, args, kwargs) {
    return this.session.call(method, args, kwargs);
  }

  publish(topic, args, kwargs) {
    this.session.publish(topic, args, kwargs);
  }

  on_challenge(session, method, extra) {
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
      throw "don't know how to authenticate using '{}'".format(method);
    }
  }

  on_open(session, details) {
    this.session = session;

    this.debug_log("Connected");
    this.debug_log("session>", session);
    this.debug_log("details>", details);

    this.session.subscribe('xchat.user.{}.msg'.format(this.session.id), ::this.on_msg).then(
      ::this.on_msg_sub,
      function (err) {
        console.error('failed to subscribe to topic>', err);
      }
    );

    // publish client info.
    this.session.publish('xchat.user.info.pub', [""]);

  }

  on_close(reason, details) {
    this.debug_log("Connection lost");
    this.debug_log("reason>", reason);
    this.debug_log("details>", details);

    this.session = null;
  }

  on_msg(args, kwargs) {
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

  on_msg_sub(sub) {
    this.debug_log('subscribed to msg topic');

    // ready.
    this.onready(this);
  }
}
