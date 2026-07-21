import type { FieldValues, Resolver, UseFormProps, UseFormReturn } from 'react-hook-form';
import type { z } from 'zod';

import { zodResolver } from '@hookform/resolvers/zod';
import { useForm } from 'react-hook-form';

interface UseAppFormProps<TFieldValues extends FieldValues, TContext, TTransformedValues> extends Omit<
    UseFormProps<TFieldValues, TContext, TTransformedValues>,
    'mode' | 'resolver' | 'reValidateMode'
> {
    // The schema's OUTPUT is the form's transformed values; its INPUT is the raw field values (they differ when
    // the schema has defaults/transforms). zodResolver maps input→output — the cast below keeps callers from
    // having to align the two by hand, matching how `useForm<Input, ctx, Output>` is used directly.
    schema: z.ZodType<TTransformedValues>;
}

// Project form convention (see the forms note in memory): validation stays SILENT until the first submit, then
// re-validates live on every change — react-hook-form's default (mode 'onSubmit' + reValidateMode 'onChange').
// This wrapper owns the timing and the zod resolver so no form can drift back to eager 'onTouched'/'onChange',
// which paints fields red on input before the user has tried to save. Callers pass `schema` in place of
// `resolver`, plus the usual defaultValues/values/resetOptions; the `Omit` makes the timing keys unreachable.
// The three generics mirror `useForm<TFieldValues, TContext, TTransformedValues>` so schemas that transform
// (input ≠ output) keep their precise typing.
export function useAppForm<TFieldValues extends FieldValues, TContext = unknown, TTransformedValues = TFieldValues>({
    schema,
    ...options
}: UseAppFormProps<TFieldValues, TContext, TTransformedValues>): UseFormReturn<
    TFieldValues,
    TContext,
    TTransformedValues
> {
    return useForm<TFieldValues, TContext, TTransformedValues>({
        mode: 'onSubmit',
        reValidateMode: 'onChange',
        ...options,
        resolver: zodResolver(schema as z.ZodType<TTransformedValues, TFieldValues>) as Resolver<
            TFieldValues,
            TContext,
            TTransformedValues
        >,
    });
}
