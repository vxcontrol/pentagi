import { style } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const buttonStyles = style([
  font.textSmRegular,
  {
    display: "block",
    textDecoration: "none",
    background: vars.color.gray3,
    border: "none",
    textAlign: "left",
    color: vars.color.gray10,
    padding: "9px 16px",
    cursor: "pointer",
    borderRadius: "0 6px 6px 0",
    flex: 1,
    width: "100px",
    textOverflow: "ellipsis",
    overflow: "hidden",
    whiteSpace: "nowrap",

    ":hover": {
      backgroundColor: vars.color.gray4,
    },
  },
]);
