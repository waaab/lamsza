/** Fixed game titles and greetings. Seeded in SQL; not edited in admin.
 * `color` is a placeholder until each game gets its own color.
 * `category` is the game kind: word, geo, picture. One per game.
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
		color: '#d4a017'
	},
	{
		slug: 'szorejto',
		title_hu: 'Szórejtő',
		description_hu: 'Találd meg a székely szót hat próbálkozásból!',
		category: 'word',
		color: '#2f7d6d'
	},
	{
		slug: 'szokereso',
		title_hu: 'Szókereső',
		description_hu: 'Keresd meg a hét székely szót a rácsban.',
		category: 'word',
		color: '#3a6fbf'
	},
	{
		slug: 'akasztofa',
		title_hu: 'Akasztófa',
		description_hu: 'Találd ki a székely szót betűnként.',
		category: 'word',
		color: '#7a4ea3'
	},
	{
		slug: 'tajszorejtveny',
		title_hu: 'Tájszórejtvény',
		description_hu: 'Fejtsd meg a székely rejtvényt, és tanulj meg közben néhány tájszót!',
		category: 'word',
		color: '#a4473b',
		prelaunch: true
	}
];

/**
 * The games a list draws before `/api/games` answers, and when it cannot be
 * reached. A prelaunch game is left out, so a switched-off game never shows
 * for a moment and then disappears; the server's list adds it once enabled.
 */
export const LISTED_GAMES = GAMES.filter((game) => !game.prelaunch);

/** @param {string} slug */
export function gameBySlug(slug) {
	return GAMES.find((game) => game.slug === slug);
}

/**
 * The games a home page shows from Játszótér's `/api/games` answer: the
 * enabled ones in its order, each with the catalog's colour and icon and the
 * server's title and description where it has them, at most `count`. A game
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
			return meta ? { ...meta, title_hu: g.title_hu || meta.title_hu, description_hu: g.description_hu || meta.description_hu } : null;
		})
		.filter((/** @type {any} */ g) => g !== null)
		.slice(0, count);
	return list.length ? list : null;
}
