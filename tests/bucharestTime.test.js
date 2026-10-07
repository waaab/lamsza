import assert from "node:assert/strict";
import test from "node:test";
import { bucharestWallToMs, bucharestYMD, noteServerDate, serverNowMs } from "../src/lib/bucharestTime.js";
import { getEventStatus, referenceTodayYMD } from "../src/lib/eventStatus.js";

const utc = (s) => Date.parse(s);

test("Bucharest wall-clock times become the right instants, both seasons", () => {
    assert.equal(bucharestWallToMs(2026, 7, 15, 23, 59), utc("2026-07-15T20:59:00Z"), "summer, UTC+3");
    assert.equal(bucharestWallToMs(2026, 1, 15, 0, 0), utc("2026-01-14T22:00:00Z"), "winter, UTC+2");
    assert.equal(bucharestWallToMs(2026, 3, 29, 2, 30), utc("2026-03-29T00:30:00Z"), "before the spring jump");
    assert.equal(bucharestWallToMs(2026, 3, 29, 4, 30), utc("2026-03-29T01:30:00Z"), "after the spring jump");
    assert.equal(bucharestWallToMs(2026, 10, 25, 5, 0), utc("2026-10-25T03:00:00Z"), "after the autumn change");
});

test("the Bucharest day turns at Bucharest midnight, not UTC midnight", () => {
    assert.equal(bucharestYMD(utc("2026-07-15T20:59:59Z")), "2026-07-15");
    assert.equal(bucharestYMD(utc("2026-07-15T22:30:00Z")), "2026-07-16", "22:30 UTC in summer is the next day");
    assert.equal(bucharestYMD(utc("2026-01-15T21:59:59Z")), "2026-01-15");
    assert.equal(bucharestYMD(utc("2026-01-15T22:00:00Z")), "2026-01-16");
    assert.equal(bucharestYMD(utc("2026-03-29T20:59:59Z")), "2026-03-29", "the short spring day");
    assert.equal(bucharestYMD(utc("2026-03-29T21:00:00Z")), "2026-03-30");
    assert.equal(bucharestYMD(utc("2026-10-25T21:59:59Z")), "2026-10-25", "the long autumn day");
    assert.equal(bucharestYMD(utc("2026-10-25T22:00:00Z")), "2026-10-26");
});

test("a one-day event ends at Bucharest midnight, whatever the visitor's zone", () => {
    const ev = { start_date: "2026-07-15", end_date: "2026-07-15", start_time: "00:00:00", end_time: "00:00:00" };
    assert.notEqual(getEventStatus(ev, new Date(utc("2026-07-15T20:30:00Z"))), "ended", "23:30 in Bucharest");
    assert.equal(getEventStatus(ev, new Date(utc("2026-07-15T21:30:00Z"))), "ended", "00:30 the next day in Bucharest");
    const winter = { start_date: "2026-01-15", end_date: "2026-01-15" };
    assert.notEqual(getEventStatus(winter, new Date(utc("2026-01-15T21:30:00Z"))), "ended");
    assert.equal(getEventStatus(winter, new Date(utc("2026-01-15T22:30:00Z"))), "ended");
});

test("today follows the server's clock from the Date header, not this device's", () => {
    noteServerDate("Wed, 15 Jul 2026 22:30:00 GMT");
    assert.equal(referenceTodayYMD(), "2026-07-16");
    assert.ok(Math.abs(serverNowMs() - utc("2026-07-15T22:30:00.5Z")) < 2000);
    noteServerDate("not a date");
    assert.equal(referenceTodayYMD(), "2026-07-16", "a bad header changes nothing");
});
