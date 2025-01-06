import { wrapperStyles } from "./Panel.css";

type PanelProps = {
  children: React.ReactNode;
  className?: string;
};

export const Panel = ({ children, className }: PanelProps) => (
  <div className={`${wrapperStyles} ${className || ''}`}>{children}</div>
);
