import vuePlugin from 'esbuild-vue'
import esbuild from 'esbuild'

esbuild.build({
  bundle: true,
  define: {
    'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'dev'),
  },
  entryPoints: ['src/main.js'],
  loader: {},
  minify: true,
  outfile: 'editor/app.js',
  plugins: [vuePlugin()],
  target: [
    'chrome87',
    'edge87',
    'es2020',
    'firefox84',
    'safari14',
  ],
})
