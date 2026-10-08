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
| Profile / account button (`tb-profile`) | Lámsza's | Fiók is the default in every public app: Lámsza, Szótár and Játszótér each have a `/fiok` page, reached from the account menu (`tb-account-menu`). This replaces Szótár's exemption of 2026-10-07. Owner's decision, 2026-10-08. |
| Settings control (`tb-settings`) | Lámsza's | A Beállítások page in every public app (`/beallitasok`), reached from the account menu; no settings button or theme dropdown in the toolbar. This replaces "Szótár and Játszótér keep their theme dropdown". Owner's decision, 2026-10-08. |
| Admin link in toolbar (`tb-admin-link`) | None | No user-facing app links to the admin app: no toolbar button, no edit link, no redirect. Admins open it by URL (WAYS_OF_WORKING R18). Owner's decision, 2026-10-07. |
| App-only toolbar buttons (`tb-extra`) | Other | keep as is |
| Toolbar while auth loads (`tb-skeleton`) | Lámsza's | Belépés shows until the session is known; the profile button then takes its place at the same size. |
| Back-to-top button (`tb-backtotop`) | Lámsza's | Back-to-top in every app, the admin app included. |
| Responsive behaviour (`tb-responsive`) | Lámsza's | Lámsza's toolbar rules for the apps with a toolbar; the admin app gets basic small-screen rules for its own shell. |
| Apps launcher (`tb-apps-launcher`) | New | A nine-dot button (`AppIcon` "apps") as the last button on the right of the toolbar in Lámsza, Szótár and Játszótér, not in the admin app. It opens a small panel (the settings dropdown's look) with each public app's icon and name, the current app highlighted. One shared component, `AppsLauncher.svelte`, and one app list, `networkApps.js` (links from `networkOrigins.js`), both synced from lamsza. Keyboard: Enter or Space opens it and focuses the first app, Escape closes it and returns focus to the button, tabbing out or a click outside closes it. It adds to `tb-cross-links`, which stay as they are. Owner's request, 2026-10-08. |
| Account menu (`tb-account-menu`) | New | The right of the toolbar in Lámsza, Szótár and Játszótér holds only two buttons. Signed out: Belépés, then the apps launcher. Signed in: the profile button, always the profile icon (never the user's photo), then the apps launcher. The profile button opens a menu in the dropdown look (panel padding 0.5rem, radius 12px; items keep their highlight inside the padding): the shown name and email, then Fiók, Beállítások, Kijelentkezés. One shared component, `AccountMenu.svelte`, with no per-app variations in appearance: the top-right area is pixel-identical in Lámsza, Szótár and Játszótér (owner, 2026-10-08); keyboard as the launcher (Enter or Space opens it on the first item, Escape closes it and returns focus, tabbing out or a click outside closes it). Owner's decision, 2026-10-08. |
| Add button in the title row (`pg-title-add`) | Lámsza's | A page's main "new item" button sits at the far right of the title row as a `btn btn-lg` with a short helper text under it, the way Lámsza's Index places "Új Bejegyzés"; not in the toolbar. Szótár's "Új szó" follows it on its word-list pages (home, /lista, /betu, /betu/<letter>, /szofaj/<type>, /kereses). Everyone sees the button; signed out it opens the sign-in dialog and then the form. On phones it wraps under the title. Owner's decision, 2026-10-08. |

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
| Icon system (`ic-system`) | Admin's | One icon system for all four apps: the shared `AppIcon.svelte` (the admin set, which also has Játszótér). Since 2026-10-08 no toolbar draws its own icons: Lámsza's inline toolbar icons and the admin app's `AdminPlusIcon` (now `AppIcon` "plus") moved into the set, pixel-identical. |
| Játszótér icon (`ic-jatszoter`) | Játszótér's | Its cross-of-tiles drawing, at the set's stroke 2 like every other icon (not the original 1.75). Confirmed by the owner, 2026-10-07. |
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
| Content dialog shell (`dlg-content`) | Lámsza's | On phones (up to 768px) the shell takes the screen's width (`calc(100vw - 1rem)`), not 70vw; the shell's own rule in `global.css` since 2026-10-07, asked for with Tájszórejtvény. |
| Close-button wording (`dlg-close-label`) | Owner's rule | A dialog that only shows content (media, notices, help, info) closes with "Bezárás". "Mégse" is only for cancelling an action, in a dialog that asks for one (confirm, forms, sign-in), next to the action button. Added by the owner on 2026-10-07. |

## Settings and profile

| Item | Baseline | Note |
|---|---|---|
| Theme choice UI (`set-theme-ui`) | Lámsza's | Only a signed-in user chooses a theme, on the Téma tab of the app's Beállítások page. The theme dropdown leaves the toolbar in Szótár and Játszótér. Owner's decision, 2026-10-08. |
| Theme storage (`set-theme-store`) | Lámsza's | Signed in: the choice is saved in that app's own account (`users.theme` in Lámsza, Szótár and Játszótér) and cached in `localStorage` so the page paints in it. Signed out: the device's theme (`prefers-color-scheme`); signing out, or a session the server no longer knows, clears the saved one. Stored per app, not shared, until single sign-on (OPEN_ITEMS). Owner's decision, 2026-10-08; replaces "Szótár and Játszótér keep localStorage". |
| Theme-init script (`set-theme-init`) | Lámsza's |  |
| Profile page (`set-profile`) | (no pick) | See `acc-page`. |
| Fiók page (`acc-page`) | New | One Fiók page in every public app, built from the shared `AccountPage.svelte` and `AccountDetails.svelte`. Its first tab, "Fiók", shows the same details everywhere (photo, Megjelenített név, names, email, Google-azonosító, last sign-in, account created); the app's own tabs follow (Lámsza: its current tabs; Szótár: Szójavaslataim; Játszótér: Eredményeim). The display name belongs to Lámsza, which is the only place to edit it; Szótár and Játszótér copy it by Google ID through Lámsza's internal account API at sign-in and when Fiók opens, and fall back to Google's name. Each app keeps its own Google sign-in. Owner's decision, 2026-10-08. |

## Error pages

| Item | Baseline | Note |
|---|---|---|
| 404 / error page (`err-page`) | Lámsza's | One shared page for the network: `ErrorPage.svelte` with the lantern illustration (`icons/ErrorLantern.svelte`, inline SVG on the `--text-muted`, `--warm-light` and `--szekely-red` tokens), "Hoppácska!", "Hiba: <code>", a Hungarian title and explanation for 400, 401, 403, 404, 408, 410 and 429 and a fallback for any other code; noindex. Every app's root `+error.svelte` renders it through `ErrorShell.svelte`, one minimal shell for the network: the Lámsza button and, in a sub-app, its own button; no footer, sidebar or right-side icons; "Vissza a főoldalra" goes to the app's home; nothing is fetched, so it works with the backend down. The app's own shell lives in a route group (`(public)` in Lámsza, `(app)` in Szótár and Játszótér) that the root error page sits outside. Owner's decisions, 2026-10-07. |
| Old brand name in titles (`err-brand`) | Szótár's | No old brand anywhere: "Székely Gugel" becomes "Lámsza". |

## Sign-in states

| Item | Baseline | Note |
|---|---|---|
| Google button component (`si-button`) | Lámsza's |  |
| Loading wording (`si-loading`) | Lámsza's |  |
| Async content reserves its space (`ld-reserve-space`) | New | Content that arrives after the page appears never pushes what is already on screen. Either the page fetches in its `+page.js` `load`, so title, badge, lead and list render together, or it shows a placeholder of the final size in the existing style (Szótár's dimmed rows, the global `.skeleton`) until the content arrives; a placeholder that may differ in size reserves at least a screen's height so the next block starts below the fold. Checked with the layout-shift measurement on a slowed API (target: no shift at all; under 0.01 per page). Owner's decision, 2026-10-08. |
| Error wording (`si-errors`) | Játszótér's |  |
| Google client id source (`si-clientid`) | Lámsza's |  |

## Admin areas

| Item | Baseline | Note |
|---|---|---|
| Admin layout (`adm-layout`) | Other | Keep the admin app's own layout. Since 2026-10-07 it has three sections on one shell (`AdminShell`): `/` ("Vezérlőpult", unchanged), `/dictionary` and `/games` (WAYS_OF_WORKING R18). Each section's sidebar uses only shared icons: first an external link to the app (`external`), then its dashboard (`dashboard`), then one entry per feature; the dashboard cards mirror the sidebar entries with their counts. The header buttons switch between the three sections and use the shared toolbar buttons (`.btn.nav-btn`: pill, 16px icons, the current section in the active state). The section dashboards are titled "Vezérlőpult Szótár" and "Vezérlőpult Játszótér". Owner's decisions, 2026-10-07. |
| English labels (`adm-english`) | Admin's | Translate the admin's English labels to Hungarian. Clarified by the owner; overrides the Admin pick. |
| Admin sign-in gate (`adm-gate`) | Játszótér's | Since the admin move (R18, 2026-10-07) only the admin app has a gate. No user-facing app links to it (no Admin button, no admin-only links); admins open it by its URL. Owner's decision, 2026-10-07. |
| Network toolbar and footer on admin pages (`adm-chrome`) | Játszótér's | Szótár and Játszótér have no admin pages since the admin move (R18, 2026-10-07); the admin app keeps its own shell. Clarified by the owner. |
