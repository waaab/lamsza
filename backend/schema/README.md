# `backend/schema/` — the runnable lamsza schema

Two generated, committed files that bring an empty Postgres to a working lamsza
database in one command:

| File | What it is |
|---|---|
| `001_schema.sql` | The whole current schema, structure only: 41 tables, the `locations` view, the `pg_slugify`/`sync_*` functions and their triggers, indexes, constraints, and the `pg_trgm` and `unaccent` extensions. |
| `002_reference.sql` | The reference rows the app cannot work without: `counties`, `geo_locations`, `settlements`, `settlement_location_types`, `historical_seats`, `county_historical_seats`, `venue_types`, `weather_desc_translations`. |

```bash
scripts/db-bootstrap.sh                     # the local dev database
scripts/db-bootstrap.sh --create            # create it first if it is missing
scripts/db-bootstrap.sh --url "$DATABASE_URL"
```

## Why this exists

`backend/migrations/` holds 40 files that were applied by hand over the life of
the project. It is a historical record, not a runnable sequence: there is no
recorded order, and several files only ever applied to a schema that has since
moved on. Before BOG-53 there was no way to create a working lamsza database
from this repo at all — which is why the root `backend` test package, a database
integration suite, could not run in CI. It got excluded by name instead.

`backend/migrations/` stays where it is. Nothing here replaces it, and old
migrations are not to be deleted; they are the only explanation of how some
columns got there.

## Generated, and committed

Generated, because nobody keeps 2700 lines of DDL correct by hand, and a
bootstrap that is subtly wrong is worse than none at all. Committed, because CI
has to build the database out of the repo, and because a reviewer should see the
schema move in a diff.

After any schema change, regenerate and commit the result with the change:

```bash
scripts/db-dump-schema.sh
```

It reads a live database — the local dev one by default — so what lands here is
what the running app actually has. There is no CI-only copy of the schema: CI
runs the same `scripts/db-bootstrap.sh` over these same two files.

## The allowlist in `002_reference.sql`

`scripts/db-dump-schema.sh` names the reference tables explicitly rather than
excluding the ones to skip. `users`, `sessions`, `admin_sessions`,
`site_settings`, `entries` and everything else user- or admin-generated must
never land in a committed file: that is personal data, and `site_settings` can
hold configuration values that should not be public.

**Adding a table to that list is a decision to publish its contents.** Read the
rows before you do it.

## Drift

Nothing checks automatically that these files still match the live dev database
— that check would need a database, and CI only has an empty one. What CI does
catch is the case that matters: if the committed schema stops being enough for
the code, the `backend` suite fails against the bootstrapped database. A column
that exists in dev and was never dumped here shows up as a red test, not as a
silent pass.
