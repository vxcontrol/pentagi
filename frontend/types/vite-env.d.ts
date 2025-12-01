/// <reference types="vite/client" />

interface ImportMeta {
    readonly env: ImportMetaEnv;
}

interface ImportMetaEnv {
    readonly VITE_APP_API_ROOT: string;
    readonly VITE_APP_LOG_LEVEL: 'DEBUG' | 'ERROR' | 'INFO' | 'WARN';
    readonly VITE_APP_SESSION_KEY: string;
}
