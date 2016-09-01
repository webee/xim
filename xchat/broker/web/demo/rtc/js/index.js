import 'webrtc-adapter';
import '../../js/common';
import {XChatClient, autobahn_debug} from '../../js/xchat_client';
import {anyUserkey} from '../../js/configs';
import {decode_ns_user} from '../../js/utils';
import {CallManager} from './call_manager';
import {XChatMsgChannel} from './message_channel';

// init.
autobahn_debug(true);
var xim_state = document.querySelector('#xim_state');
var callingButton = document.querySelector('#callingButton');
callingButton.onclick = calling;


var user = decode_ns_user(document.location.hash.substr(1) || "test:test");
var sToken = document.location.search.substr(1);
var wsuri = (document.location.protocol === "http:" ? "ws:" : "wss:") + "//" + document.location.host + "/ws";


var xchatClient = new XChatClient({
  user, sToken, wsuri,
  key: anyUserkey,
  debug_log: console.log,
  onready: xchatClient => {
    console.log("xim is ready");
    // 可以开始呼叫了!!
    callingButton.disabled = false;
  },
  onerror: err => {
    alert("xim error: {}".format(err))
  },
  onstatechange: state => {
    xim_state.innerText = state;
  },
  onclose: () => {
    console.log("xim is closed");
  }
});

var callManager = new CallManager(new XChatMsgChannel(xchatClient, "xrtc"), { onringing });


var answerButton = document.querySelector('#answerButton');
var hangupButton = document.querySelector('#hangupButton');
answerButton.onclick = answer;
hangupButton.onclick = hangup;


function onringing(sess) {
  if (session !== null) {
    console.log("another call coming");
    sess.hangup("busy");
    callManager.removeSession(sess.id);
    return;
  }
  console.log("calling coming...");

  session = sess;
  session.onhangup = on_hangup;
  session.onaddstream = on_add_stream;

  // states.
  // 呼叫界面消失
  callingButton.disabled = true;
  document.querySelector("#toCallArea").style.display = "none";

  // 通话界面显示
  answerButton.disabled = false;
  hangupButton.disabled = false;
  document.querySelector("#callingArea").style.display = "";
};

function cur_callee() {
  // caller perspective.
  return document.querySelector("#callee").value;
}

var session = null;
var localVideo = document.getElementById('localVideo');
var remoteVideo = document.getElementById('remoteVideo');

localVideo.addEventListener('loadedmetadata', function () {
  console.log('Local video videoWidth: {}px, videoHeight: {}px'.format(this.videoWidth, this.videoHeight));
});

remoteVideo.addEventListener('loadedmetadata', function () {
  console.log('Remote video videoWidth: {}px, videoHeight: {}px'.format(this.videoWidth, this.videoHeight));
});

remoteVideo.onresize = function () {
  console.log('Remote video size changed to {}x{}'.format(remoteVideo.videoWidth, remoteVideo.videoHeight));
};


// 呼叫
function calling() {
  // states.
  callingButton.disabled = true;

  callManager.calling(cur_callee(), { onhangup: on_hangup, onaddstream: on_add_stream }).then(s=> {
    session = s;
    localVideo.srcObject = session.stream;


    // states.
    document.querySelector("#toCallArea").style.display = "none";
    callingButton.disabled = true;

    document.querySelector("#callingArea").style.display = "";
    answerButton.disabled = true;
    hangupButton.disabled = false;
  }).catch(err=> {
    console.error(err);

    // states.
    callingButton.disabled = false;
  });
}


// 接听
function answer() {
  // states.
  answerButton.disabled = true;

  callManager.answer(session).then(s=> {
    localVideo.srcObject = s.stream;
  }).catch(err=> {
    console.log("answer error:", err);
  });
}


// 挂断/取消
function hangup() {
  if (session) {
    session.hangup();

    end();
  }
}

function end() {
  if (session) {
    callManager.removeSession(session);
    session.close();
    session = null;
  }

  localVideo.src = "";
  remoteVideo.src = "";

  // states.
  callingButton.disabled = false;
  document.querySelector("#toCallArea").style.display = "";

  document.querySelector("#callingArea").style.display = "none";
  answerButton.disabled = true;
  hangupButton.disabled = true;
}


function on_add_stream(stream) {
  remoteVideo.srcObject = stream;
}

function on_hangup(state, reason) {
  end();
}
