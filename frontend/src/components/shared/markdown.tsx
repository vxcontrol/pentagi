import bash from 'highlight.js/lib/languages/bash';
import c from 'highlight.js/lib/languages/c';
import csharp from 'highlight.js/lib/languages/csharp';
import dockerfile from 'highlight.js/lib/languages/dockerfile';
import go from 'highlight.js/lib/languages/go';
import graphql from 'highlight.js/lib/languages/graphql';
import http from 'highlight.js/lib/languages/http';
import java from 'highlight.js/lib/languages/java';
import javascript from 'highlight.js/lib/languages/javascript';
import json from 'highlight.js/lib/languages/json';
import kotlin from 'highlight.js/lib/languages/kotlin';
import lua from 'highlight.js/lib/languages/lua';
import markdown from 'highlight.js/lib/languages/markdown';
import nginx from 'highlight.js/lib/languages/nginx';
import php from 'highlight.js/lib/languages/php';
import python from 'highlight.js/lib/languages/python';
import sql from 'highlight.js/lib/languages/sql';
import xml from 'highlight.js/lib/languages/xml';
import yaml from 'highlight.js/lib/languages/yaml';
import { common, createLowlight } from 'lowlight';
import { isValidElement, type ReactElement, type ReactNode, useCallback, useMemo } from 'react';
import ReactMarkdown, { type Components } from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import rehypeSlug from 'rehype-slug';
import remarkGfm from 'remark-gfm';

const lowlight = createLowlight();
lowlight.register('bash', bash);
lowlight.register('c', c);
lowlight.register('csharp', csharp);
lowlight.register('dockerfile', dockerfile);
lowlight.register('go', go);
lowlight.register('graphql', graphql);
lowlight.register('http', http);
lowlight.register('java', java);
lowlight.register('javascript', javascript);
lowlight.register('json', json);
lowlight.register('kotlin', kotlin);
lowlight.register('lua', lua);
lowlight.register('markdown', markdown);
lowlight.register('nginx', nginx);
lowlight.register('php', php);
lowlight.register('python', python);
lowlight.register('sql', sql);
lowlight.register('xml', xml);
lowlight.register('yaml', yaml);

interface MarkdownProps {
    children: string;
    className?: string;
    searchValue?: string;
}

/**
 * Scan `content` line-by-line for fenced code blocks opened with backtick
 * fences (```). When the body of a block contains a line whose leading
 * backtick run is long enough to close the current fence, the outer fence is
 * extended by one backtick beyond the longest such inner run so the parser
 * treats the body as literal content instead of closing the block early.
 *
 * Only backtick fences are normalised; tilde fences (~) are left untouched
 * because remark-gfm already handles them correctly.
 */
function preprocessMarkdownFences(content: string): string {
    const lines = content.split('\n');
    const result: string[] = [];
    let i = 0;

    while (i < lines.length) {
        const line = lines[i];
        // Match an opening backtick fence: optional indent (≤3 spaces), then 3+ backticks, then optional info string.
        const openMatch = /^( {0,3})(```+)([ \t].*)?$/.exec(line);

        if (!openMatch) {
            result.push(line);
            i++;
            continue;
        }

        const indent = openMatch[1] ?? '';
        const fence = openMatch[2];
        const info = openMatch[3] ?? '';
        const fenceLen = fence.length;

        // Scan ahead: find the closing fence and track the longest backtick
        // run on any single line inside the block body.
        let closeIdx = -1;
        let maxInnerLen = 0;

        for (let j = i + 1; j < lines.length; j++) {
            // A closing fence must consist solely of backticks (≥ fenceLen) with
            // optional trailing whitespace and optional leading indent (≤3 spaces).
            const closeMatch = /^( {0,3})(```+)\s*$/.exec(lines[j]);
            if (closeMatch && closeMatch[2].length >= fenceLen) {
                closeIdx = j;
                break;
            }
            // Track the longest leading backtick run inside the block body.
            const innerMatch = /^( {0,3})(```+)/.exec(lines[j]);
            if (innerMatch) {
                maxInnerLen = Math.max(maxInnerLen, innerMatch[2].length);
            }
        }

        if (closeIdx === -1) {
            // No closing fence found — emit as-is and move on.
            result.push(line);
            i++;
            continue;
        }

        // If any inner run is long enough to act as a closing fence, extend
        // the outer fence to one backtick beyond the longest inner run.
        const effectiveFenceLen = maxInnerLen >= fenceLen ? maxInnerLen + 1 : fenceLen;
        const effectiveFence = '`'.repeat(effectiveFenceLen);

        result.push(`${indent}${effectiveFence}${info}`);
        for (let j = i + 1; j < closeIdx; j++) {
            result.push(lines[j]);
        }
        // Preserve the original closing fence indent.
        const closeIndent = /^( {0,3})/.exec(lines[closeIdx])?.[1] ?? '';
        result.push(`${closeIndent}${effectiveFence}`);

        i = closeIdx + 1;
    }

    return result.join('\n');
}

const textElements = [
    'p',
    'span',
    'div',
    'h1',
    'h2',
    'h3',
    'h4',
    'h5',
    'h6',
    'a',
    'li',
    'ul',
    'ol',
    'blockquote',
    'table',
    'thead',
    'tbody',
    'tr',
    'td',
    'th',
    'strong',
    'em',
    'b',
    'i',
    'u',
    's',
    'del',
    'ins',
    'mark',
    'small',
    'sub',
    'sup',
    'dl',
    'dt',
    'dd',
];

const escapeRegExp = (string: string): string => {
    return string.replaceAll(/[.*+?^${}()|[\]\\]/g, '\\$&');
};

function Markdown({ children, className, searchValue }: MarkdownProps) {
    const processedSearch = useMemo(() => {
        const trimmedSearch = searchValue?.trim();

        if (!trimmedSearch) {
            return null;
        }

        return {
            escaped: escapeRegExp(trimmedSearch),
            regex: new RegExp(`(${escapeRegExp(trimmedSearch)})`, 'gi'),
            trimmed: trimmedSearch,
        };
    }, [searchValue]);

    const createHighlightedText = useCallback(
        (text: string) => {
            if (!processedSearch) {
                return text;
            }

            const parts = text.split(processedSearch.regex);

            return parts.map((part, index) => {
                if (part.toLowerCase() === processedSearch.trimmed.toLowerCase()) {
                    return (
                        <span
                            key={`highlight-${index}`}
                            style={{
                                backgroundColor: 'rgba(255, 255, 0, 0.15)',
                                borderRadius: '2px',
                                boxShadow: 'inset 0 0 0 1px rgba(255, 255, 0, 0.25)',
                                padding: '0px 1px',
                            }}
                        >
                            {part}
                        </span>
                    );
                }

                return part;
            });
        },
        [processedSearch],
    );

    const processTextNode = useMemo(() => {
        const hasChildrenProp = (node: unknown): node is ReactElement<{ children?: ReactNode }> =>
            isValidElement(node) && (node.props as { children?: ReactNode }).children !== undefined;

        const fn = (nodeChildren: ReactNode): ReactNode => {
            if (!processedSearch) {
                return nodeChildren;
            }

            if (typeof nodeChildren === 'string') {
                return createHighlightedText(nodeChildren);
            }

            if (Array.isArray(nodeChildren)) {
                return nodeChildren.map((child, index) => {
                    if (typeof child === 'string') {
                        return createHighlightedText(child);
                    }

                    if (hasChildrenProp(child)) {
                        return {
                            ...child,
                            key: child.key || `processed-${index}`,
                            props: {
                                ...child.props,
                                children: fn(child.props.children),
                            },
                        };
                    }

                    return child;
                });
            }

            if (hasChildrenProp(nodeChildren)) {
                return {
                    ...nodeChildren,
                    props: {
                        ...nodeChildren.props,
                        children: fn(nodeChildren.props.children),
                    },
                };
            }

            return nodeChildren;
        };

        return fn;
    }, [processedSearch, createHighlightedText]);

    const createComponentRenderer = useCallback(
        (ComponentName: string) => {
            const Component = ComponentName as React.ElementType;

            const Renderer = ({
                children: nodeChildren,
                ...props
            }: React.PropsWithChildren<Record<string, unknown>>) => {
                const processedChildren = processTextNode(nodeChildren);

                return <Component {...props}>{processedChildren}</Component>;
            };

            Renderer.displayName = `Highlighted(${ComponentName})`;

            return Renderer;
        },
        [processTextNode],
    );

    const customComponents = useMemo(() => {
        const components: Components = {};

        if (processedSearch) {
            textElements.forEach((element) => {
                (components as Record<string, ReturnType<typeof createComponentRenderer>>)[element] =
                    createComponentRenderer(element);
            });

            // Code blocks pass through untouched — highlighting injected `<span>`s into a `<code>`
            // breaks rehype-highlight's tokenization and the rendered listing.
            components.code = ({ children: nodeChildren, ...props }) => {
                return <code {...props}>{nodeChildren}</code>;
            };

            components.pre = ({ children: nodeChildren, ...props }) => {
                return <pre {...props}>{nodeChildren}</pre>;
            };
        }

        components.table = ({ children: nodeChildren, ...props }) => (
            <div className="overflow-x-auto">
                <table {...props}>{processedSearch ? processTextNode(nodeChildren) : nodeChildren}</table>
            </div>
        );

        return components;
    }, [processedSearch, createComponentRenderer, processTextNode]);

    const processedChildren = useMemo(() => preprocessMarkdownFences(children), [children]);

    return (
        <div className={`prose prose-sm dark:prose-invert max-w-none ${className || ''}`}>
            <ReactMarkdown
                components={customComponents}
                rehypePlugins={[
                    [
                        rehypeHighlight,
                        {
                            detect: true,
                            languages: {
                                ...common,
                                bash,
                                c,
                                csharp,
                                dockerfile,
                                go,
                                graphql,
                                http,
                                java,
                                javascript,
                                json,
                                kotlin,
                                lua,
                                markdown,
                                nginx,
                                php,
                                python,
                                sql,
                                xml,
                                yaml,
                            },
                        },
                    ],
                    rehypeSlug,
                ]}
                remarkPlugins={[remarkGfm]}
            >
                {processedChildren}
            </ReactMarkdown>
        </div>
    );
}

export default Markdown;
