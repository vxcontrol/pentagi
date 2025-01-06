import * as DropdownMenu from "@radix-ui/react-dropdown-menu";
import { useEffect } from "react";

import {
  Dropdown,
  dropdownMenuContentStyles,
  dropdownMenuItemStyles,
} from "@/components/Dropdown/Dropdown";
import { dropdownMenuItemIconStyles } from "@/components/Dropdown/Dropdown.css";
import { Icon } from "@/components/Icon/Icon";
import { ProvidersQuery } from "@/generated/graphql";

import { buttonStyles } from "./ProviderSelector.css";

type ProviderSelectorProps = {
  availableProviders: ProvidersQuery["providers"];
  selectedProvider?: string;
  activeProvider?: string;
  setSelectedProvider: (provider: string) => void;
};

export const ProviderSelector = ({
  availableProviders = [],
  selectedProvider,
  activeProvider,
  setSelectedProvider,
}: ProviderSelectorProps) => {
  // Automatically select the first available providers
  useEffect(() => {
    if (!activeProvider && availableProviders[0]) {
      setSelectedProvider(availableProviders[0]);
    }
  }, [activeProvider, availableProviders, setSelectedProvider]);

  const handleValueChange = (value: string) => {
    const newProvider = availableProviders.find(
      (provider) => provider === value,
    );

    if (!newProvider) return;

    setSelectedProvider(newProvider);
  };

  const dropdownContent = (
    <DropdownMenu.Content className={dropdownMenuContentStyles} sideOffset={5}>
      <DropdownMenu.RadioGroup
        value={selectedProvider}
        onValueChange={handleValueChange}
      >
        {availableProviders.length > 0 ? (
          availableProviders.map((provider) => (
            <DropdownMenu.RadioItem
              key={provider}
              className={dropdownMenuItemStyles}
              value={provider}
            >
              <DropdownMenu.ItemIndicator
                className={dropdownMenuItemIconStyles}
              >
                <Icon.Check />
              </DropdownMenu.ItemIndicator>
              {provider}
            </DropdownMenu.RadioItem>
          ))
        ) : (
          <DropdownMenu.Item disabled className={dropdownMenuItemStyles}>
            No available providers
          </DropdownMenu.Item>
        )}
      </DropdownMenu.RadioGroup>
    </DropdownMenu.Content>
  );

  return (
    <Dropdown content={dropdownContent}>
      <div className={buttonStyles}>{activeProvider || "No provider"}</div>
    </Dropdown>
  );
};
