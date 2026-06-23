#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
tmp_dir="$(mktemp -d)"

cleanup() {
  if [[ "${SK_SMOKE_KEEP_TMP:-}" == "1" ]]; then
    printf 'Keeping smoke directory: %s\n' "$tmp_dir"
    return
  fi
  rm -rf "$tmp_dir"
}
trap cleanup EXIT

bin="${SK_SMOKE_BIN:-$tmp_dir/sk}"
home="$tmp_dir/home"
cache="$tmp_dir/cache"

mkdir -p "$home" "$cache"

if [[ -z "${SK_SMOKE_BIN:-}" ]]; then
  (cd "$repo_root" && go build -o "$bin" ./main.go)
fi

run_sk() {
  printf '\n$ sk'
  printf ' %q' "$@"
  printf '\n'
  HOME="$home" XDG_CACHE_HOME="$cache" "$bin" "$@"
}

printf 'Smoke HOME: %s\n' "$home"
printf 'Smoke cache: %s\n' "$cache"

run_sk --help >/dev/null
run_sk doctor
run_sk doctor --registry
run_sk search testing
run_sk search --category testing
run_sk install docx --name smoke-docx
run_sk doctor

printf 'Registry smoke passed without mutating the real HOME.\n'
