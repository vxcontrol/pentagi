import { useState } from "react";
import { Outlet } from "react-router-dom";

import { Sidebar } from "@/components/Sidebar/Sidebar";
import { useFlowsQuery, useProvidersQuery } from "@/generated/graphql";

import { wrapperStyles } from "./AppLayout.css";

export const AppLayout = () => {
  const [{ data }] = useFlowsQuery();
  const [{ data: providersData }] = useProvidersQuery();
  const [isSidebarCollapsed, setIsSidebarCollapsed] = useState(false);

  const sidebarItems =
    data?.flows?.map((flow) => ({
      id: flow.id,
      title: flow.title,
      done: flow.status === "finished",
    })) ?? [];

  return (
    <div className={wrapperStyles}>
      <Sidebar
        items={sidebarItems}
        availableProviders={providersData?.providers ?? []}
        isCollapsed={isSidebarCollapsed}
        onToggleCollapse={() => setIsSidebarCollapsed(prev => !prev)}
      />
      <Outlet />
    </div>
  );
};
