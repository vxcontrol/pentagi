import { style } from "@vanilla-extract/css";

import { vars } from "@/styles/theme.css";

export const wrapperStyles = style({
  backgroundColor: vars.color.gray2,
  border: `1px solid ${vars.color.gray4}`,
  borderRadius: 6,
  padding: 16,
});
