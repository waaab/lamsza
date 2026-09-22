import { test } from 'node:test';
import assert from 'node:assert/strict';
import {
	showListingPhotos,
	showListingHours,
	showListingDeliveryHours,
	showListingRatings,
	showListingTodayHours
} from '../src/lib/entryPublicExtras.js';

test('showListingPhotos returns false for unclaimed entry', () => {
	const entry = { claimed: false, photos: [{ id: 1 }] };
	assert.equal(showListingPhotos(entry), false);
});

test('showListingPhotos returns false for claimed entry with empty photos', () => {
	const entry = { claimed: true, photos: [] };
	assert.equal(showListingPhotos(entry), false);
});

test('showListingPhotos returns true for claimed entry with photos', () => {
	const entry = { claimed: true, photos: [{ id: 1 }] };
	assert.equal(showListingPhotos(entry), true);
});

test('showListingHours returns false for unclaimed entry', () => {
	const entry = { claimed: false, hours: { mon: { open: '09:00', close: '17:00' } } };
	assert.equal(showListingHours(entry), false);
});

test('showListingHours returns false for claimed entry with no hours', () => {
	const entry = { claimed: true, hours: null };
	assert.equal(showListingHours(entry), false);
});

test('showListingHours returns true for claimed entry with hours', () => {
	const entry = { claimed: true, hours: { mon: { open: '09:00', close: '17:00' } } };
	assert.equal(showListingHours(entry), true);
});

test('showListingDeliveryHours returns false for unclaimed entry', () => {
	const entry = { claimed: false, delivery_hours: { mon: { open: '09:00', close: '17:00' } } };
	assert.equal(showListingDeliveryHours(entry), false);
});

test('showListingDeliveryHours returns true for claimed entry with delivery hours', () => {
	const entry = { claimed: true, delivery_hours: { mon: { open: '09:00', close: '17:00' } } };
	assert.equal(showListingDeliveryHours(entry), true);
});

test('showListingRatings returns false for unclaimed entry', () => {
	const entry = { claimed: false, ratings_enabled: true };
	assert.equal(showListingRatings(entry), false);
});

test('showListingRatings returns false for claimed entry without ratings_enabled', () => {
	const entry = { claimed: true, ratings_enabled: false };
	assert.equal(showListingRatings(entry), false);
});

test('showListingRatings returns true for claimed entry with ratings_enabled', () => {
	const entry = { claimed: true, ratings_enabled: true };
	assert.equal(showListingRatings(entry), true);
});

test('showListingTodayHours returns false for unclaimed entry', () => {
	const entry = { claimed: false, hours: { mon: { open: '09:00', close: '17:00' } } };
	assert.equal(showListingTodayHours(entry), false);
});

test('showListingTodayHours returns true for claimed entry with hours', () => {
	const entry = { claimed: true, hours: { mon: { open: '09:00', close: '17:00' } } };
	assert.equal(showListingTodayHours(entry), true);
});
