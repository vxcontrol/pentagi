import type { MarkdownStorage } from 'tiptap-markdown';

declare module '@tiptap/core' {
    interface Storage {
        markdown: MarkdownStorage;
    }
}
