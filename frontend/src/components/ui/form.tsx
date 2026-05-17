import * as LabelPrimitive from '@radix-ui/react-label';
import { Slot } from '@radix-ui/react-slot';
import * as React from 'react';
import {
    Controller,
    type ControllerProps,
    type FieldPath,
    type FieldValues,
    FormProvider,
    useFormContext,
} from 'react-hook-form';

import { Label } from '@/components/ui/label';
import { cn } from '@/lib/utils';

const Form = FormProvider;

type FormFieldContextValue<
    TFieldValues extends FieldValues = FieldValues,
    TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
> = {
    name: TName;
};

const FormFieldContext = React.createContext<FormFieldContextValue>({} as FormFieldContextValue);

const FormField = <
    TFieldValues extends FieldValues = FieldValues,
    TName extends FieldPath<TFieldValues> = FieldPath<TFieldValues>,
>({
    ...props
}: ControllerProps<TFieldValues, TName>) => {
    return (
        <FormFieldContext.Provider value={{ name: props.name }}>
            <Controller {...props} />
        </FormFieldContext.Provider>
    );
};

const useFormField = () => {
    const fieldContext = React.useContext(FormFieldContext);
    const itemContext = React.useContext(FormItemContext);
    const { formState, getFieldState } = useFormContext();

    const fieldState = getFieldState(fieldContext.name, formState);

    if (!fieldContext) {
        throw new Error('useFormField should be used within <FormField>');
    }

    const { id } = itemContext;

    return {
        formDescriptionId: `${id}-form-item-description`,
        formItemId: `${id}-form-item`,
        formMessageId: `${id}-form-item-message`,
        id,
        name: fieldContext.name,
        ...fieldState,
    };
};

type FormItemContextValue = {
    id: string;
};

const FormItemContext = React.createContext<FormItemContextValue>({} as FormItemContextValue);

const FormItem = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLDivElement> & { ref?: React.Ref<HTMLDivElement> }) => {
    const id = React.useId();

    return (
        <FormItemContext.Provider value={{ id }}>
            <div
                className={cn('flex flex-col gap-2', className)}
                ref={ref}
                {...props}
            />
        </FormItemContext.Provider>
    );
};

FormItem.displayName = 'FormItem';

const FormLabel = ({
    className,
    ref,
    ...props
}: React.ComponentPropsWithoutRef<typeof LabelPrimitive.Root> & {
    ref?: React.Ref<React.ElementRef<typeof LabelPrimitive.Root>>;
}) => {
    const { error, formItemId } = useFormField();

    return (
        <Label
            className={cn(error && 'text-destructive', className)}
            htmlFor={formItemId}
            ref={ref}
            {...props}
        />
    );
};

FormLabel.displayName = 'FormLabel';

const FormControl = ({
    ref,
    ...props
}: React.ComponentPropsWithoutRef<typeof Slot> & {
    ref?: React.Ref<React.ElementRef<typeof Slot>>;
}) => {
    const { error, formDescriptionId, formItemId, formMessageId } = useFormField();

    return (
        <Slot
            aria-describedby={!error ? `${formDescriptionId}` : `${formDescriptionId} ${formMessageId}`}
            aria-invalid={!!error}
            id={formItemId}
            ref={ref}
            {...props}
        />
    );
};

FormControl.displayName = 'FormControl';

const FormDescription = ({
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLParagraphElement> & { ref?: React.Ref<HTMLParagraphElement> }) => {
    const { formDescriptionId } = useFormField();

    return (
        <p
            className={cn('text-muted-foreground text-[0.8rem]', className)}
            id={formDescriptionId}
            ref={ref}
            {...props}
        />
    );
};

FormDescription.displayName = 'FormDescription';

const FormMessage = ({
    children,
    className,
    ref,
    ...props
}: React.HTMLAttributes<HTMLParagraphElement> & { ref?: React.Ref<HTMLParagraphElement> }) => {
    const { error, formMessageId } = useFormField();
    const body = error ? String(error?.message) : children;

    if (!body) {
        return null;
    }

    return (
        <p
            className={cn('text-destructive text-[0.8rem] font-medium', className)}
            id={formMessageId}
            ref={ref}
            {...props}
        >
            {body}
        </p>
    );
};

FormMessage.displayName = 'FormMessage';

export { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage, useFormField };
