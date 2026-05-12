import type { ReactNode } from 'react';

import { LibraryBig } from 'lucide-react';

import { Breadcrumb, BreadcrumbItem, BreadcrumbList, BreadcrumbPage } from '@/components/ui/breadcrumb';
import { Separator } from '@/components/ui/separator';
import { SidebarTrigger } from '@/components/ui/sidebar';

interface KnowledgeHeaderProps {
    isNew: boolean;
    knowledgeName: null | string;
    saveButton?: ReactNode;
}

export const KnowledgeHeader = ({ isNew, knowledgeName, saveButton }: KnowledgeHeaderProps) => (
    <header className="bg-background sticky top-0 z-10 flex h-12 shrink-0 items-center gap-2 border-b px-4">
        <SidebarTrigger className="-ml-1" />
        <Separator
            className="mr-2 h-4"
            orientation="vertical"
        />
        <Breadcrumb>
            <BreadcrumbList>
                <BreadcrumbItem>
                    <LibraryBig className="size-4 shrink-0" />
                    <BreadcrumbPage className="max-w-[240px] truncate">
                        {isNew ? 'New knowledge' : (knowledgeName ?? 'Knowledge')}
                    </BreadcrumbPage>
                </BreadcrumbItem>
            </BreadcrumbList>
        </Breadcrumb>
        {saveButton ? <div className="ml-auto flex items-center gap-2">{saveButton}</div> : null}
    </header>
);
