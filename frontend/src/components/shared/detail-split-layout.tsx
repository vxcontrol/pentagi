import type { ReactNode } from 'react';

import { Card, CardContent } from '@/components/ui/card';
import { ResizableHandle, ResizablePanel, ResizablePanelGroup } from '@/components/ui/resizable';

interface DetailSplitLayoutProps {
    content: ReactNode;
    contentClassName?: string;
    panel: ReactNode;
}

export function DetailSplitLayout({
    content,
    contentClassName = 'flex h-full min-h-0 flex-col overflow-y-auto p-4',
    panel,
}: DetailSplitLayoutProps) {
    return (
        <div className="flex min-h-0 w-full max-w-full flex-1 overflow-hidden">
            <ResizablePanelGroup
                className="w-full"
                orientation="horizontal"
            >
                <ResizablePanel
                    defaultSize="45%"
                    minSize={390}
                >
                    <div className="bg-card h-full min-h-0 overflow-y-auto">
                        <Card className="mx-auto min-h-full w-full max-w-2xl rounded-none border-0">
                            <CardContent className="flex flex-col gap-6 py-6">{panel}</CardContent>
                        </Card>
                    </div>
                </ResizablePanel>
                <ResizableHandle withHandle />
                <ResizablePanel
                    defaultSize="55%"
                    minSize={390}
                >
                    <div className={contentClassName}>{content}</div>
                </ResizablePanel>
            </ResizablePanelGroup>
        </div>
    );
}
