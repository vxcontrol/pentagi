import type { ReactNode } from 'react';

import { Check, Ellipsis, LibraryBig, Loader2, Pencil, Trash, X } from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import type { KnowledgeDocumentFragmentFragment } from '@/graphql/types';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Button } from '@/components/ui/button';
import {
    DropdownMenu,
    DropdownMenuContent,
    DropdownMenuItem,
    DropdownMenuSeparator,
    DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import { InputGroup, InputGroupAddon, InputGroupButton, InputGroupInput } from '@/components/ui/input-group';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useKnowledges } from '@/providers/knowledges-provider';

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
    const [isEditingTitle, setIsEditingTitle] = useState(false);
    const [isRenaming, setIsRenaming] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const editingInputRef = useRef<HTMLInputElement>(null);

    const knowledgeId = knowledge?.id ?? null;
    // Title source-of-truth is the server-side `question`. We intentionally do
    // not read it from the form draft below — the inline rename flow in this
    // header writes through `updateKnowledge`, which refreshes `knowledge` via
    // the cache, and the form picks up the new value separately.
    const knowledgeName = knowledge?.question ?? null;
    const canShowActions = !isNew && !!knowledge;

    // Reset inline-edit state when navigating between documents so the input
    // doesn't carry over a stale draft from a previous document.
    useEffect(() => {
        setIsEditingTitle(false);
    }, [knowledgeId]);

    // Focus and select the rename input when the inline editor opens. We can't
    // rely on `autoFocus` here: the input mounts inside the same render cycle
    // that closes the Radix DropdownMenu, and the dropdown's own focus restore
    // (scheduled via `requestAnimationFrame`) wins the race against React's
    // autoFocus effect. Defer our focus to the next frame so it lands *after*
    // Radix has finished its restore.
    useEffect(() => {
        if (!isEditingTitle) {
            return;
        }

        const id = requestAnimationFrame(() => {
            const input = editingInputRef.current;

            if (!input) {
                return;
            }

            input.focus();
            input.select();
        });

        return () => cancelAnimationFrame(id);
    }, [isEditingTitle]);

    const handleRenameStart = useCallback(() => {
        setIsEditingTitle(true);
    }, []);

    const handleRenameCancel = useCallback(() => {
        setIsEditingTitle(false);
    }, []);

    const handleRenameSave = useCallback(async () => {
        const newQuestion = editingInputRef.current?.value.trim();

        if (!knowledge || !newQuestion) {
            return;
        }

        if (newQuestion === knowledge.question) {
            setIsEditingTitle(false);

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
            setIsEditingTitle(false);
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsRenaming(false);
        }
    }, [knowledge, updateKnowledge]);

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
                                <InputGroup className="h-8 w-64 max-w-full">
                                    <InputGroupInput
                                        className="text-foreground"
                                        defaultValue={knowledgeName ?? ''}
                                        onKeyDown={(event) => {
                                            if (event.key === 'Enter') {
                                                event.preventDefault();
                                                handleRenameSave();

                                                return;
                                            }

                                            if (event.key === 'Escape') {
                                                event.preventDefault();
                                                handleRenameCancel();
                                            }
                                        }}
                                        placeholder="Knowledge question"
                                        ref={editingInputRef}
                                    />
                                    <InputGroupAddon
                                        align="inline-end"
                                        className="gap-0 pr-2"
                                    >
                                        <InputGroupButton
                                            aria-label="Save"
                                            disabled={isRenaming}
                                            onClick={() => handleRenameSave()}
                                        >
                                            {isRenaming ? <Loader2 className="animate-spin" /> : <Check />}
                                        </InputGroupButton>
                                        <InputGroupButton
                                            aria-label="Cancel"
                                            onClick={() => handleRenameCancel()}
                                        >
                                            <X />
                                        </InputGroupButton>
                                    </InputGroupAddon>
                                </InputGroup>
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
                                onCloseAutoFocus={(event) => {
                                    // Radix returns focus to the trigger on close. When the
                                    // selected action mounts the rename input, prevent that
                                    // restore so our deferred focus actually wins.
                                    if (isEditingTitle) {
                                        event.preventDefault();
                                    }
                                }}
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
