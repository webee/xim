import {SignalingChannel} from './signaling_channel';


export class RTCManager {
  // 实时通话策略控制与流程管理
  constructor(xchatClient) {
    this.xchatClient = xchatClient;
    this.xchatClient.addMsgListener(this.on_rtc_notify.bind(this), "chat_notify", "rtc");
    this.signaling_channels = {};

    // listeners.
    this.oncalling = null;
  }

  new_signal_channel(role, chat_id, peer_id) {
    let signaling_channel = new SignalingChannel(role, (msg=> {
      return this.xchatClient.sendNotify(chat_id, JSON.stringify(msg), "rtc");
    }).bind(this), peer_id);
    this.signaling_channels[signaling_channel.id] = signaling_channel;
    return signaling_channel;
  }

  removeSignalChannel(signalChannel) {
    delete this.signaling_channels[signalChannel.id];
  }

  // 呼叫
  calling(chat_id) {
    let signaling_channel = this.new_signal_channel("caller", chat_id);
    signaling_channel.calling();
    return signaling_channel;
  }

  on_rtc_notify(_, notify) {
    let chat_id = notify.chat_id;
    let user = notify.user;
    let msg = JSON.parse(notify.msg);

    var signaling_channel = null;
    if (msg.type === 'calling') {
      // {type: "calling", id: 1234567890}
      if (this.oncalling) {
        if (!this.signaling_channels[msg.id]) {
          // self.
          signaling_channel = this.new_signal_channel("callee", chat_id, msg.id);
          this.oncalling(signaling_channel);
        }
      }
    } else {
      let peer_id = msg.peer_id;
      if (peer_id !== undefined) {
        signaling_channel = this.signaling_channels[msg.peer_id];
      }
    }
    if (signaling_channel) {
      signaling_channel.onMsg(msg);
    }
  }
}