import { style, globalStyle } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const messagesWrapper = style({
  position: "relative",
  display: "flex",
  flexDirection: "column",
  height: "calc(100dvh - 72px)",
  maxHeight: "calc(100dvh - 72px)",
  overflow: "hidden",
  margin: "8px",
});

export const titleStyles = style([
  font.textSmSemibold,
  {
    position: "relative",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    color: vars.color.gray11,
    textAlign: "center",
    padding: "8px 0",
    borderBottom: `1px solid ${vars.color.gray5}`,
    zIndex: 2,
    backgroundColor: vars.color.gray2,
  },
]);

export const taskTreeButtonStyles = style([
  font.textSmSemibold,
  {
    position: "absolute",
    left: "16px",
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    background: vars.color.gray3,
    border: "none",
    color: vars.color.success9,
    cursor: "pointer",
    borderRadius: "8px",
    padding: 0,
    height: "32px",
    width: "32px",
    marginLeft: "8px",
    marginBottom: "8px",
    flexShrink: 0,
    scale: 1.3,

    ":hover": {
      color: vars.color.primary9,
      backgroundColor: vars.color.gray4,
    },
  },
]);

globalStyle(`${taskTreeButtonStyles} svg`, {
  width: "16px",
  height: "16px",
  scale: "1.3",
  display: "block",
  margin: "auto",
  position: "absolute",
  top: "12px",
  left: "12px",
});

export const titleContentStyles = style({
  display: "flex",
  alignItems: "center",
  gap: 12,
});

export const taskTreeContainerStyles = style({
  backgroundColor: vars.color.gray2,
  borderBottom: `1px solid ${vars.color.gray5}`,
  transition: "height 0.3s ease",
});

export const messagesListWrapper = style({
  display: "flex",
  flexDirection: "column",
  gap: 22,
  paddingTop: "16px",
  paddingLeft: "0px",
  paddingRight: "8px",
  paddingBottom: "24px",
  overflowY: "auto",
  overflowX: "hidden",
  flex: 1,
  minHeight: 0,
  marginBottom: 120,
});

export const modelStyles = style({
  color: vars.color.gray10,
});

export const newMessageTextarea = style([
  font.textSmMedium,
  {
    position: "absolute",
    bottom: 0,
    left: 0,
    right: 0,
    backgroundColor: vars.color.gray4,
    border: `1px solid ${vars.color.gray5}`,
    borderRadius: "0 0 6px 6px",
    width: "100%",
    height: 120,
    color: vars.color.gray12,
    padding: 16,
    flexShrink: 0,
    boxShadow: `0 -20px 30px 10px ${vars.color.gray2}`,
    resize: "none",
    zIndex: 1,

    ":focus": {
      outline: "none",
      borderColor: vars.color.primary5,
    },

    ":disabled": {
      backgroundColor: vars.color.gray3,
      borderColor: vars.color.gray4,
    },
  },
]);
