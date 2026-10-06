import test from "node:test";
import assert from "node:assert/strict";

import { failureMessage, probeApi, resolveApiBase } from "../scripts/preflight-api.js";

/**
 * The preflight is the only thing standing between a release and a build cut
 * against a backend nobody checked. These tests pin the two ways it can rot:
 * probing the wrong host, or treating a bad answer as a good one.
 */

test("resolveApiBase falls back to the local Go backend", () => {
	assert.equal(resolveApiBase({}), "http://127.0.0.1:3001");
});

test("resolveApiBase prefers API_BASE_URL and strips the trailing slash", () => {
	const base = resolveApiBase({
		API_BASE_URL: "https://lamsza.com/",
		VITE_API_BASE_URL: "http://127.0.0.1:9999",
	});
	assert.equal(base, "https://lamsza.com");
});

test("resolveApiBase uses VITE_API_BASE_URL when the server-only name is unset", () => {
	assert.equal(resolveApiBase({ VITE_API_BASE_URL: "https://staging.lamsza.com" }), "https://staging.lamsza.com");
});

test("failureMessage names the health URL and the port to start", () => {
	const message = failureMessage("http://127.0.0.1:3001", "The request failed: fetch failed");
	assert.match(message, /http:\/\/127\.0\.0\.1:3001\/api\/health/);
	assert.match(message, /port 3001/);
	assert.match(message, /API_BASE_URL=/);
});

test("probeApi fails on a closed port instead of throwing", async () => {
	// Port 1 is privileged and never listening, so this is a connection refusal.
	const result = await probeApi("http://127.0.0.1:1");
	assert.equal(result.ok, false);
	assert.match(result.detail, /The request failed/);
});

test("probeApi fails on a non-200 answer", async () => {
	const { createServer } = await import("node:http");
	const server = createServer((_req, res) => {
		res.writeHead(503, { "content-type": "application/json" });
		res.end('{"ok":false,"db":"down"}');
	});
	await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
	try {
		const { port } = /** @type {import("node:net").AddressInfo} */ (server.address());
		const result = await probeApi(`http://127.0.0.1:${port}`);
		assert.equal(result.ok, false);
		assert.match(result.detail, /HTTP 503/);
	} finally {
		await new Promise((resolve) => server.close(resolve));
	}
});

test("probeApi accepts a 200 answer", async () => {
	const { createServer } = await import("node:http");
	const server = createServer((req, res) => {
		assert.equal(req.url, "/api/health");
		res.writeHead(200, { "content-type": "application/json" });
		res.end('{"ok":true,"db":"up"}');
	});
	await new Promise((resolve) => server.listen(0, "127.0.0.1", resolve));
	try {
		const { port } = /** @type {import("node:net").AddressInfo} */ (server.address());
		assert.deepEqual(await probeApi(`http://127.0.0.1:${port}`), { ok: true });
	} finally {
		await new Promise((resolve) => server.close(resolve));
	}
});
