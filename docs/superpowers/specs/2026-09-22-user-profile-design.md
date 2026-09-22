# User profile

Date: 2026-09-22

## Problem

A signed-in Google account stores a name, an email, and login timestamps. Theme, homepage link slot count, personal homepage links, and recently opened listings live only in this browser. Directory listings have no owner. The public **Ellenőrzött** badge is the `claimed` column, so “someone owns this” and “an admin checked this” are the same flag. The homepage weather widget and events ticker follow one site-wide settlement (`my_location_slug`, default Csíkszereda).

## Goal

A signed-in page at `/profil` where this Google account’s preferences, listings, links, history, and favorite places are saved. The header theme control stays in sync with the account. The homepage, while signed in, follows that account. Admins get an in-app queue for unpublished listings and member requests.

Hungarian UI. No public profile URL.

## Non-goals

- Email to admins
- Ownership transfer
- Storing the raw Google token, or claims we do not have a label for
- Homepage widgets for favorite listings or favorite events
- A cap on how many favorites a user may save
- Changing weather providers, cache TTL, or icon style (those stay site settings)

## Profile page

`/profil` requires a session. A signed-out visitor gets the existing Google sign-in dialog. The name in the header opens this page.

Tabs:

1. **Profil** — read-only Google data: photo, given name, family name, display name, email, locale, Google account id, last login, account created. Only non-empty fields are shown.
2. **Beállítások** — theme (`light`, `dark`, `system`) and homepage slot count (existing range 7–14, default 7).
3. **Bejegyzéseim** — listings you own, listings where you are a member, your unpublished listings, and your pending join requests. Create a listing here.
4. **Linkjeim** — add, edit, delete, and reorder the homepage quick links (title, URL, color).
5. **Előzmények** — listings you opened, newest first. Remove one, or clear all.
6. **Kedvenc helyek** — four groups: Települések, Látnivalók, Bejegyzések, Események. Remove one here. Add them from the public pages.

## Google fields

Known fields: photo URL, given name, family name, display name, email, locale, and the Google account id (`sub`).

Every sign-in overwrites those fields with the current token. If the token omits a known field, the stored value is cleared. `sub` is the account key and does not change. Extra claims are ignored. The raw token is not stored.

## Favorites and the homepage

A favorite is one of: settlement, látnivaló (attraction), listing, or event. The public page for each has a working favorite control. The event-page heart saves a favorite. Toggling again removes it. A duplicate add is ignored. A favorite whose record was deleted disappears from the list.

Signed in, the homepage:

- Renders one weather widget and one events ticker per favorite settlement, in the order they were added.
- Renders one weather widget per favorite látnivaló that has coordinates, in the order added. A látnivaló without coordinates stays on the profile only.
- Uses the account’s links and slot count.
- If there is no favorite settlement, weather and events fall back to the admin default settlement.

Signed out, the homepage keeps today’s behavior: admin default settlement, and links plus slot count from this browser.

Favorite listings and events do not add homepage widgets.

## Listing ownership

A listing has at most one owner and any number of members. Admin can still edit or delete any listing from the admin panel.

**Claim.** On a published listing with no owner, the signed-in user becomes the owner immediately. If an owner already exists, the user may request to join as a member from that public listing. Unpublished listings have no public claim or join action. That request waits for an admin. A second request while one is pending is ignored. Rejecting a request deletes the pending row, and the user may request again.

**Create.** Any signed-in user may create a listing. Required fields: name, settlement, category, type. Other listing content fields are optional. The creator is the owner. `published` is false until an admin publishes it. The owner sees it on **Bejegyzéseim** as waiting. It is absent from the public site and from search.

**Members.** An approved member may edit listing content: name, category, type, settlement, URL, phone, address, notes, languages, hours, delivery hours, and photos. A member cannot delete the listing, remove people, change the owner, publish, or set verified. The owner may remove a member immediately. There is no ownership transfer in this version.

**Delete.** Only the owner may delete the listing from the profile. Members and pending requests go with it. Admin delete from the admin panel still works.

**Badges.** **Claimed** is shown on the public listing when it has an active owner. It is not a stored column. **Ellenőrzött** is the `verified` flag, set only by an admin. Publishing does not set `verified`. The public page shows the two badges independently. Existing `claimed` values migrate to `verified`. Existing listings stay `published`.

Two people claiming the same free listing: one transaction makes the first the owner; the other becomes a pending member request.

## Admin queue

The queue is the set of unpublished listings plus pending member requests. It is not a separate notification table. The admin button in the public header shows the count of those rows. The queue itself is a section in the existing admin panel, where an admin can publish a listing or approve or reject a join request. No email is sent.

## Account data

On the user row: photo URL, given name, family name, locale, theme, homepage slot count. Theme and slot count are null until set. Display name and email stay on the row as they do today.

Separate tables:

- **Links** — title, URL, color, order, per user.
- **History** — the same item shape as today’s browser history (`slug`, `name`, `category`, `location`, `photo`), newest first, at most 12. The profile lists every stored row. The listing page still shows at most 8 and hides the listing that is open.
- **Favorites** — user, entity type, entity id, `created_at`.
- **Listing people** — listing, user, role (`owner` or `member`), status (`active` or `pending`).

On the listing: `verified` (migrated from `claimed`) and `published` (true for rows that already exist).

**First import.** Once per account, when `prefs_imported_at` is null, copy this browser’s theme, slot count, quick links, and listing history onto the account if the account value is still empty. Then set `prefs_imported_at`, including when the browser had nothing to copy. Later deletes do not import again. If the copy fails, sign-in still succeeds and `prefs_imported_at` stays null so a later visit can retry. Favorites start empty. Signed-out browsers keep using local storage.

While signed in, the header theme control and **Beállítások** read and write the account theme. While signed out, the header keeps writing this browser’s theme.

While signed in, opening a listing prepends history on the account. While signed out, it keeps writing this browser’s history. This replaces the “history stays on this device only” rule in the listing-page spec for signed-in users. Signed-out visitors are unchanged.

## Failures

Profile reads and account writes require a session. A failed save keeps the previous value and shows an error. A member who tries to delete a listing or remove a person is refused. An unpublished listing is not public. A cleared Google field is shown as blank.

## Tests

- Profile APIs reject a missing session.
- Sign-in stores photo, given name, family name, and locale, and clears a known field the token omits.
- Theme saved on the account is the theme the header applies while signed in.
- The one-time import copies browser theme, slot count, links, and history only while `prefs_imported_at` is null, and does not run again after the user deletes those rows.
- Claim with no owner is immediate. Claim when an owner exists stays pending. A concurrent second claim becomes pending.
- A created listing is unpublished until an admin publishes it. Publishing does not set `verified`.
- The public listing can show Claimed and Ellenőrzött independently.
- The owner can remove a member. A member cannot delete the listing.
- Favorites accept a settlement, a látnivaló, a listing, and an event. A deleted target disappears.
- Signed in, homepage weather and events follow favorite settlements, and látnivalók with coordinates get weather. With no favorite settlement, the admin default is used.
- Signed out, the homepage uses the admin default settlement.
- The admin nav count equals unpublished listings plus pending member requests.
