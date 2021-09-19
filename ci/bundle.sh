#!/bin/bash
set -euxo pipefail

outfile=${1:-}
[[ -n $outfile ]] || {
	echo "Missing parameters: $0 <outfile> [libraries]"
	exit 1
}
shift

libs=("$@")

IFS=$','
exec curl -sSfLo "${outfile}" "https://cdn.jsdelivr.net/combine/${libs[*]}"
