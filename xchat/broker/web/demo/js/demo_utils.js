export class DemoUtils {
  constructor(xchatClient) {
    this.maxMsgCount = 100;
    this.do_update = true;
    this.xchatClient = xchatClient;
    this.cur_user_full_name = "*{}*".format(this.xchatClient.user.full_user);
  }

  get_chat_id() {
    return document.getElementById("chat_id").value;
  }

  get_domain() {
    return document.getElementById("domain").value;
  }

  get_send_msg() {
    return document.getElementById("send").value;
  }

  clear_send_msg() {
    document.getElementById("send").value = "";
  }

  do_send_msg(chat_id, domain, msg, show, onsuccess, onfail, onerror) {
    this.xchatClient.sendMsg(chat_id, msg, domain).then(
      res=> {
        console.log("res:", res);
        var args = res.args;
        if (args[0]) {
          if (show) {
            this.newMsg("chat", {
              chat_id: chat_id,
              user: this.cur_user_full_name,
              id: args[1],
              ts: args[2],
              msg: msg
            });
          }
          if (onsuccess) {
            onsuccess();
          }
        } else {
          console.error("error:" + args[2]);
          if (onfail) {
            onfail(args[2]);
          }
        }
      }).catch(
      err=> {
        console.error("err:", err);
        if (onerror) {
          onerror(err);
        }
      }
    );
  }

  sendMsg() {
    var chat_id = this.get_chat_id();
    var content = this.get_send_msg();
    var domain = this.get_domain();
    if (content.length === 0) {
      alert("消息不能为空");
      return
    }

    this.do_send_msg(chat_id, domain, content, true, ::this.clear_send_msg, function (fail) {
      setTimeout(()=>alert("error:" + fail), 0);
    }, function (err) {
      setTimeout(()=>alert("send error:" + err), 0);
    });
  }

  testSendMsg(s, chat_id, i, n, show, domain) {
    if (i < n) {
      var content = s.format(i);
      this.do_send_msg(chat_id, domain, content, show, (function () {
        this.testSendMsg(s, chat_id, i + 1, n, show);
      }).bind(this));
    }
  }

  testSendNotify(s, chat_id, i, n, interval) {
    if (i < n) {
      var content = s + i;
      this.xchatClient.publish('xchat.user.notify.pub', [chat_id, content]);
      interval = interval || 100;
      setTimeout((function () {
        this.testSendNotify(s, chat_id, i + 1, n, interval);
      }).bind(this), interval);
    }
  }

  testSendUserNotify(user, domain, s, i, n, interval) {
    if (i < n) {
      var content = s + i;
      this.xchatClient.publish('xchat.user.usernotify.pub', [user, content, domain]);
      interval = interval || 100;
      setTimeout((function () {
        this.testSendUserNotify(user, domain, s, i + 1, n, interval);
      }).bind(this), interval);
    }
  }


  testPing(...args) {
    if (args.length === 0) {
      args = ["net", 0, 0];
    }

    this.xchatClient.call('xchat.ping', args).then(res=>console.log("res:", res), err=>console.error("err:", err));
  }

  newMsg(kind, msg) {
    if (this.do_update) {
      var msg_div = document.querySelector("#msg");
      var p = document.createElement("p");
      if (kind === "chat") {
        p.innerHTML = msg.ts + ": [" + kind + "]" + msg.chat_id + "@" + msg.user + "#" + msg.id + " 「" + msg.msg + "」";
      } else if (kind === "chat_notify") {
        p.innerHTML = msg.ts + ": [" + kind + "]" + msg.chat_id + "@" + msg.user + " 「" + msg.msg + "」";
      } else if (kind === "user_notify") {
        p.innerHTML = msg.ts + ": [" + kind + "]" + " 「" + msg.msg + "」";
      } else {
        p.innerHTML = JSON.stringify(msg);
      }
      msg_div.insertBefore(p, msg_div.firstChild);
      if (msg_div.childElementCount > this.maxMsgCount) {
        msg_div.removeChild(msg_div.lastElementChild);
      }
    }
  }

  update_on() {
    this.do_update = true;
    return do_update;
  }

  update_off() {
    this.do_update = false;
    return do_update;
  }
}
