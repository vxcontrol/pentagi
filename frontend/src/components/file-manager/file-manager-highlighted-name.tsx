import { Fragment, type ReactNode } from 'react';

import { cn } from '@/lib/utils';

interface FileManagerHighlightedNameProps {
    className?: string;
    name: string;
    query?: string;
}

export const FileManagerHighlightedName = ({ className, name, query }: FileManagerHighlightedNameProps) => {
    if (!query?.trim()) {
        return <span className={cn('truncate', className)}>{name}</span>;
    }

    const lower = name.toLowerCase();
    const lowerQuery = query.toLowerCase();
    const segments: ReactNode[] = [];
    let cursor = 0;

    while (cursor < name.length) {
        const matchIndex = lower.indexOf(lowerQuery, cursor);

        if (matchIndex < 0) {
            segments.push(<Fragment key={`t-${cursor}`}>{name.slice(cursor)}</Fragment>);
            break;
        }

        if (matchIndex > cursor) {
            segments.push(<Fragment key={`p-${cursor}`}>{name.slice(cursor, matchIndex)}</Fragment>);
        }

        segments.push(
            <mark
                className="bg-primary/20 text-foreground rounded-[2px]"
                key={`m-${matchIndex}`}
            >
                {name.slice(matchIndex, matchIndex + lowerQuery.length)}
            </mark>,
        );

        cursor = matchIndex + lowerQuery.length;
    }

    return <span className={cn('truncate', className)}>{segments}</span>;
};
