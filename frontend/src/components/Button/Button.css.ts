import { globalStyle, styleVariants } from "@vanilla-extract/css";
import { style } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const baseStyles = style([
  font.textSmSemibold,
  {
    display: "flex",
    borderRadius: 8,
    cursor: "pointer",
    border: "1px solid transparent",
    transition: "background 0.15s",
    alignItems: "center",
  },
]);

export const buttonStyles = styleVariants({
  Primary: [
    baseStyles,
    {
      color: vars.color.primary1,
      backgroundColor: vars.color.primary9,
      ":hover": {
        backgroundColor: vars.color.primary10,
      },
      ":disabled": {
        backgroundColor: vars.color.primary3,
        color: vars.color.primary8,
        cursor: "not-allowed",
      },
    },
  ],
  Secondary: [
    baseStyles,
    {
      color: vars.color.gray12,
      backgroundColor: vars.color.gray3,
      border: `1px solid ${vars.color.gray7}`,
      boxShadow: vars.shadow.xs,
      ":hover": {
        backgroundColor: vars.color.gray4,
      },
      ":disabled": {
        border: `1px solid ${vars.color.gray5}`,
        color: vars.color.gray8,
        cursor: "not-allowed",
      },
    },
  ],
  Danger: [
    baseStyles,
    {
      color: vars.color.error9,
      backgroundColor: vars.color.error2,
      ":hover": {
        backgroundColor: vars.color.error3,
      },
      ":disabled": {
        backgroundColor: vars.color.error5,
        cursor: "not-allowed",
      },
    },
  ],
});

export const buttonSizesStyles = styleVariants({
  Small: {
    padding: "4px 8px",
  },
  Medium: {
    padding: "8px 14px",
  },
});

export const buttonIconStyles = style({
  display: "flex",
  marginRight: 8,
});

globalStyle(`${buttonIconStyles} svg`, {
  // TODO make it different for each size
  width: 16,
  height: 16,
});
