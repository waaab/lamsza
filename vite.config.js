import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		allowedHosts: ['.test', 'localhost', '127.0.0.1', '::1'],
		proxy: {
			'/api': {
				target: 'http://localhost:3000',
				changeOrigin: true
			}
		}
	}
});

console.log('Vite config loaded — allowedHosts:', ['.test', 'localhost', '127.0.0.1', '::1']);