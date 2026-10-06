import test from "node:test";
import assert from "node:assert/strict";

import { canReachApi } from "../src/lib/api.js";

/**
 * The build must not need the Go backend. A component fetch started during
 * prerender is never awaited by the renderer, so it cannot change `dist/`, but a
 * dead backend on port 3001 makes `npm run build` print an ECONNREFUSED stack per
 * page and still exit 0. Gate component fetches on canReachApi() so the build
 * stays quiet and reproducible.
 */
test("canReachApi is false with no window (SSR / prerender)", () => {
	assert.equal(typeof globalThis.window, "undefined");
	assert.equal(canReachApi(), false);
});

test("canReachApi is true once a window exists (browser)", () => {
	globalThis.window = { location: { origin: "https://lamsza.com" } };
	try {
		assert.equal(canReachApi(), true);
	} finally {
		delete globalThis.window;
	}
});
