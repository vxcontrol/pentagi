import { globalStyle } from "@vanilla-extract/css";

import { vars } from "./theme.css";

globalStyle("html, body", {
  margin: 0,
  padding: 0,
  color: vars.color.gray12,
});

globalStyle("*", {
  fontFamily: "Inter var, sans-serif",
  WebkitFontSmoothing: "antialiased",
  boxSizing: "border-box",
});

// Styles for webkit browsers (Chrome, Safari, newer versions of Edge)
globalStyle('*::-webkit-scrollbar', {
  width: '8px',
  height: '8px',
});

globalStyle('*::-webkit-scrollbar-track', {
  background: 'transparent',
});

globalStyle('*::-webkit-scrollbar-thumb', {
  backgroundColor: vars.color.gray6,
  borderRadius: '4px',
});

globalStyle('*::-webkit-scrollbar-thumb:hover', {
  backgroundColor: vars.color.gray7,
});

// Styles for Firefox
globalStyle('*', {
  scrollbarWidth: 'thin',
  scrollbarColor: `${vars.color.gray6} transparent`,
});

// Disable style change on hover
globalStyle('*::-webkit-scrollbar-thumb:active', {
  backgroundColor: vars.color.gray7,
});

globalStyle('*::-webkit-scrollbar-thumb:window-inactive', {
  backgroundColor: vars.color.gray6,
});
