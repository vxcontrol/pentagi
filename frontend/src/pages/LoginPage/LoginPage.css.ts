import { style } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const wrapperStyles = style({
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  minHeight: "100vh",
  padding: "16px",
  background: vars.color.gray1,
});

export const formStyles = style({
  width: "100%",
  maxWidth: "400px",
  padding: "32px",
  display: "flex",
  flexDirection: "column",
  gap: "24px",
  margin: "0 24px",
  background: vars.color.gray2,
  borderRadius: "12px",
  border: `1px solid ${vars.color.gray6}`,
});

export const titleStyles = style([
  font.textLgSemibold,
  {
    textAlign: "center",
    color: vars.color.gray12,
  },
]);

export const inputGroupStyles = style({
  display: "flex",
  flexDirection: "column",
  gap: "8px",
});

export const labelStyles = style([
  font.textSmMedium,
  {
    color: vars.color.gray11,
  },
]);

export const inputStyles = style([
  font.textSmRegular,
  {
    padding: "8px 12px",
    borderRadius: "6px",
    border: `1px solid ${vars.color.gray7}`,
    background: vars.color.gray3,
    color: vars.color.gray12,
    transition: "border-color 0.2s",

    ":focus": {
      outline: "none",
      borderColor: vars.color.primary9,
    },

    "::placeholder": {
      color: vars.color.gray9,
    },

    ":disabled": {
      opacity: 0.5,
      cursor: "not-allowed",
    },
  },
]);

export const errorMessageStyles = style([
  font.textSmMedium,
  {
    color: vars.color.red9,
    padding: "8px 12px",
    background: vars.color.red3,
    borderRadius: "6px",
  },
]);

export const buttonStyles = style({
  display: "flex",
  alignItems: "center",
  justifyContent: "center",
  width: "100%",
});

export const dividerStyles = style({
  display: 'flex',
  alignItems: 'center',
  gap: '16px',
  margin: '8px 0',
});

export const dividerLineStyles = style({
  flex: 1,
  height: '1px',
  background: vars.color.gray6,
});

export const dividerTextStyles = style([
  font.textSmMedium,
  {
    color: vars.color.gray9,
  },
]);

export const socialButtonsStyles = style({
  display: 'flex',
  flexDirection: 'column',
  gap: '12px',
});

export const socialButtonStyles = style([
  font.textSmMedium,
  {
    display: 'flex',
    alignItems: 'center',
    justifyContent: 'center',
    gap: '8px',
    width: '100%',
    padding: '8px 16px',
    borderRadius: '6px',
    border: `1px solid ${vars.color.gray7}`,
    background: vars.color.gray3,
    color: vars.color.gray12,
    cursor: 'pointer',
    transition: 'all 0.2s',

    ':hover': {
      background: vars.color.gray4,
    },

    ':disabled': {
      opacity: 0.5,
      cursor: 'not-allowed',
    },
  },
]); 