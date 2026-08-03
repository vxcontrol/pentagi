import * as React from 'react';
import { useImperativeHandle } from 'react';

import { cn } from '@/lib/utils';

export type TextareaRef = {
    focus: () => void;
    maxHeight: number;
    minHeight: number;
    textarea: HTMLTextAreaElement;
};

type TextareaProps = React.ComponentProps<'textarea'> & {
    autoSize?: boolean;
    maxHeight?: number;
    minHeight?: number;
};

interface UseTextareaProps {
    enabled?: boolean;
    maxHeight?: number;
    minHeight?: number;
    textareaRef: React.MutableRefObject<HTMLTextAreaElement | null>;
    triggerAutoSize: string;
}

function Textarea({
    autoSize = true,
    className,
    maxHeight = 118,
    minHeight = 38,
    onChange,
    ref,
    value,
    ...props
}: Omit<TextareaProps, 'ref'> & { ref?: React.Ref<TextareaRef> }) {
    const textareaRef = React.useRef<HTMLTextAreaElement | null>(null);
    const [triggerAutoSize, setTriggerAutoSize] = React.useState('');

    useTextarea({
        enabled: autoSize,
        maxHeight,
        minHeight,
        textareaRef,
        triggerAutoSize: triggerAutoSize,
    });

    useImperativeHandle(ref, () => ({
        focus: () => textareaRef?.current?.focus(),
        maxHeight,
        minHeight,
        textarea: textareaRef.current as HTMLTextAreaElement,
    }));

    React.useEffect(() => {
        // eslint-disable-next-line react-hooks/set-state-in-effect -- syncs internal auto-size trigger with controlled value prop
        setTriggerAutoSize(value as string);
    }, [props?.defaultValue, value]);

    return (
        <textarea
            className={cn(
                'aria-invalid:ring-destructive/20 dark:aria-invalid:ring-destructive/40 aria-invalid:border-destructive border-input dark:bg-input/30 placeholder:text-muted-foreground focus-visible:ring-ring flex w-full min-w-0 resize-none rounded-md border bg-transparent px-3 py-2 text-base shadow-xs focus-visible:ring-1 focus-visible:outline-hidden disabled:cursor-not-allowed disabled:opacity-50 md:text-sm',
                className,
            )}
            data-slot="textarea"
            ref={textareaRef}
            {...props}
            onChange={(e) => {
                setTriggerAutoSize(e.target.value);
                onChange?.(e);
            }}
            value={value}
        />
    );
}

function useTextarea({
    enabled = true,
    maxHeight = Number.MAX_SAFE_INTEGER,
    minHeight = 0,
    textareaRef,
    triggerAutoSize,
}: UseTextareaProps) {
    const initRef = React.useRef(true);

    React.useEffect(() => {
        const offsetBorder = 0;
        const textareaElement = textareaRef.current;

        if (!enabled || !textareaElement) {
            return;
        }

        if (initRef.current) {
            textareaElement.style.minHeight = `${minHeight + offsetBorder}px`;

            if (maxHeight > minHeight) {
                textareaElement.style.maxHeight = `${maxHeight}px`;
            }

            initRef.current = false;
        }

        textareaElement.style.height = `${minHeight + offsetBorder}px`;
        const scrollHeight = textareaElement.scrollHeight;
        textareaElement.style.height = scrollHeight > maxHeight ? `${maxHeight}px` : `${scrollHeight + offsetBorder}px`;
    }, [enabled, triggerAutoSize, maxHeight, minHeight, textareaRef]);
}

export { Textarea };
