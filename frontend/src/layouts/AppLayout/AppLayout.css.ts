import { style, globalStyle } from "@vanilla-extract/css";

import { vars } from "@/styles/theme.css";

// Block scrolling at body level
globalStyle('body', {
  overflow: 'hidden',
  margin: 0,
  height: '100vh',
});

export const wrapperStyles = style({
  display: "flex",
  height: "100vh",
  maxHeight: "100vh",
  overflow: "hidden",
  transition: "all 0.3s ease",
  backgroundColor: vars.color.gray1,
  position: "fixed",
  top: 0,
  left: 0,
  right: 0,
  bottom: 0,
});
