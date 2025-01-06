import { globalStyle, style } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const triggerStyles = style({
  all: "unset",
  borderRadius: 6,

  selectors: {
    '&[data-state="open"]': {
      backgroundColor: vars.color.gray3,
    },
  },
});

export const dropdownMenuContentStyles = style({
  minWidth: 220,
  backgroundColor: vars.color.gray3,
  border: `1px solid ${vars.color.gray4}`,
  borderRadius: 6,
  padding: 3,
  boxShadow: `0 0 10px 2px #12121187`,
});

export const dropdownMenuSubContentStyles = dropdownMenuContentStyles;

export const dropdownMenuItemStyles = style([
  font.textSmMedium,
  {
    display: "flex",
    borderRadius: 3,
    alignItems: "center",
    height: 32,
    padding: "0 3px",
    position: "relative",
    paddingLeft: 32,
    userSelect: "none",
    outline: "none",
    color: vars.color.gray12,
    cursor: "pointer",

    selectors: {
      "&[data-highlighted]": {
        backgroundColor: vars.color.gray4,
      },
    },
  },
]);

export const dropdownMenuItemIconStyles = style({
  position: "absolute",
  left: 8,
  top: 8,
  color: vars.color.gray9,
  width: 16,
  height: 16,
});

globalStyle(`${dropdownMenuItemStyles}:hover ${dropdownMenuItemIconStyles}`, {
  color: vars.color.primary9,
});

export const dropdownMenuSubTriggerStyles = dropdownMenuItemStyles;

export const dropdownMenuSeparatorStyles = style({
  height: 1,
  backgroundColor: vars.color.gray4,
  margin: 5,
});

export const dropdownMenuRightSlotStyles = style({
  display: "flex",
  marginLeft: "auto",
  paddingLeft: 20,
  top: 4,
  color: vars.color.gray9,
});
