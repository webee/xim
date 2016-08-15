import autobahn from 'autobahn';
import KJUR from 'jsrsasign';
import {trace, trace_objs} from './common';


export function autobahn_debug(debug) {
  autobahn.AUTOBAHN_DEBUG = debug;
}

export class XChatClient {
  constructor(user, sToken, wsuri, key) {
    this.user = user;
    this.sToken = sToken;
    this.wsuri = wsuri;
    this.key = key;

    this.session = null;
    this.onready = null;
    this.onmsg = null;
    this.msgSub = null;
    this.msg_listeners = [];

    this.connection = new autobahn.Connection({
      url: this.wsuri,
      realm: "xchat",
      authmethods: ["xjwt"],
      authid: this.on_challenge(null, "jwt", null),
      //onchallenge: this.on_challenge.bind(this),
    });

    this.connection.onopen = this.on_open.bind(this);
    this.connection.onclose = this.on_close.bind(this);
  }

  addMsgListener(fn, kind, domain) {
    this.msg_listeners.push({
      kind: kind,
      domain: domain,
      fn: fn
    });
  }

  open() {
    this.connection.open();
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
    return
  }

  call(method, args, kwargs) {
    return this.session.call(method, args, kwargs);
  }

  publish(topic, args, kwargs) {
    this.session.publish(topic, args, kwargs);
  }

  on_challenge(session, method, extra) {
    trace_objs("on_challenge>", method, extra);
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

    trace("Connected");
    trace_objs("session>", session);
    trace_objs("details>", details);

    this.session.subscribe('xchat.user.{}.msg'.format(this.session.id), this.on_msg.bind(this)).then(
      this.on_msg_sub.bind(this),
      function (err) {
        console.error('failed to subscribe to topic>', err);
      }
    );
    // publish client info.
    this.session.publish('xchat.user.info.pub', [""]);

    // ready.
    if (this.onready) {
      this.onready();
    }
  }

  on_msg(args, kwargs) {
    var kind = args[0];
    var msgs = args[1];
    if (this.onmsg) {
      this.onmsg(kind, msgs);
    }

    msgs.forEach(msg=> {
        this.msg_listeners.forEach(l=> {
            if (l.kind && l.kind !== kind) {
              return
            }
            if (l.domain && l.domain !== msg.domain) {
              return
            }
            l.fn(kind, msg);
          }
        );
      }
    );
  }

  on_msg_sub(sub) {
    this.msgSub = sub;
    trace('subscribed to msg topic');
  }

  on_close(reason, details) {
    trace("Connection lost");
    trace_objs("reason>", reason);
    trace_objs("details>", details);

    this.msgSub = null;
    this.session = null;
  }
}
