/**
 * Public változásnapló. Keep the newest release first.
 * The footer version is PUBLIC_CHANGELOG[0].version.
 *
 * @type {Array<{
 *   version: string,
 *   date: string,
 *   items: Array<string | { lead: string, text: string }>,
 * }>}
 */
export const PUBLIC_CHANGELOG = [
    {
        version: "1.2.0",
        date: "2026. szeptember 24.",
        items: [
            "Eseménynaptár, helyszínekkel.",
            "Városok, falvak, megyék és történelmi székek.",
            "Google-belépés, a Fiók oldal és a felhasználói beállítások.",
            "Kedvenc helyek.",
            "Település beállítás: ha ki van választva, a kezdőlap időjárása és eseménysora ezt a települést használja.",
            "Bejegyzés átvétele és szerkesztése: fotó, nyitvatartás, értékelés, ha a bejegyzésnek gazdája van.",
            "Weboldal beküldése, és külön Weboldalak lista az indexen.",
            "Keresés szűrése szolgáltatásokra és weboldalakra.",
        ],
    },
    {
        version: "1.1.0",
        date: "2026. március 6.",
        items: [
            {
                lead: "Moduláris felépítés",
                text: "A háttérrendszer mostantól modulokra oszlik, így könnyebben fejleszthető és karbantartható.",
            },
            {
                lead: "Funkció-kapcsolók",
                text: "Bizonyos funkciók (pl. események, hírek) mostantól egyetlen kattintással kikapcsolhatóak a beállításaokban, ha nincs rájuk szükség.",
            },
            {
                lead: "Rendszertakarítás",
                text: "A belső kódstruktúra optimalizálva lett a gyorsabb és megbízhatóbb működés érdekében.",
            },
            "Helyreigazított automatikus tesztelési folyamat.",
        ],
    },
    {
        version: "1.0.0",
        date: "2026. március 1.",
        items: [
            "Teljes újraindítás - a Székely Gugel él és virul!",
            "Helyi kereső: kereshetsz orvosra, iskolára, mesteremberre és hivatalra",
            "Időjárás jelenlegi helyi adatok alapján",
            "Friss hírek erdélyi forrásokból",
            "Gyorslinkek rács",
            "Székely mondás naponta",
            "Sötét / Világos / Rendszer alapú témavalásztó",
            "Szolgáltatások oldal kategóriák szerint szűrhető",
        ],
    },
    {
        version: "0.9.5",
        date: "2026. február 28.",
        items: [
            "Alap elrendezés és fejléc kialakítva",
            "Keresősáv hozzáadva Google, Bing, DuckDuckGo és Yandex gombokkal",
            "Székely mondások megjelennek a főoldal alján",
            "Teljesen reszponzív, mobil-barát elrendezés",
        ],
    },
    {
        version: "0.9.0",
        date: "2026. február 20.",
        items: [
            "Kezdeti dizájn és alapok letéve",
            "Székely színvilág és tipográfia meghatározva",
            "SvelteKit + Go backend elindítva",
        ],
    },
    {
        version: "0.8.0",
        date: "2026. február 12.",
        items: ["PostgreSQL adatbázis felállítva", "Első API végpontok elkészültek"],
    },
    {
        version: "0.7.0",
        date: "2026. február 1.",
        items: ["A projekt elindult - „Na lámsza, csináljuk meg!\""],
    },
];

export const APP_VERSION = PUBLIC_CHANGELOG[0].version;
