module.exports = {
  devServer: {
    client : { webSocketURL : 'https://dev.nettica.com' },
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
