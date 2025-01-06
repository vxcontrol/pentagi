import { style, globalStyle } from "@vanilla-extract/css";

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
    flexShrink: 0,
  },
]);

export const wrapperStyles = style({
  backgroundColor: vars.color.gray2,
  borderRadius: 8,
  border: `1px solid ${vars.color.gray3}`,
  overflow: "hidden",
  // maxHeight: "calc(100vh - 200px)",
  display: "flex",
  flexDirection: "column",
  position: "relative",
});

export const galleryWrapperStyles = style({
  flex: 1,
  overflow: "auto",
  padding: "16px",
  position: "relative",
  display: "flex",
  flexDirection: "column",
  marginBottom: "16px",
  maxHeight: "calc(100vh - 150px)",
});

export const galleryContainerStyles = style({
  flex: 1,
  minHeight: 0,
  overflow: "auto",
  padding: "4px",
});

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

// Global styles to disable focus on gallery elements
globalStyle('[role="grid"]', {
  outline: 'none !important',
  border: 'none !important',
});

globalStyle('[role="grid"] *', {
  outline: 'none !important',
  border: 'none !important',
});

globalStyle('.ReactGridGallery', {
  outline: 'none !important',
  border: 'none !important',
});

globalStyle('.ReactGridGallery_tile', {
  outline: 'none !important',
  border: 'none !important',
});

globalStyle('.ReactGridGallery_tile-viewport', {
  outline: 'none !important',
  border: 'none !important',
});

// Styles to disable focus on images
globalStyle('img', {
  outline: 'none !important',
  WebkitTapHighlightColor: 'transparent',
});

// Disable outline for all interactive elements inside the gallery
globalStyle('.ReactGridGallery [tabindex]', {
  outline: 'none !important',
  border: 'none !important',
});
