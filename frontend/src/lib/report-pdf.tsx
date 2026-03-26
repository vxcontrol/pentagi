import { Document, Page, pdf, StyleSheet, Text, View } from '@react-pdf/renderer';
import { marked } from 'marked';

import { Log } from './log';

// PDF styles for @react-pdf/renderer - Enhanced beautiful styles
const pdfStyles = StyleSheet.create({
    bold: {
        fontWeight: 'bold',
    },
    code: {
        color: '#dc2626',
        fontFamily: 'Courier',
        fontSize: 9,
        fontWeight: 'bold',
    },
    codeBlock: {
        backgroundColor: '#1e293b',
        borderColor: '#334155',
        borderRadius: 4,
        borderWidth: 1,
        color: '#e2e8f0',
        fontFamily: 'Courier',
        fontSize: 8.5,
        lineHeight: 1.4,
        marginBottom: 8,
        marginTop: 4,
        padding: 8,
    },
    h1: {
        color: '#0f172a',
        fontSize: 16,
        fontWeight: 'bold',
        marginBottom: 10,
        marginTop: 0,
    },
    h2: {
        borderBottomColor: '#e2e8f0',
        borderBottomWidth: 1,
        color: '#1e293b',
        fontSize: 14,
        fontWeight: 'bold',
        marginBottom: 8,
        marginTop: 12,
        paddingBottom: 4,
    },
    h3: {
        color: '#334155',
        fontSize: 13,
        fontWeight: 'bold',
        marginBottom: 6,
        marginTop: 10,
    },
    h4: {
        color: '#475569',
        fontSize: 12,
        fontWeight: 'bold',
        marginBottom: 5,
        marginTop: 8,
    },
    h5: {
        color: '#64748b',
        fontSize: 11,
        fontWeight: 'bold',
        marginBottom: 4,
        marginTop: 6,
    },
    h6: {
        color: '#94a3b8',
        fontSize: 10,
        fontWeight: 'bold',
        marginBottom: 4,
        marginTop: 6,
    },
    hr: {
        borderBottomColor: '#cbd5e1',
        borderBottomWidth: 1,
        marginBottom: 12,
        marginTop: 12,
    },
    italic: {
        fontStyle: 'italic',
    },
    link: {
        color: '#2563eb',
        fontWeight: 'semibold',
        textDecoration: 'underline',
    },
    list: {
        marginBottom: 8,
        marginLeft: 0,
        marginTop: 6,
    },
    listBullet: {
        color: '#64748b',
        fontSize: 9,
        marginRight: 8,
        minWidth: 20,
    },
    listContent: {
        color: '#334155',
        flex: 1,
        fontSize: 10,
        lineHeight: 1.5,
    },
    listItem: {
        alignItems: 'flex-start',
        flexDirection: 'row',
        marginBottom: 4,
        marginLeft: 16,
    },
    page: {
        backgroundColor: '#ffffff',
        color: '#334155',
        fontFamily: 'Helvetica',
        fontSize: 10,
        lineHeight: 1.5,
        padding: 40,
    },
    paragraph: {
        color: '#475569',
        lineHeight: 1.6,
        marginBottom: 8,
        textAlign: 'justify',
    },
});

// Map of emoji to text replacements for PDF rendering
const emojiMap: Record<string, string> = {
    '⏳': '[WAIT]',
    '⚠️': '[WARN]',
    '⚡': '[RUN]',
    '✅': '[OK]',
    '✨': '[NEW]',
    '❌': '[FAIL]',
    '🎯': '[TARGET]',
    '🐛': '[BUG]',
    '💡': '[IDEA]',
    '📊': '[DATA]',
    '📝': '[NOTE]',
    '🔍': '[SEARCH]',
    '🔐': '[SEC]',
    '🔧': '[TOOL]',
    '🚀': '[START]',
};

// Replace emojis with text equivalents for PDF
const replaceEmojis = (text: string): string => {
    let result = text;

    for (const [emoji, replacement] of Object.entries(emojiMap)) {
        result = result.replaceAll(emoji, replacement);
    }

    return result;
};

// Inline token types for rich text formatting
interface InlineToken {
    bold?: boolean;
    code?: boolean;
    italic?: boolean;
    link?: string;
    text: string;
    type: 'text';
}

// Parsed content interface with support for inline formatting
interface ParsedContent {
    content?: string;
    inlineTokens?: InlineToken[];
    items?: Array<{ inlineTokens: InlineToken[]; raw: string }>;
    level?: number;
    ordered?: boolean;
    type: string;
}

// Parse inline markdown formatting (bold, italic, code, links)
const parseInlineTokens = (text: string): InlineToken[] => {
    const tokens: InlineToken[] = [];
    const inlineTokens = marked.lexer(text, { breaks: false });

    // If lexer returns a paragraph token, use its tokens property
    const firstToken = inlineTokens[0];

    if (firstToken && firstToken.type === 'paragraph' && 'tokens' in firstToken) {
        const paragraphTokens =
            (firstToken as { tokens?: unknown[] }).tokens?.filter((t): t is Record<string, unknown> => {
                return typeof t === 'object' && t !== null;
            }) || [];

        paragraphTokens.forEach((token) => {
            switch (token.type) {
                case 'codespan': {
                    tokens.push({
                        code: true,
                        text: replaceEmojis(String(token.text || '')),
                        type: 'text',
                    });
                    break;
                }

                case 'em': {
                    tokens.push({
                        italic: true,
                        text: replaceEmojis(String(token.text || '')),
                        type: 'text',
                    });
                    break;
                }

                case 'link': {
                    tokens.push({
                        link: String(token.href || ''),
                        text: replaceEmojis(String(token.text || '')),
                        type: 'text',
                    });
                    break;
                }

                case 'strong': {
                    tokens.push({
                        bold: true,
                        text: replaceEmojis(String(token.text || '')),
                        type: 'text',
                    });
                    break;
                }

                case 'text': {
                    tokens.push({
                        text: replaceEmojis(String(token.text || '')),
                        type: 'text',
                    });
                    break;
                }

                default: {
                    // For other inline types, try to extract text
                    if ('text' in token) {
                        tokens.push({
                            text: replaceEmojis(String(token.text || '')),
                            type: 'text',
                        });
                    }
                }
            }
        });
    } else {
        // Fallback: return plain text
        tokens.push({
            text: replaceEmojis(text),
            type: 'text',
        });
    }

    return tokens;
};

// Parse markdown using marked library and convert tokens
const parseMarkdownTokens = (markdown: string): ParsedContent[] => {
    const tokens = marked.lexer(markdown);
    const result: ParsedContent[] = [];

    const processToken = (token: Record<string, unknown>): void => {
        switch (token.type) {
            case 'code': {
                result.push({
                    content: replaceEmojis(String(token.text || '')),
                    type: 'code',
                });
                break;
            }

            case 'heading': {
                result.push({
                    inlineTokens: parseInlineTokens(String(token.text || '')),
                    level: Number(token.depth || 1),
                    type: 'heading',
                });
                break;
            }

            case 'hr': {
                result.push({ type: 'hr' });
                break;
            }

            case 'list': {
                const tokenItems = (
                    Array.isArray(token.items) ? token.items : []
                ) as Array<Record<string, unknown>>;
                const items = tokenItems.map((item) => ({
                    inlineTokens: parseInlineTokens(String(item.text || '')),
                    raw: String(item.text || ''),
                }));
                result.push({
                    items,
                    ordered: Boolean(token.ordered),
                    type: 'list',
                });
                break;
            }

            case 'paragraph': {
                result.push({
                    inlineTokens: parseInlineTokens(String(token.text || '')),
                    type: 'paragraph',
                });
                break;
            }

            case 'space': {
                // Skip empty lines
                break;
            }

            default: {
                // For other types, try to extract text if available
                if ('text' in token && typeof token.text === 'string') {
                    result.push({
                        inlineTokens: parseInlineTokens(token.text),
                        type: 'paragraph',
                    });
                }
            }
        }
    };

    tokens.forEach((token) => processToken(token as Record<string, unknown>));

    return result;
};

// Helper function to render inline tokens with formatting
const renderInlineTokens = (tokens: InlineToken[], keyPrefix: string) => {
    return tokens.map((token, idx) => {
        const textContent = token.text;

        // Collect all applicable styles
        const appliedStyles = [];

        if (token.code) {
            appliedStyles.push(pdfStyles.code);
        }

        if (token.bold) {
            appliedStyles.push(pdfStyles.bold);
        }

        if (token.italic) {
            appliedStyles.push(pdfStyles.italic);
        }

        if (token.link) {
            appliedStyles.push(pdfStyles.link);
        }

        // If we have any styles, wrap in Text component
        if (appliedStyles.length > 0) {
            return (
                <Text
                    key={`${keyPrefix}-inline-${idx}`}
                    style={appliedStyles}
                >
                    {textContent}
                </Text>
            );
        }

        // Return plain text without wrapper
        return textContent;
    });
};

// Render parsed markdown as React PDF components
const renderPDFContent = (parsed: ParsedContent[]) => {
    const elements = parsed
        .map((item, index) => {
            switch (item.type) {
                case 'code': {
                    if (!item.content) {
                        return null;
                    }

                    return (
                        <Text
                            key={`code-${index}`}
                            style={pdfStyles.codeBlock}
                        >
                            {item.content}
                        </Text>
                    );
                }

                case 'heading': {
                    if (!item.inlineTokens || item.inlineTokens.length === 0) {
                        return null;
                    }

                    const style =
                        item.level === 1
                            ? pdfStyles.h1
                            : item.level === 2
                              ? pdfStyles.h2
                              : item.level === 3
                                ? pdfStyles.h3
                                : item.level === 4
                                  ? pdfStyles.h4
                                  : item.level === 5
                                    ? pdfStyles.h5
                                    : pdfStyles.h6;

                    return (
                        <Text
                            key={`heading-${index}`}
                            style={style}
                        >
                            {renderInlineTokens(item.inlineTokens, `heading-${index}`)}
                        </Text>
                    );
                }

                case 'hr': {
                    return (
                        <View
                            key={`hr-${index}`}
                            style={pdfStyles.hr}
                        />
                    );
                }

                case 'list': {
                    if (!item.items || item.items.length === 0) {
                        return null;
                    }

                    return (
                        <View
                            key={`list-${index}`}
                            style={pdfStyles.list}
                        >
                            {item.items.map((listItem, li) => (
                                <View
                                    key={`li-${index}-${li}`}
                                    style={pdfStyles.listItem}
                                >
                                    <Text style={pdfStyles.listBullet}>{item.ordered ? `${li + 1}.` : '•'}</Text>
                                    <Text style={pdfStyles.listContent}>
                                        {renderInlineTokens(listItem.inlineTokens, `li-${index}-${li}`)}
                                    </Text>
                                </View>
                            ))}
                        </View>
                    );
                }

                case 'paragraph': {
                    if (!item.inlineTokens || item.inlineTokens.length === 0) {
                        return null;
                    }

                    return (
                        <Text
                            key={`para-${index}`}
                            style={pdfStyles.paragraph}
                        >
                            {renderInlineTokens(item.inlineTokens, `para-${index}`)}
                        </Text>
                    );
                }

                default: {
                    return null;
                }
            }
        })
        .filter((el) => el !== null);

    return elements;
};

// PDF Document component
const PDFReportDocument = ({ content }: { content: string }) => {
    const parsed = parseMarkdownTokens(content);
    const elements = renderPDFContent(parsed);

    return (
        <Document>
            <Page
                size="A4"
                style={pdfStyles.page}
            >
                {elements}
            </Page>
        </Document>
    );
};

// Main function to generate PDF from markdown
export const generatePDFFromMarkdownNew = async (content: string, fileName: string): Promise<void> => {
    try {
        const doc = <PDFReportDocument content={content} />;
        const blob = await pdf(doc).toBlob();

        // Download
        const url = URL.createObjectURL(blob);
        const link = document.createElement('a');
        link.href = url;
        link.download = `${fileName}.pdf`;
        link.style.display = 'none';
        document.body.appendChild(link);
        link.click();
        link.remove();
        URL.revokeObjectURL(url);
    } catch (error) {
        Log.error('Failed to generate PDF:', error);
        throw error;
    }
};

// Generate PDF as blob
export const generatePDFBlobNew = async (content: string): Promise<Blob> => {
    try {
        const doc = <PDFReportDocument content={content} />;

        return await pdf(doc).toBlob();
    } catch (error) {
        Log.error('Failed to generate PDF blob:', error);
        throw error;
    }
};
