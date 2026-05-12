import { type Control, useWatch } from 'react-hook-form';

import type {
    KnowledgeAnswerType as KnowledgeAnswerTypeT,
    KnowledgeGuideType as KnowledgeGuideTypeT,
} from '@/graphql/types';

import { MarkdownEditor } from '@/components/shared/markdown-editor';
import { FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { InputGroup, InputGroupTextareaAutosize } from '@/components/ui/input-group';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';
import { KnowledgeAnswerType, KnowledgeDocType, KnowledgeGuideType } from '@/graphql/types';

import type { FormValues } from './knowledge-form';

// `<Select>` option lists. Co-located with the controls they feed because no
// other module needs them.
const docTypeValues = [KnowledgeDocType.Answer, KnowledgeDocType.Guide, KnowledgeDocType.Code] as const;
const guideTypeValues = Object.values(KnowledgeGuideType) as KnowledgeGuideTypeT[];
const answerTypeValues = Object.values(KnowledgeAnswerType) as KnowledgeAnswerTypeT[];

interface KnowledgeMetaFieldsProps {
    control: Control<FormValues>;
    isNew: boolean;
    isSaving: boolean;
}

export const KnowledgeMetaFields = ({ control, isNew, isSaving }: KnowledgeMetaFieldsProps) => {
    // Targeted subscription: only this component re-renders when docType changes,
    // not the whole form. The full-form `useWatch` from the original code
    // re-rendered on every keystroke in the markdown editor.
    const docType = useWatch({ control, name: 'docType' });

    return (
        <div className="flex flex-col gap-4">
            <div className="grid gap-4 sm:grid-cols-2">
                <FormField
                    control={control}
                    name="docType"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>Document type</FormLabel>
                            {isNew ? (
                                <Select
                                    disabled={isSaving}
                                    onValueChange={field.onChange}
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
                            ) : (
                                <div className="border-input bg-muted/30 text-muted-foreground flex h-9 items-center rounded-md border px-3 text-sm">
                                    {field.value || '—'}
                                </div>
                            )}
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
                                <FormControl>
                                    <Input
                                        disabled={isSaving}
                                        placeholder="e.g. python, go, typescript"
                                        {...field}
                                    />
                                </FormControl>
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
};

interface KnowledgeContentFieldProps {
    control: Control<FormValues>;
    /** When `true`, the editor stretches to fill its parent (desktop split view). */
    fillParent?: boolean;
    isSaving: boolean;
    showLabel?: boolean;
}

export const KnowledgeContentField = ({
    control,
    fillParent = false,
    isSaving,
    showLabel = false,
}: KnowledgeContentFieldProps) => (
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
