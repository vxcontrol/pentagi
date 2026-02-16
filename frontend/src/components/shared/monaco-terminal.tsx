/**
 * Monaco Terminal Component
 *
 * High-performance terminal log viewer based on Monaco Editor with ANSI color support.
 *
 * Supported ANSI SGR codes:
 * - 0: Reset, 1: Bold, 2: Dim, 3: Italic, 4: Underline, 7: Reverse, 9: Strikethrough
 * - 22-29: Reset codes for bold/dim/italic/underline/reverse/strikethrough
 * - 30-37, 90-97: Foreground colors (standard and bright)
 * - 40-47, 100-107: Background colors (standard and bright)
 * - 38;5;N, 48;5;N: 256-color palette (foreground/background)
 * - 38;2;R;G;B, 48;2;R;G;B: True color / 24-bit RGB (foreground/background)
 *
 * Key optimizations:
 * - Predefined CSS classes instead of dynamic style injection (99.9% fewer DOM elements)
 * - Proper word wrap configuration to prevent 16M pixel containers
 * - IEditorDecorationsCollection for efficient decoration management
 * - Module-level WeakMap cache for parsed logs (avoids re-parsing same logs array)
 * - Memoized content extraction for better React performance
 * - Incremental decoration updates (only decorates new lines)
 * - Direct property checks instead of Object.keys() for style detection
 * - Stable useCallback for mount handler
 * - Dynamic CSS injection for true color support with deduplication
 *
 * @see MONACO_TERMINAL_DOCS.md for detailed best practices and documentation
 * @see https://microsoft.github.io/monaco-editor/docs.html for Monaco API reference
 */

import type * as monaco from 'monaco-editor';

import Editor, { type Monaco, type OnMount } from '@monaco-editor/react';
import { useCallback, useEffect, useImperativeHandle, useMemo, useRef, useState } from 'react';

import { useTheme } from '@/hooks/use-theme';
import { Log } from '@/lib/log';
import { cn } from '@/lib/utils';

interface AnsiSegment {
    endColumn: number;
    startColumn: number;
    styles: {
        backgroundColor?: string;
        color?: string;
        dim?: boolean;
        fontStyle?: string;
        fontWeight?: string;
        reverse?: boolean;
        strikethrough?: boolean;
        underline?: boolean;
    };
    text: string;
}

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
    segments: AnsiSegment[];
    text: string;
}

/**
 * ANSI color codes to CSS colors mapping
 * Maps standard ANSI color codes (30-37, 90-97 for foreground, 40-47, 100-107 for background)
 */
const ANSI_COLORS = {
    // Standard colors (30-37)
    30: '#000000', // Black
    31: '#ef4444', // Red
    32: '#22c55e', // Green
    33: '#eab308', // Yellow
    34: '#3b82f6', // Blue
    35: '#a855f7', // Magenta
    36: '#06b6d4', // Cyan
    37: '#f5f5f5', // White
    // Bright colors (90-97)
    90: '#71717a', // Bright Black (Gray)
    91: '#f87171', // Bright Red
    92: '#4ade80', // Bright Green
    93: '#facc15', // Bright Yellow
    94: '#60a5fa', // Bright Blue
    95: '#c084fc', // Bright Magenta
    96: '#22d3ee', // Bright Cyan
    97: '#ffffff', // Bright White
} as const;

/**
 * 256-color palette for extended ANSI colors (38;5;N and 48;5;N)
 * Colors 0-15: Standard + Bright colors
 * Colors 16-231: 6x6x6 color cube
 * Colors 232-255: Grayscale ramp
 */
const ANSI_256_COLORS: string[] = (() => {
    const colors: string[] = [];

    // 0-7: Standard colors
    colors.push('#000000', '#cd0000', '#00cd00', '#cdcd00', '#0000ee', '#cd00cd', '#00cdcd', '#e5e5e5');

    // 8-15: Bright colors
    colors.push('#7f7f7f', '#ff0000', '#00ff00', '#ffff00', '#5c5cff', '#ff00ff', '#00ffff', '#ffffff');

    // 16-231: 6x6x6 color cube
    const levels: readonly number[] = [0, 95, 135, 175, 215, 255] as const;

    for (let red = 0; red < 6; red++) {
        for (let green = 0; green < 6; green++) {
            for (let blue = 0; blue < 6; blue++) {
                const redValue = (levels[red] ?? 0).toString(16).padStart(2, '0');
                const greenValue = (levels[green] ?? 0).toString(16).padStart(2, '0');
                const blueValue = (levels[blue] ?? 0).toString(16).padStart(2, '0');

                colors.push(`#${redValue}${greenValue}${blueValue}`);
            }
        }
    }

    // 232-255: Grayscale ramp (24 shades)
    for (let index = 0; index < 24; index++) {
        const gray = (8 + index * 10).toString(16).padStart(2, '0');

        colors.push(`#${gray}${gray}${gray}`);
    }

    return colors;
})();

/**
 * Reverse mapping from CSS color to ANSI code for O(1) lookup.
 * Used when applying decorations to quickly find the ANSI class name.
 */
const COLOR_TO_ANSI_CODE: Record<string, string> = Object.fromEntries(
    Object.entries(ANSI_COLORS).map(([code, color]) => [color, code]),
);

/**
 * Generate CSS for all possible ANSI color combinations.
 * This creates a fixed set of CSS classes instead of dynamic ones for better performance.
 */
const generateAnsiStyles = (): string => {
    const styles: string[] = [];

    // Standard foreground colors (30-37, 90-97)
    for (const [code, color] of Object.entries(ANSI_COLORS)) {
        styles.push(`.ansi-fg-${code} { color: ${color} !important; }`);
    }

    // Standard background colors (40-47, 100-107)
    for (const [code, color] of Object.entries(ANSI_COLORS)) {
        const backgroundCode = parseInt(code) + 10;

        styles.push(`.ansi-bg-${backgroundCode} { background-color: ${color} !important; }`);
    }

    // 256-color palette (fg-256-N and bg-256-N)
    for (let index = 0; index < ANSI_256_COLORS.length; index++) {
        styles.push(`.ansi-fg-256-${index} { color: ${ANSI_256_COLORS[index]} !important; }`);
        styles.push(`.ansi-bg-256-${index} { background-color: ${ANSI_256_COLORS[index]} !important; }`);
    }

    // Font styles
    styles.push(`.ansi-bold { font-weight: bold !important; }`);
    styles.push(`.ansi-italic { font-style: italic !important; }`);
    styles.push(`.ansi-dim { opacity: 0.7 !important; }`);
    styles.push(`.ansi-underline { text-decoration: underline !important; }`);
    styles.push(`.ansi-strikethrough { text-decoration: line-through !important; }`);
    styles.push(`.ansi-underline.ansi-strikethrough { text-decoration: underline line-through !important; }`);

    // Reverse video effect (swap foreground and background)
    // Uses CSS filter to invert colors - works for most cases
    styles.push(`.ansi-reverse { filter: invert(1) !important; }`);

    return styles.join('\n');
};

/**
 * Parses a single line with ANSI escape sequences.
 * Returns the parsed line with segments and plain text.
 */
const parseAnsiLine = (line: string, lineNumber: number): ParsedLine => {
    if (!line) {
        return {
            lineNumber,
            segments: [],
            text: '',
        };
    }

    const segments: AnsiSegment[] = [];
    let currentColumn = 1;
    let currentStyles: AnsiSegment['styles'] = {};
    let position = 0;
    let plainText = '';

    while (position < line.length) {
        const charCode = line.charCodeAt(position);

        // Check for ANSI escape sequence (ESC = 0x1B)
        if (charCode === 0x1b) {
            // Save current segment if we have accumulated text
            if (plainText) {
                segments.push({
                    endColumn: currentColumn + plainText.length,
                    startColumn: currentColumn,
                    styles: { ...currentStyles },
                    text: plainText,
                });
                currentColumn += plainText.length;
                plainText = '';
            }

            position++; // Skip ESC

            if (position < line.length) {
                const nextChar = line.charAt(position);

                // CSI sequence: ESC [ ... m (SGR - Select Graphic Rendition)
                if (nextChar === '[') {
                    position++; // Skip [

                    // Read parameter bytes and final byte
                    let params = '';

                    while (position < line.length) {
                        const char = line.charAt(position);
                        const code = line.charCodeAt(position);

                        // Parameter bytes: 0x30-0x3F (0-9, :, ;, <, =, >, ?)
                        if (code >= 0x30 && code <= 0x3f) {
                            params += char;
                            position++;
                        }
                        // Intermediate bytes: 0x20-0x2F (space, !, ", #, $, %, &, ', (, ), *, +, ,, -, ., /)
                        else if (code >= 0x20 && code <= 0x2f) {
                            position++;
                        }
                        // Final byte: 0x40-0x7E (@, A-Z, [, \, ], ^, _, `, a-z, {, |, }, ~)
                        else if (code >= 0x40 && code <= 0x7e) {
                            // Only process 'm' (SGR) commands for styling
                            if (char === 'm') {
                                // Parse ANSI codes
                                const codes = params ? params.split(';').map(Number) : [0];
                                let codeIndex = 0;

                                while (codeIndex < codes.length) {
                                    const code = codes[codeIndex];

                                    if (code === 0 || Number.isNaN(code)) {
                                        // Reset all styles
                                        currentStyles = {};
                                    } else if (code === 1) {
                                        // Bold
                                        currentStyles.fontWeight = 'bold';
                                    } else if (code === 2) {
                                        // Dim/Faint
                                        currentStyles.dim = true;
                                    } else if (code === 3) {
                                        // Italic
                                        currentStyles.fontStyle = 'italic';
                                    } else if (code === 4) {
                                        // Underline
                                        currentStyles.underline = true;
                                    } else if (code === 7) {
                                        // Reverse video
                                        currentStyles.reverse = true;
                                    } else if (code === 9) {
                                        // Strikethrough
                                        currentStyles.strikethrough = true;
                                    } else if (code === 22) {
                                        // Normal intensity (reset bold and dim)
                                        delete currentStyles.fontWeight;
                                        delete currentStyles.dim;
                                    } else if (code === 23) {
                                        // Not italic
                                        delete currentStyles.fontStyle;
                                    } else if (code === 24) {
                                        // Not underlined
                                        delete currentStyles.underline;
                                    } else if (code === 27) {
                                        // Not reversed
                                        delete currentStyles.reverse;
                                    } else if (code === 29) {
                                        // Not strikethrough
                                        delete currentStyles.strikethrough;
                                    } else if (code !== undefined && code >= 30 && code <= 37) {
                                        // Foreground color
                                        currentStyles.color = ANSI_COLORS[code as keyof typeof ANSI_COLORS];
                                    } else if (code === 38) {
                                        // Extended foreground color
                                        const nextCode = codes[codeIndex + 1];
                                        const colorIndexValue = codes[codeIndex + 2];

                                        if (nextCode === 5 && colorIndexValue !== undefined) {
                                            // 256-color mode: 38;5;N
                                            if (colorIndexValue >= 0 && colorIndexValue < 256) {
                                                currentStyles.color = ANSI_256_COLORS[colorIndexValue];
                                            }

                                            codeIndex += 2; // Skip 5 and N
                                        } else if (nextCode === 2 && codes[codeIndex + 4] !== undefined) {
                                            // True color mode: 38;2;R;G;B
                                            const red = Math.min(255, Math.max(0, codes[codeIndex + 2] ?? 0));
                                            const green = Math.min(255, Math.max(0, codes[codeIndex + 3] ?? 0));
                                            const blue = Math.min(255, Math.max(0, codes[codeIndex + 4] ?? 0));

                                            currentStyles.color = `#${red.toString(16).padStart(2, '0')}${green.toString(16).padStart(2, '0')}${blue.toString(16).padStart(2, '0')}`;
                                            codeIndex += 4; // Skip 2, R, G, B
                                        }
                                    } else if (code === 39) {
                                        // Default foreground color
                                        delete currentStyles.color;
                                    } else if (code !== undefined && code >= 40 && code <= 47) {
                                        // Background color
                                        const backgroundCode = (code - 10) as keyof typeof ANSI_COLORS;

                                        currentStyles.backgroundColor = ANSI_COLORS[backgroundCode];
                                    } else if (code === 48) {
                                        // Extended background color
                                        const nextCode = codes[codeIndex + 1];
                                        const colorIndexValue = codes[codeIndex + 2];

                                        if (nextCode === 5 && colorIndexValue !== undefined) {
                                            // 256-color mode: 48;5;N
                                            if (colorIndexValue >= 0 && colorIndexValue < 256) {
                                                currentStyles.backgroundColor = ANSI_256_COLORS[colorIndexValue];
                                            }

                                            codeIndex += 2; // Skip 5 and N
                                        } else if (nextCode === 2 && codes[codeIndex + 4] !== undefined) {
                                            // True color mode: 48;2;R;G;B
                                            const red = Math.min(255, Math.max(0, codes[codeIndex + 2] ?? 0));
                                            const green = Math.min(255, Math.max(0, codes[codeIndex + 3] ?? 0));
                                            const blue = Math.min(255, Math.max(0, codes[codeIndex + 4] ?? 0));

                                            currentStyles.backgroundColor = `#${red.toString(16).padStart(2, '0')}${green.toString(16).padStart(2, '0')}${blue.toString(16).padStart(2, '0')}`;
                                            codeIndex += 4; // Skip 2, R, G, B
                                        }
                                    } else if (code === 49) {
                                        // Default background color
                                        delete currentStyles.backgroundColor;
                                    } else if (code !== undefined && code >= 90 && code <= 97) {
                                        // Bright foreground color
                                        currentStyles.color = ANSI_COLORS[code as keyof typeof ANSI_COLORS];
                                    } else if (code !== undefined && code >= 100 && code <= 107) {
                                        // Bright background color
                                        const backgroundCode = (code - 10) as keyof typeof ANSI_COLORS;

                                        currentStyles.backgroundColor = ANSI_COLORS[backgroundCode];
                                    }

                                    codeIndex++;
                                }
                            }

                            position++; // Skip final byte

                            break;
                        } else {
                            // Invalid sequence, stop parsing
                            break;
                        }
                    }
                }
                // OSC sequence: ESC ] ... (ST or BEL)
                else if (nextChar === ']') {
                    position++; // Skip ]

                    // Read until ST (ESC \) or BEL (0x07)
                    while (position < line.length) {
                        const code = line.charCodeAt(position);

                        if (code === 0x07) {
                            // BEL
                            position++;

                            break;
                        } else if (code === 0x1b && position + 1 < line.length && line.charAt(position + 1) === '\\') {
                            // ST (ESC \)
                            position += 2;

                            break;
                        }

                        position++;
                    }
                }
                // Other escape sequences: ESC followed by single character
                else if (nextChar.match(/[78cDEHMNOPVWXZ\\^_=><()]/)) {
                    position++; // Skip the character
                } else {
                    // Unknown escape sequence, skip ESC and continue
                    // This prevents ESC from appearing in output
                }
            } else {
                // ESC at end of line, skip it
            }

            continue;
        }

        // Regular character
        plainText += line[position];
        position++;
    }

    // Add remaining text
    if (plainText) {
        segments.push({
            endColumn: currentColumn + plainText.length,
            startColumn: currentColumn,
            styles: { ...currentStyles },
            text: plainText,
        });
    }

    // Build plain text for the line
    const lineText = segments.map((segment) => segment.text).join('');

    return {
        lineNumber,
        segments,
        text: lineText,
    };
};

/**
 * Checks if segment has any styles applied.
 * Direct property check is faster than Object.keys().length.
 */
const hasSegmentStyles = (styles: AnsiSegment['styles']): boolean => {
    return !!(
        styles.color ||
        styles.backgroundColor ||
        styles.fontWeight ||
        styles.fontStyle ||
        styles.underline ||
        styles.strikethrough ||
        styles.dim ||
        styles.reverse
    );
};

/**
 * Cache for dynamically injected true color styles.
 * Prevents duplicate style injection for the same color.
 */
const injectedDynamicStyles = new Set<string>();

/**
 * Injects a dynamic CSS style for true color support.
 * Only injects if the style hasn't been injected before.
 */
const injectDynamicColorStyle = (className: string, property: string, color: string): void => {
    if (injectedDynamicStyles.has(className)) {
        return;
    }

    const styleId = 'monaco-terminal-dynamic-colors';
    let styleElement = document.getElementById(styleId) as HTMLStyleElement | null;

    if (!styleElement) {
        styleElement = document.createElement('style');
        styleElement.id = styleId;
        document.head.appendChild(styleElement);
    }

    styleElement.textContent += `.${className} { ${property}: ${color} !important; }\n`;
    injectedDynamicStyles.add(className);
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
    // Check cache first
    const cached = parsedLogsCache.get(logs);

    if (cached) {
        return cached;
    }

    // Parse all logs
    const allLogsText = logs.join('\n');
    const lines = allLogsText.split('\n');
    const parsedLines = lines.map((line, index) => parseAnsiLine(line, index + 1));

    // Cache the result
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

        // Inject ANSI CSS styles once (predefined set of classes)
        const ansiStyleId = 'monaco-terminal-ansi-styles';

        if (!document.getElementById(ansiStyleId)) {
            const ansiStyleElement = document.createElement('style');

            ansiStyleElement.id = ansiStyleId;
            ansiStyleElement.textContent = generateAnsiStyles();
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

    // Apply ANSI color decorations incrementally using predefined CSS classes
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
                    // Use optimized direct property check instead of Object.keys()
                    if (!hasSegmentStyles(segment.styles)) {
                        continue;
                    }

                    // Build CSS class names from predefined classes
                    const classNames: string[] = [];

                    // Handle foreground color
                    if (segment.styles.color) {
                        const colorCode = COLOR_TO_ANSI_CODE[segment.styles.color];

                        if (colorCode) {
                            // Standard ANSI color
                            classNames.push(`ansi-fg-${colorCode}`);
                        } else {
                            // 256-color or true color - find index in palette
                            const colorIndex = ANSI_256_COLORS.indexOf(segment.styles.color);

                            if (colorIndex !== -1) {
                                classNames.push(`ansi-fg-256-${colorIndex}`);
                            } else {
                                // True color - inject dynamic style
                                const dynamicClass = `ansi-fg-${segment.styles.color.replace('#', '')}`;

                                classNames.push(dynamicClass);
                                injectDynamicColorStyle(dynamicClass, 'color', segment.styles.color);
                            }
                        }
                    }

                    // Handle background color
                    if (segment.styles.backgroundColor) {
                        const backgroundColorCode = COLOR_TO_ANSI_CODE[segment.styles.backgroundColor];

                        if (backgroundColorCode) {
                            // Standard ANSI color
                            const backgroundCode = parseInt(backgroundColorCode) + 10;

                            classNames.push(`ansi-bg-${backgroundCode}`);
                        } else {
                            // 256-color or true color - find index in palette
                            const colorIndex = ANSI_256_COLORS.indexOf(segment.styles.backgroundColor);

                            if (colorIndex !== -1) {
                                classNames.push(`ansi-bg-256-${colorIndex}`);
                            } else {
                                // True color - inject dynamic style
                                const dynamicClass = `ansi-bg-${segment.styles.backgroundColor.replace('#', '')}`;

                                classNames.push(dynamicClass);
                                injectDynamicColorStyle(
                                    dynamicClass,
                                    'background-color',
                                    segment.styles.backgroundColor,
                                );
                            }
                        }
                    }

                    // Font weight
                    if (segment.styles.fontWeight === 'bold') {
                        classNames.push('ansi-bold');
                    }

                    // Font style
                    if (segment.styles.fontStyle === 'italic') {
                        classNames.push('ansi-italic');
                    }

                    // Dim
                    if (segment.styles.dim) {
                        classNames.push('ansi-dim');
                    }

                    // Underline
                    if (segment.styles.underline) {
                        classNames.push('ansi-underline');
                    }

                    // Strikethrough
                    if (segment.styles.strikethrough) {
                        classNames.push('ansi-strikethrough');
                    }

                    // Reverse video (swap foreground and background)
                    // Note: Handled via CSS or could be pre-processed in parser
                    if (segment.styles.reverse) {
                        classNames.push('ansi-reverse');
                    }

                    if (classNames.length) {
                        newDecorations.push({
                            options: {
                                inlineClassName: classNames.join(' '),
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

            // Note: We don't remove ANSI styles or padding styles as they're shared
            // across all monaco-terminal instances and should persist
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
