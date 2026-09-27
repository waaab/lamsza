# Listing profile

Date: 2026-09-27

Related: `docs/superpowers/specs/2026-09-23-listing-public-claimed-design.md`, `docs/superpowers/specs/2026-09-22-user-profile-design.md`

This spec replaces the public-field rules in the 2026-09-23 listing spec. Ratings stay as that spec defines them. Claim and membership rules below replace the immediate claim used today.

## Problem

`/bejegyzes/[slug]` shows the same full profile for every listing. Gazdátlan and Átvéve, and Nem ellenőrzött and Ellenőrzött, do not change which fields a reader sees. Claim takes effect immediately. A suggestion does not reach an admin as a preview of changed fields.

## Goal

A reader gets a useful short profile. Átvéve opens phone and social profiles. Ellenőrzött opens photos and the full services and introduction text. Logged-in users can suggest corrections, ask to own a gazdátlan listing, or ask to join an Átvéve listing. Hungarian UI.

## Non-goals

- Comments, review replies, and changes to the ratings rules in the 2026-09-23 spec
- Choosing delivery for types other than restaurants
- A message on a claim or membership request
- Letting the sender withdraw a waiting claim or suggestion
- Editors for photos on the public URL
- Changing how a logged-in user creates a new listing

## Marks

All four marks stay on the header.

| Mark | Meaning | Color |
|---|---|---|
| Gazdátlan | No active owner | Grey |
| Átvéve | An owner manages the listing | Blue |
| Nem ellenőrzött | An admin has not confirmed the public facts | Grey |
| Ellenőrzött | An admin has confirmed the public facts | Blue |

Átvéve and Ellenőrzött are independent. A gazdátlan listing stays on the short view even when it is already Ellenőrzött.

A listing the user creates is Átvéve at once, because that user is the owner. Sajátnak jelölöm is only for a listing that has no active owner.

## Short view

Shown for every gazdátlan listing, and for an Átvéve listing that is still Nem ellenőrzött.

- **Szolgáltatások:** the stored tag list, clamped to two lines.
- **Bemutatkozás:** the stored notes, clamped to four lines.
- **Helyszín:** town and street when stored. A map pin and **Útvonal** when coordinates are stored. Coordinates are not printed.
- **Nyitvatartás:** shown when the hours switch is on, including an empty week. An admin can turn the switch on for any listing. The owner, and an accepted member, can turn it on once the listing is Átvéve.
- **Kiszállítás:** the same switch. A type that does not offer delivery never shows the block. Restaurants offer delivery. The Lámsza.com listing does not. Every other type does not show delivery in this version.
- **Sidebar:** website and languages, when those values are stored.

An empty services or introduction section keeps the existing empty placeholder. A phone row, a social row, a website row, or a languages row with no stored value is omitted. The hours and delivery blocks are omitted while their switches are off.

## Átvéve adds

Social profile links, each a label and a URL, and the phone number, when stored. Services and the introduction stay clamped.

## Ellenőrzött adds

When the listing is also Átvéve: the photo gallery, when at least one photo is stored, and the full services list and introduction text. A gazdátlan listing keeps the clamped text and hides photos even when it is Ellenőrzött.

## Header and sidebar

The heart sits at the top right of the header. Only a logged-in user sees it. The old action row under the profile is removed.

**Sajátnak jelölöm** sits in the sidebar on a gazdátlan listing. Once the listing is Átvéve, that spot becomes **Tagság kérése**. Either label is visible only to a logged-in user who is not an active owner or an active member.

**Javaslat módosításra** is for a logged-in user who is not an active owner or an active member. The active owner sees **Szerkesztés** in that same sidebar spot. It opens the existing listing editor and saves the listing directly. It does not create a suggestion. A waiting suggestion leaves **Szerkesztés** in place for the owner.

An accepted member does not see **Javaslat módosításra** on the public page. They edit from **Bejegyzéseim**, which already has **Szerkesztés**.

A logged-out visitor sees none of these actions and does not see the heart.

## Javaslat módosításra

This section is the visitor suggestion. The active owner uses **Szerkesztés** instead, as described above.

The form opens filled with the stored public values, including phone, social profiles, and the full services and introduction text when the public page hides or clamps them. The user edits what is wrong. The request stores only the changed fields, plus an optional note.

Fields on the form:

- name
- services
- introduction
- settlement and street address, using the stored location and address fields
- opening hours, when the hours switch is on
- delivery hours, when the delivery switch is on and the type offers delivery
- website
- phone
- social profiles, each a label and a URL
- languages

A cleared field is a change. Accept writes that empty value. Accepting a new name keeps the existing slug.

Photos are not on the form.

Only one suggestion can be open for a listing. Every eligible logged-in user sees that it is waiting, and the action stays disabled. The sender cannot withdraw it. A second submit while one is open is rejected, and the page shows the waiting state.

An admin previews every changed field and the note, then accepts or denies the whole request. Accept writes the submitted values onto those fields. Deny discards the request and the action returns.

## Sajátnak jelölöm

The request records who is asking and which listing. It has no note. An admin previews that pair and accepts or denies it.

Accept makes that user the active owner. The listing becomes Átvéve. Deny removes the request. The listing stays Gazdátlan until accept.

Only one request can be open. While it is waiting, every eligible logged-in user sees that a request is already open, and the action stays disabled. The sender cannot withdraw it.

## Tagság kérése

The request records who is asking and which listing. It has no note. Several requests can be open at once. Someone who already sent one sees that their request is waiting. Other eligible users still see **Tagság kérése**.

The owner accepts or denies each request from **Bejegyzéseim**. An admin can accept or deny it from the listing queue. Accept makes that user an active member. Deny removes the request.

An accepted member edits the listing directly, including the hours and delivery switches, using the same fields as the owner. Only the owner can delete the listing. Admins keep their existing admin tools.

A user with a pending membership is not yet a member, so they may still send **Javaslat módosításra** when no suggestion is open.

## Queues

- Suggestions: a new admin queue. Each item shows the changed fields, the note, and Accept / Deny.
- Claims: a new admin queue for a pending owner. Each item shows the user and the listing, and Accept / Deny.
- Membership: the owner handles it on **Bejegyzéseim**. The admin handles it in the existing members queue.

The admin queue count includes open suggestions and open claims.
