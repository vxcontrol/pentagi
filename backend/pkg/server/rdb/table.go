package rdb

import (
	"crypto/md5" //nolint:gosec
	"encoding/hex"
	"errors"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// TableFilter is auxiliary struct to contain method of filtering
//
//nolint:lll
type TableFilter struct {
	Value    any    `form:"value" json:"value,omitempty" binding:"required" swaggertype:"object"`
	Field    string `form:"field" json:"field,omitempty" binding:"required"`
	Operator string `form:"operator" json:"operator,omitempty" binding:"oneof='<' '<=' '>=' '>' '=' '!=' 'like' 'not like' 'in',omitempty" default:"like" enums:"<,<=,>=,>,=,!=,like,not like,in"`
}

// TableSort is auxiliary struct to contain method of sorting
type TableSort struct {
	Prop  string `form:"prop" json:"prop,omitempty" binding:"omitempty"`
	Order string `form:"order" json:"order,omitempty" binding:"oneof=ascending descending,required_with=Prop,omitempty" enums:"ascending,descending"`
}

// TableQuery is main struct to contain input params
//
//nolint:lll
type TableQuery struct {
	// Number of page (since 1)
	Page int `form:"page" json:"page" binding:"min=1,required" default:"1" minimum:"1"`
	// Amount items per page (min -1, max 1000, -1 means unlimited)
	Size int `form:"pageSize" json:"pageSize" binding:"min=-1,max=1000" default:"5" minimum:"-1" maximum:"1000"`
	// Type of request
	Type string `form:"type" json:"type" binding:"oneof=sort filter init page size,required" default:"init" enums:"sort,filter,init,page,size"`
	// Sorting result on server e.g. {"prop":"...","order":"..."}
	//   field order is "ascending" or "descending" value
	//   order is required if prop is not empty
	Sort []TableSort `form:"sort[]" json:"sort[],omitempty" binding:"omitempty,dive" swaggertype:"array,string"`
	// Filtering result on server e.g. {"value":[...],"field":"...","operator":"..."}
	//   field is the unique identifier of the table column, different for each endpoint
	//   value should be integer or string or array type, "value":123 or "value":"string" or "value":[123,456]
	//   operator value should be one of <,<=,>=,>,=,!=,like,not like,in
	//   default operator value is 'like' or '=' if field is 'id' or '*_id' or '*_at'
	Filters []TableFilter `form:"filters[]" json:"filters[],omitempty" binding:"omitempty,dive" swaggertype:"array,string"`
	// Field to group results by
	Group string `form:"group" json:"group,omitempty" binding:"omitempty" swaggertype:"string"`
	// non input arguments
	table      string                                `form:"-" json:"-"`
	groupField string                                `form:"-" json:"-"`
	sqlMappers map[string]any                        `form:"-" json:"-"`
	sqlFind    func(out any) func(*gorm.DB) *gorm.DB `form:"-" json:"-"`
	sqlFilters []func(*gorm.DB) *gorm.DB             `form:"-" json:"-"`
	sqlOrders  []func(*gorm.DB) *gorm.DB             `form:"-" json:"-"`
}

// Init is function to set table name and sql mapping to data columns
func (q *TableQuery) Init(table string, sqlMappers map[string]any) error {
	q.table = table
	q.sqlFind = func(out any) func(db *gorm.DB) *gorm.DB {
		return func(db *gorm.DB) *gorm.DB {
			return db.Find(out)
		}
	}
	q.sqlMappers = make(map[string]any)
	q.sqlOrders = append(q.sqlOrders, func(db *gorm.DB) *gorm.DB {
		return db.Order("id DESC")
	})
	for k, v := range sqlMappers {
		switch t := v.(type) {
		case string:
			t = q.DoConditionFormat(t)
			if isNumbericField(k) {
				q.sqlMappers[k] = t
			} else {
				q.sqlMappers[k] = "LOWER(" + t + "::text)"
			}
		case func(q *TableQuery, db *gorm.DB, value any) *gorm.DB:
			q.sqlMappers[k] = t
		default:
			continue
		}
	}
	if q.Group != "" {
		var ok bool
		q.groupField, ok = q.sqlMappers[q.Group].(string)
		if !ok {
			return errors.New("wrong field for grouping")
		}
	}
	return nil
}

// DoConditionFormat is auxiliary function to prepare condition to the table
func (q *TableQuery) DoConditionFormat(cond string) string {
	cond = strings.ReplaceAll(cond, "{{type}}", q.Type)
	cond = strings.ReplaceAll(cond, "{{table}}", q.table)
	cond = strings.ReplaceAll(cond, "{{page}}", strconv.Itoa(q.Page))
	cond = strings.ReplaceAll(cond, "{{size}}", strconv.Itoa(q.Size))
	return cond
}

// SetFilters is function to set custom filters to build target SQL query
func (q *TableQuery) SetFilters(sqlFilters []func(*gorm.DB) *gorm.DB) {
	q.sqlFilters = sqlFilters
}

// SetFind is function to set custom find function to build target SQL query
func (q *TableQuery) SetFind(find func(out any) func(*gorm.DB) *gorm.DB) {
	q.sqlFind = find
}

// SetOrders is function to set custom ordering to build target SQL query
func (q *TableQuery) SetOrders(sqlOrders []func(*gorm.DB) *gorm.DB) {
	q.sqlOrders = sqlOrders
}

// Mappers is getter for private field (SQL find funcction to use it in custom query)
func (q *TableQuery) Find(out any) func(*gorm.DB) *gorm.DB {
	return q.sqlFind(out)
}

// Mappers is getter for private field (SQL mappers fields to table ones)
func (q *TableQuery) Mappers() map[string]any {
	return q.sqlMappers
}

// Table is getter for private field (table name)
func (q *TableQuery) Table() string {
	return q.table
}

// Ordering is function to get order of data rows according with input params
func (q *TableQuery) Ordering() func(db *gorm.DB) *gorm.DB {
	var sortItems []TableSort

	for _, sort := range q.Sort {
		var t TableSort

		switch sort.Order {
		case "ascending":
			t.Order = "ASC"
		case "descending":
			t.Order = "DESC"
		}

		if v, ok := q.sqlMappers[sort.Prop]; ok {
			if s, ok := v.(string); ok {
				t.Prop = s
			}
		}

		if t.Prop != "" && t.Order != "" {
			sortItems = append(sortItems, t)
		}
	}

	return func(db *gorm.DB) *gorm.DB {
		for _, sort := range sortItems {
			// sort.Prop comes from server-side whitelist (q.sqlMappers)
			// sort.Order is validated to be only "ASC" or "DESC"
			db = db.Order(sort.Prop + " " + sort.Order)
		}
		for _, order := range q.sqlOrders {
			db = order(db)
		}
		return db
	}
}

// Paginate is function to navigate between pages according with input params
func (q *TableQuery) Paginate() func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if q.Page <= 0 && q.Size >= 0 {
			return db.Limit(q.Size)
		} else if q.Page > 0 && q.Size >= 0 {
			offset := (q.Page - 1) * q.Size
			return db.Offset(offset).Limit(q.Size)
		}
		return db
	}
}

// GroupBy is function to group results by some field
func (q *TableQuery) GroupBy(total *uint64, result any) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		return db.Group(q.groupField).Where(q.groupField+" IS NOT NULL").Count(total).Pluck(q.groupField, result)
	}
}

// DataFilter is function to build main data filter from filters input params
func (q *TableQuery) DataFilter() func(db *gorm.DB) *gorm.DB {
	type item struct {
		op string
		v  any
	}

	fl := make(map[string][]item)
	setFilter := func(field, operator string, value any) {
		if operator == "" {
			operator = "like" // nolint:goconst
		}
		fvalue := []item{}
		if fv, ok := fl[field]; ok {
			fvalue = fv
		}
		switch tvalue := value.(type) {
		case string, float64, bool:
			fl[field] = append(fvalue, item{operator, tvalue})
		case []any:
			fl[field] = append(fvalue, item{operator, tvalue})
		}
	}
	patchOperator := func(f *TableFilter) {
		switch f.Operator {
		case "<", "<=", ">=", ">", "=", "!=", "in":
		case "not like":
			if isNumbericField(f.Field) {
				f.Operator = "!="
			}
		default:
			f.Operator = "like"
			if isNumbericField(f.Field) {
				f.Operator = "="
			}
		}
	}

	for _, f := range q.Filters {
		f := f

		patchOperator(&f)
		if _, ok := q.sqlMappers[f.Field]; ok {
			if v, ok := f.Value.(string); ok && v != "" {
				vs := v
				if slices.Contains([]string{"like", "not like"}, f.Operator) {
					vs = "%" + strings.ToLower(vs) + "%"
				}
				setFilter(f.Field, f.Operator, vs)
			}
			if v, ok := f.Value.(float64); ok {
				setFilter(f.Field, f.Operator, v)
			}
			if v, ok := f.Value.(bool); ok {
				setFilter(f.Field, f.Operator, v)
			}
			if v, ok := f.Value.([]any); ok && len(v) != 0 {
				var vi []any
				for _, ti := range v {
					if ts, ok := ti.(string); ok {
						vi = append(vi, strings.ToLower(ts))
					}
					if ts, ok := ti.(float64); ok {
						vi = append(vi, ts)
					}
					if ts, ok := ti.(bool); ok {
						vi = append(vi, ts)
					}
				}
				if len(vi) != 0 {
					setFilter(f.Field, "in", vi)
				}
			}
		}
	}

	return func(db *gorm.DB) *gorm.DB {
		doFilter := func(db *gorm.DB, k, s string, v any) *gorm.DB {
			switch t := q.sqlMappers[k].(type) {
			case string:
				return db.Where(t+s, v)
			case func(q *TableQuery, db *gorm.DB, value any) *gorm.DB:
				return t(q, db, v)
			default:
				return db
			}
		}
		for k, f := range fl {
			for _, it := range f {
				if _, ok := it.v.([]any); ok {
					db = doFilter(db, k, " "+it.op+" (?)", it.v)
				} else {
					db = doFilter(db, k, " "+it.op+" ?", it.v)
				}
			}
		}
		for _, filter := range q.sqlFilters {
			db = filter(db)
		}
		return db
	}
}

// Query is function to retrieve table data according with input params
func (q *TableQuery) Query(db *gorm.DB, result any,
	funcs ...func(*gorm.DB) *gorm.DB) (uint64, error) {
	var total uint64
	err := ApplyToChainDB(
		ApplyToChainDB(db.Table(q.Table()), funcs...).Scopes(q.DataFilter()).Count(&total),
		q.Ordering(),
		q.Paginate(),
		q.Find(result),
	).Error
	return uint64(total), err
}

// QueryGrouped is function to retrieve grouped data according with input params
func (q *TableQuery) QueryGrouped(db *gorm.DB, result any,
	funcs ...func(*gorm.DB) *gorm.DB) (uint64, error) {
	if _, ok := q.sqlMappers[q.Group]; !ok {
		return 0, errors.New("group field not found")
	}

	var total uint64
	err := ApplyToChainDB(
		ApplyToChainDB(db.Table(q.Table()), funcs...).Scopes(q.DataFilter()),
		q.GroupBy(&total, result),
	).Error
	return uint64(total), err
}

// ApplyToChainDB is function to extend gorm method chaining by custom functions
func ApplyToChainDB(db *gorm.DB, funcs ...func(*gorm.DB) *gorm.DB) (tx *gorm.DB) {
	for _, f := range funcs {
		db = f(db)
	}
	return db
}

// EncryptPassword is function to prepare user data as a password
func EncryptPassword(password string) (hpass []byte, err error) {
	hpass, err = bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return
}

// MakeMD5Hash is function to generate common hash by value
func MakeMD5Hash(value, salt string) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05.000000000")
	hash := md5.Sum([]byte(currentTime + value + salt)) // nolint:gosec
	return hex.EncodeToString(hash[:])
}

// MakeUserHash is function to generate user hash from name
func MakeUserHash(name string) string {
	currentTime := time.Now().Format("2006-01-02 15:04:05.000000000")
	return MakeMD5Hash(name+currentTime, "248a8bd896595be1319e65c308a903c568afdb9b")
}

// MakeUuidStrFromHash is function to convert format view from hash to UUID
func MakeUuidStrFromHash(hash string) (string, error) {
	hashBytes, err := hex.DecodeString(hash)
	if err != nil {
		return "", err
	}
	userIdUuid, err := uuid.FromBytes(hashBytes)
	if err != nil {
		return "", err
	}
	return userIdUuid.String(), nil
}

func isNumbericField(field string) bool {
	return strings.HasSuffix(field, "_id") || strings.HasSuffix(field, "_at") || field == "id"
}
