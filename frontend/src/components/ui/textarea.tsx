import * as React from 'react';
import { useImperativeHandle } from 'react';

import { cn } from '@/lib/utils';

interface UseTextareaProps {
    maxHeight?: number;
    minHeight?: number;
    textareaRef: React.MutableRefObject<HTMLTextAreaElement | null>;
    triggerAutoSize: string;
}

const useTextarea = ({
    maxHeight = Number.MAX_SAFE_INTEGER,
    minHeight = 0,
    textareaRef,
    triggerAutoSize,
}: UseTextareaProps) => {
    const [init, setInit] = React.useState(true);

    React.useEffect(() => {
        const offsetBorder = 0;
        const textareaElement = textareaRef.current;

        if (!textareaElement) {
            return;
        }

        if (init) {
            textareaElement.style.minHeight = `${minHeight + offsetBorder}px`;

            if (maxHeight > minHeight) {
                textareaElement.style.maxHeight = `${maxHeight}px`;
            }

            setInit(false);
        }

        textareaElement.style.height = `${minHeight + offsetBorder}px`;
        const scrollHeight = textareaElement.scrollHeight;
        textareaElement.style.height = scrollHeight > maxHeight ? `${maxHeight}px` : `${scrollHeight + offsetBorder}px`;
    }, [textareaRef.current, triggerAutoSize]);
};

type TextareaProps = React.TextareaHTMLAttributes<HTMLTextAreaElement> & {
    maxHeight?: number;
    minHeight?: number;
};

type TextareaRef = {
    maxHeight: number;
    minHeight: number;
    textarea: HTMLTextAreaElement;
};

const Textarea = React.forwardRef<TextareaRef, TextareaProps>(
    (
        { className, maxHeight = 118, minHeight = 38, onChange, value, ...props }: TextareaProps,
        ref: React.Ref<TextareaRef>,
    ) => {
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
            setTriggerAutoSize(value as string);
        }, [props?.defaultValue, value]);

        return (
            <textarea
                className={cn(
                    'flex w-full resize-none rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-xs placeholder:text-muted-foreground focus-visible:outline-hidden focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
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
    },
);
Textarea.displayName = 'Textarea';

export { Textarea };
