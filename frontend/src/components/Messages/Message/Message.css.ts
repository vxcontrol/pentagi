import { globalStyle, style, styleVariants } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const wrapperStyles = style({
  display: "flex",
  gap: 12,
});

export const avatarStyles = style({
  borderRadius: "50%",
  border: `1px solid ${vars.color.gray4}`,
});

export const rightColumnStyles = style({
  display: "flex",
  flexDirection: "column",
  gap: 6,
  flex: 1,
});

export const timeStyles = style([
  font.textXsRegular,
  {
    color: vars.color.gray8,
  },
]);

const messageStylesBase = style([
  font.textSmRegular,
  {
    padding: "10px 14px",
    borderRadius: "0 8px 8px 8px",
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    cursor: "pointer",
    color: vars.color.primary12,
  },
]);

export const messageStyles = styleVariants({
  Input: [
    messageStylesBase,
    {
      border: `1px solid ${vars.color.gray3}`,
      background: vars.color.gray1,
      cursor: "auto",
      color: vars.color.gray12,
    },
  ],
  Regular: [
    messageStylesBase,
    {
      border: `1px solid ${vars.color.gray4}`,
      background: vars.color.gray3,
      ":hover": {
        background: vars.color.gray5,
        border: `1px solid ${vars.color.gray6}`,
      },
    },
  ],
  Failed: [
    messageStylesBase,
    {
      border: `1px solid ${vars.color.error3}`,
      background: vars.color.error1,
      ":hover": {
        background: vars.color.error2,
        border: `1px solid ${vars.color.error6}`,
      },
    },
  ],
});

export const contentStyles = style({
  display: "flex",
  gap: 10,
  alignItems: "center",
});

globalStyle(`${messageStyles} button`, {
  opacity: 0,
});

const iconStylesBase = style({
  height: 16,
});

export const iconStyles = styleVariants({
  Regular: [iconStylesBase],
  Failed: [iconStylesBase],
});

globalStyle(`${iconStyles.Regular} svg`, {
  width: 16,
  height: 16,
  color: vars.color.primary10,
});

globalStyle(`${iconStyles.Failed} svg`, {
  width: 16,
  height: 16,
  color: vars.color.error9,
});

export const outputStyles = style([
  font.textSmRegular,
  {
    padding: "10px 14px",
    borderRadius: 8,
    display: "flex",
    justifyContent: "space-between",
    alignItems: "center",
    color: vars.color.gray11,
    marginTop: -2,
    border: `1px solid ${vars.color.gray3}`,
    background: vars.color.gray2,
  },
]);
