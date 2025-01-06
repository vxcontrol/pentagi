import { globalFontFace, style } from "@vanilla-extract/css";

const inter = "Inter var";

globalFontFace(inter, {
  src: 'url("/Inter-roman.var.woff2")',
  fontWeight: "100 900",
  fontDisplay: "swap",
  fontStyle: "normal",
});

const displayMd = style({
  fontSize: "36px",
  lineHeight: "44px",
  letterSpacing: "-0.02em",
});

const displayXs = style({
  fontSize: "24px",
  lineHeight: "32px",
});

const textMd = style({
  fontSize: "16px",
  lineHeight: "24px",
});

const textSm = style({
  fontSize: "14px",
  lineHeight: "20px",
});

const textXs = style({
  fontSize: "12px",
  lineHeight: "18px",
});

const textLg = style({
  fontSize: "18px",
  lineHeight: "28px",
});

// TODO figure out if it's possible to combine these
// fonts styles in a less manual way
const displayMdSemiBold = style([
  displayMd,
  {
    fontWeight: 600,
  },
]);

const displayXsSemibold = style([
  displayXs,
  {
    fontWeight: 600,
  },
]);

const textXsRegular = style([
  textXs,
  {
    fontWeight: 400,
  },
]);

const textXsMedium = style([
  textXs,
  {
    fontWeight: 500,
  },
]);

const textXsSemibold = style([
  textXs,
  {
    fontWeight: 600,
  },
]);

const textMdRegular = style([
  textMd,
  {
    fontWeight: 400,
  },
]);

const textMdMedium = style([
  textMd,
  {
    fontWeight: 500,
  },
]);

const textMdSemibold = style([
  textMd,
  {
    fontWeight: 600,
  },
]);

const textSmRegular = style([
  textSm,
  {
    fontWeight: 400,
  },
]);

const textSmMedium = style([
  textSm,
  {
    fontWeight: 500,
  },
]);

const textSmSemibold = style([
  textSm,
  {
    fontWeight: 600,
  },
]);

const textLgSemibold = style([
  textLg,
  {
    fontWeight: 600,
  },
]);

const textLgRegular = style([
  textLg,
  {
    fontWeight: 400,
  },
]);

const textLgMedium = style([
  textLg,
  {
    fontWeight: 500,
  },
]);

export const font = {
  displayMdSemiBold,
  displayXsSemibold,
  textLgMedium,
  textLgRegular,
  textLgSemibold,
  textMdMedium,
  textMdRegular,
  textMdSemibold,
  textSmMedium,
  textSmRegular,
  textSmSemibold,
  textXsMedium,
  textXsRegular,
  textXsSemibold,
};
