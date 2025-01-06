import { style } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const wrapperStyles = style([
  font.textSmSemibold,
  {
    borderRadius: 6,

    ":hover": {
      color: vars.color.gray12,
      backgroundColor: vars.color.gray2,
    },
  },
]);

export const linkStyles = style({
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  color: vars.color.gray11,
  textDecoration: "none",
  padding: "9px 16px",

  selectors: {
    "&.active": {
      color: vars.color.gray12,
      backgroundColor: vars.color.gray2,
    },
  },
});

export const checkIconStyles = style({
  color: vars.color.success9,
});
