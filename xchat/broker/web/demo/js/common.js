import format from 'string-format';

format.extend(String.prototype);

export function trace(s, ...args) {
  let text = s.format(...args);
  // This function is used for logging.
  if (text[text.length - 1] === '\n') {
    text = text.substring(0, text.length - 1);
  }
  if (window.performance) {
    var now = (window.performance.now() / 1000).toFixed(3);
    console.log(now + ': ' + text);
  } else {
    console.log(text);
  }
}

export function trace_objs(obj, ...args) {
  if (window.performance) {
    var now = (window.performance.now() / 1000).toFixed(3);
    console.log(now + ': ', obj, ...args);
  } else {
    console.log(text);
  }
}
