import type { Control } from 'react-hook-form';

import { GripVertical } from 'lucide-react';

import type { KnowledgeDocumentFragmentFragment } from '@/graphql/types';

import { Badge } from '@/components/ui/badge';
import { Card, CardContent } from '@/components/ui/card';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';

import type { FormValues } from './knowledge-form';

import { KnowledgeContentField, KnowledgeMetaFields } from './knowledge-form-controls';

interface KnowledgeFormLayoutProps {
    control: Control<FormValues>;
    isNew: boolean;
    isSaving: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
}

interface KnowledgeIntroBlockProps {
    isNew: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
}

export function KnowledgeFormLayoutDesktop({ control, isNew, isSaving, knowledge }: KnowledgeFormLayoutProps) {
    return (
        <div className="flex min-h-0 w-full max-w-full flex-1 overflow-hidden">
            <ResizablePanelGroup
                className="w-full"
                direction="horizontal"
            >
                <ResizablePanel
                    defaultSize={45}
                    minSize={30}
                >
                    <div className="h-full min-h-0 overflow-y-auto">
                        <Card className="mx-auto min-h-full w-full max-w-2xl rounded-none border-0">
                            <CardContent className="flex flex-col gap-6 py-6">
                                <KnowledgeIntroBlock
                                    isNew={isNew}
                                    knowledge={knowledge}
                                />
                                <KnowledgeMetaFields
                                    control={control}
                                    isNew={isNew}
                                    isSaving={isSaving}
                                />
                            </CardContent>
                        </Card>
                    </div>
                </ResizablePanel>
                <ResizableHandle withHandle>
                    <GripVertical className="size-4" />
                </ResizableHandle>
                <ResizablePanel
                    defaultSize={55}
                    minSize={30}
                >
                    <div className="flex h-full min-h-0 flex-col overflow-hidden p-4">
                        <KnowledgeContentField
                            control={control}
                            fillParent
                            isSaving={isSaving}
                        />
                    </div>
                </ResizablePanel>
            </ResizablePanelGroup>
        </div>
    );
}

export function KnowledgeFormLayoutMobile({ control, isNew, isSaving, knowledge }: KnowledgeFormLayoutProps) {
    return (
        <div className="flex min-w-0 flex-1 items-start justify-center p-4">
            <Card className="w-full max-w-3xl">
                <CardContent className="flex flex-col gap-6 pt-6">
                    <KnowledgeIntroBlock
                        isNew={isNew}
                        knowledge={knowledge}
                    />
                    <KnowledgeMetaFields
                        control={control}
                        isNew={isNew}
                        isSaving={isSaving}
                    />
                    <KnowledgeContentField
                        control={control}
                        isSaving={isSaving}
                        showLabel
                    />
                </CardContent>
            </Card>
        </div>
    );
}

function KnowledgeIntroBlock({ isNew, knowledge }: KnowledgeIntroBlockProps) {
    return (
        <div className="flex flex-col gap-4">
            <div className="text-center">
                <h1 className="text-2xl font-semibold">
                    {isNew ? 'Create a new knowledge document' : 'Edit knowledge document'}
                </h1>
                <p className="text-muted-foreground mt-2">
                    {isNew
                        ? 'Add an entry to the vector knowledge base'
                        : 'Edits to content or metadata will trigger re-embedding'}
                </p>
            </div>

            {!isNew && knowledge ? (
                <div className="text-muted-foreground flex flex-wrap items-center gap-2 text-xs">
                    <Badge variant={knowledge.manual ? 'secondary' : 'outline'}>
                        {knowledge.manual ? 'manual' : 'agent'}
                    </Badge>
                    {knowledge.flowId ? <Badge variant="outline">flow #{knowledge.flowId}</Badge> : null}
                    {knowledge.taskId ? <Badge variant="outline">task #{knowledge.taskId}</Badge> : null}
                    {knowledge.subtaskId ? <Badge variant="outline">subtask #{knowledge.subtaskId}</Badge> : null}
                    <span>·</span>
                    <span>
                        chunk {knowledge.partSize} of {knowledge.totalSize}
                    </span>
                </div>
            ) : null}
        </div>
    );
}
