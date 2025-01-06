import { style, styleVariants } from "@vanilla-extract/css";
import { vars } from "@/styles/theme.css";
import { font } from "@/styles/font.css";
import { StatusType } from "@/generated/graphql";

export const taskItemStyles = {
  wrapper: style({
    marginBottom: "8px",
  }),
  
  status: styleVariants({
    [StatusType.Starting]: {
      color: vars.color.gray9,
    },
    [StatusType.Running]: {
      color: vars.color.primary9,
    },
    [StatusType.Finished]: {
      color: vars.color.success9,
    },
    [StatusType.Failed]: {
      color: vars.color.error9,
    },
    [StatusType.Waiting]: {
      color: vars.color.warning9,
    },
  }),
  
  description: style([
    font.textSmRegular,
    {
      marginTop: "4px",
      marginLeft: "24px",
      color: vars.color.gray11,
      padding: "8px",
      backgroundColor: vars.color.gray3,
      borderRadius: "4px",
    },
  ]),
};

export const taskTitleStyles = style([
  font.textSmSemibold,
  {
    display: "flex",
    alignItems: "center",
    gap: "8px",
    cursor: "pointer",
    padding: "4px",
    borderRadius: "4px",
    
    ":hover": {
      backgroundColor: vars.color.gray3,
    },
  },
]);

export const subtaskListStyles = style({
  marginLeft: "24px",
  marginTop: "8px",
});

export const preWrapper = style({
  whiteSpace: 'pre-wrap',
  wordWrap: 'break-word',
  maxWidth: '100%'
}); 