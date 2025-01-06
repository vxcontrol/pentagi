import { formatDistanceToNowStrict } from "date-fns";
import { useState } from "react";
import Markdown from "markdown-to-jsx";
import { markdownHeaderStyles } from "@/styles/markdown.css";
import type { MarkdownToJSX } from "markdown-to-jsx";

import logoPng from "@/assets/logo.png";
import mePng from "@/assets/me.png";
import { Icon } from "@/components/Icon/Icon";
import { MessageLogType } from "@/generated/graphql";

import {
  avatarStyles,
  contentStyles,
  iconStyles,
  messageStyles,
  outputStyles,
  rightColumnStyles,
  timeStyles,
  wrapperStyles,
} from "./Message.css";

type MessageProps = {
  id: string;
  type: MessageLogType;
  message: string;
  result: string;
  createdAt: string;
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

export const Message = ({
  type,
  message,
  result,
  createdAt,
}: MessageProps) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const toggleExpand = () => {
    setIsExpanded((prev) => !prev);
  };

  const isUserInput = type === MessageLogType.Input;

  return (
    <div className={wrapperStyles}>
      <img
        src={isUserInput ? mePng : logoPng}
        alt="avatar"
        className={avatarStyles}
        width="40"
        height="40"
      />
      <div className={rightColumnStyles}>
        <div className={timeStyles}>
          {formatDistanceToNowStrict(new Date(createdAt), { addSuffix: true })}
        </div>
        <div
          className={isUserInput ? messageStyles.Input : messageStyles.Regular}
          onClick={toggleExpand}
        >
          <div className={contentStyles}>
            {isUserInput ? (
              <Markdown options={{ overrides: MarkdownOverrides }}>
                {message}
              </Markdown>
            ) : (
              <>
                <span className={iconStyles.Regular}>
                  {getIcon(type)}
                </span>
                <div>{message}</div>
              </>
            )}
          </div>
        </div>
        {isExpanded && result && (
          <div className={outputStyles}>
            <Markdown options={{ overrides: MarkdownOverrides }}>
              {result}
            </Markdown>
          </div>
        )}
      </div>
    </div>
  );
};

const getIcon = (type: MessageLogType) => {
  switch (type) {
    case MessageLogType.Browser:
      return <Icon.Browser />;
    case MessageLogType.Terminal:
      return <Icon.Terminal />;
    case MessageLogType.File:
      return <Icon.Code />;
    case MessageLogType.Advice:
      return <Icon.MessageHelp />;
    case MessageLogType.Ask:
      return <Icon.MessageQuestion />;
    case MessageLogType.Done:
      return <Icon.CheckCircle />;
    case MessageLogType.Thoughts:
      return <Icon.Brain />;
    case MessageLogType.Search:
      return <Icon.Search />;
    case MessageLogType.Input:
      return <Icon.User />;
    default:
      return null;
  }
};
