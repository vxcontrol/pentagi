// Deliberately does NOT re-export the heavy MarkdownEditor value — that would statically pull the tiptap chunk
// into any route importing a light util from here. Consume the mode-switching MarkdownEditorField instead (it
// owns the lazy() boundary, so importing it is chunk-free until rich mode renders).
// For the raw rich editor standalone, import './markdown-editor' by path — that eagerly bundles tiptap into the
// route, so wrap it in lazy(() => import('./markdown-editor')) yourself if you need the chunk deferred.
export { MarkdownEditorField } from './markdown-editor-field';
export type { MarkdownEditorFieldHandle } from './markdown-editor-field';
export { VARIABLE_RE, variableUseRegex } from './markdown-editor-variable-syntax';
export { EditorViewModeToggle } from './markdown-editor-view-mode';
export type { EditorViewMode } from './markdown-editor-view-mode';
