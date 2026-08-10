import { SquareMenu, Type } from 'lucide-react';

import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { useI18n } from '@/hooks/use-i18n';

// 'rich' reflows whitespace on save (tiptap); 'raw' is a byte-exact textarea over the source.
export type EditorViewMode = 'raw' | 'rich';

interface EditorViewModeToggleProps {
    className?: string;
    mode: EditorViewMode;
    onModeChange: (mode: EditorViewMode) => void;
    rawTooltip?: string;
}

export function EditorViewModeToggle({ className, mode, onModeChange, rawTooltip }: EditorViewModeToggleProps) {
    const { t } = useI18n();

    return (
        <Tabs
            className={className}
            onValueChange={(value) => onModeChange(value as EditorViewMode)}
            value={mode}
        >
            <TabsList className="dark:bg-background h-7 p-0.5">
                <TabsTrigger
                    aria-label={t('markdownEditor.richEditor')}
                    className="dark:data-[state=active]:bg-card h-6 px-2"
                    value="rich"
                >
                    <SquareMenu className="size-4" />
                </TabsTrigger>
                <TabsTrigger
                    aria-label={t('markdownEditor.rawSource')}
                    className="dark:data-[state=active]:bg-card h-6 px-2"
                    title={rawTooltip}
                    value="raw"
                >
                    <Type className="size-4" />
                </TabsTrigger>
            </TabsList>
        </Tabs>
    );
}
