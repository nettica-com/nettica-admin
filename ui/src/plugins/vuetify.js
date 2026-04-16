import 'vuetify/styles'
import { createVuetify } from 'vuetify'
import * as components from 'vuetify/components'
import * as directives from 'vuetify/directives'


export default createVuetify({
  components: {
    ...components,
  },
  directives,
  icons: {
    defaultSet: 'mdi',
  },
  theme: {
    defaultTheme: 'light',
    themes: {
      light: {
        colors: {
          primary: '#336699',
          success: '#004000',
          error: '#400000',
        },
      },
      dark: {
        colors: {
          primary: '#336699',
          success: '#004000',
          error: '#400000',
        },
      },
    },
  },
})
