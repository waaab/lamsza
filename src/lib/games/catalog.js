/** Fixed game titles and greetings. Seeded in SQL; not edited in admin.
 * `color` is a placeholder until each game gets its own color.
 * `category` is the game kind: word, geo, picture. One per game.
 * `szekely_words`: the game is played with Szótár's Székely words (its words
 * come from the dictionary); Szótár's home lists only these. Rovásfejtő is
 * played with Székely proverbs in rovás, not dictionary words.
 * `prelaunch` marks a game whose row ships switched off: lists draw it only
 * once the server says it is enabled (LISTED_GAMES). Drop the flag at the
 * game's release.
 */

export const GAMES = [
	{
		slug: 'rovasfejto',
		title_hu: 'Rovásfejtő',
		description_hu: 'Fejtsd meg a székely közmondást a rovásírásból!',
		category: 'word',
		color: '#8c5a3c'
	},
	{
		slug: 'kaptar',
		title_hu: 'Kaptár',
		description_hu: 'Hány szót találsz ma?',
		category: 'word',
		szekely_words: true,
		color: '#d4a017'
	},
	{
		slug: 'szorejto',
		title_hu: 'Szórejtő',
		description_hu: 'Találd meg a székely szót hat próbálkozásból!',
		category: 'word',
		szekely_words: true,
		color: '#2f7d6d'
	},
	{
		slug: 'szokereso',
		title_hu: 'Szókereső',
		description_hu: 'Keresd meg a hét székely szót a rácsban.',
		category: 'word',
		szekely_words: true,
		color: '#3a6fbf'
	},
	{
		slug: 'akasztofa',
		title_hu: 'Akasztófa',
		description_hu: 'Találd ki a székely szót betűnként.',
		category: 'word',
		szekely_words: true,
		color: '#7a4ea3'
	},
	{
		slug: 'tajszorejtveny',
		title_hu: 'Tájszórejtvény',
		description_hu: 'Fejtsd meg a székely rejtvényt, és tanulj meg közben néhány tájszót!',
		category: 'word',
		szekely_words: true,
		color: '#a4473b'
	}
];

/**
 * The games a list draws before `/api/games` answers, and when it cannot be
 * reached. A prelaunch game is left out, so a switched-off game never shows
 * for a moment and then disappears; the server's list adds it once enabled.
 */
export const LISTED_GAMES = GAMES.filter((game) => !game.prelaunch);

/**
 * Whether a game is played with Szótár's Székely words (catalog
 * `szekely_words`); a game the catalog does not know is not.
 * @param {{ slug?: string, szekely_words?: boolean }} game
 */
export function playsWithSzekelyWords(game) {
	return Boolean(game?.szekely_words ?? gameBySlug(String(game?.slug ?? ''))?.szekely_words);
}

/** @param {string} slug */
export function gameBySlug(slug) {
	return GAMES.find((game) => game.slug === slug);
}

/**
 * The games a home page shows from Játszótér's `/api/games` answer: the
 * enabled ones in its order, at most `count`, each as the catalog has it
 * (colour, title and line). The catalog is the one source of a game's card
 * text, the same as on its game page, so a card reads the same while the
 * page loads and after (owner, 2026-10-09). A game
 * the catalog does not know is left out (it would have no icon). Null when
 * the answer lists none, so the caller keeps LISTED_GAMES.
 *
 * @param {unknown} answer `/api/games` JSON
 * @param {number} count
 */
export function featuredGames(answer, count) {
	const rows = Array.isArray(/** @type {any} */ (answer)?.games) ? /** @type {any} */ (answer).games : [];
	const list = rows
		.filter((/** @type {any} */ g) => g?.enabled)
		.sort((/** @type {any} */ a, /** @type {any} */ b) => (a.sort_order ?? 0) - (b.sort_order ?? 0))
		.map((/** @type {any} */ g) => {
			const meta = gameBySlug(g.slug);
			return meta ? { ...meta } : null;
		})
		.filter((/** @type {any} */ g) => g !== null)
		.slice(0, count);
	return list.length ? list : null;
}

/**
 * Lámsza's "Játszótér · Mai kihívások": the games from Játszótér's
 * `/api/games` answer that have a daily challenge today (`daily_today`), in
 * its order, as featuredGames gives them. A Játszótér that does not send the
 * field yet lists every game. An empty list when no game has one today; null
 * when the answer lists no games, so the caller keeps LISTED_GAMES.
 *
 * @param {unknown} answer `/api/games` JSON
 */
export function todaysChallenges(answer) {
	const rows = /** @type {any} */ (answer)?.games;
	if (!Array.isArray(rows) || !rows.some((/** @type {any} */ g) => g?.enabled)) return null;
	const today = rows.filter((/** @type {any} */ g) => g?.daily_today !== false);
	return featuredGames({ games: today }, Infinity) ?? [];
}
