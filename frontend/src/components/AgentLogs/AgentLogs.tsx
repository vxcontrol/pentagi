import { useEffect, useRef, useMemo, Suspense } from "react";
import Markdown from "markdown-to-jsx";
import { Icon } from "@/components/Icon/Icon";
import { Tooltip } from "@/components/Tooltip/Tooltip";
import type { AgentLog } from "@/generated/graphql";
import {
  wrapperStyles,
  headerStyles,
  logListStyles,
  logItemStyles,
  agentIconsStyles,
  taskStyles,
  resultStyles,
  metaInfoStyles,
  emptyStateStyles,
  arrowIconStyles,
  logHeaderStyles,
  collapsedContentStyles,
  expandedContentStyles,
} from "./AgentLogs.css";
import { markdownHeaderStyles } from "@/styles/markdown.css";
import type { MarkdownToJSX } from "markdown-to-jsx";

type AgentLogsProps = {
  logs: AgentLog[];
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

const MarkdownOverrides: MarkdownToJSX.Overrides = {
  h1: { component: "h1", props: { className: markdownHeaderStyles.h1 } },
  h2: { component: "h2", props: { className: markdownHeaderStyles.h2 } },
  h3: { component: "h3", props: { className: markdownHeaderStyles.h3 } },
  h4: { component: "h4", props: { className: markdownHeaderStyles.h4 } },
  h5: { component: "h5", props: { className: markdownHeaderStyles.h5 } },
  h6: { component: "h6", props: { className: markdownHeaderStyles.h6 } },
  a: { component: "a", props: { className: markdownHeaderStyles.link } },
};

type LogItemProps = {
  log: AgentLog;
  isExpanded: boolean;
  onToggle: () => void;
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
      </div>
      <div className={arrowIconStyles}>
        {isExpanded ? <Icon.ArrowUp /> : <Icon.ArrowDown />}
      </div>
    </div>

    <div className={isExpanded ? expandedContentStyles : collapsedContentStyles}>
      <div className={taskStyles}>
        <Suspense fallback={<div>Loading...</div>}>
          {isExpanded && (
            <Markdown options={{ overrides: MarkdownOverrides }}>
              {log.task}
            </Markdown>
          )}
        </Suspense>
      </div>
      
      <div className={resultStyles}>
        <Suspense fallback={<div>Loading...</div>}>
          {isExpanded && (
            <Markdown options={{ overrides: MarkdownOverrides }}>
              {log.result}
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

export const AgentLogs = ({ logs, expandedLogs, onToggleLog }: AgentLogsProps) => {
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
        <div className={headerStyles}>Agent Logs</div>
        <div className={emptyStateStyles}>
          No agent logs available
        </div>
      </div>
    );
  }

  return (
    <div className={wrapperStyles}>
      <div className={headerStyles}>
        Agent Logs ({logs.length} entries)
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

export default AgentLogs;