import { describe, expect, it } from 'vitest';

import { TAG_RE } from './markdown-editor-tag-highlight';
import { findVariableUseRanges, VARIABLE_RE, variableUseRegex } from './markdown-editor-variable-syntax';

const matches = (re: RegExp, text: string) => [...text.matchAll(re)].map((m) => m[0]);

describe('VARIABLE_RE — {{ go-template actions }}', () => {
    it.each([
        ['{{.Var}}', ['{{.Var}}']],
        ['{{- if .X}}', ['{{- if .X}}']],
        ['{{end}}', ['{{end}}']],
        ['{{.X | upper}}', ['{{.X | upper}}']],
        ['{{printf "%d" .N}}', ['{{printf "%d" .N}}']],
        ['{{- .B -}}', ['{{- .B -}}']],
        ['a {{.X}} b {{.Y}} c', ['{{.X}}', '{{.Y}}']],
    ])('matches %s', (input, expected) => {
        expect(matches(VARIABLE_RE, input)).toEqual(expected);
    });

    it.each(['{single brace}', '${shell}', 'open {{ no close', 'a }} before {{ open'])('does NOT match %s', (input) => {
        expect(matches(VARIABLE_RE, input)).toEqual([]);
    });
});

describe('TAG_RE — <xml-like tags>', () => {
    it.each([
        ['<container_environment>', ['<container_environment>']],
        ['</language_policy>', ['</language_policy>']],
        ['<specialist name="searcher">', ['<specialist name="searcher">']],
        ['<tool/>', ['<tool/>']],
        ['<a_b-c>', ['<a_b-c>']],
    ])('matches %s', (input, expected) => {
        expect(matches(TAG_RE, input)).toEqual(expected);
    });

    it.each([
        'count a < b here', // less-than, space after <
        'loop x<5 times', // digit after <
        'rating <3',
        'empty <>',
        'spaced < tag>',
        '<https://example.com>', // markdown autolink — must NOT be a tag
        '<!-- a comment -->',
    ])('does NOT match %s', (input) => {
        expect(matches(TAG_RE, input)).toEqual([]);
    });

    it('documented false-positive: a single-letter <b> in prose is highlighted (harmless, view-only)', () => {
        expect(matches(TAG_RE, 'if a<b> then')).toEqual(['<b>']);
    });
});

describe('findVariableUseRanges — block-first {{ … .Var … }} ranges (ReDoS-safe)', () => {
    it('returns each use as {index, length}', () => {
        expect(findVariableUseRanges('a {{.Foo}} b {{ .Foo | upper }} c {{.Bar}}', 'Foo')).toEqual([
            { index: 2, length: 8 },
            { index: 13, length: 18 },
        ]);
    });

    it('stays linear on an unclosed {{ with many anchors (the old lazy regex froze here)', () => {
        const pathological = `{{ ${' .Foo'.repeat(20000)}`;
        const start = performance.now();
        const ranges = findVariableUseRanges(pathological, 'Foo');

        expect(performance.now() - start).toBeLessThan(200);
        expect(ranges).toEqual([]);
    });
});

describe('variableUseRegex — escapes the interpolated name', () => {
    it('treats a dot in the name as literal, not any-char', () => {
        expect(variableUseRegex('Foo').test('.Foo')).toBe(true);
        expect(variableUseRegex('F.o').test('.Fxo')).toBe(false);
    });
});
