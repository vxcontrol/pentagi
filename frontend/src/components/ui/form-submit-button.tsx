import { Loader2 } from 'lucide-react';
import { useFormContext } from 'react-hook-form';

import { Button } from '@/components/ui/button';

interface FormSubmitButtonProps extends React.ComponentProps<typeof Button> {
    /**
     * When true (the default), the button is disabled as soon as the form has
     * been submitted once with invalid values — i.e. mirrors the "click first,
     * see errors, then re-disable" UX our forms already use. Pass `false` for
     * flows that must accept clicks even with invalid state (e.g. multi-step
     * forms where submit also runs validation manually).
     */
    requireValid?: boolean;
}

/**
 * Submit button that reads the pending/validity state of the nearest
 * react-hook-form context. Drops in next to a <Form {...form}> wrapper without
 * any prop drilling — the button shows a spinner while form.handleSubmit(...)
 * is awaiting and disables itself once the form is dirty-but-invalid.
 *
 * This is the RHF-flavoured analogue of React 19's useFormStatus — same idea
 * (child component subscribes to the form's status), different source
 * (useFormContext vs the form-action runtime) because our forms run through
 * react-hook-form's handleSubmit, not native <form action>.
 */
function FormSubmitButton({ children, requireValid = true, ...props }: FormSubmitButtonProps) {
    const { formState } = useFormContext();
    const { isSubmitted, isSubmitting, isValid } = formState;
    const isDisabled = isSubmitting || (requireValid && isSubmitted && !isValid);

    return (
        <Button
            disabled={isDisabled}
            type="submit"
            {...props}
        >
            {isSubmitting && <Loader2 className="size-4 animate-spin" />}
            {children}
        </Button>
    );
}

export { FormSubmitButton };
