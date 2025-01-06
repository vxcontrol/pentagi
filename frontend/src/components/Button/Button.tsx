import { forwardRef } from "react";

import {
  buttonIconStyles,
  buttonSizesStyles,
  buttonStyles,
} from "./Button.css";

export type ButtonProps = {
  children: React.ReactNode;
  icon?: React.ReactNode;
  disabled?: boolean;
  hierarchy?: "primary" | "secondary" | "danger";
  size?: "small" | "medium";
} & React.ButtonHTMLAttributes<HTMLButtonElement>;

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  (
    {
      icon = null,
      disabled = false,
      children,
      hierarchy = "primary",
      size = "medium",
      className,
      ...rest
    },
    ref,
  ) => (
    <button
      ref={ref}
      className={
        (hierarchy === "primary"
          ? buttonStyles.Primary
          : hierarchy === "secondary"
            ? buttonStyles.Secondary
            : buttonStyles.Danger) +
        " " +
        (size === "small"
          ? buttonSizesStyles.Small
          : buttonSizesStyles.Medium) +
        " " +
        className
      }
      disabled={disabled}
      {...rest}
    >
      {icon && <div className={buttonIconStyles}>{icon}</div>}
      {children}
    </button>
  ),
);
