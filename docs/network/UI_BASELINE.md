# UI baseline

The owner's decisions on how the four apps look and behave, made on 2026-10-07
from the visual inventory (https://claude.ai/artifact/S6jA3Uk64emqQctVv7RcTJ, which
shows each difference with the exact values in every app). For every item: the
chosen baseline, and the owner's note or clarification where there is one.

This file is the reference for UI work across the network: a change that touches
one of these items follows the baseline here, and a new decision is added here in
the same commit (WAYS_OF_WORKING R3, R10). Nothing here assumes Lámsza is the
reference; each item names its own.


## Design tokens

| Item | Baseline | Note |
|---|---|---|
| Info note colours, dark theme (`tok-info-dark`) | Lámsza's |  |
| Info note background, light theme (`tok-info-light`) | Lámsza's |  |
| --thin-grey token (`tok-thin-grey`) | Lámsza's |  |
| App-only tokens (`tok-app-specific`) | Játszótér's |  |

## Shared CSS rules

| Item | Baseline | Note |
|---|---|---|
| global.css as a whole (`css-fork`) | Lámsza's |  |
| Base .btn (`css-btn`) | Játszótér's | Only Játszótér's `user-select: none`. No font-family or font-size on `.btn`, so a `<button>` keeps the browser's own button font, as on Lámsza (owner's decision after seeing step 2). |
| Hover lift on buttons (`css-hover-lift`) | Lámsza's |  |
| .link-dialog box (`css-link-dialog`) | Lámsza's |  |
| .note text colour (`css-note`) | Lámsza's |  |
| .info-box p padding and variants (`css-infobox`) | Lámsza's |  |
| .search-input right padding (`css-search`) | Lámsza's |  |
| .page-title layout (`css-page-title`) | Keep each app's own |  |
| .sidebar-heading case (`css-sidebar-heading`) | Lámsza's |  |

## Typography

| Item | Baseline | Note |
|---|---|---|
| typography.css / component-typography.css (`typo-files`) | Keep each app's own |  |
| Web font loaded (`typo-webfont`) | Játszótér's |  |
| Button font size (`typo-btn-size`) | Lámsza's | `<button>`s use the browser's default button font; `<a class="btn">` links use the site font. |
| Admin page title size (`typo-admin-title`) | Admin's |  |

## Header and toolbar

| Item | Baseline | Note |
|---|---|---|
| Header structure (`tb-structure`) | Lámsza's | Applies to the apps with a toolbar. The admin app keeps its own shell (no network toolbar or footer). Clarified by the owner. |
| The "Lámsza" link (`tb-lamsza-link`) | Lámsza's |  |
| Own-app home button (`tb-own-home`) | Other | keep as is |
| Left-side sections (`tb-left-items`) | Other | keep as is |
| Links to the sister apps (`tb-cross-links`) | Other | keep as is |
| Profile / account button (`tb-profile`) | Lámsza's |  |
| Settings control (`tb-settings`) | Lámsza's | Szótár and Játszótér keep their theme dropdown (localStorage); Lámsza keeps its settings page. Clarified by the owner. |
| Admin link in toolbar (`tb-admin-link`) | Other | keep as is |
| App-only toolbar buttons (`tb-extra`) | Other | keep as is |
| Toolbar while auth loads (`tb-skeleton`) | Lámsza's |  |
| Back-to-top button (`tb-backtotop`) | Lámsza's | Back-to-top in every app, the admin app included. |
| Responsive behaviour (`tb-responsive`) | Lámsza's | Lámsza's toolbar rules for the apps with a toolbar; the admin app gets basic small-screen rules for its own shell. |

## Footer

| Item | Baseline | Note |
|---|---|---|
| Footer present (`ft-presence`) | (no pick) | keep as is |
| Footer line one (`ft-line1`) | Other | keep as is |
| Policy links (`ft-policy`) | Other | All apps should have the same policy links pointing to the same main domain pages. |
| Version / changelog link (`ft-version`) | Lámsza's |  |
| Extra footer stats (`ft-stats`) | Other | keep as is |

## Icons

| Item | Baseline | Note |
|---|---|---|
| Icon system (`ic-system`) | Admin's | One icon system for all four apps: the shared `AppIcon.svelte` (the admin set, which also has Játszótér). |
| Játszótér icon (`ic-jatszoter`) | Játszótér's |  |
| Rendered icon sizes (`ic-sizes`) | Other | keep as is |
| Settings gear attributes (`ic-gear`) | Lámsza's |  |
| aria-hidden on the magnifier (`ic-aria`) | Lámsza's |  |
| Mondások quote icon class (`ic-quote`) | Lámsza's |  |

## Dialogs

| Item | Baseline | Note |
|---|---|---|
| Sign-in dialog (`dlg-signin`) | Lámsza's |  |
| Sign-in dialog states (`dlg-signin-states`) | Játszótér's |  |
| Confirm dialog (`dlg-confirm`) | Szótár's |  |
| Notice / alert dialog (`dlg-notice`) | Lámsza's |  |
| Content dialog shell (`dlg-content`) | Lámsza's |  |

## Settings and profile

| Item | Baseline | Note |
|---|---|---|
| Theme choice UI (`set-theme-ui`) | Other | keep as is |
| Theme storage (`set-theme-store`) | Lámsza's | Account storage stays in Lámsza and Admin only; Szótár and Játszótér keep localStorage. Clarified by the owner. |
| Theme-init script (`set-theme-init`) | Lámsza's |  |
| Profile page (`set-profile`) | (no pick) | keep as is |

## Error pages

| Item | Baseline | Note |
|---|---|---|
| 404 / error page (`err-page`) | Lámsza's |  |
| Old brand name in titles (`err-brand`) | Szótár's | No old brand anywhere: "Székely Gugel" becomes "Lámsza". |

## Sign-in states

| Item | Baseline | Note |
|---|---|---|
| Google button component (`si-button`) | Lámsza's |  |
| Loading wording (`si-loading`) | Lámsza's |  |
| Error wording (`si-errors`) | Játszótér's |  |
| Google client id source (`si-clientid`) | Lámsza's |  |

## Admin areas

| Item | Baseline | Note |
|---|---|---|
| Admin layout (`adm-layout`) | Other | keep as is |
| English labels (`adm-english`) | Admin's | Translate the admin's English labels to Hungarian. Clarified by the owner; overrides the Admin pick. |
| Admin sign-in gate (`adm-gate`) | Játszótér's |  |
| Network toolbar and footer on admin pages (`adm-chrome`) | Játszótér's | Szótár's and Játszótér's admin pages keep the network toolbar and footer; the admin app keeps its own shell. Clarified by the owner. |
