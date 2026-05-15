import {
    ChevronDown,
    ChevronLeft,
    ChevronRight,
    Copy,
    Download,
    Ellipsis,
    ExternalLink,
    GitFork,
    GripVertical,
    Loader2,
    NotepadText,
    Pause,
    PencilLine,
    Star,
    Trash,
} from 'lucide-react';
import { useCallback, useEffect, useMemo, useState } from 'react';
import { useNavigate, useSearchParams } from 'react-router-dom';
import { toast } from 'sonner';

import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
import { ProviderIcon } from '@/components/icons/provider-icon';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { DetailNavigationSheet, DetailNavigationToolbar, useNavigation } from '@/components/shared/detail-navigation';
import { HeaderButton } from '@/components/shared/header-button';
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
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import FlowCentralTabs from '@/features/flows/flow-central-tabs';
import FlowTabs from '@/features/flows/flow-tabs';
import { useFlowDetailNavigation } from '@/features/flows/use-flow-detail-navigation';
import { ResultType, StatusType, useRenameFlowMutation } from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { useFlowTabDetection } from '@/hooks/use-flow-tab-detection';
import { Log } from '@/lib/log';
import { copyToClipboard, downloadTextFile, generateFileName, generateReport } from '@/lib/report';
import { mergeHrefWithSearchParams } from '@/lib/url-params';
import { formatName } from '@/lib/utils/format';
import { useFavorites } from '@/providers/favorites-provider';
import { useFlow } from '@/providers/flow-provider';
import { type Flow as FlowItem, useFlows } from '@/providers/flows-provider';

const FlowReportDropdown = () => {
    const { flowData, flowId } = useFlow();
    const flow = flowData?.flow;
    const tasks = flowData?.tasks ?? [];

    // Check if flow is available for report generation
    const isReportDisabled = !flow || !flowId;

    // Report export handlers
    const handleCopyToClipboard = async () => {
        if (isReportDisabled) {
            return;
        }

        const reportContent = generateReport(tasks, flow);
        const success = await copyToClipboard(reportContent);

        if (success) {
            toast.success('Report copied to clipboard');
        } else {
            Log.error('Failed to copy report to clipboard');
            toast.error('Failed to copy report to clipboard');
        }
    };

    const handleDownloadMD = () => {
        if (isReportDisabled || !flow) {
            return;
        }

        try {
            // Generate report content
            const reportContent = generateReport(tasks, flow);

            // Generate file name
            const baseFileName = generateFileName(flow);
            const fileName = `${baseFileName}.md`;

            // Download file
            downloadTextFile(reportContent, fileName, 'text/markdown; charset=UTF-8');
        } catch (error) {
            Log.error('Failed to download markdown report:', error);
        }
    };

    const handleDownloadPDF = () => {
        if (isReportDisabled || !flow || !flowId) {
            return;
        }

        // Open new tab (not popup) with report page and download flag
        const url = `/flows/${flowId}/report?download=true&silent=true`;
        window.open(url, '_blank');
    };

    const handleOpenWebView = () => {
        if (isReportDisabled || !flowId) {
            return;
        }

        // Open new tab with report page for web viewing
        const url = `/flows/${flowId}/report`;
        window.open(url, '_blank');
    };

    return (
        <DropdownMenu>
            <DropdownMenuTrigger asChild>
                <HeaderButton
                    className="shrink-0"
                    disabled={isReportDisabled}
                    endIcon={<ChevronDown className="opacity-50" />}
                    icon={<NotepadText />}
                    label="Report"
                    variant="ghost"
                />
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
                <DropdownMenuItem
                    className="flex items-center gap-2"
                    disabled={isReportDisabled}
                    onClick={handleOpenWebView}
                >
                    <ExternalLink className="size-4" />
                    Open web view
                </DropdownMenuItem>
                <DropdownMenuItem
                    className="flex items-center gap-2"
                    disabled={isReportDisabled}
                    onClick={handleCopyToClipboard}
                >
                    <Copy className="size-4" />
                    Copy to clipboard
                </DropdownMenuItem>
                <DropdownMenuItem
                    className="flex items-center gap-2"
                    disabled={isReportDisabled}
                    onClick={handleDownloadMD}
                >
                    <Download className="size-4" />
                    Download MD
                </DropdownMenuItem>
                <DropdownMenuItem
                    className="flex items-center gap-2"
                    disabled={isReportDisabled}
                    onClick={handleDownloadPDF}
                >
                    <Download className="size-4" />
                    Download PDF
                </DropdownMenuItem>
            </DropdownMenuContent>
        </DropdownMenu>
    );
};

const Flow = () => {
    const { isDesktop, isMobile } = useBreakpoint();
    const navigate = useNavigate();
    const [searchParams] = useSearchParams();

    const { flowData, flowError, flowId, isLoading: isFlowLoading } = useFlow();
    const { deleteFlow, finishFlow } = useFlows();
    const { isFavoriteFlow, toggleFavoriteFlow } = useFavorites();

    const flow = flowData?.flow;
    const flowTitle = flow?.title ?? '';
    const isFlowRunning = flow ? ![StatusType.Failed, StatusType.Finished].includes(flow.status) : false;

    // Walk the same `?q=` filtered subset the list page renders. The hook
    // also restores the filter from `localStorage` when the page is opened
    // via a bookmark, so the toolbar stays in lockstep with the user's last
    // persisted intent.
    const { toolbarProps: flowToolbarProps } = useFlowDetailNavigation(flowId);

    // Mirror what `<DetailNavigationToolbar>` computes internally so the
    // mobile menu items (Previous / Open list / Next) and the sheet trigger
    // share the same filtered subset and ordering as the desktop toolbar.
    const mobileNav = useNavigation<FlowItem>({
        currentId: flowToolbarProps.currentId,
        getId: flowToolbarProps.getId,
        getSearchableText: flowToolbarProps.getSearchableText ?? flowToolbarProps.getLabel,
        items: flowToolbarProps.items,
        query: flowToolbarProps.filter,
    });
    const [isMobileNavSheetOpen, setIsMobileNavSheetOpen] = useState(false);

    const mobileNavGoTo = useCallback(
        (id: null | string) => {
            if (!id) {
                return;
            }

            const target = mobileNav.filteredItems.find((item) => String(flowToolbarProps.getId(item)) === id);

            if (!target) {
                return;
            }

            navigate(mergeHrefWithSearchParams(flowToolbarProps.getHref(target), searchParams), { replace: true });
        },
        [flowToolbarProps, mobileNav.filteredItems, navigate, searchParams],
    );

    const mobileNavSelectItem = useCallback(
        (item: FlowItem) => {
            setIsMobileNavSheetOpen(false);
            navigate(mergeHrefWithSearchParams(flowToolbarProps.getHref(item), searchParams), { replace: true });
        },
        [flowToolbarProps, navigate, searchParams],
    );

    const mobilePositionLabel = useMemo(
        () =>
            mobileNav.total === 0 || mobileNav.currentIndex === -1
                ? `–/${mobileNav.total}`
                : `${mobileNav.currentIndex + 1}/${mobileNav.total}`,
        [mobileNav.currentIndex, mobileNav.total],
    );

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
    const [renameFlowMutation, { loading: isRenameLoading }] = useRenameFlowMutation();

    // Redirect to flows list if there's an error loading flow data or flow not found
    useEffect(() => {
        if (flowError || (!isFlowLoading && !flowData?.flow)) {
            navigate('/flows', { replace: true });
        }
    }, [flowError, flowData, isFlowLoading, navigate]);

    const handleFlowRenameSave = useCallback(async () => {
        const newTitle = editingInputRef.current?.value.trim();

        if (!flowId || !newTitle) {
            return;
        }

        try {
            const { data } = await renameFlowMutation({
                variables: {
                    flowId,
                    title: newTitle,
                },
            });

            if (data?.renameFlow === ResultType.Success) {
                toast.success('Flow renamed successfully');
                handleFlowRenameCancel();
            }
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to rename flow';
            toast.error(errorMessage);
        }
    }, [editingInputRef, flowId, handleFlowRenameCancel, renameFlowMutation]);

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
                navigate('/flows', { replace: true });
            }
        } finally {
            setIsDeleting(false);
        }
    }, [flow, deleteFlow, navigate]);

    // Desktop: side panel defaults to 'terminal'
    const [desktopTabsTab, setDesktopTabsTab] = useState<string>('terminal');

    // Mobile: use the same auto-detection logic as FlowCentralTabs
    const { handleTabChange: handleMobileTabChange, resolvedTab: mobileAutoTab } = useFlowTabDetection();

    const activeTabsTab = isDesktop ? desktopTabsTab : mobileAutoTab;
    const handleTabsTabChange = isDesktop ? setDesktopTabsTab : handleMobileTabChange;

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
            <header className="bg-background sticky top-0 z-10 flex h-12 w-full shrink-0 items-center gap-2 border-b transition-[width,height] ease-linear group-has-data-[collapsible=icon]/sidebar-wrapper:h-12">
                <div className="flex w-full items-center justify-between gap-2 px-4">
                    <div className="flex min-w-0 flex-1 items-center gap-2">
                        <SidebarTrigger className="-ml-1 shrink-0" />
                        <Separator
                            className="mr-2 h-4 shrink-0"
                            orientation="vertical"
                        />
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
                                            className="w-64 min-w-0 max-w-full flex-1"
                                            defaultValue={flowTitle}
                                            inputRef={editingInputRef}
                                            onCancel={handleFlowRenameCancel}
                                            onSave={handleFlowRenameSave}
                                            placeholder="Flow title"
                                        />
                                    ) : flow ? (
                                        <Tooltip>
                                            <TooltipTrigger asChild>
                                                <BreadcrumbPage
                                                    className="min-w-0 cursor-text select-none truncate"
                                                    onDoubleClick={handleFlowRenameStart}
                                                >
                                                    {flowTitle || 'Select a flow'}
                                                </BreadcrumbPage>
                                            </TooltipTrigger>
                                            <TooltipContent>Double-click to rename</TooltipContent>
                                        </Tooltip>
                                    ) : (
                                        <BreadcrumbPage className="min-w-0 truncate">
                                            {flowTitle || 'Select a flow'}
                                        </BreadcrumbPage>
                                    )}
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
                    <div className="flex shrink-0 items-center gap-2">
                        {flow && !isMobile && (
                            <DetailNavigationToolbar<FlowItem>
                                {...flowToolbarProps}
                                renderItem={(item, isCurrent) => (
                                    <>
                                        <FlowStatusIcon
                                            className="size-3 shrink-0"
                                            status={item.status}
                                        />
                                        <span className={isCurrent ? 'truncate font-medium' : 'truncate'}>
                                            {item.title || `Flow #${item.id}`}
                                        </span>
                                        <Badge
                                            className="ml-auto shrink-0 font-mono text-[10px]"
                                            variant="outline"
                                        >
                                            #{item.id}
                                        </Badge>
                                    </>
                                )}
                                sheetIcon={<GitFork className="size-4" />}
                                sheetTitle="Flows"
                            />
                        )}
                        {flowId && !isMobile && (
                            <Button
                                className="shrink-0"
                                onClick={() => toggleFavoriteFlow(flowId)}
                                size="icon"
                                variant="ghost"
                            >
                                <Star className={isFavoriteFlow(flowId) ? 'fill-yellow-500 stroke-yellow-500' : ''} />
                            </Button>
                        )}
                        {!!(flowData?.tasks ?? [])?.length && <FlowReportDropdown />}
                        {flow && (
                            <DropdownMenu>
                                <DropdownMenuTrigger asChild>
                                    <Button
                                        aria-label="Flow actions"
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
                                    {isMobile && mobileNav.total > 0 && (
                                        <>
                                            {/* Single row that mirrors the desktop toolbar: label on
                                                the left, prev / position / next button group on the
                                                right. `onSelect={preventDefault}` stops the menu from
                                                closing on label clicks; the inner buttons own their
                                                own click handlers and don't bubble into menu-item
                                                selection. Position button doubles as the sheet
                                                trigger, matching the toolbar's middle-button role. */}
                                            <DropdownMenuItem
                                                className="cursor-default hover:bg-transparent focus:bg-transparent"
                                                onSelect={(event) => event.preventDefault()}
                                            >
                                                <GitFork className="size-4" />
                                                Flows
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
                                                        aria-label="Open flows list"
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
                                            {flowId && (
                                                <DropdownMenuItem onClick={() => toggleFavoriteFlow(flowId)}>
                                                    <Star
                                                        className={
                                                            isFavoriteFlow(flowId)
                                                                ? 'size-4 fill-yellow-500 stroke-yellow-500'
                                                                : 'size-4'
                                                        }
                                                    />
                                                    {isFavoriteFlow(flowId)
                                                        ? 'Remove from favorites'
                                                        : 'Add to favorites'}
                                                </DropdownMenuItem>
                                            )}
                                            <DropdownMenuSeparator />
                                        </>
                                    )}
                                    <DropdownMenuItem onClick={handleFlowRenameStart}>
                                        <PencilLine className="size-3" />
                                        Rename
                                    </DropdownMenuItem>
                                    {isFlowRunning && (
                                        <DropdownMenuItem
                                            disabled={isFinishing}
                                            onClick={() => handleFlowFinish()}
                                        >
                                            {isFinishing ? (
                                                <>
                                                    <Loader2 className="animate-spin" />
                                                    Finishing...
                                                </>
                                            ) : (
                                                <>
                                                    <Pause />
                                                    Finish
                                                </>
                                            )}
                                        </DropdownMenuItem>
                                    )}
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
                </div>
            </header>
            {isMobile && flow && (
                <DetailNavigationSheet<FlowItem>
                    currentId={flowToolbarProps.currentId}
                    currentIndex={mobileNav.currentIndex}
                    getId={flowToolbarProps.getId}
                    getLabel={flowToolbarProps.getLabel}
                    items={mobileNav.filteredItems}
                    onItemSelect={mobileNavSelectItem}
                    onOpenChange={setIsMobileNavSheetOpen}
                    open={isMobileNavSheetOpen}
                    renderItem={(item, isCurrent) => (
                        <>
                            <FlowStatusIcon
                                className="size-3 shrink-0"
                                status={item.status}
                            />
                            <span className={isCurrent ? 'truncate font-medium' : 'truncate'}>
                                {item.title || `Flow #${item.id}`}
                            </span>
                            <Badge
                                className="ml-auto shrink-0 font-mono text-[10px]"
                                variant="outline"
                            >
                                #{item.id}
                            </Badge>
                        </>
                    )}
                    sheetIcon={<GitFork className="size-4" />}
                    sheetTitle="Flows"
                    total={mobileNav.total}
                />
            )}
            <div className="relative flex h-[calc(100dvh-3rem)] w-full max-w-full flex-1">
                {isFlowLoading && (
                    <div className="bg-background/50 absolute inset-0 z-50 flex items-center justify-center">
                        <Loader2 className="size-16 animate-spin" />
                    </div>
                )}
                {isDesktop ? (
                    <ResizablePanelGroup
                        className="w-full"
                        direction="horizontal"
                    >
                        <ResizablePanel
                            defaultSize={50}
                            minSize={30}
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
                            defaultSize={50}
                            minSize={30}
                        >
                            {tabsCard}
                        </ResizablePanel>
                    </ResizablePanelGroup>
                ) : (
                    tabsCard
                )}
            </div>
            <ConfirmationDialog
                cancelText="Cancel"
                confirmText="Delete"
                handleConfirm={handleFlowDelete}
                handleOpenChange={setIsDeleteDialogOpen}
                isOpen={isDeleteDialogOpen}
                itemName={flow?.title}
                itemType="flow"
            />
        </>
    );
};

export default Flow;
