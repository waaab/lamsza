#!/usr/bin/env bash
# Regenerate backend/schema/ from a live lamsza database.
#
# backend/schema/*.sql is generated and committed. Generated, because nobody is
# going to keep a 2700-line schema correct by hand and a wrong bootstrap is
# worse than none. Committed, because CI has to create the database from the
# repo and a developer has to be able to read the diff when the schema moves.
#
# Run this after a schema change, and commit the result with the change:
#
#   scripts/db-dump-schema.sh                      # from the local dev database
#   scripts/db-dump-schema.sh --url "$DATABASE_URL"
#
# Two files come out:
#
#   001_schema.sql     structure only - tables, the locations view, the
#                      pg_slugify/sync_* functions, triggers, indexes,
#                      constraints, pg_trgm and unaccent.
#   002_reference.sql  the reference rows the app cannot work without:
#                      counties, geo_locations, settlements,
#                      settlement_location_types, historical_seats,
#                      county_historical_seats, venue_types,
#                      weather_desc_translations, and the directory and
#                      event catalogs (entry_types, entry_categories,
#                      catalog_event_types, catalog_event_subtypes).
#
# REFERENCE_TABLES below is deliberately an allowlist, not an exclusion list.
# users, sessions, admin_sessions, site_settings, entries and everything else
# user- or admin-generated must never land in a committed file: that is
# personal data and, in site_settings, configuration that can hold keys. Adding
# a table here is a decision to publish its contents - check the rows first.

set -euo pipefail

here=$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)
repo=$(cd -- "$here/.." && pwd)
# shellcheck source=lib/pg-url.sh
source "$here/lib/pg-url.sh"

DEFAULT_URL="postgres://lamsza_user:lamsza_password@localhost:5433/lamsza?sslmode=disable"

REFERENCE_TABLES=(
	counties
	geo_locations
	settlements
	settlement_location_types
	historical_seats
	county_historical_seats
	venue_types
	weather_desc_translations
	# The directory and event catalogs. The backend seeds them on boot
	# (MigrateDirectoryCatalog, events.Migrate), but lamsza-admin runs no DDL
	# and no seeding, so without them a bootstrapped database has no
	# categories, entry types or event types for admin to work with. Labels
	# only, no personal data. Both seeds are idempotent: the directory seed
	# runs only on empty tables, the event seed uses ON CONFLICT DO NOTHING.
	entry_types
	entry_categories
	catalog_event_types
	catalog_event_subtypes
)

url=${DATABASE_URL:-$DEFAULT_URL}
container=""

while [[ $# -gt 0 ]]; do
	case $1 in
	--url)
		url=$2
		shift 2
		;;
	--container)
		container=$2
		shift 2
		;;
	-h | --help)
		sed -n '2,29p' "${BASH_SOURCE[0]}" | sed 's/^# \{0,1\}//'
		exit 0
		;;
	*)
		echo "unknown argument: $1" >&2
		exit 2
		;;
	esac
done

pg_url_parse "$url"
runner=$(pg_pick_runner "$container" pg_dump psql)
out="$repo/backend/schema"
mkdir -p "$out"

echo "dumping     $PGUSER@$PGHOST:$PGPORT/$PGDATABASE  (runner: $runner)"

banner() {
	cat <<-EOF
		-- $1
		--
		-- GENERATED FILE - do not edit by hand.
		-- Regenerate with scripts/db-dump-schema.sh and commit the diff.
		-- $2

	EOF
}

{
	banner "backend/schema/001_schema.sql - the whole lamsza schema, structure only." \
		"Apply it with scripts/db-bootstrap.sh, which is what CI uses."
	pg_run "$runner" pg_dump --schema-only --no-owner --no-acl --no-comments
} >"$out/001_schema.sql"

{
	banner "backend/schema/002_reference.sql - reference rows the app needs to work." \
		"Allowlisted tables only; see scripts/db-dump-schema.sh for why."
	# --column-inserts so the file survives a column being added or reordered,
	# and so a reviewer can read the diff. 2>/dev/null drops pg_dump's warning
	# about the self-referencing settlements.parent_id and
	# entry_categories.parent_id FKs; the rows load fine because the foreign
	# keys are checked at the end of each INSERT statement, and a table of up
	# to 100 rows arrives in one statement. A larger self-referencing table
	# would need its parents dumped first.
	pg_run "$runner" pg_dump --data-only --no-owner --no-acl \
		--column-inserts --rows-per-insert=100 \
		"${REFERENCE_TABLES[@]/#/--table=public.}" 2>/dev/null
} >"$out/002_reference.sql"

wc -l "$out"/001_schema.sql "$out"/002_reference.sql
