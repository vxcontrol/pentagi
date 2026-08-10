import type { ReactNode } from 'react';

import { useMutation } from '@apollo/client/react';
import {
    ChevronDown,
    Copy,
    Download,
    Ellipsis,
    ExternalLink,
    GitFork,
    GripVertical,
    NotepadText,
    Pause,
    PencilLine,
    Star,
    Trash,
} from 'lucide-react';
import { startTransition, useCallback, useEffect, useOptimistic, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
import { ProviderIcon } from '@/components/icons/provider-icon';
import { AppHeader, AppHeaderAction, AppHeaderActions, AppHeaderContent } from '@/components/layouts/app/app-header';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import {
    DetailNavigationButtons,
    DetailNavigationSheet,
    DetailNavigationToolbar,
} from '@/components/shared/detail-navigation';
import { ErrorState } from '@/components/shared/error-state';
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
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { Spinner } from '@/components/ui/spinner';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import FlowCentralTabs from '@/features/flows/flow-central-tabs';
import FlowTabs from '@/features/flows/flow-tabs';
import { useFlowDetailNavigation } from '@/features/flows/use-flow-detail-navigation';
import { RenameFlowDocument, ResultType, StatusType } from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { useFlowTabDetection } from '@/hooks/use-flow-tab-detection';
import { useI18n } from '@/hooks/use-i18n';
import { Log } from '@/lib/log';
import { copyToClipboard, downloadTextFile, generateFileName, generateReport } from '@/lib/report';
import { routes } from '@/lib/routes';
import { cn } from '@/lib/utils';
import { formatName } from '@/lib/utils/format';
import { useFavorites } from '@/providers/favorites-provider';
import { useFlow } from '@/providers/flow-provider';
import { type Flow as FlowItem, useFlows } from '@/providers/flows-provider';

const renderFlowItem = (item: FlowItem, isCurrent: boolean, t: (key: string, opts?: Record<string, unknown>) => string): ReactNode => (
    <>
        <FlowStatusIcon
            className="size-3 shrink-0"
            status={item.status}
        />
        <span className={cn('min-w-0 flex-1 truncate', isCurrent && 'font-medium')}>
            {item.title || t('flow.flowNumber', { id: item.id })}
        </span>
        <Badge
            className="ml-auto shrink-0 font-mono text-[10px]"
            variant="outline"
        >
            #{item.id}
        </Badge>
    </>
);

function Flow() {
    const { isDesktop, isMobile } = useBreakpoint();
    const navigate = useNavigate();
    const { t } = useI18n();

    const { flowData, flowId, flowLoadError, isFlowMissing, isLoading: isFlowLoading, refetchFlow } = useFlow();
    const { deleteFlow, finishFlow } = useFlows();
    const { isFavoriteFlow, toggleFavoriteFlow } = useFavorites();

    const flow = flowData?.flow;
    const actualFlowTitle = flow?.title ?? '';
    const [flowTitle, setOptimisticFlowTitle] = useOptimistic(actualFlowTitle, (_current, next: string) => next);
    const isFlowRunning = flow ? ![StatusType.Failed, StatusType.Finished].includes(flow.status) : false;

    const flowNav = useFlowDetailNavigation(flowId);

    const {
        handleDropdownCloseAutoFocus,
        inputRef: editingInputRef,
        isEditing: isEditingTitle,
        startEdit: handleFlowRenameStart,
        stopEdit: handleFlowRenameCancel,
    } = useInlineEdit({ resetKey: flowId });

    const [isFinishing, setIsFinishing] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const [renameFlowMutation, { loading: isRenameLoading }] = useMutation(RenameFlowDocument);

    useEffect(() => {
        if (isFlowMissing) {
            navigate(routes.flows, { replace: true });
        }
    }, [isFlowMissing, navigate]);

    const handleFlowRenameSave = useCallback(async () => {
        const newTitle = editingInputRef.current?.value.trim();

        if (!flowId || !newTitle) {
            return;
        }

        startTransition(async () => {
            setOptimisticFlowTitle(newTitle);

            try {
                const { data } = await renameFlowMutation({
                    variables: {
                        flowId,
                        title: newTitle,
                    },
                });

                if (data?.renameFlow === ResultType.Success) {
                    toast.success(t('flow.renameSuccess'));
                    handleFlowRenameCancel();
                }
            } catch (error) {
                const errorMessage = error instanceof Error ? error.message : t('flow.renameFailed');
                toast.error(errorMessage);
            }
        });
    }, [editingInputRef, flowId, handleFlowRenameCancel, renameFlowMutation, setOptimisticFlowTitle, t]);

    const handleFlowFinish = useCallback(async () => {
        if (!flow) {
            return;
        }

        setIsFinishing(true);

        try {
            await finishFlow(flow);
        } finally {
            setIsFinishing(false);
        }
    }, [flow, finishFlow]);

    const handleFlowDelete = useCallback(async () => {
        if (!flow) {
            return;
        }

        setIsDeleting(true);

        try {
            const success = await deleteFlow(flow);

            if (success) {
                navigate(routes.flows, { replace: true });
            }
        } finally {
            setIsDeleting(false);
        }
    }, [flow, deleteFlow, navigate]);

    const [desktopTabsTab, setDesktopTabsTab] = useState<string>('terminal');

    const { handleTabChange: handleMobileTabChange, resolvedTab: mobileAutoTab } = useFlowTabDetection();

    const activeTabsTab = isDesktop ? desktopTabsTab : mobileAutoTab;
    const handleTabsTabChange = isDesktop ? setDesktopTabsTab : handleMobileTabChange;

    if (flowLoadError) {
        return (
            <>
                <AppHeader>
                    <AppHeaderContent>
                        <Breadcrumb className="min-w-0 flex-1">
                            <BreadcrumbList className="min-w-0 flex-nowrap">
                                <BreadcrumbItem className="min-w-0">
                                    <BreadcrumbPage>{t('flow.flowBreadcrumb')}</BreadcrumbPage>
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </AppHeaderContent>
                </AppHeader>
                <div className="flex flex-1 flex-col gap-4 p-4">
                    <ErrorState
                        message={flowLoadError.message}
                        onRetry={refetchFlow}
                        title={t('flow.errorLoadingFlow')}
                    />
                </div>
            </>
        );
    }

    const tabsCard = (
        <div className="flex h-[calc(100dvh-3rem)] max-w-full flex-col rounded-none border-0">
            <div className="flex-1 overflow-auto py-4 pr-0 pl-4">
                <FlowTabs
                    activeTab={activeTabsTab}
                    onTabChange={handleTabsTabChange}
                />
            </div>
        </div>
    );

    return (
        <>
            <AppHeader>
                <AppHeaderContent>
                    <Breadcrumb className="min-w-0 flex-1">
                        <BreadcrumbList className="min-w-0 flex-nowrap">
                            <BreadcrumbItem className="min-w-0 gap-2">
                                {flow && (
                                    <>
                                        <FlowStatusIcon
                                            status={flow.status}
                                            tooltip={formatName(flow.status)}
                                        />

                                        <ProviderIcon
                                            provider={flow.provider}
                                            tooltip={formatName(flow.provider.name)}
                                        />
                                    </>
                                )}
                                {isEditingTitle && flow ? (
                                    <InlineEditInput
                                        busy={isRenameLoading}
                                        className="w-64 max-w-full min-w-0 flex-1"
                                        defaultValue={flowTitle}
                                        inputRef={editingInputRef}
                                        onCancel={handleFlowRenameCancel}
                                        onSave={handleFlowRenameSave}
                                        placeholder={t('flow.flowTitle')}
                                    />
                                ) : flow ? (
                                    <Tooltip>
                                        <TooltipTrigger asChild>
                                            <BreadcrumbPage
                                                className="max-w-64 min-w-0 cursor-text truncate select-none"
                                                onDoubleClick={handleFlowRenameStart}
                                            >
                                                {flowTitle || t('flow.selectFlow')}
                                            </BreadcrumbPage>
                                        </TooltipTrigger>
                                        <TooltipContent>{t('flow.doubleClickRename')}</TooltipContent>
                                    </Tooltip>
                                ) : (
                                    <BreadcrumbPage className="min-w-0 truncate">
                                        {flowTitle || t('flow.selectFlow')}
                                    </BreadcrumbPage>
                                )}
                            </BreadcrumbItem>
                        </BreadcrumbList>
                    </Breadcrumb>
                </AppHeaderContent>
                <AppHeaderActions>
                    {!!(flowData?.tasks ?? [])?.length && <FlowReportDropdown />}
                    {!isMobile && (
                        <DetailNavigationToolbar<FlowItem>
                            controller={flowNav}
                            renderItem={(item, isCurrent) => renderFlowItem(item, isCurrent, t)}
                            sheetIcon={<GitFork className="size-4" />}
                            sheetTitle={t('common.flows')}
                        />
                    )}
                    {flowId && !isMobile && (
                        <Button
                            aria-label={t('flow.toggleFavorite')}
                            aria-pressed={isFavoriteFlow(flowId)}
                            className="shrink-0"
                            disabled={isFlowLoading}
                            onClick={() => toggleFavoriteFlow(flowId)}
                            size="icon"
                            variant="ghost"
                        >
                            <Star className={isFavoriteFlow(flowId) ? 'fill-yellow-500 stroke-yellow-500' : ''} />
                        </Button>
                    )}
                    <DropdownMenu>
                        <DropdownMenuTrigger asChild>
                            <Button
                                aria-label={t('flow.flowActions')}
                                className="size-8 p-0"
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
                            {isMobile && (
                                <>
                                    {/* onSelect={preventDefault} stops the Radix menu from closing on label
                                                clicks; DetailNavigationButtons owns its own click handlers. */}
                                    <DropdownMenuItem
                                        className="cursor-default hover:bg-transparent focus:bg-transparent"
                                        onSelect={(event) => event.preventDefault()}
                                    >
                                        <GitFork />
                                        {t('common.flows')}
                                        <div className="-my-1.5 -mr-2 ml-auto flex items-center">
                                            <DetailNavigationButtons<FlowItem>
                                                controller={flowNav}
                                                sheetTitle={t('common.flows')}
                                                size="sm"
                                            />
                                        </div>
                                    </DropdownMenuItem>
                                    {flowId && (
                                        <DropdownMenuItem
                                            disabled={isFlowLoading}
                                            onClick={() => toggleFavoriteFlow(flowId)}
                                        >
                                            <Star
                                                className={
                                                    isFavoriteFlow(flowId)
                                                        ? 'size-4 fill-yellow-500 stroke-yellow-500'
                                                        : 'size-4'
                                                }
                                            />
                                            {isFavoriteFlow(flowId)
                                                ? t('flow.removeFromFavorites')
                                                : t('flow.addToFavorites')}
                                        </DropdownMenuItem>
                                    )}
                                    <DropdownMenuSeparator />
                                </>
                            )}
                            <DropdownMenuItem
                                disabled={isFlowLoading}
                                onClick={handleFlowRenameStart}
                            >
                                <PencilLine className="size-3" />
                                {t('common.rename')}
                            </DropdownMenuItem>
                            {isFlowRunning && (
                                <DropdownMenuItem
                                    disabled={isFinishing}
                                    onClick={() => handleFlowFinish()}
                                >
                                    {isFinishing ? (
                                        <>
                                            <Spinner variant="circle" />
                                            {t('flow.finishing')}
                                        </>
                                    ) : (
                                        <>
                                            <Pause />
                                            {t('common.finish')}
                                        </>
                                    )}
                                </DropdownMenuItem>
                            )}
                            <DropdownMenuSeparator />
                            <DropdownMenuItem
                                disabled={isDeleting || isFlowLoading}
                                onClick={() => setIsDeleteDialogOpen(true)}
                            >
                                {isDeleting ? (
                                    <>
                                        <Spinner variant="circle" />
                                        {t('flow.deleting')}
                                    </>
                                ) : (
                                    <>
                                        <Trash />
                                        {t('common.delete')}
                                    </>
                                )}
                            </DropdownMenuItem>
                        </DropdownMenuContent>
                    </DropdownMenu>
                </AppHeaderActions>
            </AppHeader>
            {isMobile && (
                <DetailNavigationSheet<FlowItem>
                    controller={flowNav}
                    renderItem={(item, isCurrent) => renderFlowItem(item, isCurrent, t)}
                    sheetIcon={<GitFork className="size-4" />}
                    sheetTitle={t('common.flows')}
                />
            )}
            <div className="relative flex h-[calc(100dvh-3rem)] w-full max-w-full flex-1">
                {isFlowLoading && (
                    <div className="bg-background/50 absolute inset-0 z-50 flex items-center justify-center">
                        <Spinner
                            className="size-16"
                            variant="circle"
                        />
                    </div>
                )}
                {isDesktop ? (
                    <ResizablePanelGroup
                        className="w-full"
                        orientation="horizontal"
                    >
                        <ResizablePanel
                            defaultSize="50%"
                            minSize={390}
                        >
                            <div className="flex h-[calc(100dvh-3rem)] max-w-full flex-col rounded-none border-0">
                                <div className="flex-1 overflow-auto py-4 pr-0 pl-4">
                                    <FlowCentralTabs />
                                </div>
                            </div>
                        </ResizablePanel>
                        <ResizableHandle withHandle>
                            <GripVertical className="size-4" />
                        </ResizableHandle>
                        <ResizablePanel
                            defaultSize="50%"
                            minSize={390}
                        >
                            {tabsCard}
                        </ResizablePanel>
                    </ResizablePanelGroup>
                ) : (
                    tabsCard
                )}
            </div>
            <ConfirmationDialog
                cancelText={t('common.cancel')}
                confirmText={t('common.delete')}
                handleConfirm={handleFlowDelete}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={flow?.title}
                itemType={t('flow.itemType')}
            />
        </>
    );
}

function FlowReportDropdown() {
    const { t } = useI18n();
    const { flowData, flowId } = useFlow();
    const flow = flowData?.flow;
    const tasks = flowData?.tasks ?? [];

    const isReportDisabled = !flow || !flowId;

    const handleCopyToClipboard = async () => {
        if (isReportDisabled) {
            return;
        }

        const reportContent = generateReport(tasks, flow);
        const success = await copyToClipboard(reportContent);

        if (success) {
            toast.success(t('flow.reportCopied'));
        } else {
            Log.error('Failed to copy report to clipboard');
            toast.error(t('flow.reportCopyFailed'));
        }
    };

    const handleDownloadMD = () => {
        if (isReportDisabled || !flow) {
            return;
        }

        try {
            const reportContent = generateReport(tasks, flow);

            const baseFileName = generateFileName(flow);
            const fileName = `${baseFileName}.md`;

            downloadTextFile(reportContent, fileName, 'text/markdown; charset=UTF-8');
        } catch (error) {
            Log.error('Failed to download markdown report:', error);
        }
    };

    const handleDownloadPDF = () => {
        if (isReportDisabled || !flow || !flowId) {
            return;
        }

        const url = `${routes.flowReport(flowId)}?download=true&silent=true`;
        window.open(url, '_blank');
    };

    const handleOpenWebView = () => {
        if (isReportDisabled || !flowId) {
            return;
        }

        const url = routes.flowReport(flowId);
        window.open(url, '_blank');
    };

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <AppHeaderAction
                    className="shrink-0"
                    disabled={isReportDisabled}
                    endIcon={<ChevronDown className="opacity-50" />}
                    icon={<NotepadText />}
                    label={t('flow.report')}
                    variant="ghost"
                />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <DropdownMenuItem
                    className="flex items-center gap-2"
                    disabled={isReportDisabled}
                    onClick={handleOpenWebView}
                >
                    <ExternalLink />
                    {t('flow.openWebView')}
                </DropdownMenuItem>
                <DropdownMenuItem
                    className="flex items-center gap-2"
                    disabled={isReportDisabled}
                    onClick={handleCopyToClipboard}
                >
                    <Copy />
                    {t('flow.copyToClipboard')}
                </DropdownMenuItem>
                <DropdownMenuItem
                    className="flex items-center gap-2"
                    disabled={isReportDisabled}
                    onClick={handleDownloadMD}
                >
                    <Download />
                    {t('flow.downloadMD')}
                </DropdownMenuItem>
                <DropdownMenuItem
                    className="flex items-center gap-2"
                    disabled={isReportDisabled}
                    onClick={handleDownloadPDF}
                >
                    <Download />
                    {t('flow.downloadPDF')}
                </DropdownMenuItem>
            </DropdownMenuContent>
        </DropdownMenu>
    );
}

export default Flow;
