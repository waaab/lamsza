import assert from 'node:assert/strict';
import { test } from 'node:test';
import { resolveNetworkOrigin, joinOriginPath } from '../src/lib/networkOrigins.js';

test('env override wins', () => {
	assert.equal(
		resolveNetworkOrigin('lamsza', {
			envOrigin: 'https://lamsza.test/',
			hostname: 'szotar.lamsza.com'
		}),
		'https://lamsza.test'
	);
});

test('hostname sniff selects .test hosts', () => {
	assert.equal(
		resolveNetworkOrigin('lamsza', { envOrigin: '', hostname: 'szotar.lamsza.test' }),
		'https://lamsza.test'
	);
	assert.equal(
		resolveNetworkOrigin('szotar', { envOrigin: '', hostname: 'lamsza.test' }),
		'https://szotar.lamsza.test'
	);
	assert.equal(
		resolveNetworkOrigin('jatszoter', { envOrigin: '', hostname: 'jatszoter.lamsza.test' }),
		'https://jatszoter.lamsza.test'
	);
});

test('localhost uses Vite strictPort origins', () => {
	// Admin 5173, Lámsza 5174, Szótár 5175, Játszótér 5176 since the port renumber.
	assert.equal(
		resolveNetworkOrigin('lamsza', { envOrigin: '', hostname: 'localhost' }),
		'http://localhost:5174'
	);
	assert.equal(
		resolveNetworkOrigin('szotar', { envOrigin: '', hostname: '127.0.0.1' }),
		'http://localhost:5175'
	);
	assert.equal(
		resolveNetworkOrigin('jatszoter', { envOrigin: '', hostname: 'localhost' }),
		'http://localhost:5176'
	);
});

test('production default when hostname is not local', () => {
	assert.equal(
		resolveNetworkOrigin('lamsza', { envOrigin: '', hostname: 'szotar.lamsza.com' }),
		'https://lamsza.com'
	);
	assert.equal(
		resolveNetworkOrigin('szotar', { envOrigin: '', hostname: '' }),
		'https://szotar.lamsza.com'
	);
});

test('joinOriginPath joins path', () => {
	assert.equal(joinOriginPath('https://lamsza.test', '/iranyelvek'), 'https://lamsza.test/iranyelvek');
	assert.equal(joinOriginPath('https://lamsza.test', ''), 'https://lamsza.test');
	assert.equal(joinOriginPath('https://lamsza.test/', 'iranyelvek'), 'https://lamsza.test/iranyelvek');
});
