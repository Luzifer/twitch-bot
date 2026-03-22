# Install Node deps on change of package.json
local_resource(
  'npm',
  cmd='npm i --include=dev',
  deps=['package.json'],
)

local_resource(
  'frontend',
  cmd='make frontend',
  deps=['src'],
  resource_deps=['npm'],
)

local_resource(
  'go-sum',
  cmd='go mod tidy',
  deps=['go.mod'],
)

local_resource(
  'server',
  deps=[
    'go.mod',
    'internal',
    'main.go',
    'pkg',
    'plugins',
  ],
  ignore=[
    'ci',
    'docs',
    'src',
    'tools',
  ],
  serve_cmd='go run -tags dev . --log-level=debug -c config.yaml',
  readiness_probe=probe(
    http_get=http_get_action(3000, path='/selfcheck'),
    initial_delay_secs=1,
  ),
  resource_deps=[
    'frontend',
    'go-sum',
  ],
)
