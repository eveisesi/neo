// Urls
const EVEONLINE_IMAGE = "https://images.evetech.net/";

const API_BASE = process.env.NUXT_APP_API_BASE
const API_URL = `http://${API_BASE}/query`
const WSS_URL = `ws://${API_BASE}/query`


export { EVEONLINE_IMAGE, API_URL, WSS_URL };
