import { basename, dirname, extname, join } from 'node:path'
import { copyFile, mkdtemp, readdir, readFile, rm, writeFile } from 'node:fs/promises'
import esbuild from 'esbuild'
import { execFileSync } from 'node:child_process'
import { fileURLToPath } from 'node:url'
import { tmpdir } from 'node:os'
import vuePlugin from 'esbuild-plugin-vue3'

const browserTargets = [
  'chrome109',
  'edge132',
  'es2020',
  'firefox115',
  'safari16',
]
const overlayDir = dirname(fileURLToPath(import.meta.url))
const outputDir = join(overlayDir, 'default')
const sourceDir = join(overlayDir, 'src')

const sourceFiles = (await readdir(sourceDir))
  .filter(file => ['.ts', '.vue'].includes(extname(file)) && !file.endsWith('.d.ts'))
  .sort()
const temporaryDir = await mkdtemp(join(tmpdir(), 'twitch-bot-overlays-'))

try {
  await esbuild.build({
    absWorkingDir: overlayDir,
    bundle: true,
    define: {
      'process.env.NODE_ENV': JSON.stringify(process.env.NODE_ENV || 'production'),
    },
    entryNames: '[name]',
    entryPoints: sourceFiles.map(file => join('src', file)),
    external: [
      './eventclient.js',
      './eventfeed.custom.js',
    ],
    format: 'esm',
    legalComments: 'none',
    minify: true,
    outdir: temporaryDir,
    plugins: [vuePlugin()],
    sourcemap: false,
    target: browserTargets,
  })

  for (const sourceFile of sourceFiles.filter(file => extname(file) === '.vue')) {
    const name = basename(sourceFile, '.vue')
    const source = await readFile(join(sourceDir, sourceFile), 'utf8')
    const title = source.match(/<overlay-title>([^<]+)<\/overlay-title>/)?.[1] || name
    const cssFile = join(temporaryDir, `${name}.css`)

    try {
      await readFile(cssFile)
    } catch (err) {
      if (!isErrnoException(err) || err.code !== 'ENOENT') {
        throw err
      }
      await writeFile(cssFile, '')
    }

    await writeFile(join(temporaryDir, `${name}.html`), `<!doctype html>
<html data-bs-theme="dark">
  <head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1">
    <title>${title}</title>
    <link rel="stylesheet" href="${name}.css">
  </head>
  <body>
    <div id="app"></div>
    <script src="${name}.js" type="module"></script>
  </body>
</html>
`)
  }

  execFileSync('pnpm', [
    'exec',
    'tsc',
    '--declaration',
    '--emitDeclarationOnly',
    '--ignoreConfig',
    '--lib',
    'DOM,ES2020',
    '--module',
    'ESNext',
    '--moduleResolution',
    'Bundler',
    '--outDir',
    temporaryDir,
    '--target',
    'ES2020',
    ...sourceFiles.filter(file => extname(file) === '.ts').map(file => join(sourceDir, file)),
  ], { cwd: overlayDir, stdio: 'inherit' })

  for (const outputFile of (await readdir(temporaryDir)).sort()) {
    await copyFile(join(temporaryDir, outputFile), join(outputDir, outputFile))
  }
} finally {
  await rm(temporaryDir, { force: true, recursive: true })
}

function isErrnoException(err: unknown): err is Error & { code?: string } {
  return err instanceof Error && 'code' in err
}
