package models

import (
	"fmt"

	"github.com/jinzhu/gorm"
)

type MsgchainType string

const (
	MsgchainTypePrimaryAgent MsgchainType = "primary_agent"
	MsgchainTypeReporter     MsgchainType = "reporter"
	MsgchainTypeGenerator    MsgchainType = "generator"
	MsgchainTypeRefiner      MsgchainType = "refiner"
	MsgchainTypeReflector    MsgchainType = "reflector"
	MsgchainTypeEnricher     MsgchainType = "enricher"
	MsgchainTypeAdviser      MsgchainType = "adviser"
	MsgchainTypeCoder        MsgchainType = "coder"
	MsgchainTypeMemorist     MsgchainType = "memorist"
	MsgchainTypeSearcher     MsgchainType = "searcher"
	MsgchainTypeInstaller    MsgchainType = "installer"
	MsgchainTypePentester    MsgchainType = "pentester"
	MsgchainTypeSummarizer   MsgchainType = "summarizer"
)

func (e *MsgchainType) Scan(src interface{}) error {
	switch s := src.(type) {
	case []byte:
		*e = MsgchainType(s)
	case string:
		*e = MsgchainType(s)
	default:
		return fmt.Errorf("unsupported scan type for MsgchainType: %T", src)
	}
	return nil
}

func (s MsgchainType) String() string {
	return string(s)
}

// Valid is function to control input/output data
func (ml MsgchainType) Valid() error {
	switch ml {
	case MsgchainTypePrimaryAgent, MsgchainTypeReporter,
		MsgchainTypeGenerator, MsgchainTypeRefiner,
		MsgchainTypeReflector, MsgchainTypeEnricher,
		MsgchainTypeAdviser, MsgchainTypeCoder,
		MsgchainTypeMemorist, MsgchainTypeSearcher,
		MsgchainTypeInstaller, MsgchainTypePentester,
		MsgchainTypeSummarizer:
		return nil
	default:
		return fmt.Errorf("invalid MsgchainType: %s", ml)
	}
}

// Validate is function to use callback to control input/output data
func (ml MsgchainType) Validate(db *gorm.DB) {
	if err := ml.Valid(); err != nil {
		db.AddError(err)
	}
}
