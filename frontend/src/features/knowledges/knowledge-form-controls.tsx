import { type Control, useFormContext, useWatch } from 'react-hook-form';

import type {
    KnowledgeAnswerType as KnowledgeAnswerTypeT,
    KnowledgeGuideType as KnowledgeGuideTypeT,
} from '@/graphql/types';

import { MarkdownEditor } from '@/components/shared/markdown-editor';
import {
    Autocomplete,
    AutocompleteContent,
    AutocompleteEmpty,
    AutocompleteGroup,
    AutocompleteInput,
    AutocompleteItem,
} from '@/components/ui/autocomplete';
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { InputGroup, InputGroupTextareaAutosize } from '@/components/ui/input-group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { KnowledgeAnswerType, KnowledgeDocType, KnowledgeGuideType } from '@/graphql/types';

import type { FormValues } from './knowledge-form';

import { KNOWLEDGE_LIMITS } from './knowledge-form';

// `<Select>` option lists. Co-located with the controls they feed because no
// other module needs them.
const docTypeValues = [KnowledgeDocType.Answer, KnowledgeDocType.Guide, KnowledgeDocType.Code] as const;
const guideTypeValues = Object.values(KnowledgeGuideType) as KnowledgeGuideTypeT[];
const answerTypeValues = Object.values(KnowledgeAnswerType) as KnowledgeAnswerTypeT[];

// The backend stores `code_lang` as a free-form string and uses it both as a
// vector-store filter and as the markdown code-block tag (see
// `backend/pkg/tools/code.go`). There is no enum to sync against — this is
// purely a UX list of common values surfaced as combobox suggestions; users
// can still type any custom identifier.
const LANGUAGES = [
    'bash',
    'c',
    'cpp',
    'csharp',
    'css',
    'dockerfile',
    'go',
    'groovy',
    'haskell',
    'html',
    'java',
    'javascript',
    'json',
    'kotlin',
    'lua',
    'markdown',
    'nginx',
    'perl',
    'php',
    'powershell',
    'python',
    'ruby',
    'rust',
    'scala',
    'shell',
    'sql',
    'swift',
    'toml',
    'typescript',
    'xml',
    'yaml',
] as const;

interface KnowledgeContentFieldProps {
    control: Control<FormValues>;
    /** When `true`, the editor stretches to fill its parent (desktop split view). */
    fillParent?: boolean;
    isSaving: boolean;
    showLabel?: boolean;
}

interface KnowledgeMetaFieldsProps {
    control: Control<FormValues>;
    isNew: boolean;
    isSaving: boolean;
}

export function KnowledgeContentField({
    control,
    fillParent = false,
    isSaving,
    showLabel = false,
}: KnowledgeContentFieldProps) {
    return (
        <FormField
            control={control}
            name="content"
            render={({ field }) => (
                <FormItem className={fillParent ? 'flex min-h-0 flex-1 flex-col' : undefined}>
                    {showLabel ? <FormLabel>Content</FormLabel> : null}
                    <FormControl>
                        <MarkdownEditor
                            className={fillParent ? 'min-h-0 flex-1' : 'min-h-[280px]'}
                            contentClassName={fillParent ? undefined : 'min-h-[240px]'}
                            disabled={isSaving}
                            onBlur={field.onBlur}
                            onChange={field.onChange}
                            placeholder="Knowledge content (will be embedded into the vector store)"
                            value={field.value}
                        />
                    </FormControl>
                    <FormMessage />
                </FormItem>
            )}
        />
    );
}

export function KnowledgeMetaFields({ control, isNew, isSaving }: KnowledgeMetaFieldsProps) {
    // Targeted subscription: only this component re-renders when docType changes,
    // not the whole form. The full-form `useWatch` from the original code
    // re-rendered on every keystroke in the markdown editor.
    const docType = useWatch({ control, name: 'docType' });
    // `setValue` is used to clear subtype fields that no longer apply when the
    // user switches docType. We do this synchronously inside `onValueChange`
    // (rather than via `useEffect`) so that a fresh load of an existing
    // document doesn't wipe its persisted subtype on first render.
    const { setValue } = useFormContext<FormValues>();

    const handleDocTypeChange = (next: KnowledgeDocType, fieldOnChange: (value: KnowledgeDocType) => void) => {
        fieldOnChange(next);
        const opts = { shouldDirty: true, shouldValidate: true };

        if (next !== KnowledgeDocType.Answer) {
            setValue('answerType', undefined, opts);
        }

        if (next !== KnowledgeDocType.Code) {
            setValue('codeLang', '', opts);
        }

        if (next !== KnowledgeDocType.Guide) {
            setValue('guideType', undefined, opts);
        }
    };

    return (
        <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
                <FormField
                    control={control}
                    name="docType"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Document type</FormLabel>
                            <Select
                                disabled={isSaving}
                                onValueChange={(value) =>
                                    handleDocTypeChange(value as KnowledgeDocType, field.onChange)
                                }
                                value={field.value}
                            >
                                <FormControl>
                                    <SelectTrigger>
                                        <SelectValue placeholder="Select type" />
                                    </SelectTrigger>
                                </FormControl>
                                <SelectContent>
                                    {docTypeValues.map((value) => (
                                        <SelectItem
                                            key={value}
                                            value={value}
                                        >
                                            {value}
                                        </SelectItem>
                                    ))}
                                </SelectContent>
                            </Select>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                {docType === KnowledgeDocType.Guide ? (
                    <FormField
                        control={control}
                        name="guideType"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Guide type</FormLabel>
                                <Select
                                    disabled={isSaving}
                                    onValueChange={field.onChange}
                                    value={field.value}
                                >
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select guide type" />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        {guideTypeValues.map((value) => (
                                            <SelectItem
                                                key={value}
                                                value={value}
                                            >
                                                {value}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                ) : null}

                {docType === KnowledgeDocType.Answer ? (
                    <FormField
                        control={control}
                        name="answerType"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Answer type</FormLabel>
                                <Select
                                    disabled={isSaving}
                                    onValueChange={field.onChange}
                                    value={field.value}
                                >
                                    <FormControl>
                                        <SelectTrigger>
                                            <SelectValue placeholder="Select answer type" />
                                        </SelectTrigger>
                                    </FormControl>
                                    <SelectContent>
                                        {answerTypeValues.map((value) => (
                                            <SelectItem
                                                key={value}
                                                value={value}
                                            >
                                                {value}
                                            </SelectItem>
                                        ))}
                                    </SelectContent>
                                </Select>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                ) : null}

                {docType === KnowledgeDocType.Code ? (
                    <FormField
                        control={control}
                        name="codeLang"
                        render={({ field }) => (
                            <FormItem>
                                <FormLabel>Code language</FormLabel>
                                {/*
                                 * `Autocomplete` is a free-text input with a
                                 * suggestion popover — the backend accepts any
                                 * string here, so the dropdown is a UX hint
                                 * rather than a closed enum.
                                 */}
                                <Autocomplete
                                    onValueChange={field.onChange}
                                    value={field.value ?? ''}
                                >
                                    <FormControl>
                                        <AutocompleteInput
                                            disabled={isSaving}
                                            maxLength={KNOWLEDGE_LIMITS.codeLang}
                                            name={field.name}
                                            onBlur={field.onBlur}
                                            placeholder="e.g. python, go, typescript"
                                            ref={field.ref}
                                        />
                                    </FormControl>
                                    <AutocompleteContent>
                                        <AutocompleteEmpty>No matching language</AutocompleteEmpty>
                                        <AutocompleteGroup>
                                            {LANGUAGES.map((lang) => (
                                                <AutocompleteItem
                                                    key={lang}
                                                    value={lang}
                                                >
                                                    {lang}
                                                </AutocompleteItem>
                                            ))}
                                        </AutocompleteGroup>
                                    </AutocompleteContent>
                                </Autocomplete>
                                <FormMessage />
                            </FormItem>
                        )}
                    />
                ) : null}
            </div>

            <FormField
                control={control}
                name="question"
                render={({ field }) => (
                    <FormItem>
                        <FormLabel>Question</FormLabel>
                        <FormControl>
                            <InputGroup className="block">
                                <InputGroupTextareaAutosize
                                    {...field}
                                    autoFocus={isNew}
                                    className="min-h-0"
                                    disabled={isSaving}
                                    maxLength={KNOWLEDGE_LIMITS.question}
                                    maxRows={6}
                                    minRows={1}
                                    placeholder="Short title or question this document answers"
                                />
                            </InputGroup>
                        </FormControl>
                        <FormMessage />
                    </FormItem>
                )}
            />

            <FormField
                control={control}
                name="description"
                render={({ field }) => (
                    <FormItem>
                        <FormLabel>Description (optional)</FormLabel>
                        <FormControl>
                            <InputGroup className="block">
                                <InputGroupTextareaAutosize
                                    {...field}
                                    className="min-h-0"
                                    disabled={isSaving}
                                    maxLength={KNOWLEDGE_LIMITS.description}
                                    maxRows={8}
                                    minRows={1}
                                    placeholder="Optional short description"
                                />
                            </InputGroup>
                        </FormControl>
                        <FormMessage />
                    </FormItem>
                )}
            />
        </div>
    );
}
