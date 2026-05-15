import type { ReactNode } from 'react';

import { ChevronLeft, ChevronRight, Ellipsis, LibraryBig, Loader2, Pencil, Trash } from 'lucide-react';
import { useCallback, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

import type { KnowledgeDocumentFragmentFragment } from '@/graphql/types';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { DetailNavigationSheet, DetailNavigationToolbar, useNavigation } from '@/components/shared/detail-navigation';
import { InlineEditInput, useInlineEdit } from '@/components/shared/inline-edit';
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
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { mergeHrefWithSearchParams } from '@/lib/url-params';
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
    const [searchParams] = useSearchParams();
    const { isMobile } = useBreakpoint();
    const { deleteKnowledge, updateKnowledge } = useKnowledges();
    const [isRenaming, setIsRenaming] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

    const knowledgeId = knowledge?.id ?? null;

    const { toolbarProps: knowledgeToolbarProps } = useKnowledgeDetailNavigation(knowledgeId);

    // Mirror what `<DetailNavigationToolbar>` computes internally so the
    // mobile menu items share the same filtered subset as the desktop toolbar.
    const mobileNav = useNavigation<Knowledge>({
        currentId: knowledgeToolbarProps.currentId,
        getId: knowledgeToolbarProps.getId,
        getSearchableText: knowledgeToolbarProps.getSearchableText ?? knowledgeToolbarProps.getLabel,
        items: knowledgeToolbarProps.items,
        query: knowledgeToolbarProps.filter,
    });
    const [isMobileNavSheetOpen, setIsMobileNavSheetOpen] = useState(false);

    const mobileNavGoTo = useCallback(
        (id: null | string) => {
            if (!id) {
                return;
            }

            const target = mobileNav.filteredItems.find(
                (item) => String(knowledgeToolbarProps.getId(item)) === id,
            );

            if (!target) {
                return;
            }

            navigate(mergeHrefWithSearchParams(knowledgeToolbarProps.getHref(target), searchParams), {
                replace: true,
            });
        },
        [knowledgeToolbarProps, mobileNav.filteredItems, navigate, searchParams],
    );

    const mobileNavSelectItem = useCallback(
        (item: Knowledge) => {
            setIsMobileNavSheetOpen(false);
            navigate(mergeHrefWithSearchParams(knowledgeToolbarProps.getHref(item), searchParams), {
                replace: true,
            });
        },
        [knowledgeToolbarProps, navigate, searchParams],
    );

    const mobilePositionLabel = useMemo(
        () =>
            mobileNav.total === 0 || mobileNav.currentIndex === -1
                ? `–/${mobileNav.total}`
                : `${mobileNav.currentIndex + 1}/${mobileNav.total}`,
        [mobileNav.currentIndex, mobileNav.total],
    );

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
    } = useInlineEdit({ resetKey: knowledgeId });

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
                <div className="flex min-w-0 flex-1 items-center gap-2">
                    <SidebarTrigger className="-ml-1 shrink-0" />
                    <Separator
                        className="mr-2 h-4 shrink-0"
                        orientation="vertical"
                    />
                    <Breadcrumb className="min-w-0 flex-1">
                        <BreadcrumbList className="min-w-0 flex-nowrap">
                            <BreadcrumbItem className="min-w-0 gap-2">
                                <LibraryBig className="size-4 shrink-0" />
                                {isEditingTitle && canShowActions ? (
                                    <InlineEditInput
                                        busy={isRenaming}
                                        className="w-64 min-w-0 max-w-full flex-1"
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
                                                className="min-w-0 cursor-text select-none truncate"
                                                onDoubleClick={handleRenameStart}
                                            >
                                                {knowledgeName ?? 'Knowledge'}
                                            </BreadcrumbPage>
                                        </TooltipTrigger>
                                        <TooltipContent>Double-click to rename</TooltipContent>
                                    </Tooltip>
                                ) : (
                                    <BreadcrumbPage className="min-w-0 truncate">
                                        {isNew ? 'New knowledge' : (knowledgeName ?? 'Knowledge')}
                                    </BreadcrumbPage>
                                )}
                            </BreadcrumbItem>
                        </BreadcrumbList>
                    </Breadcrumb>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                    {canShowActions && !isMobile && (
                        <DetailNavigationToolbar<Knowledge>
                            {...knowledgeToolbarProps}
                            renderItem={(item, isCurrent) => (
                                <>
                                    <Badge
                                        className="shrink-0 text-[10px] whitespace-nowrap"
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
                                {isMobile && mobileNav.total > 0 && (
                                    <>
                                        <DropdownMenuItem
                                            className="cursor-default hover:bg-transparent focus:bg-transparent"
                                            onSelect={(event) => event.preventDefault()}
                                        >
                                            <LibraryBig className="size-4" />
                                            Knowledges
                                            <div className="-my-1.5 -mr-2 ml-auto flex items-center">
                                                <Button
                                                    aria-label="Previous"
                                                    className="size-7 rounded-r-none border-r-0 p-0"
                                                    disabled={!mobileNav.prevId}
                                                    onClick={() => mobileNavGoTo(mobileNav.prevId)}
                                                    size="icon"
                                                    variant="outline"
                                                >
                                                    <ChevronLeft />
                                                </Button>
                                                <Button
                                                    aria-label="Open knowledges list"
                                                    className="h-7 min-w-12 rounded-none border-x px-2 font-mono text-xs tabular-nums"
                                                    disabled={mobileNav.currentIndex === -1}
                                                    onClick={() => setIsMobileNavSheetOpen(true)}
                                                    variant="outline"
                                                >
                                                    {mobilePositionLabel}
                                                </Button>
                                                <Button
                                                    aria-label="Next"
                                                    className="size-7 rounded-l-none border-l-0 p-0"
                                                    disabled={!mobileNav.nextId}
                                                    onClick={() => mobileNavGoTo(mobileNav.nextId)}
                                                    size="icon"
                                                    variant="outline"
                                                >
                                                    <ChevronRight />
                                                </Button>
                                            </div>
                                        </DropdownMenuItem>
                                        <DropdownMenuSeparator />
                                    </>
                                )}
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
            {isMobile && canShowActions && (
                <DetailNavigationSheet<Knowledge>
                    currentId={knowledgeToolbarProps.currentId}
                    currentIndex={mobileNav.currentIndex}
                    getId={knowledgeToolbarProps.getId}
                    getLabel={knowledgeToolbarProps.getLabel}
                    items={mobileNav.filteredItems}
                    onItemSelect={mobileNavSelectItem}
                    onOpenChange={setIsMobileNavSheetOpen}
                    open={isMobileNavSheetOpen}
                    renderItem={(item, isCurrent) => (
                        <>
                            <Badge
                                className="shrink-0 text-[10px] whitespace-nowrap"
                                variant="outline"
                            >
                                {item.docType}
                            </Badge>
                            <span className={isCurrent ? 'truncate font-medium' : 'truncate'}>{item.question}</span>
                        </>
                    )}
                    sheetIcon={<LibraryBig className="size-4" />}
                    sheetTitle="Knowledges"
                    total={mobileNav.total}
                />
            )}
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
