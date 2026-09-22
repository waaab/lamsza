import assert from "node:assert/strict";
import test from "node:test";
import { buildImportPayload, meToAuthState } from "../src/lib/accountPrefs.js";

test("buildImportPayload copies browser prefs", () => {
    const payload = buildImportPayload({
        theme: "dark",
        slots: 9,
        links: [{ title: "RMDSZ", url: "https://rmdsz.ro", bg_color: "#fff" }],
        history: [{ slug: "kavezo", name: "Kávézó", category: "", location: "", photo: "" }],
    });
    assert.equal(payload.theme, "dark");
    assert.equal(payload.quicklink_slots, 9);
    assert.equal(payload.links.length, 1);
    assert.equal(payload.history[0].slug, "kavezo");
});

test("meToAuthState maps google fields", () => {
    const state = meToAuthState({
        name: "Anna",
        email: "anna@example.test",
        is_admin: false,
        picture: "https://example.test/a.jpg",
        given_name: "Anna",
        family_name: "Kiss",
        locale: "hu",
        google_sub: "sub-1",
        theme: null,
        quicklink_slots: null,
        prefs_imported_at: null,
    });
    assert.equal(state.loggedIn, true);
    assert.equal(state.picture, "https://example.test/a.jpg");
    assert.equal(state.theme, null);
});
