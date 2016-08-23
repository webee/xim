import 'webrtc-adapter';
import {CallSession} from './session';


export class CallManager {
  // 实时通话策略控制与流程管理
  constructor(xchatClient, config) {
    this.xchatClient = xchatClient;
    this.user = xchatClient.user.full_user;
    this.xchatClient.subscribeMsg(::this.on_user_rtc_notify, "user_notify", "rtc");
    this.sessions = {};

    // listeners.
    this.oncalling = config.oncalling;
  }

  new_session(role, peer, config) {
    let mgr = this;
    let session = new CallSession({
      role,
      peer,
      user: mgr.user,
      stream: config.stream,
      peer_id: config.peer_id,
      onhangup: config.onhangup,
      onaddstream: config.onaddstream
    }, msg => {
      return mgr.xchatClient.sendUserNotify(peer, JSON.stringify(msg), "rtc", { is_ns_user: true });
    });
    this.sessions[session.id] = session;
    return session;
  }

  removeSession(session) {
    delete this.sessions[session.id];
  }

  // 呼叫=>dial
  calling(user, config) {
    let mgr = this;
    return new Promise((resolve, reject)=> {
      navigator.mediaDevices.getUserMedia({ audio: true, video: true }).then(stream => {
        console.log('Received local stream');
        let session = mgr.new_session("caller", user, config);
        session.stream = stream;
        session.calling();

        resolve(session);
      }).catch(err => {
        reject(new Error("getUserMedia() error: {}".format(err.name)));
      });
    });
  }

  answer(session) {
    return new Promise((resolve, reject)=> {
      navigator.mediaDevices.getUserMedia({ audio: true, video: true }).then(stream => {
        console.log('Received local stream');
        session.stream = stream;
        session.answer();

        resolve(session);
      }).catch(err => {
        reject(new Error("getUserMedia() error: {}".format(err.name)));
      });
    });
  }

  on_user_rtc_notify(_, notify) {
    let msg = JSON.parse(notify.msg);

    var session = null;
    if (msg.type === 'calling') {
      // {type: "calling", user: "test:test", id: 1234567890}
      if (this.oncalling) {
        if (!this.sessions[msg.id]) {
          // self.
          session = this.new_session("callee", msg.user, { peer_id: msg.id });
          this.oncalling(session);
        }
      }
    } else {
      let peer_id = msg.peer_id;
      if (peer_id !== undefined) {
        session = this.sessions[msg.peer_id];
      }
    }
    if (session) {
      session.onMsg(msg);
    }
  }
}