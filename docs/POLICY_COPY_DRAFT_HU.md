# Policy page copy — draft for owner review

**Status: DRAFT. Not approved. Do not paste into the admin app until the owner signs it off.**

This file holds the Hungarian text for the four `/iranyelvek` pages. It is written
against what the code actually does today (see *Evidence* at the bottom), not against
a generic template. Every `【…】` marker is a fact only the owner can supply.

How to publish, once approved: open the admin app, edit the `pages` row for each slug
below, paste the HTML/Markdown body into the `content` field. The pages read that field
at runtime.

| Page | Slug | Status |
| --- | --- | --- |
| Irányelvek (index) | `iranyelvek` | draft below |
| Adatvédelem | `iranyelvek/adatvedelem` | draft below |
| Sütik | `iranyelvek/sutik` | draft below |
| Feltételek | `iranyelvek/feltetelek` | draft below |

## Owner decisions needed before publishing

1. **Adatkezelő** — the legal identity behind Lámsza: full name or company name,
   registered address, registration/tax number. A privacy policy without a named
   controller is not valid under the GDPR.
2. **Contact address** for data requests (e.g. `adatvedelem@lamsza.com`). It must be a
   mailbox somebody actually reads.
3. **Jurisdiction and supervisory authority.** The draft assumes Romania and names
   **ANSPDCP**. If the controller is established in Hungary instead, this becomes
   **NAIH** and the address changes.
4. **Hosting location.** The droplet is at DigitalOcean; confirm the region so the
   draft can say whether data stays inside the EEA.
5. **Server log retention.** Nginx access logs hold visitor IP addresses. Confirm the
   logrotate window (nginx default is 14 days) so the retention table is truthful.
6. **Account deletion.** There is no delete-my-account endpoint in the backend today.
   Either (a) publish the draft as written — erasure on e-mail request, handled by hand —
   or (b) build the endpoint first. Option (a) is lawful but means somebody has to
   answer those mails within 30 days.

---

## 1. `iranyelvek` — Irányelvek (index page)

```html
<p>
  Itt gyűjtöttük össze, hogyan működik a Lámsza, mit kezdünk az adataiddal, és
  milyen feltételekkel használhatod az oldalt. Három dokumentum, mindegyik rövid.
</p>

<h2>Adatvédelmi tájékoztató</h2>
<p>
  Milyen adatot kezelünk rólad, miért, milyen jogalapon, mennyi ideig, és hogyan
  kérheted a törlésüket.
  <a href="/iranyelvek/adatvedelem">Adatvédelmi tájékoztató &rarr;</a>
</p>

<h2>Sütik és helyi tárolás</h2>
<p>
  Melyik sütit és melyik böngészőben tárolt adatot mire használjuk, és melyiket
  kapcsolhatod ki.
  <a href="/iranyelvek/sutik">Süti tájékoztató &rarr;</a>
</p>

<h2>Felhasználási feltételek</h2>
<p>
  Mire használhatod a Lámszát, mit vállalunk, és mit nem.
  <a href="/iranyelvek/feltetelek">Felhasználási feltételek &rarr;</a>
</p>

<h2>Kapcsolat</h2>
<p>
  Adatvédelmi kérdésekben: <a href="mailto:【e-mail】">【e-mail】</a>
</p>
```

---

## 2. `iranyelvek/adatvedelem` — Adatvédelmi tájékoztató

```html
<p><strong>Hatályos: 【dátum】. Utolsó módosítás: 【dátum】.</strong></p>

<h2>1. Ki kezeli az adataidat?</h2>
<p>
  Adatkezelő: <strong>【adatkezelő neve】</strong><br>
  Székhely: 【cím】<br>
  Nyilvántartási szám: 【szám】<br>
  E-mail: <a href="mailto:【e-mail】">【e-mail】</a>
</p>
<p>
  A Lámsza egy székelyföldi startlap, kereső és helyi címtár. Adatvédelmi
  kérdésben a fenti e-mail címen érsz el minket.
</p>

<h2>2. Mit nem teszünk</h2>
<p>
  Nem használunk analitikai vagy reklámkövetést. Nincs Google Analytics, nincs
  hirdetési pixel, nincs profilalkotás, és nem adunk el adatot senkinek.
</p>

<h2>3. Milyen adatot kezelünk, miért és milyen jogalapon?</h2>

<h3>3.1 Böngészés bejelentkezés nélkül</h3>
<p>
  A böngészéshez nem kell fiók. A webszerver technikai naplót vezet: IP-cím,
  időpont, a kért oldal, a böngésző típusa. Ez az üzemeltetéshez és a
  visszaélések kivédéséhez kell.
</p>
<ul>
  <li><strong>Jogalap:</strong> jogos érdek — GDPR 6. cikk (1) f) pont.</li>
  <li><strong>Megőrzés:</strong> 【naplómegőrzés, pl. 14 nap】, utána automatikusan törlődik.</li>
</ul>

<h3>3.2 Bejelentkezés Google-fiókkal</h3>
<p>
  A bejelentkezés Google-fiókkal történik. A Google-tól a következőket kapjuk meg
  és tároljuk: a Google-fiók állandó azonosítója, e-mail cím, név, keresztnév,
  családi név, profilkép hivatkozása, nyelvi beállítás.
</p>
<ul>
  <li><strong>Jogalap:</strong> a szolgáltatás nyújtása, azaz szerződés teljesítése — GDPR 6. cikk (1) b) pont.</li>
  <li><strong>Megőrzés:</strong> amíg a fiókod él. Törlési kérésre töröljük.</li>
</ul>
<p>
  A bejelentkezési űrlapot a Google szolgáltatja. A Google saját adatkezelő ebben
  a lépésben; a Google adatkezeléséről a
  <a href="https://policies.google.com/privacy" target="_blank" rel="noopener">Google adatvédelmi irányelvei</a>
  szólnak. A Google bejelentkezési szkript csak akkor töltődik be, amikor
  megnyitod a bejelentkezési ablakot — addig nem.
</p>

<h3>3.3 Bejelentkezési munkamenet</h3>
<p>
  Bejelentkezés után egy munkamenet-sütit kapsz (<code>lamsza_session</code>).
  A szerveren csak a süti értékének titkosított lenyomatát tároljuk, magát az
  értéket nem.
</p>
<ul>
  <li><strong>Jogalap:</strong> szerződés teljesítése — GDPR 6. cikk (1) b) pont.</li>
  <li><strong>Megőrzés:</strong> 30 nap, vagy kijelentkezésig.</li>
</ul>

<h3>3.4 A saját beállításaid</h3>
<p>
  Ha bejelentkezve állítod be, a szerveren tároljuk: megjelenítési mód (világos
  vagy sötét), a gyorslinkek száma, a saját gyorslinkjeid, a kedvenceid, és a
  megnyitott címtár-bejegyzések előzménye.
</p>
<ul>
  <li><strong>Jogalap:</strong> szerződés teljesítése — GDPR 6. cikk (1) b) pont.</li>
  <li><strong>Megőrzés:</strong> amíg a fiókod él, vagy amíg te magad törlöd őket.</li>
</ul>

<h3>3.5 Amit beküldesz</h3>
<p>
  Ha bejegyzést javasolsz az indexbe, meglévő bejegyzést igényelsz magadnak,
  véleményt írsz, vagy látnivalót javasolsz, akkor azt kezeljük, amit beírsz:
  a bejegyzés adatait, a kapcsolattartási adatokat és a fiókodhoz tartozó
  azonosítót. Az index nyilvános, tehát amit beküldesz és jóváhagyunk, az
  mindenki számára látható lesz.
</p>
<ul>
  <li><strong>Jogalap:</strong> szerződés teljesítése és jogos érdek — GDPR 6. cikk (1) b) és f) pont.</li>
  <li><strong>Megőrzés:</strong> amíg a bejegyzés az indexben van.</li>
</ul>

<h3>3.6 A böngészőben tárolt adatok</h3>
<p>
  Néhány adat nem jut el hozzánk, csak a saját böngésződben marad. Ezek
  felsorolását és a kikapcsolásuk módját a
  <a href="/iranyelvek/sutik">süti tájékoztató</a> tartalmazza.
</p>

<h2>4. Kinek adjuk át?</h2>
<p>Nincs adatkereskedelem. A következő feldolgozók és szolgáltatók vesznek részt:</p>
<ul>
  <li><strong>【hosting szolgáltató】</strong> — a szerver üzemeltetése, 【régió】.</li>
  <li><strong>Google Ireland Limited</strong> — a Google-fiókos bejelentkezés.</li>
  <li>Hatóság részére, ha jogszabály kötelez rá.</li>
</ul>
<p>
  Az időjárás- és híradatokat a szerverünk kéri le a forrásoktól. Ezekhez a
  szolgáltatókhoz a te adataid nem jutnak el, csak a település, amelyre az
  időjárást kérjük.
</p>

<h2>5. Milyen jogaid vannak?</h2>
<p>A GDPR szerint kérheted:</p>
<ul>
  <li>a tájékoztatást arról, milyen adatot kezelünk rólad (hozzáférés);</li>
  <li>a hibás adat javítását;</li>
  <li>az adataid törlését;</li>
  <li>a kezelés korlátozását;</li>
  <li>a jogos érdeken alapuló kezelés elleni tiltakozást;</li>
  <li>az adataid kiadását gépi formátumban (adathordozhatóság);</li>
  <li>a hozzájárulás visszavonását, ha hozzájárulás volt a jogalap.</li>
</ul>
<p>
  Írj a <a href="mailto:【e-mail】">【e-mail】</a> címre. Legkésőbb egy hónapon belül
  válaszolunk. A fiók és a hozzá tartozó adatok törlése jelenleg kérésre,
  kézzel történik.
</p>

<h2>6. Panasz</h2>
<p>
  Ha úgy látod, hogy jogsértően kezeljük az adataidat, panaszt tehetsz a
  felügyeleti hatóságnál:<br>
  <strong>【Autoritatea Națională de Supraveghere a Prelucrării Datelor cu Caracter
  Personal (ANSPDCP)】</strong><br>
  【cím】 — <a href="https://www.dataprotection.ro" target="_blank" rel="noopener">dataprotection.ro</a>
</p>
<p>Bírósághoz is fordulhatsz.</p>

<h2>7. Gyermekek</h2>
<p>
  A Lámsza nem gyermekeknek szól. Tudatosan nem gyűjtünk adatot 16 év alatti
  személyekről. Ha ilyen adat mégis hozzánk kerül, töröljük.
</p>

<h2>8. Módosítás</h2>
<p>
  Ha változtatunk ezen a tájékoztatón, a tetején lévő dátum is változik.
  Érdemi változásról a bejelentkezett felhasználókat külön is értesítjük.
</p>
```

---

## 3. `iranyelvek/sutik` — Süti tájékoztató

> **Keep this list in sync with `src/lib/stores/consent.js`.** The banner is only
> honest if the two match. If a category or a storage key changes in the code, this
> page has to change with it.

```html
<p><strong>Hatályos: 【dátum】.</strong></p>

<p>
  A Lámsza kétféle helyen tárol adatot a böngésződben: sütiben (ezt a szerver is
  látja) és helyi tárolóban (<code>localStorage</code>, <code>sessionStorage</code>
  — ez nem hagyja el a gépedet). Mindkettőt itt soroljuk fel.
</p>

<h2>1. Szükséges — nem kapcsolható ki</h2>
<p>Enélkül az oldal nem működik vagy elfelejti, amit te magad állítottál be.</p>
<table>
  <thead>
    <tr><th>Név</th><th>Hol</th><th>Mire</th><th>Meddig</th></tr>
  </thead>
  <tbody>
    <tr>
      <td><code>lamsza_session</code></td><td>süti</td>
      <td>Bejelentkezési munkamenet. HttpOnly, így a szkriptek nem olvassák.</td>
      <td>30 nap</td>
    </tr>
    <tr>
      <td><code>lamsza_auth_session</code></td><td>sessionStorage</td>
      <td>A bejelentkezett felhasználó adatai, hogy ne kelljen minden oldalon újra lekérni.</td>
      <td>a lap bezárásáig</td>
    </tr>
    <tr>
      <td><code>theme</code></td><td>localStorage</td>
      <td>A világos vagy sötét mód, amit te választottál.</td>
      <td>amíg nem törlöd</td>
    </tr>
    <tr>
      <td><code>user_quick_links</code></td><td>localStorage</td>
      <td>A saját gyorslinkjeid a főoldalon.</td>
      <td>amíg nem törlöd</td>
    </tr>
    <tr>
      <td><code>quick_links_display_count</code></td><td>localStorage</td>
      <td>Hány gyorslinket jelenítünk meg.</td>
      <td>amíg nem törlöd</td>
    </tr>
    <tr>
      <td><code>lamsza_cookie_consent</code></td><td>localStorage</td>
      <td>Ez a döntés, amit az alábbi kategóriákról hoztál.</td>
      <td>amíg nem törlöd</td>
    </tr>
  </tbody>
</table>

<h2>2. Kényelmi — kikapcsolható</h2>
<table>
  <thead>
    <tr><th>Név</th><th>Hol</th><th>Mire</th><th>Meddig</th></tr>
  </thead>
  <tbody>
    <tr>
      <td><code>lamsza_entry_history</code></td><td>localStorage</td>
      <td>A legutóbb megnyitott címtár-bejegyzések, hogy könnyen visszatalálj.</td>
      <td>amíg nem törlöd</td>
    </tr>
  </tbody>
</table>

<h2>3. Gyorsítótár — kikapcsolható</h2>
<p>
  Ezek csak másolatok a már letöltött tartalomról, hogy az oldal gyorsabban
  nyíljon és kevesebb adatot kelljen letölteni. Nem azonosítanak téged.
</p>
<table>
  <thead>
    <tr><th>Név</th><th>Hol</th><th>Mire</th><th>Meddig</th></tr>
  </thead>
  <tbody>
    <tr>
      <td><code>hirek_cache</code>, <code>news_cache…</code></td><td>localStorage</td>
      <td>A hírlista másolata.</td><td>amíg nem törlöd</td>
    </tr>
    <tr>
      <td><code>weather_cache_…</code></td><td>localStorage</td>
      <td>Az időjárás-adatok másolata településenként.</td><td>amíg nem törlöd</td>
    </tr>
    <tr>
      <td><code>promoted_links_cache</code></td><td>localStorage</td>
      <td>A kiemelt linkek másolata.</td><td>amíg nem törlöd</td>
    </tr>
  </tbody>
</table>

<h2>4. Nincs követés</h2>
<p>
  Nem használunk analitikai, hirdetési vagy közösségi média követőt. Nincs Google
  Analytics, nincs Facebook-pixel, nincs harmadik féltől származó követőszkript.
</p>

<h2>5. Harmadik fél</h2>
<p>
  Egyetlen külső szkriptet töltünk be: a Google bejelentkezési űrlapját, és azt is
  csak akkor, amikor megnyitod a bejelentkezési ablakot. Ekkor a Google saját sütit
  helyezhet el a saját domainjén. Lásd a
  <a href="https://policies.google.com/technologies/cookies" target="_blank" rel="noopener">Google süti tájékoztatóját</a>.
</p>
<p>
  A keresőgombok (Google, Bing, DuckDuckGo) sima hivatkozások: csak akkor
  irányítanak az adott keresőhöz, ha rákattintasz. Addig semmit nem töltünk be tőlük.
</p>

<h2>6. Hogyan állíthatod át?</h2>
<p>
  A lap alján a <strong>Süti beállítások</strong> hivatkozással bármikor módosíthatod
  a döntésedet. Ha kikapcsolsz egy kategóriát, az ahhoz tartozó adatokat töröljük a
  böngésződből. A böngésző saját beállításaiban is törölhetsz minden sütit és helyi
  tárolót — ekkor viszont a saját beállításaid is elvesznek.
</p>

<h2>7. További tudnivalók</h2>
<p>
  Hogy mit kezelünk a szerveren, azt az
  <a href="/iranyelvek/adatvedelem">adatvédelmi tájékoztató</a> írja le.
</p>
```

---

## 4. `iranyelvek/feltetelek` — Felhasználási feltételek

```html
<p><strong>Hatályos: 【dátum】.</strong></p>

<h2>1. Ki üzemelteti és mire jó?</h2>
<p>
  A Lámszát <strong>【üzemeltető neve】</strong> üzemelteti. A Lámsza egy székelyföldi
  startlap: kereső, helyi címtár (index), hírgyűjtő, események, időjárás, valamint a
  <a href="https://szotar.lamsza.com">szótár</a> és a
  <a href="https://jatszoter.lamsza.com">játszótér</a>. A szolgáltatás ingyenes.
</p>
<p>
  Az oldal használatával elfogadod ezeket a feltételeket. Ha nem fogadod el, ne
  használd az oldalt.
</p>

<h2>2. Fiók</h2>
<ul>
  <li>Böngészéshez nem kell fiók. Beküldéshez, kedvencekhez és saját gyorslinkekhez kell.</li>
  <li>A bejelentkezés Google-fiókkal történik. A Google-fiókod biztonságáért te felelsz.</li>
  <li>Egy személy egy fiókot használ. Más nevében nem jelentkezhetsz be.</li>
  <li>Ha megszegi valaki ezeket a feltételeket, a fiókját felfüggeszthetjük vagy törölhetjük.</li>
</ul>

<h2>3. Mit küldhetsz be?</h2>
<p>
  Beküldhetsz címtár-bejegyzést, igényelhetsz meglévő bejegyzést, írhatsz véleményt,
  javasolhatsz látnivalót. Amit beküldesz, azért te felelsz. Vállalod, hogy:
</p>
<ul>
  <li>a megadott adatok a tudomásod szerint valósak;</li>
  <li>jogosult vagy a beküldésre — ha egy bejegyzést magadnak igényelsz, valóban hozzá tartozol;</li>
  <li>nem küldesz be jogsértő, megtévesztő, gyűlölködő vagy mások jogát sértő tartalmat;</li>
  <li>nem küldesz be reklámot a címtár rendeltetésén kívül.</li>
</ul>
<p>
  A beküldött tartalmat moderáljuk. Jóváhagyás nélkül nem jelenik meg, és
  bármikor módosíthatjuk vagy eltávolíthatjuk, külön indoklás nélkül is.
</p>
<p>
  A beküldéssel nem korlátozott, díjmentes engedélyt adsz nekünk arra, hogy a
  tartalmat a Lámszán megjelenítsük, tároljuk és a megjelenítéshez szükséges
  mértékben szerkesszük. A szerzői jogod a tiéd marad.
</p>

<h2>4. Idegen tartalom és hivatkozások</h2>
<p>
  A híreket külső hírforrásokból gyűjtjük be, és sok kimenő hivatkozást
  tartalmazunk. Ezekre nincs ráhatásunk. A külső oldalak tartalmáért,
  elérhetőségéért és adatkezeléséért nem felelünk. Egy hivatkozás nem jelenti
  azt, hogy a tartalmát jóváhagyjuk.
</p>
<p>
  A címtár, a szótár és az időjárás adatai tájékoztató jellegűek. Pontosságukra
  nem vállalunk garanciát. Döntést ne csak ezekre alapozz.
</p>

<h2>5. Hogyan nem használhatod?</h2>
<ul>
  <li>Nem terhelheted túl a szolgáltatást automatizált kérésekkel.</li>
  <li>Nem gyűjtheted le a tartalmat gépi úton az engedélyünk nélkül.</li>
  <li>Nem próbálhatsz hozzáférni mások fiókjához vagy a rendszer nem nyilvános részeihez.</li>
  <li>Nem kerülheted meg a biztonsági vagy moderálási megoldásokat.</li>
</ul>

<h2>6. Szellemi alkotások</h2>
<p>
  A Lámsza neve, kinézete, szerkezete és az általunk készített tartalom a mi
  oltalmunk alatt áll. Rövid idézet a forrás és a visszahivatkozás megjelölésével
  rendben van. Ezen túl kérdezz rá.
</p>

<h2>7. Felelősség</h2>
<p>
  A szolgáltatás „ahogy van" alapon működik. Nem vállalunk garanciát a folyamatos
  elérhetőségre, a hibamentességre, sem arra, hogy minden adat pontos. Az
  ingyenességre tekintettel a felelősségünket a jogszabály által megengedett
  legszűkebb körre korlátozzuk. A szándékos károkozásért és az életet, testi
  épséget, egészséget sértő szerződésszegésért való felelősséget nem zárjuk ki.
</p>

<h2>8. Adatvédelem</h2>
<p>
  Az adataid kezelését az
  <a href="/iranyelvek/adatvedelem">adatvédelmi tájékoztató</a> és a
  <a href="/iranyelvek/sutik">süti tájékoztató</a> írja le.
</p>

<h2>9. Módosítás és megszűnés</h2>
<p>
  Ezeket a feltételeket módosíthatjuk; a hatályos változat dátuma mindig a lap
  tetején áll. Érdemi változásnál a bejelentkezett felhasználókat értesítjük. A
  szolgáltatás egészét vagy egy részét bármikor megszüntethetjük.
</p>

<h2>10. Alkalmazandó jog</h2>
<p>
  Ezekre a feltételekre 【Románia】 joga vonatkozik, a jogviták elbírálására a
  【illetékes bíróság】 jogosult. Ha fogyasztó vagy, a lakóhelyed szerinti
  kötelező fogyasztóvédelmi szabályok ettől függetlenül érvényesek.
</p>

<h2>11. Kapcsolat</h2>
<p><a href="mailto:【e-mail】">【e-mail】</a></p>
```

---

## Evidence — what the code actually does (verified 6 Oct 2026)

The copy above was written from these facts, not from a template. If the code changes,
check this list again.

**Cookies set by the server** — one only:
`lamsza_session`, `backend/internal/auth/auth.go:440` — `HttpOnly`, `Secure` behind TLS,
`SameSite=Lax`, TTL 30 days (`sessionTTL`, `auth.go:24`). The server stores only a
SHA-256 hash of the token (`sessions.token_hash`, `auth.go:123`).

**Browser storage** — the keys in `src/lib/stores/consent.js`, cross-checked against the
writers: `theme` (`stores/theme.js:36`), `user_quick_links` and
`quick_links_display_count` (`routes/(public)/+page.svelte:27`,
`quickLinksDisplay.js:29`), `lamsza_auth_session` (`stores/auth.js:48`),
`lamsza_entry_history` (`entryHistory.js:58`), news and weather and promoted-link caches
(`NewsWidget.svelte:48`, `WeatherWidget.svelte:97`, `hirek/+page.svelte:249`,
`+page.svelte:266`).

**Personal data in the database:** `users` (`auth.go:76`) holds `google_sub`, `email`,
`name`, `given_name`, `family_name`, `picture`, `locale`, `theme`, `quicklink_slots`,
`last_login_at`, `created_at`. Plus `sessions`, `user_links`, `user_favorites`,
`user_entry_history`, `entry_suggestions`, `attraction_suggestions`,
`attraction_contributors`, `entry_reviews`, `entry_members`.

**Third-party code in the browser:** exactly one script —
`https://accounts.google.com/gsi/client` (`GoogleSignIn.svelte:62`), injected on mount of
`GoogleSignIn`, which only renders inside `{#if open}` in `SignInDialog.svelte:28`. So it
loads when the visitor opens the sign-in dialog, not on page load. No Google Fonts, no
analytics, no tag manager — `src/app.html` pulls in no external resource.

**Outbound search buttons** (`SearchEngine.svelte:623-625`) are plain `<a>` links to
Google, Bing and DuckDuckGo. They load nothing until clicked.

**Weather and news providers** (`api.open-meteo.com`, `api.openweathermap.org`,
`api.weatherapi.com`, RSS feeds) are called from the Go backend, so the visitor's IP
never reaches them.

**Gap:** there is no account-deletion handler anywhere in `backend/internal/account/`.
Erasure requests have to be served by hand today. There is also no job that purges
expired rows from `sessions` — only sign-out deletes a row (`auth.go:433`).
