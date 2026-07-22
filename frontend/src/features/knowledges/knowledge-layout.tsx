import type { ReactNode } from 'react';

import type { KnowledgeDocumentFragmentFragment } from '@/graphql/types';

import { cn } from '@/lib/utils';

import { KnowledgeHeader } from './knowledge-header';

interface KnowledgeLayoutProps {
    children: ReactNode;
    className?: string;
    isLoading?: boolean;
    isNew: boolean;
    knowledge?: KnowledgeDocumentFragmentFragment | null;
    saveButton?: ReactNode;
}

/**
 * Shared layout shell for the knowledge page in non-form branches
 * (loading, not-found). `KnowledgeForm` owns its own `<form>` root and
 * renders the header inline because the form must be the parent of every
 * input.
 */
export function KnowledgeLayout({
    children,
    className,
    isLoading,
    isNew,
    knowledge,
    saveButton,
}: KnowledgeLayoutProps) {
    return (
        <div className={cn('flex min-h-[100dvh] flex-col', className)}>
            <KnowledgeHeader
                isLoading={isLoading}
                isNew={isNew}
                knowledge={knowledge}
                saveButton={saveButton}
            />
            {children}
        </div>
    );
}
