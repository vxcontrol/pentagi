import { ProvidersQuery } from "@/generated/graphql";
import { MenuItem, MenuItemProps } from "./MenuItem/MenuItem";
import { NewFlow } from "./NewFlow/NewFlow";
import { logout } from '@/api/auth';
import { Button } from '@/components/Button/Button';
import { Icon } from '@/components/Icon/Icon';
import { 
  sidebarStyles, 
  headerStyles, 
  scrollableContentStyles,
  footerStyles,
  logoutButtonStyles 
} from "./Sidebar.css";

interface SidebarProps {
  items: Array<MenuItemProps>;
  availableProviders: ProvidersQuery["providers"];
  isCollapsed: boolean;
  onToggleCollapse: () => void;
}

export const Sidebar = ({ items, availableProviders, isCollapsed, onToggleCollapse }: SidebarProps) => {
  const handleLogout = () => {
    logout('/login');
  };

  return (
    <div className={sidebarStyles[isCollapsed ? "true" : "false"]}>
      <div className={headerStyles}>
        <NewFlow 
          availableProviders={availableProviders}
          isCollapsed={isCollapsed}
          onToggleCollapse={onToggleCollapse}
        />
      </div>
      <div className={scrollableContentStyles} style={{ visibility: isCollapsed ? "hidden" : "visible" }}>
        {items.map((item) => (
          <MenuItem key={item.id} {...item} />
        ))}
      </div>
      <div className={footerStyles}>
        <Button
          className={logoutButtonStyles}
          onClick={handleLogout}
          hierarchy="secondary"
          size="small"
        >
          <Icon.Logout /> {!isCollapsed && 'Logout'}
        </Button>
      </div>
    </div>
  );
};
