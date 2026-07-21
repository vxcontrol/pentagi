import type { AriaAttributes, Ref } from 'react';

import { Loader2 } from 'lucide-react';
import { lazy, Suspense, useImperativeHandle, useRef } from 'react';

import type { TextareaRef } from '@/components/ui/textarea';

import { Textarea } from '@/components/ui/textarea';
import { cn } from '@/lib/utils';

import type { MarkdownEditorHandle } from './markdown-editor';
import type { EditorViewMode } from './markdown-editor-view-mode';

import { MARKDOWN_EDITOR_WRAPPER_CLASS } from './markdown-editor-styles';
import { insertAtTextareaCaret, selectNextTextareaUse } from './markdown-editor-textarea';

// Static-importing MarkdownEditor would merge the tiptap chunk into every route that pulls a util from the barrel.
const MarkdownEditor = lazy(() => import('./markdown-editor').then((module) => ({ default: module.MarkdownEditor })));

/**
 * Imperative handle exposed via `ref`. Every method works in BOTH modes — the raw textarea and the rich editor
 * each implement them — so a consumer drives the field the same way whether it renders raw source or rich.
 */
export interface MarkdownEditorFieldHandle {
    focus: () => void;
    insertAtCursor: (text: string) => void;
    selectNextUse: (variable: string) => boolean;
}

interface MarkdownEditorFieldProps extends Pick<AriaAttributes, 'aria-describedby' | 'aria-invalid' | 'aria-label'> {
    className?: string;
    disabled?: boolean;
    id?: string;
    mode: EditorViewMode;
    onBlur?: () => void;
    onChange: (value: string) => void;
    placeholder?: string;
    ref?: Ref<MarkdownEditorFieldHandle>;
    value: string;
}

export function MarkdownEditorField({
    'aria-describedby': ariaDescribedby,
    'aria-invalid': ariaInvalid,
    'aria-label': ariaLabel,
    className,
    disabled,
    id,
    mode,
    onBlur,
    onChange,
    placeholder,
    ref,
    value,
}: MarkdownEditorFieldProps) {
    const rawRef = useRef<TextareaRef>(null);
    const richRef = useRef<MarkdownEditorHandle>(null);

    useImperativeHandle(
        ref,
        () =>
            mode === 'raw'
                ? {
                      focus: () => rawRef.current?.focus(),
                      // A disabled field must not be mutated through the handle (a variable click mid-save
                      // would splice the value and dirty the form); the rich branch guards on editor.isEditable.
                      insertAtCursor: (text) => {
                          if (!disabled && rawRef.current) {
                              insertAtTextareaCaret(rawRef.current.textarea, text, onChange);
                          }
                      },
                      selectNextUse: (variable) =>
                          rawRef.current ? selectNextTextareaUse(rawRef.current.textarea, variable) : false,
                  }
                : {
                      focus: () => richRef.current?.focus(),
                      insertAtCursor: (text) => richRef.current?.insertAtCursor(text),
                      selectNextUse: (variable) => richRef.current?.selectNextUse(variable) ?? false,
                  },
        [mode, onChange, disabled],
    );

    const boxClassName = cn('h-[calc(100dvh-5rem)] xl:h-auto xl:min-h-0 xl:flex-1', className);

    if (mode === 'raw') {
        return (
            <Textarea
                aria-describedby={ariaDescribedby}
                aria-invalid={ariaInvalid}
                aria-label={ariaLabel}
                autoSize={false}
                // Raw config is applied LAST so a consumer `className` can't override the byte-exact source styling.
                className={cn(boxClassName, 'aria-invalid:border-destructive resize-none font-mono text-sm')}
                disabled={disabled}
                id={id}
                onBlur={onBlur}
                onChange={(event) => onChange(event.target.value)}
                placeholder={placeholder}
                ref={rawRef}
                value={value}
            />
        );
    }

    return (
        <Suspense
            fallback={
                <div
                    aria-busy="true"
                    aria-describedby={ariaDescribedby}
                    aria-invalid={ariaInvalid}
                    aria-label={ariaLabel}
                    className={cn(
                        MARKDOWN_EDITOR_WRAPPER_CLASS,
                        'items-center justify-center',
                        disabled && 'pointer-events-none opacity-60',
                        boxClassName,
                    )}
                    id={id}
                >
                    <Loader2 className="text-muted-foreground size-5 animate-spin" />
                </div>
            }
        >
            <MarkdownEditor
                aria-describedby={ariaDescribedby}
                aria-invalid={ariaInvalid}
                aria-label={ariaLabel}
                className={boxClassName}
                disabled={disabled}
                id={id}
                onBlur={onBlur}
                onChange={onChange}
                placeholder={placeholder}
                ref={richRef}
                value={value}
            />
        </Suspense>
    );
}
