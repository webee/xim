import axios from 'axios';

function gen_random_id() {
  return Math.random() * 2**53;
}

export class CallSession {
  constructor(config, send_msg) {
    this.role = config.role;
    this.user = config.user;
    this.stream = null;
    this.remoteStream = null;
    this.peer = config.peer;

    this.send_msg = send_msg;
    this.state = "init";

    this.id = gen_random_id();
    this.peer_id = config.peer_id || null;

    this.ringing_peers = {};

    this.pc = null;

    // listeners.
    this.onhangup = config.onhangup;
    this.onaddstream = config.onaddstream;

    this.onsignalingstatechange = ()=> {
    };
    this.oniceconnectionstatechange = ()=> {
    };
  }

  onMsg(msg) {
    console.log("rtc msg:", msg);
    switch (msg.type) {
      case "calling":
        // 呼叫: {type: "calling", id: 1234567890}
        // TODO: check message structure.
        this.on_calling(msg.id);
        break;
      case "ringing":
        // 响铃: {type: "ringing", peer_id: 1234567890, id: 9876543210}
        this.on_ringing(msg.id);
        break;
      case "ok":
        // 接听: {type: "ok", peer_id: 1234567890, id: 9876543210}
        // 就绪: {type: "ok", peer_id: 9876543210, id: 1234567890}
        this.on_ok(msg.id);
        break;
      case "hangup":
        // 错误: {type: "hangup", peer_id: 12345, id: 67890, reason: "error"}
        // 拒接: {type: "hangup", peer_id: 12345, id: 67890, reason: "refuse"}
        // 繁忙: {type: "hangup", peer_id: 12345, id: 67890, reason: "busy"}
        // 取消: {type: "hangup", peer_id: 12345, id: 67890, reason: "cancel"}
        // 挂断: {type: "hangup", peer_id: 12345, id: 67890, reason: "hangup"}
        this.on_hangup(msg.id, msg.reason);
        break;
      case "rtc":
        // rtc 信令
        this.on_rtc(msg.sub_type, msg.msg);
    }
  }

  on_rtc(type, msg) {
    if (this.state === "ready") {
      switch (type) {
        case "candidate":
          this._on_candidate(msg);
          break;
        case "sdp":
          this._on_sdp(msg);
          break;
      }
    }
  }

  transfer_state(pre_state, state) {
    if (pre_state === this.state) {
      console.log("state: {} => {}".format(pre_state, state));
      this.state = state;
      return true;
    }
    return false;
  }

  on_calling(peer_id) {
    if (this.role === "callee") {
      // 接到呼叫: b:init->ringing, do=>回复ringing消息
      if (this.transfer_state("init", "ringing")) {
        this.peer_id = peer_id;
        let ringing_msg = { type: "ringing", peer_id: peer_id, id: this.id };
        this.send_msg(ringing_msg);
      }
    }
  }

  on_ringing(peer_id) {
    if (this.role === "caller") {
      // 呼叫对象响铃: a:[calling|ringing]->ringing
      if (this.transfer_state("calling", "ringing") || this.transfer_state("ringing", "ringing")) {
        this.ringing_peers[peer_id] = "ringing";
      }
    }
  }

  on_ok(peer_id) {
    let session = this;
    // 1. 呼叫对象接听: a:ringing->ready, a:calling->ready, do=>回复ok消息
    // 2. 呼叫者ready: b:ringing->ready
    if (this.role === 'caller') {
      if (this.transfer_state("ringing", "ready") || this.transfer_state("calling", "ready")) {
        // send ok
        this.peer_id = peer_id;
        let ok_msg = { type: "ok", peer_id: peer_id, id: this.id };
        this.send_msg(ok_msg);

        this.createPeerConnection().then(pc=> {
          session.pc = pc;
          pc.addStream(session.stream);
          pc.createOffer({
            offerToReceiveAudio: 1,
            offerToReceiveVideo: 1
          }).then(::session._setAndSendSdp).catch(::session._createOfferOrAnswerError);
        }).catch(err=> {
          console.log(err);
        });

        // notify other ringing peers.
        delete this.ringing_peers[peer_id];
        let reason = "busy";
        for (let peer_id in this.ringing_peers) {
          let hangup_msg = { type: 'hangup', peer_id: peer_id, id: this.id, reason: reason };
          this.send_msg(hangup_msg);
        }
      }
    } else if (this.role === 'callee') {
    }
  }

  on_hangup(peer_id, reason) {
    // 1. 呼叫对象拒绝: a:ringing->end, a:calling->end
    // 2. 呼叫者取消: b:ringing->end, b:calling->end
    // 3. 挂断: x:ready->end
    var do_end = false;
    if (this.role === 'caller') {
      if (reason === "busy" || reason === "refuse") {
        delete this.ringing_peers[peer_id];
        if (Object.keys(this.ringing_peers).length === 0) {
          do_end = true;
        }
      } else {
        do_end = true;
      }
    } else if (this.role === 'callee') {
      do_end = true;
    }
    if (do_end) {
      if (this.transfer_state(this.state, "end")) {
        this.onhangup(this.state, reason);
      }
    }
  }

  // 呼叫
  calling() {
    if (this.state !== "init") {
      return;
    }

    let session = this;
    let calling_msg = { type: 'calling', user: this.user, id: this.id };
    this.send_msg(calling_msg).then(res=> {
      session.transfer_state("init", "calling");
    }).catch(err=> {
      console.log("calling error:", err);
      // end
      if (this.transfer_state(this.state, "end")) {
        this.onhangup(this.state, "error");
      }
    });
  }

  // 接听
  answer() {
    if (this.state !== "ringing") {
      return;
    }

    let session = this;
    this.createPeerConnection().then(pc=> {
      session.pc = pc;
      pc.addStream(session.stream);
      session.ok(function () {
      }, function () {
      });
    }).catch(err=> {
      console.log(err);
    });
  }

  ok(callback, errCallback) {
    if (this.state !== "ringing") {
      return;
    }

    let ok_msg = { type: 'ok', peer_id: this.peer_id, id: this.id };
    this.send_msg(ok_msg).then((res=> {
      if (this.transfer_state("ringing", "ready")) {
        callback();
      }
    }).bind(this)).catch((err=> {
      console.log("send msg error:", err);
      errCallback();
    }).bind(this));
  }

  hangup(reason) {
    var reason = reason;
    if (!reason) {
      if (this.state === "init") {
        reason = "busy";
      } else if (this.state === "ready") {
        reason = "hangup";
      } else if (this.state === "calling") {
        reason = "cancel";
      } else if (this.state === "ringing") {
        if (this.role === "caller") {
          reason = "cancel";
          for (let peer_id in this.ringing_peers) {
            let hangup_msg = { type: 'hangup', peer_id: peer_id, id: this.id, reason: reason };
            this.send_msg(hangup_msg);
          }
        } else {
          reason = "refuse"
        }
      }
    }

    if (this.peer_id) {
      let session = this;
      let hangup_msg = { type: 'hangup', peer_id: this.peer_id, id: this.id, reason: reason };
      this.send_msg(hangup_msg).then(res=> {
        session.transfer_state(this.state, "end");
      }).catch(err=> {
        console.log("send msg error:", err);
      });
    }
  }

  close() {
    if (this.pc) {
      this.pc.close();
      this.pc = null;
    }
    if (this.stream) {
      this.stream.getTracks().forEach(t=>t.stop());
      this.stream = null;
    }
    if (this.remoteStream) {
      this.remoteStream.getTracks().forEach(t=>t.stop());
      this.remoteStream = null;
    }
  }

  // 发送ice candidate.
  sendIceCandidate(candidate) {
    let msg = { type: 'rtc', sub_type: 'candidate', peer_id: this.peer_id, msg: candidate };
    this.send_msg(msg);
  }

  // 发送session description.
  sendSdp(sdp) {
    let sdp_msg = { type: 'rtc', sub_type: 'sdp', peer_id: this.peer_id, msg: sdp };
    this.send_msg(sdp_msg);
  }

  createPeerConnection() {
    let session = this;
    return new Promise((resolve, reject)=> {
      axios.get('//t.xchat.engdd.com/xrtc/api/iceconfig').then(res=> {
        let iceServers = res.data.iceServers;
        console.log("ice servers: ", iceServers);

        try {
          resolve(session._do_create_pc({ iceServers: iceServers }));
        } catch (e) {
          reject(new Error("Failed to create PeerConnection, exception: {}".format(e)));
        }
      }).catch(err=> {
        console.log("fetch turn servers error: {}".format(err));
        let iceServers = [{
          urls: ["stun:t.turn.engdd.com:3478"]
        }];
        console.log("ice servers: ", iceServers);

        try {
          resolve(session._do_create_pc({ iceServers: iceServers }));
        } catch (e) {
          reject(new Error("Failed to create PeerConnection, exception: {}".format(e)));
        }
      });
    });
  }

  _do_create_pc(iceServers) {
    let session = this;
    let pc = new RTCPeerConnection(iceServers);
    pc.onicecandidate = ::this._on_ice_candidate;
    pc.onsignalingstatechange = function (event) {
      session.onsignalingstatechange(pc.signalingState);
    };

    pc.oniceconnectionstatechange = function (event) {
      session.oniceconnectionstatechange(pc.iceConnectionState);
      if (pc.iceConnectionState === "disconnected") {
        session.onhangup();
      }
    };

    pc.onaddstream = ::this._on_add_stream;
    pc.onremovestream = ::this._on_stream_remove;
    return pc;
  }

  _on_ice_candidate(event) {
    console.log('ice candidate event: ', event);
    if (event.candidate) {
      this.sendIceCandidate(event.candidate);
    } else {
      console.log('End of candidates.');
    }
  }

  _on_add_stream(event) {
    console.log('Remote stream added.');
    this.remoteStream = event.stream;
    this.onaddstream(this.remoteStream);
  }

  _on_stream_remove(event) {
    console.log('Remote stream removed. Event: ', event);
    this.remoteStream = null;
  }

  _setAndSendSdp(sdp) {
    let session = this;
    this.pc.setLocalDescription(sdp, function () {
      console.log('setLocalAndSendMessage sending message', sdp);
      session.sendSdp(sdp);
    })
  }

  _createOfferOrAnswerError(err) {
    console.log('create offer or answer error: ', err);
  }

  _on_candidate(candidate) {
    this.pc.addIceCandidate(new RTCIceCandidate(candidate));
  }

  _on_sdp(sdp) {
    if (sdp.type === "offer") {
      this.pc.setRemoteDescription(new RTCSessionDescription(sdp));
      console.log('Sending answer to peer.');
      this.pc.createAnswer().then(::this._setAndSendSdp, ::this._createOfferOrAnswerError);
    } else if (sdp.type === "answer") {
      this.pc.setRemoteDescription(new RTCSessionDescription(sdp));
    }
  }
}