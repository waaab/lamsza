#!/usr/bin/env bash
#
# Tests scripts/sync-shared-frontend.sh: per-app shares from the "consumers"
# map, the generated manifests, and what --check catches.
#
# Every case builds a throwaway projects root in a temp dir (a lamsza copy of
# the script plus three app folders) and never touches the real repos.
#
#   scripts/tests/sync-shared-frontend.test.sh

set -uo pipefail

HERE="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SCRIPT="$HERE/../sync-shared-frontend.sh"
[ -f "$SCRIPT" ] || { echo "missing: $SCRIPT" >&2; exit 2; }

PASS=0
FAIL=0
OUT=""
CODE=0

# A fake projects root: lamsza with three modules, a manifest that gives admin
# all of them and szotar / jatszoter only the first, and three empty apps.
make_tree() {
	local dir
	dir="$(mktemp -d)"
	mkdir -p "$dir/lamsza/scripts" "$dir/lamsza/src/lib" "$dir/lamsza/src/styles"
	cp "$SCRIPT" "$dir/lamsza/scripts/"
	echo "export const a = 1;" > "$dir/lamsza/src/lib/a.js"
	echo "export const b = 2;" > "$dir/lamsza/src/lib/b.js"
	echo "body { margin: 0 }" > "$dir/lamsza/src/styles/global.css"
	cat > "$dir/lamsza/shared-frontend-modules.json" <<-'JSON'
	{
	  "consumers": {
	    "lamsza-admin": "all",
	    "lamsza-szotar": ["src/lib/a.js"],
	    "lamsza-jatszoter": ["src/lib/a.js", "src/styles/global.css"]
	  },
	  "modules": {
	    "src/lib/a.js": "sha256:x",
	    "src/lib/b.js": "sha256:x",
	    "src/styles/global.css": "sha256:x"
	  }
	}
	JSON
	for app in lamsza-admin lamsza-szotar lamsza-jatszoter; do
		mkdir -p "$dir/$app/frontend/src"
	done
	printf '%s' "$dir"
}

run() {
	local dir="$1"
	shift
	OUT="$(env -u LAMSZA_ADMIN_ROOT -u LAMSZA_SZOTAR_ROOT -u LAMSZA_JATSZOTER_ROOT \
		"$@" bash "$dir/lamsza/scripts/sync-shared-frontend.sh" ${ARGS:-} 2>&1)"
	CODE=$?
}

ok() {
	if eval "$2"; then
		PASS=$((PASS + 1))
	else
		FAIL=$((FAIL + 1))
		echo "FAIL: $1"
		echo "  output: $OUT"
	fi
}

# 1. Each app gets exactly its share, and its own manifest lists only that.
t="$(make_tree)"
ARGS="" run "$t"
ok "sync exits 0" '[ "$CODE" = 0 ]'
ok "admin gets every module" '[ -f "$t/lamsza-admin/frontend/src/lib/b.js" ] && [ -f "$t/lamsza-admin/frontend/src/styles/global.css" ]'
ok "szotar gets only its module" '[ -f "$t/lamsza-szotar/frontend/src/lib/a.js" ] && [ ! -e "$t/lamsza-szotar/frontend/src/lib/b.js" ]'
ok "jatszoter gets global.css" '[ -f "$t/lamsza-jatszoter/frontend/src/styles/global.css" ]'
ok "szotar manifest lists one module" \
	'[ "$(node -e "console.log(Object.keys(require(process.argv[1]).modules).join(\",\"))" "$t/lamsza-szotar/frontend/shared-frontend-modules.json")" = "src/lib/a.js" ]'
ok "lamsza manifest keeps the consumers map" \
	'node -e "process.exit(require(process.argv[1]).consumers[\"lamsza-szotar\"].length === 1 ? 0 : 1)" "$t/lamsza/shared-frontend-modules.json"'
ok "hashes are real" '! grep -q "sha256:x" "$t/lamsza/shared-frontend-modules.json"'
ARGS="--check" run "$t"
ok "--check passes right after a sync" '[ "$CODE" = 0 ] && [[ "$OUT" == *"in step"* ]]'

# 2. A hand edit in one app is caught and named.
echo "// local edit" >> "$t/lamsza-jatszoter/frontend/src/styles/global.css"
ARGS="--check" run "$t"
ok "--check fails on drift" '[ "$CODE" != 0 ]'
ok "drift names the app" '[[ "$OUT" == *"DRIFT: src/styles/global.css differs between lamsza and lamsza-jatszoter"* ]]'
rm -rf "$t"

# 3. An owner edit without a sync leaves the apps' manifests stale.
t="$(make_tree)"
ARGS="" run "$t"
echo "export const a = 3;" > "$t/lamsza/src/lib/a.js"
cp "$t/lamsza/src/lib/a.js" "$t/lamsza-szotar/frontend/src/lib/a.js"
cp "$t/lamsza/src/lib/a.js" "$t/lamsza-jatszoter/frontend/src/lib/a.js"
cp "$t/lamsza/src/lib/a.js" "$t/lamsza-admin/frontend/src/lib/a.js"
ARGS="--check" run "$t"
ok "--check fails on a stale app manifest" '[ "$CODE" != 0 ] && [[ "$OUT" == *"STALE: lamsza-szotar"* ]]'
rm -rf "$t"

# 4. A consumer list naming an unknown module is refused before anything is copied.
t="$(make_tree)"
node -e '
  const fs = require("fs"); const p = process.argv[1];
  const m = JSON.parse(fs.readFileSync(p)); m.consumers["lamsza-szotar"].push("src/lib/nope.js");
  fs.writeFileSync(p, JSON.stringify(m));
' "$t/lamsza/shared-frontend-modules.json"
ARGS="" run "$t"
ok "unknown module is refused" '[ "$CODE" != 0 ] && [[ "$OUT" == *"src/lib/nope.js"* ]]'
ok "nothing copied" '[ ! -e "$t/lamsza-admin/frontend/src/lib/a.js" ]'
rm -rf "$t"

# 5. A missing app stops the run; a LAMSZA_*_ROOT override finds it elsewhere.
t="$(make_tree)"
mv "$t/lamsza-szotar" "$t/elsewhere"
ARGS="" run "$t"
ok "missing app stops the run" '[ "$CODE" != 0 ] && [[ "$OUT" == *"lamsza-szotar frontend not found"* ]]'
ARGS="" run "$t" LAMSZA_SZOTAR_ROOT="$t/elsewhere"
ok "LAMSZA_SZOTAR_ROOT override works" '[ "$CODE" = 0 ] && [ -f "$t/elsewhere/frontend/src/lib/a.js" ]'
rm -rf "$t"

echo "sync-shared-frontend: $PASS passed, $FAIL failed"
[ "$FAIL" = 0 ]
