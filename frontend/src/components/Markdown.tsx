import 'highlight.js/styles/atom-one-dark.css';

import dockerfile from 'highlight.js/lib/languages/dockerfile';
import graphql from 'highlight.js/lib/languages/graphql';
import http from 'highlight.js/lib/languages/http';
import nginx from 'highlight.js/lib/languages/nginx';
import yaml from 'highlight.js/lib/languages/yaml';
import { common, createLowlight } from 'lowlight';
import ReactMarkdown from 'react-markdown';
import rehypeHighlight from 'rehype-highlight';
import remarkGfm from 'remark-gfm';

const lowlight = createLowlight();
lowlight.register('http', http);
lowlight.register('nginx', nginx);
lowlight.register('dockerfile', dockerfile);
lowlight.register('yaml', yaml);
lowlight.register('graphql', graphql);

interface MarkdownProps {
    children: string;
    className?: string;
}

const Markdown = ({ children, className }: MarkdownProps) => {
    return (
        <ReactMarkdown
            className={`prose prose-sm max-w-none dark:prose-invert ${className}`}
            remarkPlugins={[remarkGfm]}
            rehypePlugins={[
                [
                    rehypeHighlight,
                    {
                        languages: {
                            ...common,
                            http,
                            nginx,
                            dockerfile,
                            yaml,
                            graphql,
                        },
                        detect: true,
                    },
                ],
            ]}
        >
            {children}
        </ReactMarkdown>
    );
};

export default Markdown;
