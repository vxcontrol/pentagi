import * as TooltipRadix from "@radix-ui/react-tooltip";

import { arrowStyles, contentStyles } from "./Tooltip.css";

type TooltipProps = {
  children: React.ReactNode;
  content: React.ReactNode;
  contentProps?: TooltipRadix.TooltipContentProps;
} & TooltipRadix.TooltipProps;

export const Tooltip = ({ children, content, contentProps }: TooltipProps) => {
  if (!content) return <>{children}</>;

  return (
    <TooltipRadix.Provider>
      <TooltipRadix.Root delayDuration={300}>
        <TooltipRadix.Trigger asChild>{children}</TooltipRadix.Trigger>
        <TooltipRadix.Portal>
          <TooltipRadix.Content
            className={contentStyles}
            sideOffset={5}
            {...contentProps}
          >
            {content}
            <TooltipRadix.Arrow className={arrowStyles} />
          </TooltipRadix.Content>
        </TooltipRadix.Portal>
      </TooltipRadix.Root>
    </TooltipRadix.Provider>
  );
};
