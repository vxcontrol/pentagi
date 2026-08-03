// Shell classes for the editor box. Kept in this tiny leaf module (no tiptap import) so MarkdownEditorField's
// lazy-load fallback can wear the same box as the mounted editor without pulling the tiptap chunk eagerly.
export const MARKDOWN_EDITOR_WRAPPER_CLASS =
    'border-input dark:bg-input/30 group/markdown-editor flex w-full flex-col overflow-hidden rounded-md border shadow-2xs outline-hidden transition-[color,box-shadow] has-[[aria-invalid=true]]:border-destructive';

// The editor's scrollable content wrapper. The link/image/table overlay handles find their dismiss-on-scroll
// parent by THIS class — a behavioral hook, not just styling — so it must stay in sync with the className that
// stamps it on the contenteditable (markdown-editor.tsx) and with getEditorScrollParent below.
export const EDITOR_CONTENT_CLASS = 'tiptap-content';

export const getEditorScrollParent = (editorDom: Element): Element | Window =>
    editorDom.closest(`.${EDITOR_CONTENT_CLASS}`) ?? window;
