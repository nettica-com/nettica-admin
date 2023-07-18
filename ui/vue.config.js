module.exports = {
  devServer: {
    client : { webSocketURL : 'https://dev.meshify.app' },
    allowedHosts: 'all',
    port: 8081,
  },
  "transpileDependencies": [
    "vuetify"
  ],
  configureWebpack: {
    devtool: 'source-map'
  }  
};
