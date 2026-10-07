import { sveltekit } from '@sveltejs/kit/vite';
import { defineConfig } from 'vite';

export default defineConfig({
	plugins: [sveltekit()],
	server: {
		port: 5174,
		strictPort: true,
		allowedHosts: ['.test', 'localhost', '127.0.0.1', '::1'],
		proxy: {
			'/api': {
				target: 'http://localhost:3001',
				changeOrigin: true
			}
		}
	}
});

