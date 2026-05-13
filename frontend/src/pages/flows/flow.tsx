import {
    Check,
    ChevronDown,
    Copy,
    Download,
    Ellipsis,
    ExternalLink,
    GripVertical,
    Loader2,
    NotepadText,
    Pause,
    PencilLine,
    Star,
    Trash,
    X,
} from 'lucide-react';
import { useCallback, useEffect, useRef, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { toast } from 'sonner';

import { FlowStatusIcon } from '@/components/icons/flow-status-icon';
import { ProviderIcon } from '@/components/icons/provider-icon';
import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { HeaderButton } from '@/components/shared/header-button';
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
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip';
import FlowCentralTabs from '@/features/flows/flow-central-tabs';
import FlowTabs from '@/features/flows/flow-tabs';
import { ResultType, StatusType, useRenameFlowMutation } from '@/graphql/types';
import { useBreakpoint } from '@/hooks/use-breakpoint';
import { useFlowTabDetection } from '@/hooks/use-flow-tab-detection';
import { Log } from '@/lib/log';
import { copyToClipboard, downloadTextFile, generateFileName, generateReport } from '@/lib/report';
import { formatName } from '@/lib/utils/format';
import { useFavorites } from '@/providers/favorites-provider';
import { useFlow } from '@/providers/flow-provider';
import { useFlows } from '@/providers/flows-provider';

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
    const { isDesktop } = useBreakpoint();
    const navigate = useNavigate();

    const { flowData, flowError, flowId, isLoading: isFlowLoading } = useFlow();
    const { deleteFlow, finishFlow } = useFlows();
    const { isFavoriteFlow, toggleFavoriteFlow } = useFavorites();

    const flow = flowData?.flow;
    const flowTitle = flow?.title ?? '';
    const isFlowRunning = flow ? ![StatusType.Failed, StatusType.Finished].includes(flow.status) : false;

    const [isEditingTitle, setIsEditingTitle] = useState(false);
    const [isFinishing, setIsFinishing] = useState(false);
    const [isDeleting, setIsDeleting] = useState(false);
    const [isDeleteDialogOpen, setIsDeleteDialogOpen] = useState(false);
    const editingInputRef = useRef<HTMLInputElement>(null);
    const [renameFlowMutation, { loading: isRenameLoading }] = useRenameFlowMutation();

    // Reset inline-edit state when navigating between flows so the input doesn't
    // carry over a stale draft from a previous flow.
    useEffect(() => {
        setIsEditingTitle(false);
    }, [flowId]);

    // Focus and select the rename input when the inline editor opens. We can't
    // rely on `autoFocus` here: the input mounts inside the same render cycle
    // that closes the Radix DropdownMenu, and the dropdown's own focus restore
    // (which it schedules via `requestAnimationFrame`) wins the race against
    // React's autoFocus effect. Defer our focus to the next frame so it lands
    // *after* Radix has finished its restore. Selecting the text lets the user
    // overwrite the title in one keystroke.
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

    // Redirect to flows list if there's an error loading flow data or flow not found
    useEffect(() => {
        if (flowError || (!isFlowLoading && !flowData?.flow)) {
            navigate('/flows', { replace: true });
        }
    }, [flowError, flowData, isFlowLoading, navigate]);

    const handleFlowRenameStart = useCallback(() => {
        setIsEditingTitle(true);
    }, []);

    const handleFlowRenameCancel = useCallback(() => {
        setIsEditingTitle(false);
    }, []);

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
                setIsEditingTitle(false);
            }
        } catch (error) {
            const errorMessage = error instanceof Error ? error.message : 'Failed to rename flow';
            toast.error(errorMessage);
        }
    }, [flowId, renameFlowMutation]);

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
                    <div className="flex items-center gap-2">
                        <SidebarTrigger className="-ml-1" />
                        <Separator
                            className="mr-2 h-4"
                            orientation="vertical"
                        />
                        <Breadcrumb>
                            <BreadcrumbList>
                                <BreadcrumbItem className="gap-2">
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
                                        <InputGroup className="h-8 w-64 max-w-full">
                                            <InputGroupInput
                                                className="text-foreground"
                                                defaultValue={flowTitle}
                                                onKeyDown={(event) => {
                                                    if (event.key === 'Enter') {
                                                        event.preventDefault();
                                                        handleFlowRenameSave();

                                                        return;
                                                    }

                                                    if (event.key === 'Escape') {
                                                        event.preventDefault();
                                                        handleFlowRenameCancel();
                                                    }
                                                }}
                                                placeholder="Flow title"
                                                ref={editingInputRef}
                                            />
                                            <InputGroupAddon
                                                align="inline-end"
                                                className="gap-0 pr-2"
                                            >
                                                <InputGroupButton
                                                    aria-label="Save"
                                                    disabled={isRenameLoading}
                                                    onClick={() => handleFlowRenameSave()}
                                                >
                                                    {isRenameLoading ? <Loader2 className="animate-spin" /> : <Check />}
                                                </InputGroupButton>
                                                <InputGroupButton
                                                    aria-label="Cancel"
                                                    onClick={() => handleFlowRenameCancel()}
                                                >
                                                    <X />
                                                </InputGroupButton>
                                            </InputGroupAddon>
                                        </InputGroup>
                                    ) : flow ? (
                                        <Tooltip>
                                            <TooltipTrigger asChild>
                                                <BreadcrumbPage
                                                    className="cursor-text select-none"
                                                    onDoubleClick={handleFlowRenameStart}
                                                >
                                                    {flowTitle || 'Select a flow'}
                                                </BreadcrumbPage>
                                            </TooltipTrigger>
                                            <TooltipContent>Double-click to rename</TooltipContent>
                                        </Tooltip>
                                    ) : (
                                        <BreadcrumbPage>{flowTitle || 'Select a flow'}</BreadcrumbPage>
                                    )}
                                </BreadcrumbItem>
                            </BreadcrumbList>
                        </Breadcrumb>
                    </div>
                    <div className="flex items-center gap-2">
                        {flowId && (
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
                                    onCloseAutoFocus={(event) => {
                                        // Radix returns focus to the trigger on close. When the
                                        // selected action mounts the rename input, prevent that
                                        // restore so the input's `autoFocus` actually wins.
                                        if (isEditingTitle) {
                                            event.preventDefault();
                                        }
                                    }}
                                >
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
