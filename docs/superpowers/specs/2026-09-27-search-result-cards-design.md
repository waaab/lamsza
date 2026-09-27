# Search result cards

Date: 2026-09-27

## Problem

The main search panel draws eight result types with four different layouts. Websites and services use the directory cards. Settlements reuse the service card, including the initials tile. Attractions, venues, events, news, and historical seats are one-off rows, and attractions and seats use their own border colors. In one column they do not read as the same list.

## Goal

Every result in the search panel is one compact row: a title, an optional meta line, and an optional one-line description. Látnivalók keep a brown heading and border. Történelmi székek keep a blue heading and border. The other six groups stay neutral. Links, filters, and which groups appear stay as they are.

## Non-goals

- Index, hírek, események, városok, falvak, megyék, and székek pages
- Homepage quick links, weather, date, coat of arms, and the admin dashboard
- Search API shape, result order, or the Index / Szolgáltatások / Weboldalak filters
- Initials, photos, ratings, tags, hours, or a news image on the search row

## Card

One component, `SearchResultCard`, used only by `SearchEngine`. `EntryCard` and `WebsiteCard` stay on the other pages.

The component receives:

- `href` and `title`
- `meta`, omitted when empty
- `description`, omitted when empty, and clipped to one line in CSS
- `accent`: `none`, `attraction`, or `seat`
- `external`, for websites and news, which open in a new tab with `rel="nofollow noopener"`
- `claimed`, for websites only. A claimed website shows the existing claim mark beside the title. Other types do not.

`SearchEngine` paints the section heading: brown for Látnivalók, blue for Történelmi székek, neutral for the rest. The card paints only its own border from `accent`. Hover changes the row background and leaves that border color alone.

A result with an empty title, or that cannot build an `href`, is not rendered.

## Lines

Text is trimmed. A joined meta line uses ` · `. If only one side exists, that side is shown alone. An empty line is left out.

| Group | Title | Meta | Description | Link |
|---|---|---|---|---|
| Weboldalak | `title` | `domain` | `description` | `url`, new tab |
| Szolgáltatások | `name` | `location` · category label | `notes` | `/bejegyzes/{slug}` |
| Látnivalók | `name` | `county_name` | `description` | `/{county_slug}-megye/{slug}` |
| Helyszínek | `name` | `settlement_name` · `kind_label` | — | `/{county_slug}-megye/{settlement_slug}/helyszin/{slug}` |
| Események | `title` | short date · `location_name` | — | `/esemenyek/{id}` |
| Települések | `name` | `county` | — | `/{county_slug}-megye/{slug}` |
| Történelmi székek | `name` | — | — | `/szekek/{slug}` |
| Hírek | `title` | `source` | — | `link`, new tab |

The service category label is `canonicalEntryCategory`. The event date is `formatDateShort` of `start_date`. If `start_date` is empty or does not parse, the date side is omitted. Settlements are no longer adapted into an `EntryCard` record.

A venue or attraction link also needs its parent slugs (`county_slug`, and for a venue `settlement_slug`). Without them, the row is skipped. A service, settlement, or seat without `slug` is skipped. An event without `id` is skipped. A website without `url`, or a news item without `link`, is skipped.

Empty groups stay hidden. The existing filters still decide which groups are shown.

## Building the lines

A plain function maps each result to `{ href, title, meta, description, accent, external, claimed }`, or to nothing when the title or the link is missing. `SearchResultCard` only renders that result. The function returns the full description string. The one-line clip is CSS on the card. The function is what the tests call.

## Tests

`node --test` covers:

- One example of each of the eight groups, including href, meta, description, accent, and external
- A joined meta line with only one side present, and no leftover separator
- A missing description omitted
- A seat with title only
- A result with no link skipped
- A result with an empty title skipped

No browser test. The search API is unchanged.
