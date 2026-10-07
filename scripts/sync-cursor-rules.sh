#!/usr/bin/env bash
# Copy the network-wide Cursor rules from this repo to the other Lámsza apps.
#
# lamsza owns these files (WAYS_OF_WORKING R11). A copy edited in place, or
# replaced in a bulk edit, is how szotar and jatszoter lost the UI-consistency
# rule on 2026-10-06. App-specific rules (lamsza's dialog-theme.mdc,
# jatszoter's loading-state.mdc) are not touched.
#
#   scripts/sync-cursor-rules.sh          # copy, then commit in each repo
#   scripts/sync-cursor-rules.sh --check  # exit 1 if any copy differs
#
# The repos are expected next to this one; set LAMSZA_PROJECTS_ROOT otherwise.

set -euo pipefail

here=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd -- "$here/.." && pwd)
root=${LAMSZA_PROJECTS_ROOT:-$(cd -- "$repo/.." && pwd)}

RULES=(lamsza-network.mdc no-emdash.mdc)
APPS=(lamsza-admin lamsza-szotar lamsza-jatszoter)

check=0
if [[ ${1:-} == --check ]]; then
	check=1
fi

drift=0
for app in "${APPS[@]}"; do
	dest="$root/$app/.cursor/rules"
	if [[ ! -d $root/$app ]]; then
		echo "missing repo: $root/$app" >&2
		exit 2
	fi
	for rule in "${RULES[@]}"; do
		src="$repo/.cursor/rules/$rule"
		if [[ $check -eq 1 ]]; then
			if ! cmp -s "$src" "$dest/$rule"; then
				echo "differs: $app/.cursor/rules/$rule"
				drift=1
			fi
		else
			mkdir -p "$dest"
			cp "$src" "$dest/$rule"
		fi
	done
done

if [[ $check -eq 1 ]]; then
	if [[ $drift -eq 1 ]]; then
		echo "cursor rules out of step; run scripts/sync-cursor-rules.sh" >&2
		exit 1
	fi
	echo "cursor rules: in step"
else
	echo "synced ${#RULES[@]} rules into ${APPS[*]}. Commit each repo."
fi
