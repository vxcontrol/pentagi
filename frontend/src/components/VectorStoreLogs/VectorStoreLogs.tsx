import { useEffect, useRef, useMemo, Suspense } from "react";
import Markdown from "markdown-to-jsx";
import { Icon } from "@/components/Icon/Icon";
import { Tooltip } from "@/components/Tooltip/Tooltip";
import type { VectorStoreLog } from "@/generated/graphql";
import {
  wrapperStyles,
  headerStyles,
  logListStyles,
  logItemStyles,
  agentIconsStyles,
  queryStyles,
  resultStyles,
  metaInfoStyles,
  emptyStateStyles,
  arrowIconStyles,
  logHeaderStyles,
  collapsedContentStyles,
  expandedContentStyles,
} from "./VectorStoreLogs.css";
import { markdownHeaderStyles } from "@/styles/markdown.css";
import type { MarkdownToJSX } from "markdown-to-jsx";

type VectorStoreLogsProps = {
  logs: VectorStoreLog[];
  expandedLogs: string[];
  onToggleLog: (logId: string) => void;
};

const getAgentIcon = (agentType: string) => {
  switch (agentType.toLowerCase()) {
    case "primary_agent":
      return <Icon.Brain />;
    case "reporter":
      return <Icon.Report />;
    case "generator":
      return <Icon.List />;
    case "refiner":
      return <Icon.Renew />;
    case "reflector":
      return <Icon.Reflect />;
    case "enricher":
      return <Icon.Enrich />;
    case "adviser":
      return <Icon.HelpCircle />;
    case "coder":
      return <Icon.Code />;
    case "memorist":
      return <Icon.Memory />;
    case "searcher":
      return <Icon.Search />;
    case "installer":
      return <Icon.Deploy />;
    case "pentester":
      return <Icon.Hack />;
    case "summarizer":
      return <Icon.Sum />;
    default:
      return <Icon.HelpCircle />;
  }
};

const getActionIcon = (action: string) => {
  switch (action.toLowerCase()) {
    case "store":
      return <Icon.Store />;
    case "retrieve":
      return <Icon.Retrieve />;
    default:
      return <Icon.Memory />;
  }
};

const MarkdownOverrides: MarkdownToJSX.Overrides = {
  h1: { component: "h1", props: { className: markdownHeaderStyles.h1 } },
  h2: { component: "h2", props: { className: markdownHeaderStyles.h2 } },
  h3: { component: "h3", props: { className: markdownHeaderStyles.h3 } },
  h4: { component: "h4", props: { className: markdownHeaderStyles.h4 } },
  h5: { component: "h5", props: { className: markdownHeaderStyles.h5 } },
  h6: { component: "h6", props: { className: markdownHeaderStyles.h6 } },
  a: { component: "a", props: { className: markdownHeaderStyles.link } },
  img: { component: "span" },
  iframe: { component: "span" },
  script: { component: "span" },
  style: { component: "span" },
  link: { component: "span" },
};

type LogItemProps = {
  log: VectorStoreLog;
  isExpanded: boolean;
  onToggle: () => void;
};

const logDescription = (log: VectorStoreLog) => {
  const filter = JSON.parse(log.filter);
  const docType = filter["doc_type"];
  const codeLang = filter["code_lang"];
  const toolName = filter["tool_name"];
  const guideType = filter["guide_type"];
  const answerType = filter["answer_type"];
  let description = " ";
  let prefix = "";
  let preposition = "";
  if (log.action === "store") {
    prefix = "stored";
    preposition = "in";
  } else {
    prefix = "retrieved";
    preposition = "from";
  }
  if (docType) {
    if (docType === "memory") {
      description += `${prefix} ${preposition} memory `;
    } else {
      description += `${prefix} ${docType} `;
    }
  }
  if (codeLang) {
    description += `on ${codeLang} language `;
  }
  if (toolName) {
    description += `by ${toolName} tool `;
  }
  if (guideType) {
    description += `about ${guideType} `;
  }
  if (answerType) {
    description += `as a ${answerType} `;
  }
  return description;
};

const truncate = (text: string, maxLength: number) => {
  if (text.length <= maxLength) return text;
  return text.slice(0, maxLength) + "...";
};

const LogItem = ({ log, isExpanded, onToggle }: LogItemProps) => (
  <div className={logItemStyles}>
    <div className={logHeaderStyles} onClick={onToggle} role="button" tabIndex={0}>
      <div className={agentIconsStyles}>
        <Tooltip content={`Initiator: ${log.initiator}`}>
          {getAgentIcon(log.initiator)}
        </Tooltip>
        <Icon.ArrowRight className={arrowIconStyles} />
        <Tooltip content={`Executor: ${log.executor}`}>
          {getAgentIcon(log.executor)}
        </Tooltip>
        <span> {logDescription(log)} </span>
        <Tooltip content={`Action: ${log.action}`}>
          {getActionIcon(log.action)}
        </Tooltip>
      </div>
      <div className={arrowIconStyles}>
        {isExpanded ? <Icon.ArrowUp /> : <Icon.ArrowDown />}
      </div>
    </div>

    <div className={isExpanded ? expandedContentStyles : collapsedContentStyles}>
      <div className={queryStyles}>
        <Suspense fallback={<div>Loading...</div>}>
          {isExpanded && (
            <Markdown options={{ overrides: MarkdownOverrides }}>
              {log.query}
            </Markdown>
          )}
        </Suspense>
      </div>
      
      <div className={resultStyles}>
        <Suspense fallback={<div>Loading...</div>}>
          {isExpanded && (
            <Markdown options={{ overrides: MarkdownOverrides }}>
              {log.action === "store" && log.filter.includes("memory") 
                ? truncate(log.result, 2000)
                : log.result
              }
            </Markdown>
          )}
        </Suspense>
      </div>

      {(log.taskId || log.subtaskId) && (
        <div className={metaInfoStyles}>
          {log.taskId && `Task ID: ${log.taskId}`}
          {log.taskId && log.subtaskId && " | "}
          {log.subtaskId && `Subtask ID: ${log.subtaskId}`}
        </div>
      )}
    </div>
  </div>
);

export const VectorStoreLogs = ({ logs, expandedLogs, onToggleLog }: VectorStoreLogsProps) => {
  const bottomRef = useRef<HTMLDivElement>(null);

  const sortedLogs = useMemo(() => {
    return [...logs].sort((a, b) => 
      parseInt(a.id) - parseInt(b.id)
    );
  }, [logs]);

  useEffect(() => {
    bottomRef.current?.scrollIntoView({ behavior: "instant" });
  }, []);

  useEffect(() => {
    if (logs.length > 0) {
      bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }
  }, [logs]);

  if (!logs.length) {
    return (
      <div className={wrapperStyles}>
        <div className={headerStyles}>Vector Store Logs</div>
        <div className={emptyStateStyles}>
          No vector store logs available
        </div>
      </div>
    );
  }

  return (
    <div className={wrapperStyles}>
      <div className={headerStyles}>
        Vector Store Logs ({logs.length} entries)
      </div>
      <div className={logListStyles}>
        {sortedLogs.map((log) => (
          <LogItem
            key={log.id}
            log={log}
            isExpanded={expandedLogs.includes(log.id)}
            onToggle={() => onToggleLog(log.id)}
          />
        ))}
        <div ref={bottomRef} />
      </div>
    </div>
  );
};

export default VectorStoreLogs; 