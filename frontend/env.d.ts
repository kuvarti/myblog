/// <reference types="vite/client" />

interface ImportMetaEnv {
	// Backend API base URL, baked in at build time. Falls back to the local
	// dev backend when unset (see BaseAPI.service.ts).
	readonly VITE_API_BASE_URL?: string
}

interface ImportMeta {
	readonly env: ImportMetaEnv
}
