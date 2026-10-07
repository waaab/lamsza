import adapter from '@sveltejs/adapter-static';

/** @type {import('@sveltejs/kit').Config} */
// Szótár's API: Lámsza's home page reads the daily mondás from it (OPEN_ITEMS,
// Mondások). A production build lists only szotar.lamsza.com; a development
// build adds the local hosts, and VITE_SZOTAR_ORIGIN when it is set in the
// environment at build time (networkOrigins.js uses the same variable).
const szotarConnectOrigins = [
	'https://szotar.lamsza.com',
	...(process.env.NODE_ENV === 'production' ? [] : ['https://szotar.lamsza.test', 'http://localhost:5175']),
	...(process.env.VITE_SZOTAR_ORIGIN ? [process.env.VITE_SZOTAR_ORIGIN.trim().replace(/\/+$/, '')] : [])
];

const config = {
	kit: {
		appDir: 'app',
		// adapter-auto only supports some environments, see https://svelte.dev/docs/kit/adapter-auto for a list.
		// If your environment is not supported, or you settled on a specific environment, switch out the adapter.
		// See https://svelte.dev/docs/kit/adapters for more information about adapters.
		adapter: adapter({
			pages: 'dist',
			assets: 'dist',
			fallback: 'app.html',
			precompress: false,
			strict: false
		}),
		prerender: {
			handleUnseenRoutes: 'ignore'
		},
		// Content-Security-Policy. The pages are static, so SvelteKit writes this
		// into a <meta http-equiv> tag in every built page and hashes its own
		// inline scripts and styles for us.
		//
		// script-src 'self' is the part that matters: it is the second lock on the
		// Markdown XSS (src/lib/markdown.js is the first) and it blocks the
		// exfiltration step as well, because connect-src names the only hosts the
		// page may talk to. Keep both lists as short as the app allows.
		//
		// Nginx also serves security headers (docs/PRODUCTION_ENVIRONMENT_NOTES.md).
		// frame-ancestors cannot work from a meta tag, so X-Frame-Options there
		// stays the control for framing.
		csp: {
			mode: 'hash',
			directives: {
				'default-src': ['self'],
				'base-uri': ['self'],
				'object-src': ['none'],
				'form-action': ['self'],
				// Google Identity Services is injected by GoogleSignIn.svelte.
				'script-src': ['self', 'https://accounts.google.com/gsi/client'],
				'style-src': ['self', 'unsafe-inline', 'https://accounts.google.com/gsi/style'],
				// Entry, event and news pictures come from feeds and from Google
				// avatars, so the host cannot be listed one by one.
				'img-src': ['self', 'data:', 'https:'],
				'font-src': ['self', 'data:'],
				'connect-src': ['self', 'https://accounts.google.com/gsi/', ...szotarConnectOrigins],
				'frame-src': ['self', 'https://accounts.google.com/gsi/'],
				'manifest-src': ['self'],
				'worker-src': ['self']
			}
		}
	}
};

export default config;
