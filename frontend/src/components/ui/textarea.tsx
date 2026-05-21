import * as React from 'react';
import { useImperativeHandle } from 'react';

import { cn } from '@/lib/utils';

type TextareaProps = React.ComponentProps<'textarea'> & {
    maxHeight?: number;
    minHeight?: number;
};

type TextareaRef = {
    maxHeight: number;
    minHeight: number;
    textarea: HTMLTextAreaElement;
};

interface UseTextareaProps {
    maxHeight?: number;
    minHeight?: number;
    textareaRef: React.MutableRefObject<HTMLTextAreaElement | null>;
    triggerAutoSize: string;
}

function Textarea({
    className,
    maxHeight = 118,
    minHeight = 38,
    onChange,
    ref,
    value,
    ...props
}: TextareaProps & { ref?: React.Ref<TextareaRef> }) {
    const textareaRef = React.useRef<HTMLTextAreaElement | null>(null);
    const [triggerAutoSize, setTriggerAutoSize] = React.useState('');

    useTextarea({
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
                'border-input placeholder:text-muted-foreground focus-visible:ring-ring flex w-full resize-none rounded-md border bg-transparent px-3 py-2 text-sm shadow-xs focus-visible:ring-1 focus-visible:outline-hidden disabled:cursor-not-allowed disabled:opacity-50',
                className,
            )}
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
    maxHeight = Number.MAX_SAFE_INTEGER,
    minHeight = 0,
    textareaRef,
    triggerAutoSize,
}: UseTextareaProps) {
    const initRef = React.useRef(true);

    React.useEffect(() => {
        const offsetBorder = 0;
        const textareaElement = textareaRef.current;

        if (!textareaElement) {
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
    }, [triggerAutoSize, maxHeight, minHeight, textareaRef]);
}

export { Textarea };
