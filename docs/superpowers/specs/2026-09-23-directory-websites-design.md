# Directory websites

Date: 2026-09-23

Related: `docs/superpowers/specs/2026-09-22-user-profile-design.md`, `docs/superpowers/specs/2026-09-23-listing-public-claimed-design.md`

## Problem

People need to add a website to the directory before any business listing exists for it. The same domain must not be added twice, including when a listing already stores that site. Search for the domain and search for the business name are different queries and must not dump the same facts twice.

## Goal

A signed-in user can add a website for free by giving a domain, a title, and a short description. An admin approves it. The approved website is a weboldal in the index and in the main search, with a direct link to the site. It has no listing page until a user claims it and creates the listing through the existing listing create and ownership rules. After that listing is public, a domain query shows the website first and the listing second. A name query shows the listing only, with the existing **Claimed** mark when the listing has an active owner.

## Non-goals

- A public page, entry type, category, or tag for an unclaimed website
- A new ownership badge or a second claim/join flow
- Fees, quotas, or email
- Banning an account from sign-in, or deleting that account’s existing listings
- Fetching a title or description from the live site
- Treating a path on a domain as a separate website

## Domain identity

One canonical key per site:

- Lowercase the host.
- Drop the scheme, port, path, query, and a leading `www`.
- The key is the registrable domain, so every subdomain collapses to that same site.

`https://www.kezdisorozo.com/menu`, `kezdisorozo.com`, and `shop.kezdisorozo.com` are one website.

The key is unique across pending submissions, approved websites, and URLs already stored on listings. A listing URL is registered under the same key, so Manifesto’s existing site already occupies `kezdisorozo.com`. Submitting that domain stops the user. If a public listing holds the key, the stop points at that listing. If an approved website holds it and no listing exists yet, the stop points at that weboldal. A pending submission holds the key until an admin rejects it.

## Add your website for free

The public directory offers one action, labeled **Add your website for free**. No payment.

1. A signed-out visitor who chooses it gets the existing sign-in. After sign-in, the visitor returns to the form. A signed-in user opens the form directly.
2. The form asks for three fields and no others: **domain**, **title**, and **short description**. It does not ask for category, type, settlement, phone, or address.
3. All three are required. The domain must normalize to a canonical key. Title and short description are plain text; HTML is stripped. Title is at most 120 characters. Short description is at most 300 characters.
4. Submit checks the key. A duplicate stops the form, as in **Domain identity**, and does not save a second row. A valid new key saves one submission: the submitter, the canonical key, the original host, the title, and the short description. The status is pending.
5. The user sees that the website is waiting for admin approval. It has no public page, and it is absent from the index and from search.
6. Saving the submission creates one admin notification. That message is where approval starts. No email is sent.
7. The admin acts from the buttons on that message. The message then disappears, and the admin queue count drops by one.

The submitter cannot edit a pending submission. A rejected domain is submitted again from the same form.

## Admin notification

Each new submission adds one notification for the admin. It counts toward the existing admin queue, beside unpublished listings and pending member requests. The message shows the submitter, the domain, the title, and the short description. It has three buttons:

- **Approve.** The website becomes a public weboldal. The index row and the website search hit show the title, the short description, the domain, and the direct link to the site.
- **Reject.** The submission is deleted and the key is freed. The same user, or another user, may submit that domain again.
- **Ban User.** The submission is deleted and the key is freed. That account cannot submit a website or claim one. Sign-in still works, and listings the account already owns stay as they are. A later attempt to add or claim a website is refused.

## Unclaimed weboldal

An approved website that nobody has claimed has no listing page. There is no `/bejegyzes/` URL for it.

The index lists it as a weboldal. The row shows the title, the short description, and the domain. Its link opens the site itself. A search whose whole query normalizes to that domain returns one hit: the website, with the title, the short description, and the direct link. A search that matches the title or the short description, and is not a domain query, returns that same website hit. There is no second hit and no **Claimed** mark, because no listing exists.

## Claim and the listing page

The listing page comes into existence when a signed-in user claims that approved website and creates the listing from it. This uses the existing create and ownership rules from the user-profile spec. It does not add a parallel claim path.

- Any signed-in user who is not banned from websites may do this, not only the person who submitted the domain.
- The domain is fixed to the approved website.
- Required fields are the existing ones: name, settlement, category, type. Other listing fields stay optional.
- The creator is the owner. **Claimed** is the existing mark for an active owner. It is not a stored column and it is not a new badge.
- The new listing stays unpublished until an admin publishes it. The owner sees it on **Bejegyzéseim** as waiting. Until it is published, the public index and search still show only the website hit.
- Publishing does not set **Ellenőrzött**. That remains the admin `verified` flag.
- Once a listing has been created from the domain, no second listing can be created for that key. While that listing is unpublished, other users get no public claim or join action. After it is published, another signed-in user uses the existing claim/join rule: the first owner is already set, so the new user becomes a pending member and waits for an admin. A second pending request from the same user is ignored. Rejecting a request deletes that row.
- Owner and member editing, removal, and delete stay on the existing **Bejegyzéseim** rules.

The domain remains a weboldal after the listing exists. The website hit still links directly to the site. The listing is a normal directory listing under the type and category the owner set.

## Search

The main search engine treats a website as its own hit, separate from a directory listing. The website hit shows the title, the short description, the domain, and a direct link to the site. The listing hit opens the listing page and, when the listing has an active owner, shows the existing **Claimed** mark.

**Domain query.** The whole query normalizes to one registered domain (`kezdisorozo.com`, `www.kezdisorozo.com`, or a subdomain).

1. The website hit, with the direct link.
2. The published listing tied to that domain, when one exists, with **Claimed** when it has an active owner.

**Listing query.** The query is a business name or other listing terms, such as `Manifesto`. The result is the published listing only, with **Claimed** when it has an active owner. The website hit is omitted. A query that matches the website title or short description follows this rule once the listing is published, and follows the unclaimed rule while no listing is public.

A word that does not normalize to a domain, such as `kezdisorozo` with no public suffix, is not a domain query. It matches an unclaimed website by title or short description, or a published listing by its own terms.

Places, events, news, and the other existing search groups stay as they are. The website hit leads the directory results only on a domain query.

## Index

Approved websites appear in the directory index as weboldal rows. Each row shows the title, the short description, and the domain, and links to the site. An unclaimed website has no internal page. After its listing is published, the website stays in the weboldal index with the same title, description, and direct link, and the listing appears under its own category and type. The two are one domain and one listing, not two copies of the business. The listing keeps its own name. Creating the listing does not rewrite the website title or short description.

## Failures

- Signed-out submit or claim: the existing sign-in requirement. No anonymous submissions. After sign-in, the visitor returns to the add form.
- Missing domain, title, or short description, or text over the length limit: the form stays open and does not save.
- A banned account submits or claims a website: the action is refused. No new notification is created.
- Duplicate key, including `www`, a subdomain, a pending submission, or a URL already on a listing: the submit stops. A public listing is named when one holds the key. An approved website with no listing is named when that is what holds the key.
- Claim of a domain that is still pending or was rejected: refused. There is no listing to create.
- A domain that already has a listing, published or not: no second listing. Public join is available only after that listing is published, through the existing claim flow.
- Unpublished listing: absent from the public index, search, and public claim/join, same as today.

## Tests

- The add form accepts only domain, title, and short description. A missing field or a too-long title or description does not save.
- A signed-out visitor is sent through sign-in and returns to the form.
- A saved submission stores the submitter, the canonical key, the title, and the short description, stays out of the index and search until approval, and creates one admin notification that raises the admin queue count.
- The notification shows the submitter, domain, title, and short description, and offers Approve, Reject, and Ban User.
- Approve publishes the weboldal and clears that notification. Reject deletes the submission, frees the key, and clears that notification.
- Ban User deletes the submission, frees the key, clears that notification, and later submit or claim attempts from that account are refused. Sign-in still works.
- `www`, the bare domain, a path, and a subdomain resolve to one key.
- A second submission of that key is refused while a submission is pending or approved, and while a listing already stores the URL.
- Rejection frees the key. A later submission can use it.
- A pending or rejected domain is absent from the index and from search.
- An approved unclaimed website appears as a weboldal with its title, short description, and direct link, and has no listing page.
- A domain query for that site returns only the website hit.
- A query that matches the unclaimed title or short description returns only the website hit.
- Creating the listing from that domain uses the existing create rules: the creator is the owner, and the listing stays unpublished until an admin publishes it.
- Before publish, search still returns only the website hit.
- After publish, a domain query returns the website hit first and the listing second. The listing shows the existing **Claimed** mark.
- A name query returns the listing only, with **Claimed** when it has an active owner, and omits the website hit.
- A second listing cannot be created for a domain that already has one. After that listing is published, a second user becomes a pending member through the existing claim flow.
- Publishing the listing does not set `verified`.
