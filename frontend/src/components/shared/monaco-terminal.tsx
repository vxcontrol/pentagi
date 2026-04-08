/**
 * Monaco Terminal Component
 *
 * High-performance terminal log viewer based on Monaco Editor with ANSI color support.
 * Uses the `anser` library for robust ANSI escape sequence parsing.
 *
 * Supported ANSI features (via anser):
 * - SGR codes: Reset, Bold, Dim, Italic, Underline, Blink, Reverse, Hidden, Strikethrough
 * - Standard and bright foreground/background colors (30-37, 40-47, 90-97, 100-107)
 * - 256-color palette (38;5;N, 48;5;N)
 * - True color / 24-bit RGB (38;2;R;G;B, 48;2;R;G;B)
 * - Reverse video (proper fg/bg color swap handled by anser)
 *
 * Key optimizations:
 * - Dynamic CSS class injection with deduplication for color support
 * - IEditorDecorationsCollection for efficient decoration management
 * - Module-level WeakMap cache for parsed logs (avoids re-parsing same logs array)
 * - Memoized content extraction for better React performance
 * - Incremental decoration updates (only decorates new lines)
 * - Stable useCallback for mount handler
 *
 * @see https://microsoft.github.io/monaco-editor/docs.html for Monaco API reference
 * @see https://github.com/IonicaBizau/anser for ANSI parser documentation
 */

import type * as monaco from 'monaco-editor';

import Editor, { type Monaco, type OnMount } from '@monaco-editor/react';
import Anser from 'anser';
import { useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react';

import { useTheme } from '@/hooks/use-theme';
import { Log } from '@/lib/log';
import { cn } from '@/lib/utils';

interface MonacoTerminalProps {
    className?: string;
    logs: string[];
    searchValue?: string;
}

interface MonacoTerminalRef {
    findNext: () => void;
    findPrevious: () => void;
}

interface ParsedLine {
    lineNumber: number;
    segments: ParsedSegment[];
    text: string;
}

interface ParsedSegment {
    className: string;
    endColumn: number;
    startColumn: number;
}

/**
 * Set of already injected CSS class names to prevent duplicate style injection.
 */
const injectedColorClasses = new Set<string>();

/**
 * Converts an RGB string like "187, 0, 0" to a hex string "bb0000".
 */
const rgbStringToHex = (rgb: string): string =>
    rgb
        .split(',')
        .map((part) => Math.min(255, Math.max(0, parseInt(part.trim(), 10))).toString(16).padStart(2, '0'))
        .join('');

/**
 * Returns the shared style element for dynamic color injection, creating it if needed.
 */
const getDynamicStyleElement = (): HTMLStyleElement => {
    const styleId = 'monaco-terminal-dynamic-colors';
    let element = document.getElementById(styleId) as HTMLStyleElement | null;

    if (!element) {
        element = document.createElement('style');
        element.id = styleId;
        document.head.appendChild(element);
    }

    return element;
};

/**
 * Ensures a CSS class for the given RGB color exists, injecting it if needed.
 * Returns the class name.
 */
const ensureColorClass = (prefix: string, rgb: string, cssProperty: string): string => {
    const hex = rgbStringToHex(rgb);
    const className = `${prefix}-${hex}`;

    if (!injectedColorClasses.has(className)) {
        getDynamicStyleElement().textContent += `.${className} { ${cssProperty}: rgb(${rgb}) !important; }\n`;
        injectedColorClasses.add(className);
    }

    return className;
};

/**
 * Maps anser decoration names to CSS class names.
 */
const DECORATION_CLASS_MAP: Record<string, string> = {
    blink: 'ansi-blink',
    bold: 'ansi-bold',
    dim: 'ansi-dim',
    hidden: 'ansi-hidden',
    italic: 'ansi-italic',
    strikethrough: 'ansi-strikethrough',
    underline: 'ansi-underline',
};

/**
 * Static CSS for text decoration and font style classes.
 */
const DECORATION_STYLES = [
    '.ansi-bold { font-weight: bold !important; }',
    '.ansi-dim { opacity: 0.7 !important; }',
    '.ansi-italic { font-style: italic !important; }',
    '.ansi-underline { text-decoration: underline !important; }',
    '.ansi-strikethrough { text-decoration: line-through !important; }',
    '.ansi-underline.ansi-strikethrough { text-decoration: underline line-through !important; }',
    '.ansi-hidden { visibility: hidden !important; }',
    '.ansi-blink { animation: ansi-blink 1s step-end infinite !important; }',
    '@keyframes ansi-blink { 50% { opacity: 0; } }',
].join('\n');

/**
 * Regex matching ANSI sequences and control characters not handled by anser.
 * Anser only processes CSI SGR (ESC[...m). This regex matches everything else:
 * - OSC sequences: ESC ] ... (BEL | ESC \)
 * - DCS sequences: ESC P ... (ESC \)
 * - Character set designations: ESC ( X, ESC ) X
 * - DEC private sequences: ESC # N
 * - Single-character escape codes: ESC followed by 7,8,D,M,E,H,c,N,O,Z,=,>,<
 * - Lone ESC not followed by [ (catch-all for any remaining non-CSI escape)
 * - Non-printable control characters (except \t and \n)
 */
const UNSUPPORTED_SEQUENCES_REGEX = new RegExp(
    [
        '\\x1b\\][\\s\\S]*?(?:\\x07|\\x1b\\\\)', // OSC: ESC ] ... (BEL | ST)
        '\\x1bP[\\s\\S]*?\\x1b\\\\', // DCS: ESC P ... ST
        '\\x1b[()][A-Z0-9]', // Character set: ESC ( X or ESC ) X
        '\\x1b#[0-9]', // DEC screen alignment: ESC # N
        '\\x1b[78DMEHcNOZ=>]', // Single-char escape sequences
        '\\x1b(?!\\[)', // Lone ESC not starting a CSI sequence
        '[\\x00-\\x08\\x0b\\x0c\\x0e-\\x1a\\x1c-\\x1f\\x7f]', // Control chars (keep \t=0x09, \n=0x0a, \r=0x0d; skip ESC=0x1b)
    ].join('|'),
    'g',
);

/**
 * Sanitizes a single line before passing to anser:
 * 1. Handles carriage returns (\r) — simulates terminal overwrite behavior
 *    by keeping only content after the last \r on the line
 * 2. Strips non-SGR escape sequences and control characters
 */
const sanitizeTerminalLine = (line: string): string => {
    // Handle carriage returns: keep only content after the last lone \r
    // This simulates terminal overwrite (e.g. progress bars)
    const lastCarriageReturn = line.lastIndexOf('\r');

    const withoutCarriageReturns = lastCarriageReturn !== -1 ? line.substring(lastCarriageReturn + 1) : line;

    return withoutCarriageReturns.replace(UNSUPPORTED_SEQUENCES_REGEX, '');
};

/**
 * Parses a single line using anser and returns structured data for Monaco decorations.
 * Anser handles all ANSI SGR codes including 256-color, true color, and reverse video.
 * Non-SGR sequences and control characters are stripped before parsing.
 */
const parseAnsiLine = (line: string, lineNumber: number): ParsedLine => {
    if (!line) {
        return { lineNumber, segments: [], text: '' };
    }

    const sanitizedLine = sanitizeTerminalLine(line);

    if (!sanitizedLine) {
        return { lineNumber, segments: [], text: '' };
    }

    const entries = Anser.ansiToJson(sanitizedLine, { remove_empty: true });
    const segments: ParsedSegment[] = [];
    let column = 1;

    for (const entry of entries) {
        if (!entry.content) {
            continue;
        }

        const startColumn = column;
        const endColumn = column + entry.content.length;

        column = endColumn;

        const classNames: string[] = [];

        if (entry.fg) {
            classNames.push(ensureColorClass('ansi-fg', entry.fg, 'color'));
        }

        if (entry.bg) {
            classNames.push(ensureColorClass('ansi-bg', entry.bg, 'background-color'));
        }

        for (const decoration of entry.decorations) {
            const decorationClass = DECORATION_CLASS_MAP[decoration];

            if (decorationClass) {
                classNames.push(decorationClass);
            }
        }

        if (classNames.length) {
            segments.push({
                className: classNames.join(' '),
                endColumn,
                startColumn,
            });
        }
    }

    const text = entries.map((entry) => entry.content).join('');

    return { lineNumber, segments, text };
};

/**
 * Module-level cache for parsed lines.
 * Uses WeakMap to avoid memory leaks - when logs array is garbage collected, cache is cleaned.
 */
const parsedLogsCache = new WeakMap<readonly string[], ParsedLine[]>();

/**
 * Parse logs array with caching.
 * Returns cached result if the same logs array reference is passed.
 */
const parseLogsWithCache = (logs: string[]): ParsedLine[] => {
    const cached = parsedLogsCache.get(logs);

    if (cached) {
        return cached;
    }

    const allLogsText = logs.join('\n');
    const lines = allLogsText.split('\n');
    const parsedLines = lines.map((line, index) => parseAnsiLine(line, index + 1));

    parsedLogsCache.set(logs, parsedLines);

    return parsedLines;
};

/**
 * Terminal component based on Monaco Editor with ANSI color support.
 * Provides a read-only code viewer with search functionality, theme support, and ANSI color rendering.
 * Compatible with the existing Terminal component API.
 */
const MonacoTerminal = ({
    className,
    logs,
    ref,
    searchValue,
}: MonacoTerminalProps & { ref?: React.RefObject<MonacoTerminalRef | null> }) => {
    const editorRef = useRef<monaco.editor.IStandaloneCodeEditor | null>(null);
    const monacoRef = useRef<Monaco | null>(null);
    const { theme } = useTheme();
    const [isEditorReady, setIsEditorReady] = useState(false);
    const prevLogsLengthRef = useRef<number>(0);
    const searchWidgetRef = useRef<null | { findNext: () => void; findPrevious: () => void }>(null);
    const decorationsCollectionRef = useRef<monaco.editor.IEditorDecorationsCollection | null>(null);

    // Parse ANSI codes from logs with module-level caching
    const parsedContent = useMemo(() => parseLogsWithCache(logs), [logs]);

    // Memoized plain text content extraction
    const content = useMemo(() => parsedContent.map((line) => line.text).join('\n'), [parsedContent]);

    // Monaco editor mount handler - wrapped in useCallback for stability
    const handleEditorDidMount: OnMount = useCallback((editor, monacoInstance) => {
        editorRef.current = editor;
        monacoRef.current = monacoInstance;
        setIsEditorReady(true);

        // Configure editor for optimal performance and word wrapping
        editor.updateOptions({
            automaticLayout: true,
            folding: false,
            glyphMargin: false,
            lineDecorationsWidth: 0,
            lineNumbers: 'on',
            lineNumbersMinChars: 5,
            minimap: { enabled: false },
            overviewRulerLanes: 0, // Disable overview ruler for better performance
            padding: { top: 4 },
            readOnly: true,
            renderLineHighlight: 'none',
            renderWhitespace: 'none',
            scrollbar: {
                alwaysConsumeMouseWheel: false,
                horizontal: 'hidden', // Hide horizontal scrollbar completely
                useShadows: false,
                vertical: 'visible',
                verticalScrollbarSize: 10,
            },
            scrollBeyondLastLine: false,
            selectOnLineNumbers: false,
            wordWrap: 'on', // Enable word wrap (simple 'on' works better than 'bounded')
            wrappingIndent: 'same',
            wrappingStrategy: 'simple', // Use simple strategy for better performance
        });

        // Inject ANSI decoration CSS styles once (static set of classes)
        const ansiStyleId = 'monaco-terminal-ansi-styles';

        if (!document.getElementById(ansiStyleId)) {
            const ansiStyleElement = document.createElement('style');

            ansiStyleElement.id = ansiStyleId;
            ansiStyleElement.textContent = DECORATION_STYLES;
            document.head.appendChild(ansiStyleElement);
        }

        // Inject padding and background styles
        const terminalStyleId = 'monaco-terminal-custom-styles';

        if (!document.getElementById(terminalStyleId)) {
            const terminalStyleElement = document.createElement('style');

            terminalStyleElement.id = terminalStyleId;
            terminalStyleElement.textContent = `
                .monaco-editor .line-numbers {
                    padding-right: 12px !important;
                }
                .monaco-editor,
                .monaco-editor .monaco-editor-background,
                .monaco-editor .margin {
                    background-color: var(--background) !important;
                }
            `;
            document.head.appendChild(terminalStyleElement);
        }

        // Store search widget reference
        searchWidgetRef.current = {
            findNext: () => {
                try {
                    editor.trigger('monaco-terminal', 'actions.find', {});
                    editor.trigger('monaco-terminal', 'editor.action.nextMatchFindAction', {});
                } catch (error: unknown) {
                    Log.error('Monaco findNext failed:', error);
                }
            },
            findPrevious: () => {
                try {
                    editor.trigger('monaco-terminal', 'actions.find', {});
                    editor.trigger('monaco-terminal', 'editor.action.previousMatchFindAction', {});
                } catch (error: unknown) {
                    Log.error('Monaco findPrevious failed:', error);
                }
            },
        };
    }, []);

    // Expose methods to parent component via ref
    useImperativeHandle(
        ref,
        () => ({
            findNext: () => {
                if (searchWidgetRef.current && editorRef.current) {
                    searchWidgetRef.current.findNext();
                }
            },
            findPrevious: () => {
                if (searchWidgetRef.current && editorRef.current) {
                    searchWidgetRef.current.findPrevious();
                }
            },
        }),
        [],
    );

    // Cache for tracking which lines have been decorated
    const decoratedLinesCountRef = useRef<number>(0);

    // Apply ANSI color decorations incrementally
    useEffect(() => {
        if (!isEditorReady || !editorRef.current || !monacoRef.current) {
            return;
        }

        const editor = editorRef.current;
        const monacoInstance = monacoRef.current;

        try {
            const decoratedCount = decoratedLinesCountRef.current;

            // If content was cleared or reduced, reset decorations
            if (parsedContent.length < decoratedCount) {
                if (decorationsCollectionRef.current) {
                    decorationsCollectionRef.current.clear();
                }

                decoratedLinesCountRef.current = 0;
            }

            // If no new lines to decorate, skip
            if (parsedContent.length <= decoratedLinesCountRef.current) {
                return;
            }

            // Build decorations only for new lines (incremental)
            const newDecorations: monaco.editor.IModelDeltaDecoration[] = [];

            for (let index = decoratedLinesCountRef.current; index < parsedContent.length; index++) {
                const line = parsedContent[index];

                if (!line) {
                    continue;
                }

                for (const segment of line.segments) {
                    newDecorations.push({
                        options: {
                            inlineClassName: segment.className,
                        },
                        range: new monacoInstance.Range(
                            line.lineNumber,
                            segment.startColumn,
                            line.lineNumber,
                            segment.endColumn,
                        ),
                    });
                }
            }

            // Apply new decorations incrementally
            if (newDecorations.length) {
                if (decorationsCollectionRef.current) {
                    // Append new decorations to existing collection
                    decorationsCollectionRef.current.append(newDecorations);
                } else {
                    // Create new collection if it doesn't exist
                    decorationsCollectionRef.current = editor.createDecorationsCollection(newDecorations);
                }
            }

            // Update decorated lines count
            decoratedLinesCountRef.current = parsedContent.length;
        } catch (error: unknown) {
            Log.error('Monaco apply decorations failed:', error);
        }
    }, [parsedContent, isEditorReady]);

    // Auto-scroll to bottom when new logs are added
    useEffect(() => {
        if (!isEditorReady || !editorRef.current) {
            return;
        }

        const editor = editorRef.current;

        // Only scroll if new logs were added (not on initial render or when logs were cleared)
        if (logs.length > prevLogsLengthRef.current && prevLogsLengthRef.current > 0) {
            try {
                const lineCount = editor.getModel()?.getLineCount() ?? 0;

                if (lineCount > 0) {
                    // Scroll to the last line (1 = Smooth scroll type)
                    editor.revealLine(lineCount, 1);
                }
            } catch (error: unknown) {
                Log.error('Monaco scroll failed:', error);
            }
        }

        prevLogsLengthRef.current = logs.length;
    }, [logs, isEditorReady]);

    // Handle search functionality
    useEffect(() => {
        if (!isEditorReady || !editorRef.current || !monacoRef.current) {
            return;
        }

        const editor = editorRef.current;

        if (searchValue?.trim()) {
            try {
                // Use Monaco's built-in find functionality
                const searchText = searchValue.trim();

                // Create find options
                const findOptions = {
                    isRegex: false,
                    matchCase: false,
                    preserveCase: false,
                    searchString: searchText,
                    wholeWord: false,
                };

                // Trigger find action with the search string
                editor.trigger('monaco-terminal', 'actions.find', findOptions);

                // Automatically jump to first match using requestAnimationFrame
                // for better timing than arbitrary setTimeout delays
                requestAnimationFrame(() => {
                    editor.trigger('monaco-terminal', 'editor.action.nextMatchFindAction', {});
                });
            } catch (error: unknown) {
                Log.error('Monaco search failed:', error);
            }
        } else {
            // Close find widget when search is cleared
            try {
                editor.trigger('monaco-terminal', 'closeFindWidget', {});
            } catch (error: unknown) {
                Log.error('Monaco close find widget failed:', error);
            }
        }
    }, [searchValue, isEditorReady]);

    // Cleanup on unmount
    useEffect(() => {
        return () => {
            // Clear decorations collection
            if (decorationsCollectionRef.current) {
                decorationsCollectionRef.current.clear();
                decorationsCollectionRef.current = null;
            }

            // Reset incremental decorations counter
            decoratedLinesCountRef.current = 0;

            // Note: We don't remove shared styles as they persist across instances
            // parsedLogsCache uses WeakMap so it's automatically cleaned up
        };
    }, []);

    return (
        <div className={cn('relative size-full overflow-hidden', className)}>
            <Editor
                defaultLanguage="plaintext"
                height="100%"
                language="plaintext"
                onMount={handleEditorDidMount}
                options={{
                    domReadOnly: true,
                    fontFamily: 'ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
                    fontSize: 12,
                    fontWeight: '600',
                }}
                theme={theme === 'dark' ? 'vs-dark' : 'vs-light'}
                value={content}
                width="100%"
            />
        </div>
    );
};

MonacoTerminal.displayName = 'MonacoTerminal';

export default MonacoTerminal;
