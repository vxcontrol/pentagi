import { globalStyle, style } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const tabsRootStyles = style({
  display: "flex",
  flexDirection: "column",
  height: "calc(100dvh - 64px)",
  maxHeight: "calc(100dvh - 64px)",
  maxWidth: 1000,
});

export const tabsListStyles = style({
  flexShrink: 0,
  display: "flex",
  gap: 16,
  paddingBottom: 4,
});

export const tabsTriggerStyles = style([
  font.textSmSemibold,
  {
    position: "relative",
    backgroundColor: vars.color.gray2,
    padding: "12px 8px",
    height: 32,
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    userSelect: "none",
    border: "none",
    color: vars.color.gray11,
    borderRadius: 8,
    gap: 6,

    selectors: {
      '&[data-state="active"]': {
        color: vars.color.gray12,
        backgroundColor: vars.color.gray3,
      },
      '&:hover:not([data-state="active"])': {
        backgroundColor: vars.color.gray3,
        cursor: "pointer",
      },
    },
  },
]);

globalStyle(`${tabsTriggerStyles}:where([data-state="active"]):before`, {
  content: "''",
  position: "absolute",
  left: 0,
  right: 0,
  bottom: -5,
  height: 2,
  backgroundColor: vars.color.primary9,
});

export const tabsContentStyles = style({
  paddingTop: 24,
  height: "calc(100dvh - 24px)",
  maxHeight: "calc(100dvh - 24px)",
});

export const terminalStyles = style({
  height: "calc(100% - 24px)",
  maxHeight: "calc(100% - 24px)",
});
