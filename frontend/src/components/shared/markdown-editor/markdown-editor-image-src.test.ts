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

describe('whitespace and unknown schemes are rejected, not laundered into the document', () => {
    // `new URL()` accepts a space, so the raw candidate persisted as `[hello](https://example.com/my page)` —
    // which reloads with the link destroyed and the brackets leaking into the text.
    it.each(['example.com/my page', 'https://example.com/a b', ' https://example.com/a\tb '])(
        'rejects a URL containing whitespace: %s',
        (value) => {
            expect(normalizeLinkUrl(value)).toBeNull();
            expect(normalizeImageSrc(value)).toBeNull();
        },
    );

    // `https://` was prefixed onto anything unrecognised, so `ftp://host/file` became `https://ftp://host/file`
    // — parsed as host `ftp`, shown as valid, and persisted as a dead link with no error.
    it.each(['ftp://example.com/a', 'chrome://settings', 'ssh://host/x', 'file:///etc/passwd'])(
        'rejects the unknown scheme %s instead of prefixing https:// onto it',
        (value) => {
            expect(normalizeLinkUrl(value)).toBeNull();
        },
    );

    // Values the guard must NOT catch: a scheme-less host:port reads like a scheme but is ordinary input.
    it.each(['example.com:8080/path', 'localhost:3000', 'localhost:3000/a', 'a.b.c:9200/_search'])(
        'still accepts the scheme-less host:port %s',
        (value) => {
            expect(normalizeLinkUrl(value)).toBe(`https://${value}`);
        },
    );

    it('still accepts a protocol-relative URL', () => {
        expect(normalizeImageSrc('//cdn.example.com/a.png')).toBe('https://cdn.example.com/a.png');
    });

    // Byte-fidelity: routing the value through `new URL().href` would rewrite all three, which is why the fix
    // rejects whitespace rather than re-encoding.
    it.each([
        ['example.com', 'https://example.com'],
        ['https://example.com/{{HOST}}/x', 'https://example.com/{{HOST}}/x'],
        ['https://example.com/путь', 'https://example.com/путь'],
    ])('keeps %s byte-identical', (input, expected) => {
        expect(normalizeLinkUrl(input)).toBe(expected);
    });
});
