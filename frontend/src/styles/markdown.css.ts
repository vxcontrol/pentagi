import { style } from "@vanilla-extract/css";
import { font } from "@/styles/font.css";
import { vars } from "@/styles/theme.css";

const markdownBaseHeaderStyle = style([
  font.textSmRegular,
  {
    margin: "0.5em 0",
    color: vars.color.gray12,
  },
]);

export const markdownH1Style = style([
  markdownBaseHeaderStyle,
  font.textMdSemibold,
]);

export const markdownH2Style = style([
  markdownBaseHeaderStyle,
  font.textSmSemibold,
]);

export const markdownH3Style = style([
  markdownBaseHeaderStyle,
  font.textSmSemibold,
]);

export const markdownH4Style = style([
  markdownBaseHeaderStyle,
  font.textSmSemibold,
]);

export const markdownH5Style = style([
  markdownBaseHeaderStyle,
  font.textSmSemibold,
]);

export const markdownH6Style = style([
  markdownBaseHeaderStyle,
  font.textSmSemibold,
]);

export const markdownHeaderStyles = {
  h1: markdownH1Style,
  h2: markdownH2Style,
  h3: markdownH3Style,
  h4: markdownH4Style,
  h5: markdownH5Style,
  h6: markdownH6Style,
  link: style({
    color: vars.color.primary9,
    textDecoration: "none",
    ":hover": {
      textDecoration: "underline",
    },
  }),
};
