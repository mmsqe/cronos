#!/usr/bin/env bash

set -eo pipefail

mkdir -p ./tmp-swagger-gen

cd proto
echo "Generate cronos swagger files"
proto_dirs=$(find ./ -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    echo "$query_file"
    buf generate --template buf.gen.swagger.yaml "$query_file"
  fi
done

cd ../third_party/proto
echo "Generate cosmos swagger files"

proto_dirs=$(find ./ethermint ./ibc -path -prune -o -name '*.proto' -print0 | xargs -0 -n1 dirname | sort | uniq)
for dir in $proto_dirs; do
  # generate swagger files (filter query files)
  query_file=$(find "${dir}" -maxdepth 1 \( -name 'query.proto' -o -name 'service.proto' \))
  if [[ ! -z "$query_file" ]]; then
    echo "$query_file"
    buf generate --template buf.gen.swagger.yaml "$query_file"
  fi
done

buf generate --template buf.gen.swagger.yaml "buf.build/cosmos/cosmos-sdk:8cb30a2c4de74dc9bd8d260b1e75e176"

cd ../..

echo "Combine swagger files"
# combine swagger files
# uses nodejs package `swagger-combine`.
# all the individual swagger files need to be configured in `config.json` for merging
swagger-combine ./client/docs/config.json -o ./client/docs/swagger-ui/swagger.yaml -f yaml --continueOnConflictingPaths true --includeDefinitions true

# clean swagger files
rm -rf ./tmp-swagger-gen
