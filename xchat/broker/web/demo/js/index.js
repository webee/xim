import {XChatClient, autobahn_debug} from './xchat_client';
import {anyUserkey} from './configs';
import {decode_ns_user} from './utils';
import {trace, trace_objs} from './common';
import {DemoUtils} from './demo_utils';

// init.
autobahn_debug(true);

var user = decode_ns_user(document.location.hash.substr(1) || "test:test");
var sToken = document.location.search.substr(1);
var wsuri = (document.location.protocol === "http:" ? "ws:" : "wss:") + "//" + document.location.host + "/ws";


var xchatClient = new XChatClient(user, sToken, wsuri, anyUserkey);

xchatClient.onready = function () {
  window.demo = new DemoUtils(xchatClient);
};

xchatClient.onmsg = function (kind, msgs) {
  for (let i = 0; i < msgs.length; i++) {
    window.demo.newMsg(kind, msgs[i]);
  }
};

xchatClient.open();
