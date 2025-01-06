import { useEffect, useRef, useMemo, forwardRef, Suspense } from "react";
import Markdown from "markdown-to-jsx";
import { Icon } from "@/components/Icon/Icon";
import { Tooltip } from "@/components/Tooltip/Tooltip";
import type { SearchLog } from "@/generated/graphql";
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
} from "./SearchLogs.css";
import { markdownHeaderStyles } from "@/styles/markdown.css";
import type { MarkdownToJSX } from "markdown-to-jsx";

type SearchLogsProps = {
  logs: SearchLog[];
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

const getEngineIcon = (engine: string) => {
  switch (engine.toLowerCase()) {
    case "tavily":
      return <Icon.Tavily />;
    case "google":
      return <Icon.Google />;
    case "traversaal":
      return <Icon.Traversaal />;
    default:
      return <Icon.Search />;
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
  log: SearchLog;
  isExpanded: boolean;
  onToggle: () => void;
};

const SafeMarkdown = forwardRef<HTMLDivElement, { children: string; className?: string }>(
  ({ children, className }, ref) => (
    <div ref={ref} className={className}>
      <Markdown 
        options={{ 
          overrides: MarkdownOverrides,
          forceBlock: true,
          disableParsingRawHTML: true,
        }}
      >
        {children}
      </Markdown>
    </div>
  )
);

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
        <span> by tool </span>
        <Tooltip content={`Engine: ${log.engine}`}>
          {getEngineIcon(log.engine)}
        </Tooltip>
      </div>
      <div className={arrowIconStyles}>
        {isExpanded ? <Icon.ArrowUp /> : <Icon.ArrowDown />}
      </div>
    </div>

    <div className={isExpanded ? expandedContentStyles : collapsedContentStyles}>
      <Suspense fallback={<div>Loading...</div>}>
        {isExpanded && (
          <SafeMarkdown className={queryStyles}>
            {log.query}
          </SafeMarkdown>
        )}
      </Suspense>
      
      <Suspense fallback={<div>Loading...</div>}>
        {isExpanded && (
          <SafeMarkdown className={resultStyles}>
            {log.result}
          </SafeMarkdown>
        )}
      </Suspense>

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

export const SearchLogs = ({ logs, expandedLogs, onToggleLog }: SearchLogsProps) => {
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
        <div className={headerStyles}>Search Logs</div>
        <div className={emptyStateStyles}>
          No search logs available
        </div>
      </div>
    );
  }

  return (
    <div className={wrapperStyles}>
      <div className={headerStyles}>
        Search Logs ({logs.length} entries)
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

export default SearchLogs; 