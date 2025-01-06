import { style, styleVariants } from "@vanilla-extract/css";
import { vars } from "@/styles/theme.css";

const baseTaskTreeStyles = style({
  transition: "max-height 0.3s ease-in-out, opacity 0.2s ease-in-out",
  overflow: "hidden",
  backgroundColor: vars.color.gray2,
  padding: "16px",
});

export const taskTreeStyles = styleVariants({
  expanded: [
    baseTaskTreeStyles,
    {
      maxHeight: "35vh",
      opacity: 1,
      overflowY: "auto",
    },
  ],
  collapsed: [
    baseTaskTreeStyles,
    {
      maxHeight: "0",
      opacity: 0,
      padding: 0,
    },
  ],
}); 