import { style, styleVariants } from "@vanilla-extract/css";

import { vars } from "@/styles/theme.css";

const baseSidebarStyles = style({
  transition: "all 0.3s ease",
  backgroundColor: vars.color.gray1,
  display: "flex",
  flexDirection: "column",
  maxHeight: "calc(100dvh - 32px)",
  overflowX: "hidden",
});

export const headerStyles = style({
  display: "flex",
  alignItems: "center",
  gap: 8,
  paddingTop: "2rem",
  paddingBottom: "1rem",
  paddingLeft: "1rem",
  paddingRight: "1rem",
  borderBottom: `1px solid ${vars.color.gray3}`,
  flexShrink: 0,
});

export const scrollableContentStyles = style({
  display: "flex",
  flexDirection: "column",
  gap: 8,
  padding: "1rem",
  overflowY: "auto",
  overflowX: "hidden",
  flex: 1,
});

export const sidebarStyles = styleVariants({
  false: [baseSidebarStyles, {
    width: "300px",
    minWidth: "200px",
  }],
  true: [baseSidebarStyles, {
    width: "64px",
    minWidth: "64px",
  }],
});

export const footerStyles = style({
  borderTop: `1px solid ${vars.color.gray3}`,
  marginTop: "auto",
});

export const logoutButtonStyles = style({
  width: "100%",
  justifyContent: "center",
  gap: "8px",
  textDecoration: "none",
  background: vars.color.gray3,
  border: "none",
  color: vars.color.gray10,
  cursor: "pointer",
  padding: "9px",
  borderRadius: "0 6px 6px 0",
  textOverflow: "ellipsis",
  overflow: "hidden",
  whiteSpace: "nowrap",

  ":hover": {
    backgroundColor: vars.color.gray4,
  },
});
