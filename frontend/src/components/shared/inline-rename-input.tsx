import { Check, Loader2, X } from 'lucide-react';
import { type KeyboardEvent, type Ref } from 'react';

import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { cn } from '@/lib/utils';

interface InlineRenameInputProps {
    /**
     * Auto-focus the input on mount. Needed for table-cell call sites that
     * switch into edit mode in-place: the parent flips a flag (e.g.
     * `editingFlowId === flow.id`) and this component is freshly mounted,
     * so focus must be requested explicitly.
     */
    autoFocus?: boolean;
    /** Disable input + Save button while a mutation is in flight. */
    busy?: boolean;
    /** Optional className passed through to the outer `<InputGroup>`. */
    className?: string;
    /** Initial value rendered inside the uncontrolled input. */
    defaultValue?: string;
    /** Ref to the underlying `<input>` element. Pair with `useInlineEditTitle().inputRef`. */
    inputRef?: Ref<HTMLInputElement>;
    onCancel: () => void;
    /**
     * Save handler. Read the latest text from the bound `inputRef` — kept
     * uncontrolled because most callers commit on Enter or click and have
     * no need to reflect every keystroke into React state.
     */
    onSave: () => void;
    placeholder?: string;
}

/**
 * Inline rename input used inside table cells and detail-page breadcrumbs.
 *
 * Pairs with {@link useInlineEditTitle} — the parent owns the open/close
 * state and supplies a ref via that hook; this component owns the
 * presentation (input + Save/Cancel addon buttons), keyboard semantics
 * (`Enter` saves, `Escape` cancels), and the loading spinner during save.
 *
 * The input is uncontrolled (`defaultValue`) to match the pattern across
 * the codebase: callers read the value at submit time from `inputRef.current`,
 * not from React state, which avoids a re-render per keystroke for a value
 * that's only relevant once.
 */
export const InlineRenameInput = ({
    autoFocus = false,
    busy = false,
    className,
    defaultValue,
    inputRef,
    onCancel,
    onSave,
    placeholder,
}: InlineRenameInputProps) => {
    const handleKeyDown = (event: KeyboardEvent<HTMLInputElement>) => {
        if (event.key === 'Enter') {
            event.preventDefault();
            onSave();

            return;
        }

        if (event.key === 'Escape') {
            event.preventDefault();
            onCancel();
        }
    };

    return (
        <InputGroup className={cn('h-8', className)}>
            <InputGroupInput
                autoFocus={autoFocus}
                className="text-foreground"
                defaultValue={defaultValue}
                onKeyDown={handleKeyDown}
                placeholder={placeholder}
                ref={inputRef}
            />
            <InputGroupAddon
                align="inline-end"
                className="gap-0 pr-2"
            >
                <InputGroupButton
                    aria-label="Save"
                    disabled={busy}
                    onClick={onSave}
                >
                    {busy ? <Loader2 className="animate-spin" /> : <Check />}
                </InputGroupButton>
                <InputGroupButton
                    aria-label="Cancel"
                    disabled={busy}
                    onClick={onCancel}
                >
                    <X />
                </InputGroupButton>
            </InputGroupAddon>
        </InputGroup>
    );
};
