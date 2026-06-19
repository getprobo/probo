#!/bin/sh
# Merge split GraphQL schema files into a single file for Relay.
# Strips "type Mutation" and "extend type Mutation { ... }" blocks,
# then appends a single "type Mutation { ... }" with all collected fields.
#
# Usage: merge-graphql-schema.sh <output> <graphql-dir>

set -eu

output="$1"
graphql_dir="$2"
base="$graphql_dir/base.graphql"

mutation_fields=$(mktemp)
schema_body=$(mktemp)
trap 'rm -f "$mutation_fields" "$schema_body"' EXIT

process_file() {
    awk -v mf="$mutation_fields" '
    /^type Mutation$/ { next }
    /^(extend )?type Mutation \{/ { skip=1; depth=1; next }
    skip {
        if (/\{/) depth++
        if (/\}/) { depth--; if (depth==0) { skip=0; next } }
        if (skip) { print >> mf; next }
    }
    { print }
    ' "$1"
}

{
    process_file "$base"
    for f in "$graphql_dir"/*.graphql; do
        [ "$f" = "$base" ] && continue
        process_file "$f"
    done
} > "$schema_body"

{
    cat "$schema_body"
    printf '\ntype Mutation {\n'
    cat "$mutation_fields"
    printf '}\n'
} > "$output"
