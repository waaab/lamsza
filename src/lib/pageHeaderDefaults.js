/** Ha az API nem elérhető, ezek a cím / bevezető (slug = `pages.slug`). */
export const PAGE_HEADER_FALLBACK = /** @type {Record<string, { title: string, greeting: string }>} */ ({
    home: {
        title: "Na Lámsza!",
        greeting: "Erdélyi magyar startlap és kereső. Az internet székely kapuja.",
    },
    esemenyek: {
        title: "Székelyföldi események",
        greeting: "Válogass a legfrissebb székelyföldi események közül.",
    },
    hirek: {
        title: "Friss hírek erdélyi forrásból",
        greeting: "Helyi hírcsatornák legfrissebb hírei időrendben.",
    },
    megyek: {
        title: "Székelyföldi Megyék",
        greeting: "Válassz megyét a települések és tartalmak böngészéséhez.",
    },
    varosok: {
        title: "Székelyföldi Városok",
        greeting: "Székelyföldi városok listája megyénként.",
    },
    falvak: {
        title: "Székelyföldi Falvak",
        greeting: "Falvak és községek listája megyénként.",
    },
    szekek: {
        title: "Székelyföld történelmi székei",
        greeting: "A székely székek és a hozzájuk kapcsolódó megyék.",
    },
    index: {
        title: "Index",
        greeting: "Minden, ami helyi, egy helyen: szakemberek, intézmények, szolgáltatások.",
    },
    "index/szolgaltatasok": {
        title: "Szolgáltatások",
        greeting: "Az index szolgáltatástípusú bejegyzései — nem a teljes címtár.",
    },
    terkep: {
        title: "Székelyföld Térkép",
        greeting: "Hamarosan érkezik az interaktív térképünk helyi adatokkal!",
    },
    valtozasnaplo: {
        title: "Változásnapló",
        greeting: "Újítások, javítások — emberi nyelven.",
    },
    iranyelvek: {
        title: "Irányelvek",
        greeting: "Adatvédelem, sütik és felhasználási feltételek — összefoglaló.",
    },
    "iranyelvek/sutik": {
        title: "Sütik",
        greeting: "Hogyan használjuk a sütiket és mire valók.",
    },
    "iranyelvek/feltetelek": {
        title: "Feltételek",
        greeting: "A szolgáltatás igénybevételének feltételei.",
    },
});
