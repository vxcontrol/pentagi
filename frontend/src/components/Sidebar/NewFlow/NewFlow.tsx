import { useLocalStorage } from "@uidotdev/usehooks";
import { useNavigate } from "react-router-dom";

import { Tooltip } from "@/components/Tooltip/Tooltip";
import { ProvidersQuery } from "@/generated/graphql";

import { ProviderSelector } from "./ProviderSelector/ProviderSelector";
import { collapseButtonStyles, linkWrapperStyles, wrapperStyles } from "./NewFlow.css";

import { Icon } from "@/components/Icon/Icon";

type NewFlowProps = {
  availableProviders: ProvidersQuery["providers"];
  isCollapsed?: boolean;
  onToggleCollapse?: () => void;
};

export const NewFlow = ({ 
  availableProviders, 
  isCollapsed = false, 
  onToggleCollapse 
}: NewFlowProps) => {
  const navigate = useNavigate();
  const [selectedProvider, setSelectedProvider] = useLocalStorage<
    string | undefined
  >("provider");
  const activeProvider = availableProviders.find(
    (provider) => provider == selectedProvider,
  );

  const handleNewFlow = () => {
    navigate("/chat/new");
  };

  const tooltipContent = activeProvider
    ? "Create a new flow"
    : "Please select a provider first";

  return (
    <div className={wrapperStyles}>
      <button onClick={onToggleCollapse} className={collapseButtonStyles}>
        {isCollapsed ? <Icon.ArrowRight /> : <Icon.ArrowLeft />}
      </button>
      {!isCollapsed && (
        <>
          <Tooltip content={tooltipContent}>
            <button
              className={linkWrapperStyles}
              onClick={handleNewFlow}
              disabled={!activeProvider}
            >
              ✨
            </button>
          </Tooltip>
          <ProviderSelector
            availableProviders={availableProviders}
            selectedProvider={selectedProvider}
            activeProvider={activeProvider}
            setSelectedProvider={setSelectedProvider}
          />
        </>
      )}
    </div>
  );
};
