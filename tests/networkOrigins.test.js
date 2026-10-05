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

test('production default when hostname is not .test', () => {
	assert.equal(
		resolveNetworkOrigin('lamsza', { envOrigin: '', hostname: 'szotar.lamsza.com' }),
		'https://lamsza.com'
	);
	assert.equal(
		resolveNetworkOrigin('szotar', { envOrigin: '', hostname: '' }),
		'https://szotar.lamsza.com'
	);
	assert.equal(
		resolveNetworkOrigin('jatszoter', { envOrigin: '', hostname: 'localhost' }),
		'https://jatszoter.lamsza.com'
	);
});

test('joinOriginPath joins path', () => {
	assert.equal(joinOriginPath('https://lamsza.test', '/iranyelvek'), 'https://lamsza.test/iranyelvek');
	assert.equal(joinOriginPath('https://lamsza.test', ''), 'https://lamsza.test');
	assert.equal(joinOriginPath('https://lamsza.test/', 'iranyelvek'), 'https://lamsza.test/iranyelvek');
});
