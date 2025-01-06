import { style } from "@vanilla-extract/css";
import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

export const wrapperStyles = style({
  backgroundColor: vars.color.gray2,
  borderRadius: 8,
  border: `1px solid ${vars.color.gray3}`,
  overflow: "hidden",
  display: "flex",
  flexDirection: "column",
  position: "relative",
});

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
    flexShrink: 0,
  },
]);

export const logListStyles = style({
  flex: 1,
  overflow: "auto",
  padding: "16px",
  display: "flex",
  flexDirection: "column",
  gap: "16px",
  maxHeight: "calc(100vh - 150px)",
});

export const logItemStyles = style({
  padding: "12px",
  backgroundColor: vars.color.gray3,
  borderRadius: "8px",
  display: "flex",
  flexDirection: "column",
  gap: "8px",
});

export const agentIconsStyles = style({
  display: "flex",
  alignItems: "center",
  gap: "8px",
  color: vars.color.gray11,
});

export const arrowIconStyles = style({
  color: vars.color.gray9,
  width: "16px",
  height: "16px",
});

export const queryStyles = style([
  font.textSmMedium,
  {
    color: vars.color.gray12,
    padding: "8px",
    backgroundColor: vars.color.gray4,
    borderRadius: "4px",
  },
]);

export const resultStyles = style([
  font.textSmRegular,
  {
    color: vars.color.gray11,
    padding: "8px",
    backgroundColor: vars.color.gray4,
    borderRadius: "4px",
  },
]);

export const metaInfoStyles = style([
  font.textXsRegular,
  {
    color: vars.color.gray9,
  },
]);

export const emptyStateStyles = style([
  font.textMdMedium,
  {
    display: "flex",
    alignItems: "center",
    justifyContent: "center",
    height: "100%",
    color: vars.color.gray9,
    padding: "32px",
  },
]);

export const logHeaderStyles = style({
  display: "flex",
  alignItems: "center",
  justifyContent: "space-between",
  cursor: "pointer",
  padding: "8px",
  borderRadius: "4px",
  backgroundColor: vars.color.gray4,
  transition: "background-color 0.2s ease",
  ":hover": {
    backgroundColor: vars.color.gray5,
  },
});

export const collapsedContentStyles = style({
  display: "none",
});

export const expandedContentStyles = style({
  display: "flex",
  flexDirection: "column",
  gap: "8px",
  marginTop: "8px",
}); 