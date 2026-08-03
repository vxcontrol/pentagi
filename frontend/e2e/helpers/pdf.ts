import { inflateSync } from 'node:zlib';

export interface PdfShape {
    /** Embedded font subsets — a document that drew no text embeds none. */
    fonts: number;
    isPdf: boolean;
    pages: number;
    /** Text-showing operators across every content stream. */
    textRuns: number;
    trailerComplete: boolean;
}

const inflateStreams = (pdf: Buffer): Buffer[] => {
    const streams: Buffer[] = [];
    const marker = Buffer.from('stream');
    const end = Buffer.from('endstream');

    let from = 0;

    for (;;) {
        const start = pdf.indexOf(marker, from);

        if (start === -1) {
            return streams;
        }

        const stop = pdf.indexOf(end, start);

        if (stop === -1) {
            return streams;
        }

        // Skip the EOL that must follow the `stream` keyword.
        const body = pdf.subarray(start + marker.length, stop);

        try {
            streams.push(inflateSync(body.subarray(body[0] === 0x0d ? 2 : 1)));
        } catch {
            // Not every stream is Flate-encoded (font programs, images) — those are not text.
        }

        from = stop + end.length;
    }
};

/**
 * Structural read of a generated PDF. It deliberately does not decode the text: each font subset
 * carries its own glyph map, so a merged decode garbles the result. The report's wording is checked
 * on the markdown export instead, which is built from the same string and whose bytes are searched
 * for the task and result text.
 */
export const inspectPdf = (pdf: Buffer): PdfShape => {
    const head = pdf.subarray(0, 5).toString();
    const tail = pdf.subarray(-1024).toString();
    const body = pdf.toString('latin1');
    const streams = inflateStreams(pdf);

    return {
        fonts: (body.match(/\/FontFile\d?/g) ?? []).length,
        isPdf: head === '%PDF-',
        pages: (body.match(/\/Type\s*\/Page[^s]/g) ?? []).length,
        textRuns: streams.reduce(
            (total, stream) => total + (stream.toString('latin1').match(/\bTJ\b|\bTj\b/g) ?? []).length,
            0,
        ),
        trailerComplete: tail.includes('%%EOF'),
    };
};
