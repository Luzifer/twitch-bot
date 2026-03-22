import esbuild from 'esbuild'
import { sassPlugin } from 'esbuild-sass-plugin'
import vuePlugin from 'esbuild-plugin-vue3'

const buildOpts = {
  assetNames: '[name]-[hash]',
  bundle: true,
  define: {
    'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'development'),
  },
  entryPoints: ['src/main.ts'],
  legalComments: 'none',
  loader: {
    '.ttf': 'file',
    '.woff2': 'file',
  },
  minify: true,
  outfile: 'editor/app.js',
  plugins: [
    sassPlugin(),
    vuePlugin(),
  ],
  target: [
    'chrome109',
    'edge132',
    'es2020',
    'firefox115',
    'safari16',
  ],
}

export { buildOpts }

console.log(process.env.NODE_ENV || 'development')
esbuild.build(buildOpts)
