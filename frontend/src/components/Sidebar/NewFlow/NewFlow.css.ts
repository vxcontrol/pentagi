import { style, globalStyle } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const wrapperStyles = style({
  display: "flex",
  alignItems: "center",
  gap: "8px",
  minHeight: "40px",
});

export const linkWrapperStyles = style([
  font.textSmSemibold,
  {
    display: "block",
    textDecoration: "none",
    background: vars.color.gray3,
    border: "none",
    textAlign: "left",
    color: vars.color.gray12,
    padding: "9px 16px",
    cursor: "pointer",
    borderRadius: "6px 0 0 6px",
    flex: 1,

    selectors: {
      "&.active": {
        color: vars.color.primary9,
        backgroundColor: vars.color.gray5,
      },
    },

    ":hover": {
      color: vars.color.primary9,
      backgroundColor: vars.color.gray4,
    },
  },
]);

export const collapseButtonStyles = style([
  font.textSmSemibold,
  {
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    background: vars.color.gray3,
    border: "none",
    color: vars.color.success9,
    cursor: "pointer",
    borderRadius: "8px",
    padding: "6px 9px",
    marginRight: "4px",
    height: "32px",
    width: "32px",
    marginLeft: "8px",
    flexShrink: 0,
    scale: 1.3,

    ":hover": {
      color: vars.color.primary9,
      backgroundColor: vars.color.gray4,
    },
  },
]);

globalStyle(`${collapseButtonStyles} > svg`, {
  width: "16px",
  height: "16px",
});
