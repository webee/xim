var path = require('path');

module.exports = {
  entry: {
    index: './demo/js/index.js',
    rtc: [
      './demo/rtc/js/index.js'
    ]
  },
  output: {
    path: path.join(__dirname, 'demo/dist'),
    filename: '[name].bundle.js'
  },
  module: {
    noParse: /node_modules\/autobahn\/autobahn.js/,
    loaders: [
      { test: /\.js$/, exclude: /node_modules/, loader: 'babel' },
      { test: /\.json$/, loader: 'json' }
    ]
  },
  node: {
    fs: "empty",
    tls: "empty"
  }
};
