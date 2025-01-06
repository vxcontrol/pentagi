import { globalStyle, style } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const headerStyles = style([
  font.textXsSemibold,
  {
    backgroundColor: vars.color.gray6,
    color: vars.color.gray11,
    padding: "8px 12px",
    borderRadius: "8px 8px 0 0",
    display: "flex",
    alignItems: "center",
    gap: 8,
  },
]);

globalStyle(`${headerStyles} svg`, {
  width: 14,
  height: 14,
});
