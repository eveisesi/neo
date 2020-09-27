import { API_URL, WSS_URL } from './util/const/urls'

export default {
  // Global page headers (https://go.nuxtjs.dev/config-head)
  head: {
    title: 'New Eden Obituary',
    meta: [
      { charset: 'utf-8' },
      { name: 'viewport', content: 'width=device-width, initial-scale=1' },
      { hid: 'description', name: 'description', content: 'The New Eden Obituary is a killboard for Eve Online chronicling the tragic deaths of New Eden demi-god capsuleers' }
    ],
    link: [
      { rel: 'icon', type: 'image/x-icon', href: '/favicon.ico' }
    ],
  },

  // Global CSS (https://go.nuxtjs.dev/config-css)
  css: [
  ],

  // Plugins to run before rendering page (https://go.nuxtjs.dev/config-plugins)
  plugins: [
    { src: '~/plugins/vue-bootstrap-typeahead', ssr: false },
    { src: '~/plugins/apollo-ws-client.js', mode: 'client' }
  ],

  // Auto import components (https://go.nuxtjs.dev/config-components)
  components: true,

  // Modules for dev and build (recommended) (https://go.nuxtjs.dev/config-modules)
  buildModules: [
  ],

  // Modules (https://go.nuxtjs.dev/config-modules)
  modules: [
    // https://go.nuxtjs.dev/bootstrap
    'bootstrap-vue/nuxt',
    // https://go.nuxtjs.dev/axios
    '@nuxtjs/axios',
    // https://www.npmjs.com/package/@nuxtjs/apollo
    '@nuxtjs/apollo'
  ],

  // Axios module configuration (https://go.nuxtjs.dev/config-axios)
  axios: {},

  apollo: {
    clientConfigs: {
      default: `~/plugins/vue-apollo-config.js`
    }
  },

  env: {
    apiURL: API_URL,
    wssURL: WSS_URL
  },

  // Build Configuration (https://go.nuxtjs.dev/config-build)
  build: {
  }
}
