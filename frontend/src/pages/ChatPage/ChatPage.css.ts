import { globalStyle, style } from "@vanilla-extract/css";

import { vars } from "@/styles/theme.css";

export const wrapperStyles = style({
  display: "flex",
  flex: 1,
  padding: 16,
  gap: 16,
  margin: 0,
  justifyContent: "space-between",
  alignItems: "flex-start",
});

export const tabsStyles = style({
  display: "flex",
  justifyContent: "space-between",
  width: "100%",
  maxWidth: 1000,
});

export const leftColumnStyles = style({
  display: "flex",
  gap: 8,
});

export const followButtonStyles = style({
  backgroundColor: vars.color.gray11,
});

globalStyle(`${followButtonStyles} > svg`, {
  width: 20,
});

export const leftPanelStyles = style({
  flex: 1,
  minWidth: 0,
});

export const rightPanelStyles = style({
  width: "50%",
  maxWidth: 1000,
  marginLeft: "auto",
});
