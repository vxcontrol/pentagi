import { cn } from '@/lib/utils';
import * as React from 'react';
import { useImperativeHandle } from 'react';

interface UseTextareaProps {
    textareaRef: React.MutableRefObject<HTMLTextAreaElement | null>;
    minHeight?: number;
    maxHeight?: number;
    triggerAutoSize: string;
}

const useTextarea = ({
    textareaRef,
    triggerAutoSize,
    maxHeight = Number.MAX_SAFE_INTEGER,
    minHeight = 0,
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

type TextareaRef = {
    textarea: HTMLTextAreaElement;
    maxHeight: number;
    minHeight: number;
};

type TextareaProps = {
    maxHeight?: number;
    minHeight?: number;
} & React.TextareaHTMLAttributes<HTMLTextAreaElement>;

const Textarea = React.forwardRef<TextareaRef, TextareaProps>(
    (
        { maxHeight = 118, minHeight = 38, className, onChange, value, ...props }: TextareaProps,
        ref: React.Ref<TextareaRef>,
    ) => {
        const textareaRef = React.useRef<HTMLTextAreaElement | null>(null);
        const [triggerAutoSize, setTriggerAutoSize] = React.useState('');

        useTextarea({
            textareaRef,
            triggerAutoSize: triggerAutoSize,
            maxHeight,
            minHeight,
        });

        useImperativeHandle(ref, () => ({
            textarea: textareaRef.current as HTMLTextAreaElement,
            focus: () => textareaRef?.current?.focus(),
            maxHeight,
            minHeight,
        }));

        React.useEffect(() => {
            setTriggerAutoSize(value as string);
        }, [props?.defaultValue, value]);

        return (
            <textarea
                className={cn(
                    'flex w-full rounded-md border border-input bg-transparent px-3 py-2 shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50 text-sm resize-none',
                    className,
                )}
                ref={textareaRef}
                {...props}
                value={value}
                onChange={(e) => {
                    setTriggerAutoSize(e.target.value);
                    onChange?.(e);
                }}
            />
        );
    },
);
Textarea.displayName = 'Textarea';

export { Textarea };
