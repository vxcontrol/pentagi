import * as Tabs from "@radix-ui/react-tabs";
import { useLocalStorage } from "@uidotdev/usehooks";
import { useState } from "react";
import { useNavigate, useParams } from "react-router-dom";

import Browser from "@/components/Browser/Browser";
import { Button } from "@/components/Button/Button";
import { Icon } from "@/components/Icon/Icon";
import { Messages } from "@/components/Messages/Messages";
import { Panel } from "@/components/Panel/Panel";
import {
  terminalStyles,
  tabsContentStyles,
  tabsListStyles,
  tabsRootStyles,
  tabsTriggerStyles,
} from "@/components/Tabs/Tabs.css";
import { Terminal } from "@/components/Terminal/Terminal";
import { Tooltip } from "@/components/Tooltip/Tooltip";
import {
  useFlowQuery,
  useFlowUpdatedSubscription,
  useCreateFlowMutation,
  useFinishFlowMutation,
  usePutUserInputMutation,
  useTaskUpdatedSubscription,
  useScreenshotAddedSubscription,
  useTerminalLogAddedSubscription,
  useMessageLogAddedSubscription,
  useMessageLogUpdatedSubscription,
  useAgentLogAddedSubscription,
  useSearchLogAddedSubscription,
  useVectorStoreLogAddedSubscription,
} from "@/generated/graphql";
import { AgentLogs } from "@/components/AgentLogs/AgentLogs";
import { SearchLogs } from "@/components/SearchLogs/SearchLogs";
import { VectorStoreLogs } from "@/components/VectorStoreLogs/VectorStoreLogs";

import {
  followButtonStyles,
  leftColumnStyles,
  tabsStyles,
  wrapperStyles,
  leftPanelStyles,
  rightPanelStyles,
} from "./ChatPage.css";

export const ChatPage = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [, createFlowMutation] = useCreateFlowMutation();
  const [, putUserInputMutation] = usePutUserInputMutation();
  const [, finishFlowMutation] = useFinishFlowMutation();
  const isNewFlow = !id || id === "new";
  const [isFollowingTabs, setIsFollowingTabs] = useLocalStorage(
    "isFollowingTabs",
    true,
  );
  const [selectedProvider] = useLocalStorage<string>(
    "provider",
    "",
  );
  const [activeTab, setActiveTab] = useState("terminal");
  const [isSubmitting, setIsSubmitting] = useState(false);
  const [expandedAgentLogs, setExpandedAgentLogs] = useState<string[]>([]);
  const [expandedSearchLogs, setExpandedSearchLogs] = useState<string[]>([]);
  const [expandedVectorStoreLogs, setExpandedVectorStoreLogs] = useState<string[]>([]);

  const [{ operation, data }] = useFlowQuery({
    pause: isNewFlow,
    variables: { id: id || "" },
    requestPolicy: 'cache-and-network',
  });

  const isStaleData = operation?.variables.id !== id;

  const messageLogs = !isStaleData ? data?.messageLogs ?? [] : [];
  const terminalLogs = !isStaleData ? data?.terminalLogs ?? [] : [];
  const name = !isStaleData ? data?.flow?.title ?? "" : "";
  const status = !isStaleData ? data?.flow?.status : undefined;
  const terminals = !isStaleData ? data?.flow?.terminals ?? [] : [];
  const terminal = terminals[0];
  const screenshots = !isStaleData ? data?.screenshots ?? [] : [];
  const agentLogs = !isStaleData ? data?.agentLogs ?? [] : [];
  const searchLogs = !isStaleData ? data?.searchLogs ?? [] : [];
  const vectorStoreLogs = !isStaleData ? data?.vectorStoreLogs ?? [] : [];

  useFlowUpdatedSubscription({
    variables: { flowId: id || "" },
    pause: isNewFlow,
  });

  useTaskUpdatedSubscription({
    variables: { flowId: id || "" },
    pause: isNewFlow,
  });

  useScreenshotAddedSubscription({
    variables: { flowId: id || "" },
    pause: isNewFlow,
  });

  useTerminalLogAddedSubscription({
    variables: { flowId: id || "" },
    pause: isNewFlow,
  });

  useMessageLogAddedSubscription({
    variables: { flowId: id || "" },
    pause: isNewFlow,
  });

  useMessageLogUpdatedSubscription({
    variables: { flowId: id || "" },
    pause: isNewFlow,
  });

  useAgentLogAddedSubscription({
    variables: { flowId: id || "" },
    pause: isNewFlow,
  });

  useSearchLogAddedSubscription({
    variables: { flowId: id || "" },
    pause: isNewFlow,
  });

  useVectorStoreLogAddedSubscription({
    variables: { flowId: id || "" },
    pause: isNewFlow,
  });

  const handleSubmit = async (message: string) => {
    if (isNewFlow && selectedProvider && message !== "") {
      try {
        setIsSubmitting(true);
        const result = await createFlowMutation({
          modelProvider: selectedProvider,
          input: message,
        });

        const flowId = result?.data?.createFlow?.id;
        if (flowId) {
          navigate(`/chat/${flowId}`, { replace: true });
        }
      } finally {
        setIsSubmitting(false);
      }
    }
    if (!isNewFlow && message !== "") {
      try {
        setIsSubmitting(true);
        const result = await putUserInputMutation({ flowId: id, input: message });
        console.log("putUserInputMutation result", result?.data?.putUserInput);
      } finally {
        setIsSubmitting(false);
      }
    }
  };

  const handleFlowStop = () => {
    if (id) {
      finishFlowMutation({ flowId: id });
    }
  };

  const handleChangeIsFollowingTabs = () => {
    setIsFollowingTabs(!isFollowingTabs);
  };

  const handleToggleAgentLog = (logId: string) => {
    setExpandedAgentLogs(prev => {
      if (prev.includes(logId)) {
        return prev.filter(id => id !== logId);
      }
      return [...prev, logId];
    });
  };

  const handleToggleSearchLog = (logId: string) => {
    setExpandedSearchLogs(prev => {
      if (prev.includes(logId)) {
        return prev.filter(id => id !== logId);
      }
      return [...prev, logId];
    });
  };

  const handleToggleVectorStoreLog = (logId: string) => {
    setExpandedVectorStoreLogs(prev => {
      if (prev.includes(logId)) {
        return prev.filter(id => id !== logId);
      }
      return [...prev, logId];
    });
  };

  return (
    <div className={wrapperStyles}>
      <Panel className={leftPanelStyles}>
        <Messages
          id={id || ""}
          logs={messageLogs}
          name={name}
          onSubmit={handleSubmit}
          flowStatus={status}
          isNew={isNewFlow}
          onFlowStop={handleFlowStop}
          provider={selectedProvider}
          isInputDisabled={(status !== "waiting" && !isNewFlow) || isSubmitting}
        />
      </Panel>
      <Panel className={rightPanelStyles}>
        <Tabs.Root
          className={tabsRootStyles}
          value={activeTab}
          onValueChange={setActiveTab}
        >
          <Tabs.List className={tabsListStyles}>
            <div className={tabsStyles}>
              <div className={leftColumnStyles}>
                <Tabs.Trigger className={tabsTriggerStyles} value="terminal">
                  Terminal
                </Tabs.Trigger>
                <Tabs.Trigger className={tabsTriggerStyles} value="browser">
                  Browser
                </Tabs.Trigger>
                <Tabs.Trigger className={tabsTriggerStyles} value="agents">
                  Agents
                </Tabs.Trigger>
                <Tabs.Trigger className={tabsTriggerStyles} value="search">
                  Search
                </Tabs.Trigger>
                <Tabs.Trigger className={tabsTriggerStyles} value="vectorstore">
                  Vector Store
                </Tabs.Trigger>
              </div>

              <Tooltip
                content={
                  <>
                    Following the active tab is{" "}
                    <b>{isFollowingTabs ? "enabled" : "disabled"}</b>
                  </>
                }
              >
                <Button
                  size="small"
                  hierarchy={isFollowingTabs ? "primary" : "secondary"}
                  className={followButtonStyles}
                  onClick={handleChangeIsFollowingTabs}
                >
                  {isFollowingTabs ? <Icon.Eye /> : <Icon.EyeOff />}
                </Button>
              </Tooltip>
            </div>
          </Tabs.List>
          <Tabs.Content className={tabsContentStyles} value="terminal">
            <Terminal
              id={isNewFlow ? "" : id}
              status={status}
              title={terminal?.image}
              className={terminalStyles}
              logs={terminalLogs}
              isRunning={terminal?.connected}
            />
          </Tabs.Content>
          <Tabs.Content className={tabsContentStyles} value="browser">
            <Browser screenshots={screenshots} />
          </Tabs.Content>
          <Tabs.Content className={tabsContentStyles} value="agents">
            <AgentLogs 
              logs={agentLogs} 
              expandedLogs={expandedAgentLogs}
              onToggleLog={handleToggleAgentLog}
            />
          </Tabs.Content>
          <Tabs.Content className={tabsContentStyles} value="search">
            <SearchLogs 
              logs={searchLogs} 
              expandedLogs={expandedSearchLogs}
              onToggleLog={handleToggleSearchLog}
            />
          </Tabs.Content>
          <Tabs.Content className={tabsContentStyles} value="vectorstore">
            <VectorStoreLogs 
              logs={vectorStoreLogs} 
              expandedLogs={expandedVectorStoreLogs}
              onToggleLog={handleToggleVectorStoreLog}
            />
          </Tabs.Content>
        </Tabs.Root>
      </Panel>
    </div>
  );
};
