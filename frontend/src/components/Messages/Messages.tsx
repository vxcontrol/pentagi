import { useEffect, useRef, useState } from "react";

import { StatusType, MessageLog } from "@/generated/graphql";

import { Button } from "../Button/Button";
import { Message } from "./Message/Message";
import { Icon } from "../Icon/Icon";
import {
  messagesListWrapper,
  messagesWrapper,
  modelStyles,
  newMessageTextarea,
  titleStyles,
  titleContentStyles,
  taskTreeContainerStyles,
  taskTreeButtonStyles,
} from "./Messages.css";
import { TaskTree } from "../TaskTree/TaskTree";

type MessagesProps = {
  logs: MessageLog[];
  name: string;
  onSubmit: (message: string) => void;
  onFlowStop: () => void;
  flowStatus?: StatusType;
  isNew?: boolean;
  provider?: string;
  isInputDisabled: boolean;
  id: string;
};

export const Messages = ({
  logs,
  name,
  flowStatus,
  onSubmit,
  isNew,
  onFlowStop,
  provider,
  isInputDisabled,
  id,
}: MessagesProps) => {
  const [isTaskTreeVisible, setIsTaskTreeVisible] = useState(false);
  const messagesRef = useRef<HTMLDivElement>(null);
  const textareaRef = useRef<HTMLTextAreaElement>(null);
  const autoScrollEnabledRef = useRef(true);

  useEffect(() => {
    if (!isInputDisabled && textareaRef.current) {
      textareaRef.current.focus();
    }
  }, [isInputDisabled]);

  const handleKeyPress = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
    if (e.key === "Enter" && !e.shiftKey) {
      e.preventDefault();

      const message = e.currentTarget.value;
      e.currentTarget.value = "";

      onSubmit(message);
    }
  };

  useEffect(() => {
    const messagesDiv = messagesRef.current;
    if (!messagesDiv) return;

    const scrollHandler = () => {
      if (
        messagesDiv.scrollTop + messagesDiv.clientHeight + 50 >=
        messagesDiv.scrollHeight
      ) {
        autoScrollEnabledRef.current = true;
      } else {
        autoScrollEnabledRef.current = false;
      }
    };

    messagesDiv.addEventListener("scroll", scrollHandler);

    return () => {
      messagesDiv.removeEventListener("scroll", scrollHandler);
    };
  }, []);

  useEffect(() => {
    const messagesDiv = messagesRef.current;
    if (!messagesDiv) return;

    if (autoScrollEnabledRef.current) {
      messagesDiv.scrollTop = messagesDiv.scrollHeight;
    }
  }, [logs]);

  const isFlowFinished = flowStatus === StatusType.Finished;

  const toggleTaskTree = () => {
    setIsTaskTreeVisible(!isTaskTreeVisible);
  };

  return (
    <div className={messagesWrapper}>
      {name && (
        <>
          <div className={titleStyles}>
            <Button
              hierarchy="secondary"
              size="small"
              onClick={toggleTaskTree}
              className={taskTreeButtonStyles}
            >
              {isTaskTreeVisible ? <Icon.ArrowUp /> : <Icon.ArrowDown />}
            </Button>
            <div className={titleContentStyles}>
              {name}
              <span className={modelStyles}>{provider ? ` - ${provider}` : ""}</span>{" "}
              {isFlowFinished ? (
                " (Finished)"
              ) : (
                <Button hierarchy="danger" size="small" onClick={onFlowStop}>
                  Finish
                </Button>
              )}
            </div>
          </div>
          <div className={taskTreeContainerStyles}>
            <TaskTree flowId={id} visibility={isTaskTreeVisible} />
          </div>
        </>
      )}
      <div className={messagesListWrapper} ref={messagesRef}>
        {logs.map((message) => (
          <Message key={message.id} {...message} />
        ))}
      </div>
      <textarea
        ref={textareaRef}
        className={newMessageTextarea}
        placeholder={
          isFlowFinished
            ? "The flow is finished."
            : isNew
              ? "Type a new message to start the flow..."
              : "Type a message..."
        }
        onKeyPress={handleKeyPress}
        disabled={isInputDisabled}
      />
    </div>
  );
};
