export class XChatMsgChannel {
  constructor(xchatClient, domain) {
    this.domain = domain || "rtc";
    this.xchatClient = xchatClient;
    this.xchatClient.subscribeMsg(::this._on_msg, "user_notify", this.domain);

    this.receivers = [];
  }

  registerReceiver(recv, user) {
    this.receivers.push({
      user: user,
      recv: recv
    });
  }

  send(user, msg) {
    return this.xchatClient.sendUserNotify(user, msg, this.domain, { is_ns_user: true });
  }

  _on_msg(_, msg) {
    this.receivers.forEach(r=> {
        if (r.user && r.user !== msg.user) {
          return
        }
        r.recv(msg.user, msg.msg);
      }
    );
  }
}
