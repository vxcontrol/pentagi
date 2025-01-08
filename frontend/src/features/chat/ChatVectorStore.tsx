import { useState } from 'react';

import Markdown from '@/components/Markdown';
import type { VectorStoreLogFragmentFragment } from '@/graphql/types';
import { VectorStoreAction } from '@/graphql/types';
import { formatDate } from '@/lib/utils/format';

import ChatAgentIcon from './ChatAgentIcon';
import ChatVectorStoreActionIcon from './ChatVectorStoreActionIcon';

const getDescription = (log: VectorStoreLogFragmentFragment) => {
    const { action, filter } = log;
    const {
        doc_type: docType,
        code_lang: codeLang,
        tool_name: toolName,
        guide_type: guideType,
        answer_type: answerType,
    } = JSON.parse(filter) || {};

    let description = '';
    const prefix = action === VectorStoreAction.Store ? 'Stored' : 'Retrieved';
    const preposition = action === VectorStoreAction.Store ? 'in' : 'from';

    if (docType) {
        if (docType === 'memory') {
            description += `${prefix} ${preposition} memory`;
        } else {
            description += `${prefix} ${docType}`;
        }
    }

    if (codeLang) {
        description += `${description ? ' on' : 'On'} ${codeLang} language`;
    }

    if (toolName) {
        description += `${description ? ' by' : 'By'} ${toolName} tool`;
    }

    if (guideType) {
        description += `${description ? ' about' : 'About'} ${guideType}`;
    }

    if (answerType) {
        description += `${description ? ' as' : 'As'} a ${answerType}`;
    }

    return description;
};

interface ChatVectorStoreProps {
    log: VectorStoreLogFragmentFragment;
}

const ChatVectorStore = ({ log }: ChatVectorStoreProps) => {
    const { executor, initiator, query, result, action, taskId, subtaskId, createdAt } = log;
    const [isDetailsVisible, setIsDetailsVisible] = useState(false);

    const description = getDescription(log);

    return (
        <div className="flex flex-col items-start">
            <div className="max-w-full rounded-lg bg-accent p-3 text-accent-foreground">
                <div className="flex flex-col">
                    <div className="cursor-pointer text-sm font-semibold">
                        <span className="inline-flex items-center gap-1">
                            <ChatVectorStoreActionIcon action={action} />
                            <span>{description}</span>
                        </span>
                    </div>

                    <Markdown className="prose-xs prose-fixed break-words">{query}</Markdown>
                </div>
                {result && (
                    <div className="text-xs text-muted-foreground">
                        <div
                            onClick={() => setIsDetailsVisible(!isDetailsVisible)}
                            className="cursor-pointer"
                        >
                            {isDetailsVisible ? 'Hide details' : 'Show details'}
                        </div>
                        {isDetailsVisible && (
                            <>
                                <div className="my-2 border-t dark:border-gray-700" />
                                <Markdown className="prose-xs prose-fixed break-words">{result}</Markdown>
                            </>
                        )}
                    </div>
                )}
            </div>
            <div className="mt-1 flex items-center gap-1 px-1 text-xs text-muted-foreground">
                <span className="flex items-center gap-0.5">
                    <ChatAgentIcon
                        type={executor}
                        className="text-muted-foreground"
                    />
                    <span className="text-muted-foreground/50">→</span>
                    <ChatAgentIcon
                        type={initiator}
                        className="text-muted-foreground"
                    />
                </span>
                <span className="text-muted-foreground/50">{formatDate(new Date(createdAt))}</span>
                {taskId && (
                    <>
                        <span className="text-muted-foreground/50">|</span>
                        <span className="text-muted-foreground/50">Task ID: {taskId}</span>
                    </>
                )}
                {subtaskId && (
                    <>
                        <span className="text-muted-foreground/50">|</span>
                        <span className="text-muted-foreground/50">Subtask ID: {subtaskId}</span>
                    </>
                )}
            </div>
        </div>
    );
};

export default ChatVectorStore;
