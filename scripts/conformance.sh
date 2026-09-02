#!/usr/bin/env bash
set -euo pipefail

binary=${1:?bridge binary is required}
repository_root=${2:?repository root is required}
output_dir=${3:?caller-owned output directory is required}

case "$output_dir" in
  "$repository_root"|"$repository_root"/*)
    echo "output must be outside the source repository" >&2
    exit 1
    ;;
esac

mkdir -p "$output_dir"
if [[ -n "$(find "$output_dir" -mindepth 1 -maxdepth 1 -print -quit)" ]]; then
  echo "output directory must be empty" >&2
  exit 1
fi

"$binary" \
  -suite-root "$repository_root/examples" \
  -inventory-root "$repository_root" \
  -output "$output_dir"

"$repository_root/scripts/assert-manifest.sh" "$output_dir/manifest.json"
test -s "$output_dir/dossier.md"
test -s "$output_dir/projected-claims.json"
test -s "$output_dir/mapping-report.json"
test "$(jq -r '.generated_artifact_count' "$output_dir/manifest.json")" = 4
test "$(jq -r '.artifact_digests_exclude_manifest' "$output_dir/manifest.json")" = true
test "$(jq -r '.operational_state' "$output_dir/manifest.json")" = SUCCESSFUL_READ_ONLY_PROJECTION
