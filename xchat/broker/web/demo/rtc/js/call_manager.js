import 'webrtc-adapter';
import {CallSession} from './call_session';
import {SignalingChannel} from './signaling_channel';


export class CallManager {
  // 实时通话策略控制与流程管理
  constructor(msgChannel, config) {
    this.msgChannel = msgChannel;
    this.msgChannel.registerReceiver(::this._on_user_msg);
    this.sessions = {};

    // listeners.
    this.onringing = config.onringing || function () {
      };
  }

  new_session(role, peer, config) {
    config.role = role;
    config.peer = peer;
    config.onringing = this.onringing;
    let session = new CallSession(new SignalingChannel(peer, this.msgChannel), config);
    this.sessions[session.id] = session;
    return session;
  }

  removeSession(session) {
    delete this.sessions[session.id];
  }

  // 呼叫=>dial
  calling(user, config) {
    let mgr = this;
    return navigator.mediaDevices.getUserMedia({ audio: true, video: true }).then(stream => {
      console.log('Received local stream');
      let session = mgr.new_session("caller", user, config);
      session.stream = stream;
      session.calling();

      return Promise.resolve(session);
    }).catch(err => {
      return Promise.reject(new Error("getUserMedia() error: {}".format(err.name)));
    });
  }

  answer(session) {
    return navigator.mediaDevices.getUserMedia({ audio: true, video: true }).then(stream => {
      console.log('Received local stream');
      session.stream = stream;
      session.answer();

      return Promise.resolve(session);
    }).catch(err => {
      return Promise.reject(new Error("getUserMedia() error: {}".format(err.name)));
    });
  }

  _on_user_msg(user, msg) {
    msg = JSON.parse(msg);

    var session;
    if (msg.type === 'calling') {
      // {type: "calling", id: 1234567890}
      if (!this.sessions[msg.peer_id]) {
        session = this.new_session("callee", user, { peer_id: msg.id });
      }
    } else {
      session = this.sessions[msg.peer_id];
    }
    if (session) {
      session.signaling_channel.recv(msg);
    }
  }
}