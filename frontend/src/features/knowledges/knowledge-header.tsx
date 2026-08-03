import type { ReactNode } from 'react';

import { Ellipsis, HatGlasses, LibraryBig, Pencil, Save, Trash } from 'lucide-react';
import { useCallback, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';
import { toast } from 'sonner';

import type { KnowledgeDocumentFragmentFragment } from '@/graphql/types';

import { AppHeader, AppHeaderAction, AppHeaderActions, AppHeaderContent } from '@/components/layouts/app/app-header';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import {
    DetailNavigationButtons,
    DetailNavigationSheet,
    DetailNavigationToolbar,
} from '@/components/shared/detail-navigation';
import { InlineEditInput, useInlineEdit } from '@/components/shared/inline-edit';
import { type EditorViewMode, EditorViewModeToggle } from '@/components/shared/markdown-editor';
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
import { Spinner } from '@/components/ui/spinner';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { routes } from '@/lib/routes';
import { cn } from '@/lib/utils';
import { type Knowledge, useKnowledges } from '@/providers/knowledges-provider';

import { useKnowledgeDetailNavigation } from './use-knowledge-detail-navigation';

interface KnowledgeHeaderProps {
    // Anonymize action — visible only to users with the `anonymize.call` privilege.
    canAnonymize?: boolean;
    isAnonymizeDisabled?: boolean;
    isAnonymizing?: boolean;
    isNew: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
    /**
     * Optional hook called right before the header navigates away after a
     * successful delete. The form mounts this header inside an unsaved-changes
     * guard, so it passes `skipNextBlock` here to suppress the "Save before
     * leaving?" dialog — there is nothing to save once the document is gone.
     */
    onAnonymize?: () => void;
    onBeforeNavigateAway?: () => void;
    onModeChange?: (mode: EditorViewMode) => void;
    saveButton?: ReactNode;
    viewMode?: EditorViewMode;
}

const renderKnowledgeItem = (item: Knowledge, isCurrent: boolean): ReactNode => (
    <>
        <Badge
            className="shrink-0 text-[10px] whitespace-nowrap"
            variant="outline"
        >
            {item.docType}
        </Badge>
        <span className={cn('min-w-0 flex-1 truncate', isCurrent && 'font-medium')}>{item.question}</span>
    </>
);

export function KnowledgeHeader({
    canAnonymize = false,
    isAnonymizeDisabled = false,
    isAnonymizing = false,
    isNew,
    knowledge,
    onAnonymize,
    onBeforeNavigateAway,
    onModeChange,
    saveButton,
    viewMode = 'rich',
}: KnowledgeHeaderProps) {
    const navigate = useNavigate();
    const { knowledgeId } = useParams();
    const { isMobile } = useBreakpoint();
    const { deleteKnowledge, renameKnowledge } = useKnowledges();
    const [isRenaming, setIsRenaming] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);

    const knowledgeNav = useKnowledgeDetailNavigation(isNew ? null : (knowledgeId ?? null));

    // Title source-of-truth is the server-side `question`. We intentionally do
    // not read it from the form draft below — the inline rename flow in this
    // header writes through `renameKnowledge`, which refreshes `knowledge` via
    // the cache, and the form picks up the new value separately.
    const knowledgeName = knowledge?.question ?? null;
    const hasKnowledge = !!knowledge;
    const isEntityPending = !hasKnowledge;
    const hasAnonymizeRow = isMobile && canAnonymize;
    const hasViewRow = !!onModeChange;
    const hasEntityRows = !isNew;
    const hasNavRow = isMobile && !isNew;

    const {
        handleDropdownCloseAutoFocus,
        inputRef: editingInputRef,
        isEditing: isEditingTitle,
        startEdit: handleRenameStart,
        stopEdit: handleRenameCancel,
    } = useInlineEdit({ resetKey: knowledge?.id ?? null });

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
            // The sibling edit form picks up the new `question` via
            // `useForm({ values })` once the cache updates — no manual sync here.
            await renameKnowledge(knowledge.id, newQuestion);
            toast.success('Knowledge renamed successfully');
            handleRenameCancel();
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsRenaming(false);
        }
    }, [editingInputRef, handleRenameCancel, knowledge, renameKnowledge]);

    const handleDelete = useCallback(async () => {
        if (!knowledge) {
            return;
        }

        setIsDeleting(true);

        try {
            await deleteKnowledge(knowledge.id);
            onBeforeNavigateAway?.();
            navigate(routes.knowledges, { replace: true });
        } catch {
            // Error already handled in provider with toast
        } finally {
            setIsDeleting(false);
        }
    }, [knowledge, deleteKnowledge, navigate, onBeforeNavigateAway]);

    return (
        <>
            <AppHeader>
                <AppHeaderContent>
                    <Breadcrumb className="min-w-0 flex-1">
                        <BreadcrumbList className="min-w-0 flex-nowrap">
                            <BreadcrumbItem className="min-w-0 gap-2">
                                <LibraryBig className="size-4 shrink-0" />
                                {isEditingTitle && hasKnowledge ? (
                                    <InlineEditInput
                                        busy={isRenaming}
                                        className="w-64 max-w-full min-w-0 flex-1"
                                        defaultValue={knowledgeName ?? ''}
                                        inputRef={editingInputRef}
                                        onCancel={handleRenameCancel}
                                        onSave={handleRenameSave}
                                        placeholder="Knowledge question"
                                    />
                                ) : hasKnowledge ? (
                                    <Tooltip>
                                        <TooltipTrigger asChild>
                                            <BreadcrumbPage
                                                className="max-w-64 min-w-0 cursor-text truncate select-none"
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
                </AppHeaderContent>
                <AppHeaderActions>
                    {!isNew && !isMobile && (
                        <DetailNavigationToolbar<Knowledge>
                            controller={knowledgeNav}
                            renderItem={renderKnowledgeItem}
                            sheetIcon={<LibraryBig className="size-4" />}
                            sheetTitle="Knowledges"
                        />
                    )}
                    {canAnonymize && !isMobile && (
                        <AppHeaderAction
                            disabled={isAnonymizeDisabled}
                            icon={isAnonymizing ? <Spinner variant="circle" /> : <HatGlasses aria-hidden="true" />}
                            label="Anonymize"
                            onClick={onAnonymize}
                            type="button"
                            variant="outline"
                        />
                    )}
                    {saveButton ?? (
                        <AppHeaderAction
                            disabled
                            icon={<Save />}
                            label={isNew ? 'Create' : 'Save'}
                            type="button"
                        />
                    )}
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
                            {hasAnonymizeRow && (
                                <DropdownMenuItem
                                    disabled={isAnonymizeDisabled}
                                    onClick={onAnonymize}
                                >
                                    {isAnonymizing ? (
                                        <>
                                            <Spinner variant="circle" />
                                            Anonymizing...
                                        </>
                                    ) : (
                                        <>
                                            <HatGlasses />
                                            Anonymize
                                        </>
                                    )}
                                </DropdownMenuItem>
                            )}
                            {hasNavRow && (
                                <>
                                    {hasAnonymizeRow && <DropdownMenuSeparator />}
                                    <DropdownMenuItem
                                        className="cursor-default hover:bg-transparent focus:bg-transparent"
                                        onSelect={(event) => event.preventDefault()}
                                    >
                                        <LibraryBig />
                                        Knowledges
                                        <div className="-my-1.5 -mr-2 ml-auto flex items-center">
                                            <DetailNavigationButtons<Knowledge>
                                                controller={knowledgeNav}
                                                sheetTitle="Knowledges"
                                                size="sm"
                                            />
                                        </div>
                                    </DropdownMenuItem>
                                </>
                            )}
                            {hasEntityRows && (
                                <>
                                    {(hasAnonymizeRow || hasNavRow) && <DropdownMenuSeparator />}
                                    <DropdownMenuItem
                                        disabled={isEntityPending}
                                        onClick={handleRenameStart}
                                    >
                                        <Pencil className="size-3" />
                                        Rename
                                    </DropdownMenuItem>
                                    <DropdownMenuSeparator />
                                </>
                            )}
                            {onModeChange ? (
                                <>
                                    {!hasEntityRows && (hasAnonymizeRow || hasNavRow) && <DropdownMenuSeparator />}
                                    <DropdownMenuItem
                                        className="cursor-default gap-4 hover:bg-transparent focus:bg-transparent"
                                        onSelect={(event) => event.preventDefault()}
                                    >
                                        View
                                        <EditorViewModeToggle
                                            className="-my-1.5 -mr-2 ml-auto"
                                            mode={viewMode}
                                            onModeChange={onModeChange}
                                            rawTooltip="Edit the raw markdown"
                                        />
                                    </DropdownMenuItem>
                                </>
                            ) : null}
                            {hasEntityRows && (
                                <>
                                    {hasViewRow && <DropdownMenuSeparator />}
                                    <DropdownMenuItem
                                        disabled={isDeleting || isEntityPending}
                                        onClick={() => setIsDeleteDialogOpen(true)}
                                    >
                                        {isDeleting ? (
                                            <>
                                                <Spinner variant="circle" />
                                                Deleting...
                                            </>
                                        ) : (
                                            <>
                                                <Trash />
                                                Delete
                                            </>
                                        )}
                                    </DropdownMenuItem>
                                </>
                            )}
                        </DropdownMenuContent>
                    </DropdownMenu>
                </AppHeaderActions>
            </AppHeader>
            {isMobile && !isNew && (
                <DetailNavigationSheet<Knowledge>
                    controller={knowledgeNav}
                    renderItem={renderKnowledgeItem}
                    sheetIcon={<LibraryBig className="size-4" />}
                    sheetTitle="Knowledges"
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
}
