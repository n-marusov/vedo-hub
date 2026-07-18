import vue from "@vitejs/plugin-vue";
import { defineConfig } from "vite";

export default defineConfig({
	plugins: [vue()],
	server: {
		host: "0.0.0.0",
		port: 3000,
		proxy: {
			"/api/v1": {
				target: "http://localhost:3001",
				changeOrigin: true,
			},
		},
	},
	resolve: {
		alias: {
			"@": "/src",
		},
	},
});
