import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import React from "react";

import {
  dropdownMenuContentStyles,
  dropdownMenuItemStyles,
  triggerStyles,
} from "./Dropdown.css";

type DropdownProps = {
  children: React.ReactNode;
  content: React.ReactNode;
} & React.ComponentProps<typeof DropdownMenu.Root>;

export const Dropdown = ({ children, content, ...rest }: DropdownProps) => {
  return (
    <DropdownMenu.Root {...rest}>
      <DropdownMenu.Trigger asChild>
        <button className={triggerStyles} aria-label="Customise options">
          {children}
        </button>
      </DropdownMenu.Trigger>
      <DropdownMenu.Portal>{content}</DropdownMenu.Portal>
    </DropdownMenu.Root>
  );
};

export { dropdownMenuItemStyles, dropdownMenuContentStyles };
