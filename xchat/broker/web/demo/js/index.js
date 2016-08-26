import './common';
import {XChatClient, autobahn_debug} from './xchat_client';
import {anyUserkey} from './configs';
import {decode_ns_user} from './utils';
import {DemoUtils} from './demo_utils';

// init.
autobahn_debug(true);
var xim_state = document.querySelector('#xim_state');

var user = decode_ns_user(document.location.hash.substr(1) || "test:test");
var sToken = document.location.search.substr(1);
var wsuri = (document.location.protocol === "http:" ? "ws:" : "wss:") + "//" + document.location.host + "/ws";


var xchatClient = new XChatClient({ user, sToken, wsuri,
  key: anyUserkey,
  debug_log: console.log,
  onmsg: (kind, msg)=> {
      window.demo.newMsg(kind, msg);
  },
  onready: (xchatClient)=> {
    console.log("xim is ready");
    window.demo = new DemoUtils(xchatClient);
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
