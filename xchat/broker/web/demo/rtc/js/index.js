import 'webrtc-adapter';
import axios from 'axios';
import {XChatClient, autobahn_debug} from '../../js/xchat_client';
import {anyUserkey} from '../../js/configs';
import {decode_ns_user} from '../../js/utils';
import {trace, trace_objs} from '../../js/common';
import {RTCManager} from './rtc_manager';

// init.
autobahn_debug(true);

var user = decode_ns_user(document.location.hash.substr(1) || "test:test");
var sToken = document.location.search.substr(1);
var wsuri = (document.location.protocol === "http:" ? "ws:" : "wss:") + "//" + document.location.host + "/ws";


var xchatClient = new XChatClient(user, sToken, wsuri, anyUserkey);
xchatClient.onready = function () {
  if (!cur_chat_id) {
    fetchUserChat();
  }
};
xchatClient.open();

var rtcManager = new RTCManager(xchatClient);


var callingButton = document.querySelector('#callingButton');

var answerButton = document.querySelector('#answerButton');
var hangupButton = document.querySelector('#hangupButton');

callingButton.onclick = calling;
answerButton.onclick = answer;
hangupButton.onclick = hangup;


rtcManager.oncalling = function (signaling_channel) {
  if (signalingChannel !== null) {
    console.log("another call coming");
    signaling_channel.hangup("busy");
    rtcManager.removeSignalChannel(signaling_channel.id);
    return;
  }
  console.log("calling coming...");

  signalingChannel = signaling_channel;
  signalingChannel.oncandidate = on_candidate;
  signalingChannel.onsdp = on_sdp_offer;
  signalingChannel.onhangup = on_caller_hangup;

  // states.
  // 呼叫界面消失
  callingButton.disabled = true;
  document.querySelector("#toCallArea").style.display = "none";

  // 通话界面显示
  answerButton.disabled = false;
  hangupButton.disabled = false;
  document.querySelector("#callingArea").style.display = "";
};

// caller perspective.
var cur_callee = document.querySelector("#callee").value;
var cur_chat_id = null;

window.fetchUserChat = function fetchUserChat() {
  var callee = document.querySelector("#callee").value;
  xchatClient.call("xchat.user.chat.new", ['user', [callee], ''], { is_ns_user: true }).then(res=> {
    trace_objs("res>", res);
    var ret = res.args[0];
    if (!ret) {
      console.error("error:", res.args[1]);
      document.querySelector("#callee").value = cur_callee;
      return
    }
    var chat = res.args[1];
    cur_callee = callee;
    cur_chat_id = chat.id;
    document.querySelector("#chat_id").innerText = chat.id;
    callingButton.disabled = false;
  }).catch(err=> {
    console.error("error:", err);
  });
};

var signalingChannel = null;
var localStream;
var remoteStream;
var pc;
var offerOptions = {
  offerToReceiveAudio: 1,
  offerToReceiveVideo: 1
};

var constraints = {
  audio: true,
  video: true
};

var sdpConstraints = {
  'mandatory': {
    'OfferToReceiveAudio': true,
    'OfferToReceiveVideo': true
  }
};


var localVideo = document.getElementById('localVideo');
var remoteVideo = document.getElementById('remoteVideo');

localVideo.addEventListener('loadedmetadata', function () {
  trace('Local video videoWidth: ' + this.videoWidth +
    'px,  videoHeight: ' + this.videoHeight + 'px');
});

remoteVideo.addEventListener('loadedmetadata', function () {
  trace('Remote video videoWidth: ' + this.videoWidth +
    'px,  videoHeight: ' + this.videoHeight + 'px');
});

remoteVideo.onresize = function () {
  trace('Remote video size changed to ' + remoteVideo.videoWidth + 'x' + remoteVideo.videoHeight);
};


// 呼叫
function calling() {
  // states.
  document.querySelector("#toCallArea").style.display = "none";
  callingButton.disabled = true;

  document.querySelector("#callingArea").style.display = "";
  answerButton.disabled = true;
  hangupButton.disabled = true;

  console.log("getUserMedia:", constraints);
  navigator.mediaDevices.getUserMedia(constraints)
    .then(function (stream) {
      console.log('Received local stream');
      localVideo.srcObject = stream;
      localStream = stream;

      signalingChannel = rtcManager.calling(cur_chat_id);
      signalingChannel.onok = on_callee_ok;
      signalingChannel.oncandidate = on_candidate;
      signalingChannel.onsdp = on_sdp_answer;
      signalingChannel.onhangup = on_callee_hangup;

      // states.
      hangupButton.disabled = false;
    })
    .catch(function (e) {
      console.log('getUserMedia() error: ' + e.name);
    });
}


// 接听
function answer() {
  answerButton.disabled = true;
  navigator.mediaDevices.getUserMedia(constraints)
    .then(function (stream) {
      localVideo.srcObject = stream;
      localStream = stream;
      createPeerConnection(pc=> {
        pc.addStream(localStream);
        signalingChannel.ok(function () {
        }, function () {
        });
      });
    })
    .catch(function (e) {
      console.log('getUserMedia() error: ' + e.name);
      answerButton.disabled = false;
    });
}


// 挂断/取消
function hangup() {
  signalingChannel.hangup();

  end();
}

function end() {
  rtcManager.removeSignalChannel(signalingChannel);
  signalingChannel = null;

  if (pc) {
    pc.close();
    pc = null;
  }
  if (localStream) {
    localStream.getTracks().forEach(t=>t.stop());
    localStream = null;
  }
  if (remoteStream) {
    remoteStream.getTracks().forEach(t=>t.stop());
    remoteStream = null;
  }

  localVideo.src = "";
  remoteVideo.src = "";

  callingButton.disabled = false;
  document.querySelector("#toCallArea").style.display = "";

  document.querySelector("#callingArea").style.display = "none";
  answerButton.disabled = true;
  hangupButton.disabled = true;
}


function on_callee_ok() {
  createPeerConnection(pc=> {
    pc.addStream(localStream);
    pc.createOffer(offerOptions).then(setLocalAndSendMessage).catch(handleCreateOfferOrAnswerError);
  });
}

function on_caller_hangup(state, reason) {
  end();
}

function on_callee_hangup(state, reason) {
  end();
}

function on_candidate(candidate) {
  pc.addIceCandidate(new RTCIceCandidate(candidate));
}

function on_sdp_offer(sdp) {
  if (sdp.type === "offer") {
    pc.setRemoteDescription(new RTCSessionDescription(sdp));
    console.log('Sending answer to peer.');
    pc.createAnswer().then(setLocalAndSendMessage, handleCreateOfferOrAnswerError);
  }
}

function on_sdp_answer(sdp) {
  if (sdp.type === "answer") {
    pc.setRemoteDescription(new RTCSessionDescription(sdp));
  }
}


function createPeerConnection(callback) {
  axios.get('//t.xchat.engdd.com/xrtc/api/iceconfig').then(res=> {
    let iceServers = res.data.iceServers;
    console.log("ice servers: ", iceServers);
    pc = doCreatePeerConnection({ iceServers: iceServers });
    callback(pc);
  }).catch(err=> {
    console.log("fetch turn servers error:", err);
  });
}

function doCreatePeerConnection(iceServers) {
  try {
    let pc = new RTCPeerConnection(iceServers);
    pc.onicecandidate = handleIceCandidate;
    pc.onnegotiationneeded = undefined;
    pc.onnegotiationneeded = function () {
    };

    pc.onsignalingstatechange = function (event) {
      document.querySelector("#state").innerText = pc.signalingState;
    };
    pc.onaddstream = handleRemoteStreamAdded;
    pc.onremovestream = handleRemoteStreamRemoved;
    console.log('Created RTCPeerConnection');
    return pc;
  } catch (e) {
    console.log('Failed to create PeerConnection, exception: ' + e.message);
    return null;
  }
}


function handleIceCandidate(event) {
  console.log('ice candidate event: ', event);
  if (event.candidate) {
    signalingChannel.sendIceCandidate(event.candidate);
  } else {
    console.log('End of candidates.');
  }
}

function handleRemoteStreamAdded(event) {
  console.log('Remote stream added.');
  remoteVideo.srcObject = event.stream;
  remoteStream = event.stream;
}

function handleRemoteStreamRemoved(event) {
  console.log('Remote stream removed. Event: ', event);
}

function setLocalAndSendMessage(sdp) {
  // Set Opus as the preferred codec in SDP if Opus is present.
  //  sessionDescription.sdp = preferOpus(sessionDescription.sdp);
  pc.setLocalDescription(sdp, function () {
    console.log('setLocalAndSendMessage sending message', sdp);
    signalingChannel.sendSdp(sdp);
  });
}

function handleCreateOfferOrAnswerError(event) {
  console.log('create offer or answer error: ', event);
}




