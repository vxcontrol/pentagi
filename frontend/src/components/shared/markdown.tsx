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
import 'highlight.js/styles/atom-one-dark.css';
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

        return components;
    }, [processedSearch, createComponentRenderer]);

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
                {children}
            </ReactMarkdown>
        </div>
    );
}

export default Markdown;
