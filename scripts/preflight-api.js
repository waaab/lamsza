/**
 * Build preflight: refuse to build when the Go backend is not reachable.
 *
 * The build itself no longer needs the backend - component fetches are gated on
 * canReachApi() (see src/lib/api.js), so `dist/` comes out byte-identical with the
 * backend up or down. This gate is about releases, not about `dist/`: a release is
 * cut against a running stack, and a build that cannot see the backend means the
 * machine cutting it is not looking at the system it claims to ship.
 *
 * Exit 1 with a message naming the port. Silence and exit 0 is what let the old
 * ECONNREFUSED noise go unnoticed.
 */
import process from "node:process";
import { pathToFileURL } from "node:url";

/** Health probe timeout. Long enough for a cold local backend, short enough to not hang CI. */
const TIMEOUT_MS = 5000;

/**
 * Where the preflight looks for the backend. Mirrors the SSR branch of
 * getApiBase() so the gate checks the same host the build would have called.
 * @param {Record<string, string | undefined>} [env]
 * @returns {string} origin with no trailing slash
 */
export function resolveApiBase(env = process.env) {
    const base = env.API_BASE_URL || env.VITE_API_BASE_URL || "http://127.0.0.1:3001";
    return String(base).replace(/\/$/, "");
}

/**
 * The failure message. Names the URL and both ways forward, because the person
 * reading it is usually mid-release and does not want to go read this file.
 * @param {string} base
 * @param {string} detail
 */
export function failureMessage(base, detail) {
    return [
        "",
        `  Build stopped: the backend did not answer at ${base}/api/health`,
        `  ${detail}`,
        "",
        "  Start the Go backend on port 3001 (~/projects/lamsza-network/start-lamsza-network.sh start), or point the",
        "  build at a live one with API_BASE_URL=https://... npm run build",
        "",
    ].join("\n");
}

/**
 * Probe /api/health. The endpoint returns 200 only when the database answers,
 * so a running process with a dead database correctly reads as down.
 * @param {string} base
 * @returns {Promise<{ok: true} | {ok: false, detail: string}>}
 */
export async function probeApi(base) {
    try {
        const response = await fetch(`${base}/api/health`, {
            signal: AbortSignal.timeout(TIMEOUT_MS),
            headers: { accept: "application/json" },
        });
        if (!response.ok) {
            return { ok: false, detail: `It replied HTTP ${response.status} instead of 200.` };
        }
        return { ok: true };
    } catch (error) {
        const reason = error instanceof Error ? error.message : String(error);
        return { ok: false, detail: `The request failed: ${reason}` };
    }
}

async function main() {
    const base = resolveApiBase();
    const result = await probeApi(base);
    if (!result.ok) {
        console.error(failureMessage(base, result.detail));
        process.exit(1);
    }
    console.log(`Preflight: backend is up at ${base}`);
}

/** Only run when invoked as a script, so the helpers above stay testable. */
if (process.argv[1] && import.meta.url === pathToFileURL(process.argv[1]).href) {
    await main();
}
