import type { ReactNode } from 'react';

import { Ellipsis, LibraryBig, Loader2, Pencil, Trash } from 'lucide-react';
import { useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import type { KnowledgeDocumentFragmentFragment } from '@/graphql/types';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { InlineRenameInput } from '@/components/shared/inline-rename-input';
import { ListNavigationToolbar } from '@/components/shared/list-navigation-toolbar';
import { Badge } from '@/components/ui/badge';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useInlineEditTitle } from '@/hooks/use-inline-edit-title';
import { type Knowledge, useKnowledges } from '@/providers/knowledges-provider';

import { useKnowledgeDetailNavigation } from './use-knowledge-detail-navigation';

interface KnowledgeHeaderProps {
    isNew: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
    /**
     * Optional hook called right before the header navigates away after a
     * successful delete. The form mounts this header inside an unsaved-changes
     * guard, so it passes `skipNextBlock` here to suppress the "Save before
     * leaving?" dialog — there is nothing to save once the document is gone.
     */
    onBeforeNavigateAway?: () => void;
    saveButton?: ReactNode;
}

export const KnowledgeHeader = ({ isNew, knowledge, onBeforeNavigateAway, saveButton }: KnowledgeHeaderProps) => {
    const navigate = useNavigate();
    const { deleteKnowledge, updateKnowledge } = useKnowledges();
    const [isRenaming, setIsRenaming] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

    const knowledgeId = knowledge?.id ?? null;

    const { toolbarProps: knowledgeToolbarProps } = useKnowledgeDetailNavigation(knowledgeId);

    // Title source-of-truth is the server-side `question`. We intentionally do
    // not read it from the form draft below — the inline rename flow in this
    // header writes through `updateKnowledge`, which refreshes `knowledge` via
    // the cache, and the form picks up the new value separately.
    const knowledgeName = knowledge?.question ?? null;
    const canShowActions = !isNew && !!knowledge;

    const {
        handleDropdownCloseAutoFocus,
        inputRef: editingInputRef,
        isEditing: isEditingTitle,
        startEdit: handleRenameStart,
        stopEdit: handleRenameCancel,
    } = useInlineEditTitle({ resetKey: knowledgeId });

    const handleRenameSave = useCallback(async () => {
        const newQuestion = editingInputRef.current?.value.trim();

        if (!knowledge || !newQuestion) {
            return;
        }

        if (newQuestion === knowledge.question) {
            handleRenameCancel();

            return;
        }

        setIsRenaming(true);

        try {
            // Backend requires `content` on update (always re-embeds). We pass
            // the server's current `content` so an inline rename never
            // accidentally overwrites unsaved edits made in the form below.
            // The sibling form picks up the new `question` automatically via
            // `useForm({ values })` — no manual sync needed here.
            await updateKnowledge(knowledge.id, {
                content: knowledge.content,
                question: newQuestion,
            });
            toast.success('Knowledge renamed successfully');
            handleRenameCancel();
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsRenaming(false);
        }
    }, [editingInputRef, handleRenameCancel, knowledge, updateKnowledge]);

    const handleDelete = useCallback(async () => {
        if (!knowledgeId) {
            return;
        }

        setIsDeleting(true);

        try {
            await deleteKnowledge(knowledgeId);
            onBeforeNavigateAway?.();
            navigate('/knowledges', { replace: true });
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsDeleting(false);
        }
    }, [knowledgeId, deleteKnowledge, navigate, onBeforeNavigateAway]);

    return (
        <>
            <header className="bg-background sticky top-0 z-10 flex h-12 shrink-0 items-center gap-2 border-b px-4">
                <SidebarTrigger className="-ml-1" />
                <Separator
                    className="mr-2 h-4"
                    orientation="vertical"
                />
                <Breadcrumb>
                    <BreadcrumbList>
                        <BreadcrumbItem className="gap-2">
                            <LibraryBig className="size-4 shrink-0" />
                            {isEditingTitle && canShowActions ? (
                                <InlineRenameInput
                                    busy={isRenaming}
                                    className="w-64 max-w-full"
                                    defaultValue={knowledgeName ?? ''}
                                    inputRef={editingInputRef}
                                    onCancel={handleRenameCancel}
                                    onSave={handleRenameSave}
                                    placeholder="Knowledge question"
                                />
                            ) : canShowActions ? (
                                <Tooltip>
                                    <TooltipTrigger asChild>
                                        <BreadcrumbPage
                                            className="max-w-[240px] cursor-text truncate select-none"
                                            onDoubleClick={handleRenameStart}
                                        >
                                            {knowledgeName ?? 'Knowledge'}
                                        </BreadcrumbPage>
                                    </TooltipTrigger>
                                    <TooltipContent>Double-click to rename</TooltipContent>
                                </Tooltip>
                            ) : (
                                <BreadcrumbPage className="max-w-[240px] truncate">
                                    {isNew ? 'New knowledge' : (knowledgeName ?? 'Knowledge')}
                                </BreadcrumbPage>
                            )}
                        </BreadcrumbItem>
                    </BreadcrumbList>
                </Breadcrumb>
                <div className="ml-auto flex items-center gap-2">
                    {canShowActions && (
                        <ListNavigationToolbar<Knowledge>
                            {...knowledgeToolbarProps}
                            renderItem={(item, isCurrent) => (
                                <>
                                    <Badge
                                        className="shrink-0 whitespace-nowrap text-[10px]"
                                        variant="outline"
                                    >
                                        {item.docType}
                                    </Badge>
                                    <span className={isCurrent ? 'truncate font-medium' : 'truncate'}>
                                        {item.question}
                                    </span>
                                </>
                            )}
                            sheetIcon={<LibraryBig className="size-4" />}
                            sheetTitle="Knowledges"
                        />
                    )}
                    {saveButton}
                    {canShowActions && (
                        <DropdownMenu>
                            <DropdownMenuTrigger asChild>
                                <Button
                                    aria-label="Knowledge actions"
                                    className="size-8 p-0"
                                    type="button"
                                    variant="ghost"
                                >
                                    <Ellipsis />
                                </Button>
                            </DropdownMenuTrigger>
                            <DropdownMenuContent
                                align="end"
                                className="min-w-24"
                                onCloseAutoFocus={handleDropdownCloseAutoFocus}
                            >
                                <DropdownMenuItem onClick={handleRenameStart}>
                                    <Pencil className="size-3" />
                                    Rename
                                </DropdownMenuItem>
                                <DropdownMenuSeparator />
                                <DropdownMenuItem
                                    disabled={isDeleting}
                                    onClick={() => setIsDeleteDialogOpen(true)}
                                >
                                    {isDeleting ? (
                                        <>
                                            <Loader2 className="size-4 animate-spin" />
                                            Deleting...
                                        </>
                                    ) : (
                                        <>
                                            <Trash className="size-4" />
                                            Delete
                                        </>
                                    )}
                                </DropdownMenuItem>
                            </DropdownMenuContent>
                        </DropdownMenu>
                    )}
                </div>
            </header>
            <ConfirmationDialog
                cancelText="Cancel"
                confirmText="Delete"
                handleConfirm={handleDelete}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={knowledgeName ?? undefined}
                itemType="knowledge document"
            />
        </>
    );
};
