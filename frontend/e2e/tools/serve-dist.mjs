// Static server for the built dist with an SPA fallback. The visual project
// runs inside the pinned Playwright container where the host-built (darwin)
// native vite binaries cannot load — this replaces `vite preview` with
// dependency-free Node. API calls never reach it: Playwright route mocks
// intercept them in the browser.
import { createReadStream, existsSync, statSync } from 'node:fs';
import { createServer } from 'node:http';
import { extname, join, normalize } from 'node:path';
import { fileURLToPath } from 'node:url';

const DIST = fileURLToPath(new URL('../../dist', import.meta.url));
const PORT = Number(process.env.PORT ?? 8100);

const MIME = {
    '.css': 'text/css',
    '.html': 'text/html',
    '.ico': 'image/x-icon',
    '.js': 'text/javascript',
    '.json': 'application/json',
    '.png': 'image/png',
    '.svg': 'image/svg+xml',
    '.webp': 'image/webp',
    '.woff2': 'font/woff2',
};

const INDEX = join(DIST, 'index.html');

if (!existsSync(INDEX)) {
    console.error(`[serve-dist] no build at ${DIST} — run \`pnpm run build\` first`);
    process.exit(2);
}

createServer((request, response) => {
    // The WHATWG URL parser already resolves any `..` segments in the pathname.
    const pathname = normalize(new URL(request.url ?? '/', 'http://localhost').pathname);
    let filePath = join(DIST, pathname);

    if (!existsSync(filePath) || statSync(filePath).isDirectory()) {
        // A miss on an asset-like path (hashed chunk, image) must 404 loudly —
        // an index.html fallback there masks a broken build as text/html.
        if (extname(pathname)) {
            response.writeHead(404, { 'content-type': 'text/plain' });
            response.end(`not found: ${pathname}`);

            return;
        }

        filePath = INDEX;
    }

    response.writeHead(200, { 'content-type': MIME[extname(filePath)] ?? 'application/octet-stream' });
    createReadStream(filePath)
        .on('error', () => response.destroy())
        .pipe(response);
}).listen(PORT, () => console.log(`[serve-dist] ${DIST} on :${PORT}`));
