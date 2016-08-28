export class SignalingChannel {
  constructor(peer, msgChannel) {
    this.peer = peer;
    this.msgChannel = msgChannel;

    this.onrecv = () => {};
  }

  send(msg) {
    return this.msgChannel.send(this.peer, msg);
  }

  recv(msg) {
    this.onrecv(msg);
  }
}