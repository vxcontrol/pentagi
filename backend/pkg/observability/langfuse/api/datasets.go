// This file was auto-generated by Fern from our API Definition.

package api

import (
	json "encoding/json"
	fmt "fmt"
	internal "pentagi/pkg/observability/langfuse/api/internal"
	time "time"
)

type CreateDatasetRequest struct {
	Name        string      `json:"name" url:"-"`
	Description *string     `json:"description,omitempty" url:"-"`
	Metadata    interface{} `json:"metadata,omitempty" url:"-"`
}

type DatasetsGetRunsRequest struct {
	// page number, starts at 1
	Page *int `json:"-" url:"page,omitempty"`
	// limit of items per page
	Limit *int `json:"-" url:"limit,omitempty"`
}

type DatasetsListRequest struct {
	// page number, starts at 1
	Page *int `json:"-" url:"page,omitempty"`
	// limit of items per page
	Limit *int `json:"-" url:"limit,omitempty"`
}

type Dataset struct {
	Id          string      `json:"id" url:"id"`
	Name        string      `json:"name" url:"name"`
	Description *string     `json:"description,omitempty" url:"description,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty" url:"metadata,omitempty"`
	ProjectId   string      `json:"projectId" url:"projectId"`
	CreatedAt   time.Time   `json:"createdAt" url:"createdAt"`
	UpdatedAt   time.Time   `json:"updatedAt" url:"updatedAt"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (d *Dataset) GetId() string {
	if d == nil {
		return ""
	}
	return d.Id
}

func (d *Dataset) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name
}

func (d *Dataset) GetDescription() *string {
	if d == nil {
		return nil
	}
	return d.Description
}

func (d *Dataset) GetMetadata() interface{} {
	if d == nil {
		return nil
	}
	return d.Metadata
}

func (d *Dataset) GetProjectId() string {
	if d == nil {
		return ""
	}
	return d.ProjectId
}

func (d *Dataset) GetCreatedAt() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.CreatedAt
}

func (d *Dataset) GetUpdatedAt() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.UpdatedAt
}

func (d *Dataset) GetExtraProperties() map[string]interface{} {
	return d.extraProperties
}

func (d *Dataset) UnmarshalJSON(data []byte) error {
	type embed Dataset
	var unmarshaler = struct {
		embed
		CreatedAt *internal.DateTime `json:"createdAt"`
		UpdatedAt *internal.DateTime `json:"updatedAt"`
	}{
		embed: embed(*d),
	}
	if err := json.Unmarshal(data, &unmarshaler); err != nil {
		return err
	}
	*d = Dataset(unmarshaler.embed)
	d.CreatedAt = unmarshaler.CreatedAt.Time()
	d.UpdatedAt = unmarshaler.UpdatedAt.Time()
	extraProperties, err := internal.ExtractExtraProperties(data, *d)
	if err != nil {
		return err
	}
	d.extraProperties = extraProperties
	d.rawJSON = json.RawMessage(data)
	return nil
}

func (d *Dataset) MarshalJSON() ([]byte, error) {
	type embed Dataset
	var marshaler = struct {
		embed
		CreatedAt *internal.DateTime `json:"createdAt"`
		UpdatedAt *internal.DateTime `json:"updatedAt"`
	}{
		embed:     embed(*d),
		CreatedAt: internal.NewDateTime(d.CreatedAt),
		UpdatedAt: internal.NewDateTime(d.UpdatedAt),
	}
	return json.Marshal(marshaler)
}

func (d *Dataset) String() string {
	if len(d.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(d.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(d); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", d)
}

type DatasetRun struct {
	// Unique identifier of the dataset run
	Id string `json:"id" url:"id"`
	// Name of the dataset run
	Name string `json:"name" url:"name"`
	// Description of the run
	Description *string     `json:"description,omitempty" url:"description,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty" url:"metadata,omitempty"`
	// Id of the associated dataset
	DatasetId string `json:"datasetId" url:"datasetId"`
	// Name of the associated dataset
	DatasetName string `json:"datasetName" url:"datasetName"`
	// The date and time when the dataset run was created
	CreatedAt time.Time `json:"createdAt" url:"createdAt"`
	// The date and time when the dataset run was last updated
	UpdatedAt time.Time `json:"updatedAt" url:"updatedAt"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (d *DatasetRun) GetId() string {
	if d == nil {
		return ""
	}
	return d.Id
}

func (d *DatasetRun) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name
}

func (d *DatasetRun) GetDescription() *string {
	if d == nil {
		return nil
	}
	return d.Description
}

func (d *DatasetRun) GetMetadata() interface{} {
	if d == nil {
		return nil
	}
	return d.Metadata
}

func (d *DatasetRun) GetDatasetId() string {
	if d == nil {
		return ""
	}
	return d.DatasetId
}

func (d *DatasetRun) GetDatasetName() string {
	if d == nil {
		return ""
	}
	return d.DatasetName
}

func (d *DatasetRun) GetCreatedAt() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.CreatedAt
}

func (d *DatasetRun) GetUpdatedAt() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.UpdatedAt
}

func (d *DatasetRun) GetExtraProperties() map[string]interface{} {
	return d.extraProperties
}

func (d *DatasetRun) UnmarshalJSON(data []byte) error {
	type embed DatasetRun
	var unmarshaler = struct {
		embed
		CreatedAt *internal.DateTime `json:"createdAt"`
		UpdatedAt *internal.DateTime `json:"updatedAt"`
	}{
		embed: embed(*d),
	}
	if err := json.Unmarshal(data, &unmarshaler); err != nil {
		return err
	}
	*d = DatasetRun(unmarshaler.embed)
	d.CreatedAt = unmarshaler.CreatedAt.Time()
	d.UpdatedAt = unmarshaler.UpdatedAt.Time()
	extraProperties, err := internal.ExtractExtraProperties(data, *d)
	if err != nil {
		return err
	}
	d.extraProperties = extraProperties
	d.rawJSON = json.RawMessage(data)
	return nil
}

func (d *DatasetRun) MarshalJSON() ([]byte, error) {
	type embed DatasetRun
	var marshaler = struct {
		embed
		CreatedAt *internal.DateTime `json:"createdAt"`
		UpdatedAt *internal.DateTime `json:"updatedAt"`
	}{
		embed:     embed(*d),
		CreatedAt: internal.NewDateTime(d.CreatedAt),
		UpdatedAt: internal.NewDateTime(d.UpdatedAt),
	}
	return json.Marshal(marshaler)
}

func (d *DatasetRun) String() string {
	if len(d.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(d.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(d); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", d)
}

type DatasetRunWithItems struct {
	// Unique identifier of the dataset run
	Id string `json:"id" url:"id"`
	// Name of the dataset run
	Name string `json:"name" url:"name"`
	// Description of the run
	Description *string     `json:"description,omitempty" url:"description,omitempty"`
	Metadata    interface{} `json:"metadata,omitempty" url:"metadata,omitempty"`
	// Id of the associated dataset
	DatasetId string `json:"datasetId" url:"datasetId"`
	// Name of the associated dataset
	DatasetName string `json:"datasetName" url:"datasetName"`
	// The date and time when the dataset run was created
	CreatedAt time.Time `json:"createdAt" url:"createdAt"`
	// The date and time when the dataset run was last updated
	UpdatedAt       time.Time         `json:"updatedAt" url:"updatedAt"`
	DatasetRunItems []*DatasetRunItem `json:"datasetRunItems,omitempty" url:"datasetRunItems,omitempty"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (d *DatasetRunWithItems) GetId() string {
	if d == nil {
		return ""
	}
	return d.Id
}

func (d *DatasetRunWithItems) GetName() string {
	if d == nil {
		return ""
	}
	return d.Name
}

func (d *DatasetRunWithItems) GetDescription() *string {
	if d == nil {
		return nil
	}
	return d.Description
}

func (d *DatasetRunWithItems) GetMetadata() interface{} {
	if d == nil {
		return nil
	}
	return d.Metadata
}

func (d *DatasetRunWithItems) GetDatasetId() string {
	if d == nil {
		return ""
	}
	return d.DatasetId
}

func (d *DatasetRunWithItems) GetDatasetName() string {
	if d == nil {
		return ""
	}
	return d.DatasetName
}

func (d *DatasetRunWithItems) GetCreatedAt() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.CreatedAt
}

func (d *DatasetRunWithItems) GetUpdatedAt() time.Time {
	if d == nil {
		return time.Time{}
	}
	return d.UpdatedAt
}

func (d *DatasetRunWithItems) GetDatasetRunItems() []*DatasetRunItem {
	if d == nil {
		return nil
	}
	return d.DatasetRunItems
}

func (d *DatasetRunWithItems) GetExtraProperties() map[string]interface{} {
	return d.extraProperties
}

func (d *DatasetRunWithItems) UnmarshalJSON(data []byte) error {
	type embed DatasetRunWithItems
	var unmarshaler = struct {
		embed
		CreatedAt *internal.DateTime `json:"createdAt"`
		UpdatedAt *internal.DateTime `json:"updatedAt"`
	}{
		embed: embed(*d),
	}
	if err := json.Unmarshal(data, &unmarshaler); err != nil {
		return err
	}
	*d = DatasetRunWithItems(unmarshaler.embed)
	d.CreatedAt = unmarshaler.CreatedAt.Time()
	d.UpdatedAt = unmarshaler.UpdatedAt.Time()
	extraProperties, err := internal.ExtractExtraProperties(data, *d)
	if err != nil {
		return err
	}
	d.extraProperties = extraProperties
	d.rawJSON = json.RawMessage(data)
	return nil
}

func (d *DatasetRunWithItems) MarshalJSON() ([]byte, error) {
	type embed DatasetRunWithItems
	var marshaler = struct {
		embed
		CreatedAt *internal.DateTime `json:"createdAt"`
		UpdatedAt *internal.DateTime `json:"updatedAt"`
	}{
		embed:     embed(*d),
		CreatedAt: internal.NewDateTime(d.CreatedAt),
		UpdatedAt: internal.NewDateTime(d.UpdatedAt),
	}
	return json.Marshal(marshaler)
}

func (d *DatasetRunWithItems) String() string {
	if len(d.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(d.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(d); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", d)
}

type PaginatedDatasetRuns struct {
	Data []*DatasetRun      `json:"data,omitempty" url:"data,omitempty"`
	Meta *UtilsMetaResponse `json:"meta,omitempty" url:"meta,omitempty"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (p *PaginatedDatasetRuns) GetData() []*DatasetRun {
	if p == nil {
		return nil
	}
	return p.Data
}

func (p *PaginatedDatasetRuns) GetMeta() *UtilsMetaResponse {
	if p == nil {
		return nil
	}
	return p.Meta
}

func (p *PaginatedDatasetRuns) GetExtraProperties() map[string]interface{} {
	return p.extraProperties
}

func (p *PaginatedDatasetRuns) UnmarshalJSON(data []byte) error {
	type unmarshaler PaginatedDatasetRuns
	var value unmarshaler
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = PaginatedDatasetRuns(value)
	extraProperties, err := internal.ExtractExtraProperties(data, *p)
	if err != nil {
		return err
	}
	p.extraProperties = extraProperties
	p.rawJSON = json.RawMessage(data)
	return nil
}

func (p *PaginatedDatasetRuns) String() string {
	if len(p.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(p.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(p); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", p)
}

type PaginatedDatasets struct {
	Data []*Dataset         `json:"data,omitempty" url:"data,omitempty"`
	Meta *UtilsMetaResponse `json:"meta,omitempty" url:"meta,omitempty"`

	extraProperties map[string]interface{}
	rawJSON         json.RawMessage
}

func (p *PaginatedDatasets) GetData() []*Dataset {
	if p == nil {
		return nil
	}
	return p.Data
}

func (p *PaginatedDatasets) GetMeta() *UtilsMetaResponse {
	if p == nil {
		return nil
	}
	return p.Meta
}

func (p *PaginatedDatasets) GetExtraProperties() map[string]interface{} {
	return p.extraProperties
}

func (p *PaginatedDatasets) UnmarshalJSON(data []byte) error {
	type unmarshaler PaginatedDatasets
	var value unmarshaler
	if err := json.Unmarshal(data, &value); err != nil {
		return err
	}
	*p = PaginatedDatasets(value)
	extraProperties, err := internal.ExtractExtraProperties(data, *p)
	if err != nil {
		return err
	}
	p.extraProperties = extraProperties
	p.rawJSON = json.RawMessage(data)
	return nil
}

func (p *PaginatedDatasets) String() string {
	if len(p.rawJSON) > 0 {
		if value, err := internal.StringifyJSON(p.rawJSON); err == nil {
			return value
		}
	}
	if value, err := internal.StringifyJSON(p); err == nil {
		return value
	}
	return fmt.Sprintf("%#v", p)
}