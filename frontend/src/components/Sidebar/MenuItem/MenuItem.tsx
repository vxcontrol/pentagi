import { NavLink } from "react-router-dom";

import { Icon } from "@/components/Icon/Icon";

import { checkIconStyles, linkStyles, wrapperStyles } from "./MenuItem.css";

export type MenuItemProps = {
  title: string;
  id: string;
  done?: boolean;
};

export const MenuItem = ({ title, id, done }: MenuItemProps) => {
  return (
    <div className={wrapperStyles}>
      <NavLink to={`/chat/${id}`} className={linkStyles}>
        {title}
        <span>{done && <Icon.Check className={checkIconStyles} />}</span>
      </NavLink>
    </div>
  );
};
