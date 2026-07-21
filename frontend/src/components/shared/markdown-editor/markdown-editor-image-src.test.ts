import { describe, expect, it } from 'vitest';

import { normalizeImageSrc, normalizeLinkUrl } from './markdown-editor-toolbar-url';

// data:image/svg+xml is rejected because SVG can carry script — unlike raster data: URLs, which pass through.
describe('normalizeImageSrc — prepend https to scheme-less src, validate protocol', () => {
    it.each([
        ['example.com/a.png', 'https://example.com/a.png'],
        ['//cdn.example.com/a.png', 'https://cdn.example.com/a.png'],
        ['  example.com/a.png  ', 'https://example.com/a.png'],
        ['http://example.com/a.png', 'http://example.com/a.png'],
        ['https://example.com/a.png?w=1', 'https://example.com/a.png?w=1'],
        ['data:image/png;base64,AAAA', 'data:image/png;base64,AAAA'],
        ['data:image/webp;base64,AAAA', 'data:image/webp;base64,AAAA'],
    ])('normalizes %s → %s', (input, expected) => {
        expect(normalizeImageSrc(input)).toBe(expected);
    });

    it.each([
        'javascript:alert(1)',
        'data:text/html,<script>alert(1)</script>',
        'data:application/pdf;base64,AA',
        'data:image/svg+xml;utf8,<svg onload=alert(1)/>',
        'data:image/svg+xml;base64,AAAA',
        'vbscript:msgbox(1)',
        'http://', // malformed → no host
        '/uploads/logo.png', // root-relative path — must NOT become https://uploads/logo.png
        '/a.png',
        '',
    ])('rejects %s', (input) => {
        expect(normalizeImageSrc(input)).toBeNull();
    });
});

// A scheme-less input like "example.com" would otherwise persist as a relative href the browser resolves against
// the current origin — normalizeLinkUrl prepends https:// to force it absolute.
describe('normalizeLinkUrl — prepend https to scheme-less input, validate protocol', () => {
    it.each([
        ['example.com', 'https://example.com'],
        ['www.example.com', 'https://www.example.com'],
        ['example.com:8080', 'https://example.com:8080'],
        ['localhost:3000', 'https://localhost:3000'],
        ['//evil.com', 'https://evil.com'],
        ['  example.com  ', 'https://example.com'],
        ['http://example.com', 'http://example.com'],
        ['https://example.com/path?a=1|2', 'https://example.com/path?a=1|2'],
        ['mailto:a@b.com', 'mailto:a@b.com'],
        ['tel:+123', 'tel:+123'],
    ])('normalizes %s → %s', (input, expected) => {
        expect(normalizeLinkUrl(input)).toBe(expected);
    });

    it.each([
        'javascript:alert(1)',
        'data:text/html,<script>alert(1)</script>',
        'vbscript:msgbox(1)',
        '#anchor', // no host once prepended → rejected
        'https://', // malformed → no host
        '/settings/profile', // root-relative path — must NOT become https://settings/profile
        '/path',
        '//', // protocol-relative with no host
        '',
        '   ',
    ])('rejects %s', (input) => {
        expect(normalizeLinkUrl(input)).toBeNull();
    });
});
