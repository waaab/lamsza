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
| Sidebar heading (`css-sidebar-heading`) | Lámsza's | One class, `.aside_heading` in the shared `global.css`, on every sidebar heading in every app, box titles and the headings inside a box alike: small (`--text-sm`), semi-bold, uppercase, 0.04em letter-spacing, `--text-muted` (the look of Lámsza's former `.index-tags-aside__heading`). The element follows the outline, not the look: a sidebar block's title is `h2`, a heading inside a block `h3`, so no level is skipped. Szótár's `.dict-aside` headings and Játszótér's `GameSidePanel` use it too; the `.widget-title` family (widget cards, RegionNav's "Oldal navigáció") is not a sidebar heading. Owner's decision, 2026-10-09. |

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
| Add button in the title row (`pg-title-add`) | Lámsza's | A page's main "new item" button sits at the far right of the title row as a `btn btn-lg` with a short helper text under it, the way Lámsza's Index places "Új Bejegyzés"; not in the toolbar. Szótár's "Új szó" follows it on its word-list pages (/lista, /betu, /betu/<letter>, /szofaj/<type>; not the home heading, owner 2026-10-09); on /kereses the suggestion sits in the sidebar's "Nem találod?" card instead (owner, 2026-10-09). Everyone sees the button; signed out it opens the sign-in dialog and then the form. On phones it wraps under the title. Owner's decision, 2026-10-08. |

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
| Create buttons (`btn-create-plus`) | New | Every "create new" button in every app shows the shared `AppIcon` "plus" before its text: Lámsza's "Új Weboldal", "Új Bejegyzés", "Új link", the forms' "Hozzáadás" and the quick links' "Új" card; Szótár's "Új szó", "Új szó javaslása", "Új jelentés"; the admin app's "Új jelentés", "Új feladvány", "Új nap", "Hozzáadás". The admin app's "Új …" create panels keep the plus after their text, where it also marks the panel as one that opens. Owner's decision, 2026-10-09. |
| Weather icons (`wx-icons`) | New | Lámsza only. Animated inline SVG for every weather element: `icons/weather/WeatherSymbol.svelte` draws all 41 MET Norway symbol codes, day and night, from a few shared parts (sun, moon, one or two clouds, rain, sleet, snow, lightning, fog); `WeatherGlyph.svelte` draws the measurements and shows their value (thermometer level, wind direction, humidity, pressure needle, UV colour, rain gauge, sun and moon rise and set, day length, moon phase). Theme tokens only (sun --warm-light, rain --szekely-blue, clouds from --text-muted), no colours of their own; CSS motion only, still under prefers-reduced-motion; fixed reserved size. The widgets use them when the admin's icon style is "svg"; the emoji style stays. Review page: `/idojaras/ikonok` (noindex). Accepted by the owner, 2026-10-09. |

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
| The hero (`hero`) | New | One hero per app, and only as the first block of its home page: `Hero.svelte` (shared), the big title and the line under it, with the app's own animated drawing above them where the app has one (Szótár: a stack of dictionary slips, the top one showing a tájszó, flipping to its meaning and sliding off to reveal the next, chosen by the owner over two book drawings, 2026-10-09; Játszótér: letter tiles falling into a crossword grid and spelling JÁTÉK; Lámsza's home has none, owner 2026-10-09). Drawings are inline SVG in the theme's colours, reserve their size, and stand still when the visitor prefers reduced motion, like the error page's lantern. Every app's drawing is cropped to its content and shown at one shared size, about 14.5rem of drawing on a desktop, scaling down on phones, so the apps' heroes match (owner, 2026-10-09). Size and spacing scale with the title (`--text-hero`, scoped to `.hero`; letter-spacing -0.0714em, the line pulled up by 0.268 of the title size, or by 0.12 when the title has a letter below the line: g, j, p, q, y, J, Q, so the line never crosses it; owner, 2026-10-09). A two-part title has its second part in the network red (`accent`: "Székely **szótár**", or within one word with `joined`: "Játszó**tér**"); Lámsza's has none. An app may put its search box in the hero, under the line (Szótár). Every other page title is a `PageHeader` (formerly `PublicPageHero`); nothing else is called hero. Owner's decision, 2026-10-09. |
| Async content reserves its space (`ld-reserve-space`) | New | Content that arrives after the page appears never pushes what is already on screen. Either the page fetches in its `+page.js` `load`, so title, badge, lead and list render together, or it shows a placeholder of the final size in the existing style (Szótár's dimmed rows, the global `.skeleton`) until the content arrives; a placeholder that may differ in size reserves at least a screen's height so the next block starts below the fold. Checked with the layout-shift measurement on a slowed API (target: no shift at all; under 0.01 per page). Owner's decision, 2026-10-08. |
| Lámsza home page (`home-lamsza`) | New | In this order: the hero with the search box and the featured category chips (`home-chips`); Gyorslinkek; the daily mondás; the "Ma" strip (date and time, weather, Szótár's word of the day); "Játszótér · Mai kihívások" (only the games with a daily challenge today, Játszótér's `/api/games` `daily_today`, as the shared game card, `game-look`; a card opens the game's page on today's challenge without starting it, the player starts it with Kezdés; the section goes when no game has one; owner, 2026-10-09); "Friss hírek Erdélyből" (six cards, at most two per source); "Közelgő események" (four cards with a date block, one row that scrolls sideways on phones); "Böngéssz kategóriák szerint" (the main categories as tiles). Each section has a small-caps heading and a "… ›" link to its full page; every async section reserves its space (`ld-reserve-space`). Owner's decision, 2026-10-09. |
| Szótár home page (`home-szotar`) | New | Szótár's own heading around the shared hero (owner, 2026-10-09): from 760px the title ("Székely **szótár**", the shared hero font, up to 6rem), the line and a quiet search stand against the container's left edge and the word-slips drawing (up to 19rem) against its right edge; on phones the shared centred column. The quiet search is small and rounded, its button a magnifier with the button look, no fill at rest in either theme, white only while it has the focus. No "Új szó" in the heading; Napi Székely Szó as a card with a big headword, its part of speech as a chip, a red play button and the word's first letter faint behind it; Napi Székely Mondás; Bemutatkozás as two cards; the alphabet with a dashed note under it; the Játszótér games played with Szótár's Székely words, as the shared game card (`game-look`). The sidebar on every Szótár page with one: Szófajok with bars by word count, Nemrég hozzáadott szavak as chips, Hangfelvételek with a play button per word. Owner's decision, 2026-10-09. |
| Szótár search results (`szotar-search`) | New | The title with the searched text (serif, the text in red); a quiet search field under it (small, rounded, magnifier, a plain "Keresés" button); a summary that jumps to the groups. The hits in two groups, "Szóként" (the headword holds the search) and "Jelentésben" (only the meaning does), each a card like Lámsza's search result cards: headword, part of speech and places, a one-line meaning, the searched text marked (case and accents ignored); the whole card opens the word. In the sidebar, above the dictionary's cards: "Nem találod?" (suggest a new word) and "Böngéssz tovább" (the letter, the full list, Mondások). No hit: what to try and the same ways out. Owner's decision, 2026-10-09. |
| Szótár Mondások (`szotar-mondasok`) | New | The title and the usual greeting line (`h2.greeting`); today's proverb (the server's) exactly as the home page's Napi Székely Mondás: a section heading with its date, the card with "Aszongya, hogy…" between rules and the quote; then "Korábbi mondások", newest first, by month (a small red month heading), each a card with the quote icon on the left and, on the right, its date on top, the quote and its meaning. No event-style date blocks. "Még több" and the end note in their plain style. Every Szótár list ends with "Elérted a lista végét." as the plain dashed note (`info-box`, default text style), the same everywhere. Owner's decisions, 2026-10-09. |
| Játszótér home page (`home-jatszoter`) | New | The hero: letter tiles falling into a small crossword grid and spelling JÁTÉK, the É flipping (the first drawing; the owner preferred it to the turning JÁTÉK board but kept the board's size, the grid as wide as its five tiles, and its replay on a tap, 2026-10-09); "Játszó**tér**"; the line. Then every enabled game in Játszótér's order, four to a row (two on phones), the first one two wide; rows wrap and the last row's tiles stretch, so there is never a hole. On hover or focus a card's animated scene plays and a round play button appears (`game-look`). A Ranglista card with a podium under the games. Until the server's list arrives, the listed games hold the grid dimmed and inert, and are then replaced, so nothing moves (`ld-reserve-space`). Owner's decision, 2026-10-09. |
| Székelyföld pages (`szf-pages`) | New | Székek, Megyék, Városok and Falvak open with the landscape header (`PageHeader landscape`): breadcrumb, title and line, a red sun with a dashed ring and a soft glow (hidden on phones), three mountain ridges in theme colours. Székek and Megyék: a schematic honeycomb map, one hexagon per seat or county in its colour, linking to its page, beside the "Oldal navigáció" cards (the toolbar's icons). Városok: tiles with the type (municípium in red). Falvak: county chips with the county's colour (`?megye=` in the address) and tiles with its dot. A county has the same colour on every page (`src/lib/szekelyfold.js`). Settlement and attraction pages have the same landscape header (`LandscapeBackdrop`), but in the sun's place the place's current weather, drawn with the weather pages' animated symbols (left of the ♡ button; on phones small above the ridges; its slot reserved while the forecast loads). The overview, coat of arms (settlements) and weather sit as cards side by side; the weather card (`PlaceWeatherCard`) shows now, wind, humidity, rain, UV, sunrise and sunset and the next three days, linking to the place's /idojaras page. A coat of arms that fails to load gives way to the default shield. Places and nearby sights as chips. Owner, 2026-10-09: the red sun had no meaning on a place's page; it stays on the four list pages. An attraction's overview has stat tiles for its elevation, area and depth when set in admin ("946 m · magasság", Hungarian number format); its activities are green ✓ chips and prohibitions red ✕ chips (drawn marks, not emoji); its description is the line under the title only. Owner's decision, 2026-10-09. |
| Home category chips (`home-chips`) | New | Under Lámsza's search box, at most six small chips (`btn btn-sm`) linking to `/index/<slug>`, chosen and ordered in the admin app (Bejegyzés kategóriák, "Kiemelt kategóriák a kezdőlapon"); any category, subcategories included. One row that scrolls sideways on phones. None chosen: the row stays empty at its height. Owner's decision, 2026-10-09. |
| Game look (`game-look`) | Játszótér's | A game's colour, icon and card are the same wherever it appears, its name and line too: they come only from the catalog, so a card reads the same while the page loads and after, and as on the game's page (owner, 2026-10-09). `src/lib/games/catalog.js`, `GameCard.svelte` and `GameScene.svelte` are shared from lamsza to Játszótér and Szótár, and every list of games (Lámsza's home, Szótár's home, Játszótér's home and its "További játékok") draws `GameCard`. The card's art is the game's animated scene (`GameScene`, after words.com's game thumbnails: Szórejtő's tiles flip to LASKA, Kaptár's cells pop in around the A, Szókereső's grid of cells, where KACOR's cells turn amber one by one as if dragged and then white as found, the game's own selection and found cells (owner, 2026-10-09), Akasztófa's KALÁN drops into its blanks, Rovásfejtő's signs get their Latin letters, Tájszórejtvény's grid fills from the clue cell), white on the game's colour. On a card the scene stands finished and plays from the start only while the card is hovered or focused, and the start button's round, text-less version fades in (`game-start`); the card itself never moves. A game page's starter shows the same scene in a game-coloured panel, looping. Reduced motion: the finished scene, still (owner, 2026-10-09). `GameIcon.svelte` is no longer drawn; it stays shared until its drawings move to `AppIcon`. A home page shows Játszótér's enabled games in its order (`featuredGames`): Lámsza all of them, Szótár only those played with its Székely words (catalog `szekely_words`: Kaptár, Szórejtő, Szókereső, Akasztófa, Tájszórejtvény; not Rovásfejtő, which is played with proverbs in rovás). Rovásfejtő's rovás letters are drawn as paths (traced from Noto Sans Old Hungarian), so no app needs a rovás font for the icon. Owner's decision, 2026-10-09. |
| Game start button (`game-start`) | Játszótér's | One start button for every game, `src/lib/games/StartButton.svelte`, shared from lamsza to Játszótér and Szótár: on a game's page a red (`--szekely-red`) pill with the play triangle and "Kezdés" ("Folytatás" when it resumes a game in progress); on a game card the same triangle in a red circle, without text, with a thin white ring so it shows on every game colour. Every game's daily is one run per level and day: once the selected level's daily is done, on Napi kihívás the button says "Teljesítve" and is off, and the start screen opens on Gyakorlás (Kaptár's replays and Tájszórejtvény's Eredmény for a finished day are gone). Owner's decision, 2026-10-09. |
| Error wording (`si-errors`) | Játszótér's |  |
| Google client id source (`si-clientid`) | Lámsza's |  |

## Admin areas

| Item | Baseline | Note |
|---|---|---|
| Admin layout (`adm-layout`) | Other | Keep the admin app's own layout. Since 2026-10-07 it has three sections on one shell (`AdminShell`): `/` ("Vezérlőpult", unchanged), `/dictionary` and `/games` (WAYS_OF_WORKING R18). Each section's sidebar uses only shared icons: first an external link to the app (`external`), then its dashboard (`dashboard`), then one entry per feature; the dashboard cards mirror the sidebar entries with their counts. The header buttons switch between the three sections and use the shared toolbar buttons (`.btn.nav-btn`: pill, 16px icons, the current section in the active state). The section dashboards are titled "Vezérlőpult Szótár" and "Vezérlőpult Játszótér". Owner's decisions, 2026-10-07. |
| English labels (`adm-english`) | Admin's | Translate the admin's English labels to Hungarian. Clarified by the owner; overrides the Admin pick. |
| Admin sign-in gate (`adm-gate`) | Játszótér's | Since the admin move (R18, 2026-10-07) only the admin app has a gate. No user-facing app links to it (no Admin button, no admin-only links); admins open it by its URL. Owner's decision, 2026-10-07. |
| Network toolbar and footer on admin pages (`adm-chrome`) | Játszótér's | Szótár and Játszótér have no admin pages since the admin move (R18, 2026-10-07); the admin app keeps its own shell. Clarified by the owner. |
