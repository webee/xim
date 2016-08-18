function gen_random_id() {
  return Math.random() * 2**53;
}

export class SignalingChannel {
  constructor(role, send_msg, peer_id) {
    this.role = role;
    this.send_msg = send_msg;
    this.state = "init";

    this.id = gen_random_id();
    this.peer_id = peer_id || null;

    this.ringing_peers = {};

    this.pc = null;

    // listeners.
    this.onok = null;
    this.oncandidate = null;
    this.onsdp = null;
    this.onhangup = null;
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
          this.oncandidate(msg);
          break;
        case "sdp":
          this.onsdp(msg);
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
    // 1. 呼叫对象接听: a:ringing->ready, a:calling->ready, do=>回复ok消息
    // 2. 呼叫者ready: b:ringing->ready
    if (this.role === 'caller') {
      if (this.transfer_state("ringing", "ready") || this.transfer_state("calling", "ready")) {
        this.peer_id = peer_id;
        let ok_msg = { type: "ok", peer_id: peer_id, id: this.id };
        this.send_msg(ok_msg);
        this.onok();

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

    let calling_msg = { type: 'calling', id: this.id };
    this.send_msg(calling_msg).then((res=> {
      this.transfer_state("init", "calling");
    }).bind(this)).catch((err=> {
      console.log("calling error:", err);
      // re generate id.
      this.id = gen_random_id();
    }).bind(this));
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
      let hangup_msg = { type: 'hangup', peer_id: this.peer_id, id: this.id, reason: reason };
      this.send_msg(hangup_msg).then((res=> {
        this.transfer_state(this.state, "end");
      }).bind(this)).catch((err=> {
        console.log("send msg error:", err);
      }).bind(this));
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
}