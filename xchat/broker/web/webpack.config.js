var webpack = require('webpack');
var path = require('path');
var env = process.env.NODE_ENV || "development";

config = {
  entry: {
    index: './demo/js/index.js',
    rtc: [
      './demo/rtc/js/index.js'
    ]
  },
  output: {
    path: path.join(__dirname, 'demo/dist'),
    filename: '[name].bundle.js',
    publicPath: '/demo/dist/'
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

if (env === "production") {
  config.plugins = [
    new webpack.optimize.OccurenceOrderPlugin(),
    new webpack.DefinePlugin({
      'process.env': {
        'NODE_ENV': env
      }
    }),
    new webpack.optimize.UglifyJsPlugin({
      compressor: {
        warnings: false
      }
    })
  ]
}

module.exports = config;
