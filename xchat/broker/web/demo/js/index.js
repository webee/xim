import autobahn from 'autobahn';
import KJUR from 'jsrsasign';
import format from 'string-format';
import {anyUserkey} from './configs';
import {decode_ns_user} from './utils';

// init.
format.extend(String.prototype);
autobahn.AUTOBAHN_DEBUG = true;

var user = decode_ns_user(document.location.hash.substr(1) || "test:test");
var sToken = document.location.search.substr(1);
var wsuri = (document.location.protocol === "http:" ? "ws:" : "wss:") + "//" + document.location.host + "/ws";


function onchallenge(session, method, extra) {
  console.log("onchallenge", method, extra);
  if (method === "jwt") {
    if (!!sToken) {
      return sToken;
    }
    // Header
    var oHeader = { alg: 'HS256', typ: 'JWT' };
    // Payload
    var oPayload = {};
    var tEnd = KJUR.jws.IntDate.get('now + 1day');
    oPayload.exp = tEnd;
    oPayload.user = user.user;
    oPayload.ns = user.ns;

    // Sign JWT, password=616161
    var sHeader = JSON.stringify(oHeader);
    var sPayload = JSON.stringify(oPayload);

    var token = KJUR.jws.JWS.sign("HS256", sHeader, sPayload, anyUserkey);

    if (!!user.ns) {
      return user.ns + ':' + token;
    }
    return token;
  } else {
    throw "don't know how to authenticate using '" + method + "'";
  }
}


// the WAMP connection to the Router
//
var maxMsgCount = 100;
var session = null;
var reply_sub = null;
var msg_sub = null;
var do_update = true;

var connection = new autobahn.Connection({
  url: wsuri,
  realm: "xchat",
  authmethods: ["xjwt"],
  authid: onchallenge(null, "jwt", null),
  //onchallenge: onchallenge,
});


// fired when connection is established and session attached
connection.onopen = function (newSession, details) {
  session = newSession;
  console.log("Connected");
  console.log("session:", session);
  console.log("details:", details);

  function onMsg(args, kwargs) {
    console.log(">>>>msg:", args, kwargs);
    var kind = args[0];
    var msgs = args[1];
    for (let i = 0; i < msgs.length; i++) {
      newMsg(kind, msgs[i]);
    }
  }

  session.subscribe('xchat.user.' + session.id + '.msg', onMsg).then(
    function (sub) {
      msg_sub = sub;
      window.msg_sub = msg_sub;
      console.log('subscribed to topic');
    },
    function (err) {
      console.error('failed to subscribe to topic', err);
    }
  );
  session.publish('xchat.user.info.pub', [""]);
};

// fired when connection was lost (or could not be established)
//
connection.onclose = function (reason, details) {
  console.log("Connection lost")
  console.log("reason:", reason)
  console.log("details:", details)
  reply_sub = null;
  msg_sub = null;
  session = null;
};

// now actually open the connection
connection.open();


// functions.
function sendMsg() {
  if (!session) {
    return
  }

  var chat_id = document.getElementById("chat_id").value;
  var txt = document.getElementById("send");
  var content = txt.value;
  if (content.length === 0) {
    alert("消息不能为空");
    return
  }

  session.call('xchat.user.msg.send', [chat_id, content]).then(
    function (res) {
      console.log("res:", res);
      var args = res.args;
      if (args[0]) {
        newMsg("chat", { chat_id: chat_id, user: "*" + user.full_user + "*", id: args[1], ts: args[2], msg: content });
        txt.value = "";
      } else {
        alert("error:" + args[2]);
      }
    },
    function (err) {
      console.error("err:", err);
      alert("send error");
    }
  );
}

function testPing() {
  var args = [];
  for (let i = 0; i < arguments.length; i++) {
    args.push(arguments[i]);
  }
  if (args.length === 0) {
    args = ["net", 0, 0];
  }

  session.call('xchat.ping', args).then(res=>console.log("res:", res), err=>console.error("err:", err));
}

function testSendMsg(s, chat_id, i, n, show) {
  if (i < n) {
    var content = s.format(i);
    session.call('xchat.user.msg.send', [chat_id, content]).then(
      res=> {
        console.log("res:", res);
        var args = res.args;
        if (args[0]) {
          if (show) {
            newMsg("chat", {
              chat_id: chat_id,
              user: "*" + user.full_user + "*",
              id: args[1],
              ts: args[2],
              msg: content
            });
          }
        } else {
          console.error("error:" + args[2]);
        }
        testSendMsg(s, chat_id, i + 1, n, show);
      }).catch(
      err=> {
        console.error("err:", err);
        testSendMsg(s, chat_id, i + 1, n, show);
      }
    );
  }
}

function testSendNotify(s, chat_id, i, n, interval) {
  if (i < n) {
    var content = s + i;
    session.publish('xchat.user.msg.pub', [chat_id, content]);
    interval = interval || 100;
    setTimeout(function () {
      testSendNotify(s, chat_id, i + 1, n, interval);
    }, interval);
  }
}

function update_on() {
  do_update = true;
  return do_update;
}

function update_off() {
  do_update = false;
  return do_update;
}

function newMsg(kind, msg) {
  if (do_update) {
    var msg_div = document.querySelector("#msg");
    var p = document.createElement("p");
    if (kind === "chat") {
      p.innerHTML = msg.ts + ": [" + kind + "]" + msg.chat_id + "@" + msg.user + "#" + msg.id + " 「" + msg.msg + "」";
    } else if (kind === "chat_notify") {
      p.innerHTML = msg.ts + ": [" + kind + "]" + msg.chat_id + "@" + msg.user + " 「" + msg.msg + "」";
    }
    msg_div.insertBefore(p, msg_div.firstChild);
    if (msg_div.childElementCount > maxMsgCount) {
      msg_div.removeChild(msg_div.lastElementChild);
    }
  }
}


// gobals.
window.testPing = testPing;
window.sendMsg = sendMsg;
window.update_off = update_off;
window.update_on = update_on;
window.testSendMsg = testSendMsg;
window.testSendNotify = testSendNotify;
