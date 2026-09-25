package providers

import (
	"pentagi/pkg/database"
	"pentagi/pkg/providers/pconfig"
)

var agentByChain = map[database.MsgchainType]pconfig.ProviderOptionsType{
	database.MsgchainTypeAdviser:       pconfig.OptionsTypeAdviser,
	database.MsgchainTypeAssistant:     pconfig.OptionsTypeAssistant,
	database.MsgchainTypeCoder:         pconfig.OptionsTypeCoder,
	database.MsgchainTypeEnricher:      pconfig.OptionsTypeEnricher,
	database.MsgchainTypeGenerator:     pconfig.OptionsTypeGenerator,
	database.MsgchainTypeInstaller:     pconfig.OptionsTypeInstaller,
	database.MsgchainTypeMemorist:      pconfig.OptionsTypeSearcher,
	database.MsgchainTypePentester:     pconfig.OptionsTypePentester,
	database.MsgchainTypePrimaryAgent:  pconfig.OptionsTypePrimaryAgent,
	database.MsgchainTypeRefiner:       pconfig.OptionsTypeRefiner,
	database.MsgchainTypeReflector:     pconfig.OptionsTypeReflector,
	database.MsgchainTypeReporter:      pconfig.OptionsTypeSimple,
	database.MsgchainTypeSearcher:      pconfig.OptionsTypeSearcher,
	database.MsgchainTypeSummarizer:    pconfig.OptionsTypeSimple,
	database.MsgchainTypeToolCallFixer: pconfig.OptionsTypeSimpleJSON,
}
