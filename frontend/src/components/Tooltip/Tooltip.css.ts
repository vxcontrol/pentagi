import { globalStyle, keyframes, style } from "@vanilla-extract/css";

import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

const slideUpAndFade = keyframes({
  "0%": { opacity: 0, transform: "translateY(2px)" },
  "100%": { opacity: 1, transform: "translateY(0)" },
});

const slideRightAndFade = keyframes({
  "0%": { opacity: 0, transform: "translateX(-2px)" },
  "100%": { opacity: 1, transform: "translateX(0)" },
});

const slideDownAndFade = keyframes({
  "0%": { opacity: 0, transform: "translateY(-2px)" },
  "100%": { opacity: 1, transform: "translateY(0)" },
});

const slideLeftAndFade = keyframes({
  "0%": { opacity: 0, transform: "translateX(2px)" },
  "100%": { opacity: 1, transform: "translateX(0)" },
});

export const contentStyles = style([
  font.textXsRegular,
  {
    borderRadius: 6,
    padding: "8px 12px",
    color: vars.color.gray12,
    backgroundColor: vars.color.gray3,
    border: `1px solid ${vars.color.gray5}`,
    boxShadow: vars.shadow.sm,
    userSelect: "none",
    animationDuration: "400ms",
    animationTimingFunction: "cubic-bezier(0.16, 1, 0.3, 1)",
    willChange: "transform, opacity",
  },
]);

export const arrowStyles = style({
  fill: vars.color.gray5,
});

globalStyle(`${contentStyles}[data-state="delayed-open"][data-side="top"]`, {
  animationName: slideUpAndFade,
});

globalStyle(`${contentStyles}[data-state="delayed-open"][data-side="right"]`, {
  animationName: slideRightAndFade,
});

globalStyle(`${contentStyles}[data-state="delayed-open"][data-side="bottom"]`, {
  animationName: slideDownAndFade,
});

globalStyle(`${contentStyles}[data-state="delayed-open"][data-side="left"]`, {
  animationName: slideLeftAndFade,
});
