#!/usr/bin/env bash
set -euo pipefail

target_file=.github/workflows/generated_workflow.yml
source_files=($(find ci/workflow-parts -name 'part_*'))
base=ci/workflow-parts/index.yaml

cp ${base} ${target_file}

for source_file in "${source_files[@]}"; do
  job_name=${source_file##*/part_}
  job_name=${job_name%%.*}
  yq -P ".jobs.${job_name} |= load(\"${source_file}\")" ${target_file} >${target_file}.new
  mv ${target_file}.new ${target_file}
done
