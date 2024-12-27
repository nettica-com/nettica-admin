const TOKEN_KEY = "token";
const CLIENT_ID_KEY = "client_id";
const SERVER_KEY = "server";
const WILD_TOKEN_KEY = "wild_token";
const WILD_SERVER_KEY = "wild_server";
const REFERER_KEY = "referer";
const STATE_KEY = "state";
const CODE_KEY = "code";
const REDIRECT_KEY = "redirect_uri";

export const getRedirect = () => {
  return window.localStorage.getItem(REDIRECT_KEY);
};

export const saveRedirect = redirect => {
  window.localStorage.setItem(REDIRECT_KEY, redirect);
};

export const destroyRedirect = () => {
  window.localStorage.removeItem(REDIRECT_KEY);
};

export const getReferer = () => {
  return window.localStorage.getItem(REFERER_KEY);
};

export const saveReferer = referer => {
  window.localStorage.setItem(REFERER_KEY, referer);
};

export const destroyReferer = () => {
  window.localStorage.removeItem(REFERER_KEY);
};

export const getState = () => {
  return window.localStorage.getItem(STATE_KEY);
};

export const saveState = state => {
  window.localStorage.setItem(STATE_KEY, state);
};

export const destroyState = () => {
  window.localStorage.removeItem(STATE_KEY);
};

export const getCode = () => {
  return window.localStorage.getItem(CODE_KEY);
};

export const saveCode = code => {
  window.localStorage.setItem(CODE_KEY, code);
};

export const destroyCode = () => {
  window.localStorage.removeItem(CODE_KEY);
};

export const getToken = () => {
  return window.localStorage.getItem(TOKEN_KEY);
};

export const getWildToken = () => {
  return window.localStorage.getItem(WILD_TOKEN_KEY);
};

export const saveToken = token => {
  window.localStorage.setItem(TOKEN_KEY, token);
};

export const saveWildToken = token => {
  window.localStorage.setItem(WILD_TOKEN_KEY, token);
};

export const destroyToken = () => {
  window.localStorage.removeItem(TOKEN_KEY);
};

export const destroyWildToken = () => {
  window.localStorage.removeItem(WILD_TOKEN_KEY);
};

export const getServer = () => {
  return window.localStorage.getItem(SERVER_KEY);
};

export const saveServer = server => {
  window.localStorage.setItem(SERVER_KEY, server);
};

export const destroyServer = () => {
  window.localStorage.removeItem(SERVER_KEY);
};

export const getWildServer = () => {
  return window.localStorage.getItem(WILD_SERVER_KEY);
};

export const saveWildServer = server => {
  window.localStorage.setItem(WILD_SERVER_KEY, server);
};

export const destroyWildServer = () => {
  window.localStorage.removeItem(WILD_SERVER_KEY);
};

export const getClientId = () => {
  return window.localStorage.getItem(CLIENT_ID_KEY);
};

export const saveClientId = token => {
  console.log("saveClientId = ", token );
  window.localStorage.setItem(CLIENT_ID_KEY, token);
};

export const destroyClientId = () => {
  window.localStorage.removeItem(CLIENT_ID_KEY);
};

export default {
  getRedirect,
  saveRedirect,
  destroyRedirect,
  getToken,
  saveToken,
  destroyToken,
  getClientId,
  saveClientId,
  destroyClientId,
  getServer,
  saveServer,
  destroyServer,
  getWildToken,
  saveWildToken,
  destroyWildToken,
  getWildServer,
  saveWildServer,
  destroyWildServer,
  getReferer,
  saveReferer,
  destroyReferer,
  getState,
  saveState,
  destroyState,
  getCode,
  saveCode,
  destroyCode
};
