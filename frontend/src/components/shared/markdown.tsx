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
import { useCallback, useMemo } from 'react';
import ReactMarkdown from 'react-markdown';
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

// List of all elements that should have text highlighting
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

// Function to escape special regex characters
const escapeRegExp = (string: string): string => {
    return string.replaceAll(/[.*+?^${}()|[\]\\]/g, '\\$&');
};

const Markdown = ({ children, className, searchValue }: MarkdownProps) => {
    // Memoize the escaped search value to avoid recalculating regex
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

    // Function to create highlighted text components with subtle highlighting
    const createHighlightedText = useCallback(
        (text: string) => {
            if (!processedSearch) {
                return text;
            }

            const parts = text.split(processedSearch.regex);

            return parts.map((part, index) => {
                // Use case-insensitive comparison to match the filtering logic
                if (part.toLowerCase() === processedSearch.trimmed.toLowerCase()) {
                    return (
                        <span
                            key={`highlight-${index}`}
                            style={{
                                // Much more subtle highlighting - very pale yellow with slight border
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

    // Optimized helper function to process text nodes recursively
    const processTextNode = useCallback(
        (nodeChildren: any): any => {
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

                    // Avoid deep cloning React elements to prevent memory leaks
                    // Only process if it's a simple object with props
                    if (child && typeof child === 'object' && child.props && child.props.children !== undefined) {
                        return {
                            ...child,
                            key: child.key || `processed-${index}`,
                            props: {
                                ...child.props,
                                children: processTextNode(child.props.children),
                            },
                        };
                    }

                    return child;
                });
            }

            // Handle React elements safely
            if (
                nodeChildren &&
                typeof nodeChildren === 'object' &&
                nodeChildren.props &&
                nodeChildren.props.children !== undefined
            ) {
                return {
                    ...nodeChildren,
                    props: {
                        ...nodeChildren.props,
                        children: processTextNode(nodeChildren.props.children),
                    },
                };
            }

            return nodeChildren;
        },
        [processedSearch, createHighlightedText],
    );

    // Create a simple component renderer factory to avoid recreating functions
    const createComponentRenderer = useCallback(
        (ComponentName: string) => {
            return ({ children: nodeChildren, ...props }: any) => {
                const processedChildren = processTextNode(nodeChildren);
                const Component = ComponentName as any;

                return <Component {...props}>{processedChildren}</Component>;
            };
        },
        [processTextNode],
    );

    // Memoize components to avoid recreating them on every render
    const customComponents = useMemo(() => {
        const components: Record<string, any> = {};

        if (processedSearch) {
            // Create components for all text elements using the factory
            textElements.forEach((element) => {
                components[element] = createComponentRenderer(element);
            });

            // Don't highlight inside code blocks and preserve their content
            components.code = ({ children: nodeChildren, ...props }: any) => {
                return <code {...props}>{nodeChildren}</code>;
            };

            components.pre = ({ children: nodeChildren, ...props }: any) => {
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
};

export default Markdown;
