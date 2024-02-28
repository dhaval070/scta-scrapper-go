// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.
// Code generated by gorm.io/gen. DO NOT EDIT.

package query

import (
	"context"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/schema"

	"gorm.io/gen"
	"gorm.io/gen/field"

	"gorm.io/plugin/dbresolver"

	"calendar-scrapper/dao/model"
)

func newNyhlMapping(db *gorm.DB, opts ...gen.DOOption) nyhlMapping {
	_nyhlMapping := nyhlMapping{}

	_nyhlMapping.nyhlMappingDo.UseDB(db, opts...)
	_nyhlMapping.nyhlMappingDo.UseModel(&model.NyhlMapping{})

	tableName := _nyhlMapping.nyhlMappingDo.TableName()
	_nyhlMapping.ALL = field.NewAsterisk(tableName)
	_nyhlMapping.Location = field.NewString(tableName, "location")
	_nyhlMapping.SurfaceID = field.NewInt32(tableName, "surface_id")

	_nyhlMapping.fillFieldMap()

	return _nyhlMapping
}

type nyhlMapping struct {
	nyhlMappingDo nyhlMappingDo

	ALL       field.Asterisk
	Location  field.String
	SurfaceID field.Int32

	fieldMap map[string]field.Expr
}

func (n nyhlMapping) Table(newTableName string) *nyhlMapping {
	n.nyhlMappingDo.UseTable(newTableName)
	return n.updateTableName(newTableName)
}

func (n nyhlMapping) As(alias string) *nyhlMapping {
	n.nyhlMappingDo.DO = *(n.nyhlMappingDo.As(alias).(*gen.DO))
	return n.updateTableName(alias)
}

func (n *nyhlMapping) updateTableName(table string) *nyhlMapping {
	n.ALL = field.NewAsterisk(table)
	n.Location = field.NewString(table, "location")
	n.SurfaceID = field.NewInt32(table, "surface_id")

	n.fillFieldMap()

	return n
}

func (n *nyhlMapping) WithContext(ctx context.Context) *nyhlMappingDo {
	return n.nyhlMappingDo.WithContext(ctx)
}

func (n nyhlMapping) TableName() string { return n.nyhlMappingDo.TableName() }

func (n nyhlMapping) Alias() string { return n.nyhlMappingDo.Alias() }

func (n nyhlMapping) Columns(cols ...field.Expr) gen.Columns { return n.nyhlMappingDo.Columns(cols...) }

func (n *nyhlMapping) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := n.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (n *nyhlMapping) fillFieldMap() {
	n.fieldMap = make(map[string]field.Expr, 2)
	n.fieldMap["location"] = n.Location
	n.fieldMap["surface_id"] = n.SurfaceID
}

func (n nyhlMapping) clone(db *gorm.DB) nyhlMapping {
	n.nyhlMappingDo.ReplaceConnPool(db.Statement.ConnPool)
	return n
}

func (n nyhlMapping) replaceDB(db *gorm.DB) nyhlMapping {
	n.nyhlMappingDo.ReplaceDB(db)
	return n
}

type nyhlMappingDo struct{ gen.DO }

func (n nyhlMappingDo) Debug() *nyhlMappingDo {
	return n.withDO(n.DO.Debug())
}

func (n nyhlMappingDo) WithContext(ctx context.Context) *nyhlMappingDo {
	return n.withDO(n.DO.WithContext(ctx))
}

func (n nyhlMappingDo) ReadDB() *nyhlMappingDo {
	return n.Clauses(dbresolver.Read)
}

func (n nyhlMappingDo) WriteDB() *nyhlMappingDo {
	return n.Clauses(dbresolver.Write)
}

func (n nyhlMappingDo) Session(config *gorm.Session) *nyhlMappingDo {
	return n.withDO(n.DO.Session(config))
}

func (n nyhlMappingDo) Clauses(conds ...clause.Expression) *nyhlMappingDo {
	return n.withDO(n.DO.Clauses(conds...))
}

func (n nyhlMappingDo) Returning(value interface{}, columns ...string) *nyhlMappingDo {
	return n.withDO(n.DO.Returning(value, columns...))
}

func (n nyhlMappingDo) Not(conds ...gen.Condition) *nyhlMappingDo {
	return n.withDO(n.DO.Not(conds...))
}

func (n nyhlMappingDo) Or(conds ...gen.Condition) *nyhlMappingDo {
	return n.withDO(n.DO.Or(conds...))
}

func (n nyhlMappingDo) Select(conds ...field.Expr) *nyhlMappingDo {
	return n.withDO(n.DO.Select(conds...))
}

func (n nyhlMappingDo) Where(conds ...gen.Condition) *nyhlMappingDo {
	return n.withDO(n.DO.Where(conds...))
}

func (n nyhlMappingDo) Order(conds ...field.Expr) *nyhlMappingDo {
	return n.withDO(n.DO.Order(conds...))
}

func (n nyhlMappingDo) Distinct(cols ...field.Expr) *nyhlMappingDo {
	return n.withDO(n.DO.Distinct(cols...))
}

func (n nyhlMappingDo) Omit(cols ...field.Expr) *nyhlMappingDo {
	return n.withDO(n.DO.Omit(cols...))
}

func (n nyhlMappingDo) Join(table schema.Tabler, on ...field.Expr) *nyhlMappingDo {
	return n.withDO(n.DO.Join(table, on...))
}

func (n nyhlMappingDo) LeftJoin(table schema.Tabler, on ...field.Expr) *nyhlMappingDo {
	return n.withDO(n.DO.LeftJoin(table, on...))
}

func (n nyhlMappingDo) RightJoin(table schema.Tabler, on ...field.Expr) *nyhlMappingDo {
	return n.withDO(n.DO.RightJoin(table, on...))
}

func (n nyhlMappingDo) Group(cols ...field.Expr) *nyhlMappingDo {
	return n.withDO(n.DO.Group(cols...))
}

func (n nyhlMappingDo) Having(conds ...gen.Condition) *nyhlMappingDo {
	return n.withDO(n.DO.Having(conds...))
}

func (n nyhlMappingDo) Limit(limit int) *nyhlMappingDo {
	return n.withDO(n.DO.Limit(limit))
}

func (n nyhlMappingDo) Offset(offset int) *nyhlMappingDo {
	return n.withDO(n.DO.Offset(offset))
}

func (n nyhlMappingDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *nyhlMappingDo {
	return n.withDO(n.DO.Scopes(funcs...))
}

func (n nyhlMappingDo) Unscoped() *nyhlMappingDo {
	return n.withDO(n.DO.Unscoped())
}

func (n nyhlMappingDo) Create(values ...*model.NyhlMapping) error {
	if len(values) == 0 {
		return nil
	}
	return n.DO.Create(values)
}

func (n nyhlMappingDo) CreateInBatches(values []*model.NyhlMapping, batchSize int) error {
	return n.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (n nyhlMappingDo) Save(values ...*model.NyhlMapping) error {
	if len(values) == 0 {
		return nil
	}
	return n.DO.Save(values)
}

func (n nyhlMappingDo) First() (*model.NyhlMapping, error) {
	if result, err := n.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.NyhlMapping), nil
	}
}

func (n nyhlMappingDo) Take() (*model.NyhlMapping, error) {
	if result, err := n.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.NyhlMapping), nil
	}
}

func (n nyhlMappingDo) Last() (*model.NyhlMapping, error) {
	if result, err := n.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.NyhlMapping), nil
	}
}

func (n nyhlMappingDo) Find() ([]*model.NyhlMapping, error) {
	result, err := n.DO.Find()
	return result.([]*model.NyhlMapping), err
}

func (n nyhlMappingDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.NyhlMapping, err error) {
	buf := make([]*model.NyhlMapping, 0, batchSize)
	err = n.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (n nyhlMappingDo) FindInBatches(result *[]*model.NyhlMapping, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return n.DO.FindInBatches(result, batchSize, fc)
}

func (n nyhlMappingDo) Attrs(attrs ...field.AssignExpr) *nyhlMappingDo {
	return n.withDO(n.DO.Attrs(attrs...))
}

func (n nyhlMappingDo) Assign(attrs ...field.AssignExpr) *nyhlMappingDo {
	return n.withDO(n.DO.Assign(attrs...))
}

func (n nyhlMappingDo) Joins(fields ...field.RelationField) *nyhlMappingDo {
	for _, _f := range fields {
		n = *n.withDO(n.DO.Joins(_f))
	}
	return &n
}

func (n nyhlMappingDo) Preload(fields ...field.RelationField) *nyhlMappingDo {
	for _, _f := range fields {
		n = *n.withDO(n.DO.Preload(_f))
	}
	return &n
}

func (n nyhlMappingDo) FirstOrInit() (*model.NyhlMapping, error) {
	if result, err := n.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.NyhlMapping), nil
	}
}

func (n nyhlMappingDo) FirstOrCreate() (*model.NyhlMapping, error) {
	if result, err := n.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.NyhlMapping), nil
	}
}

func (n nyhlMappingDo) FindByPage(offset int, limit int) (result []*model.NyhlMapping, count int64, err error) {
	result, err = n.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = n.Offset(-1).Limit(-1).Count()
	return
}

func (n nyhlMappingDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = n.Count()
	if err != nil {
		return
	}

	err = n.Offset(offset).Limit(limit).Scan(result)
	return
}

func (n nyhlMappingDo) Scan(result interface{}) (err error) {
	return n.DO.Scan(result)
}

func (n nyhlMappingDo) Delete(models ...*model.NyhlMapping) (result gen.ResultInfo, err error) {
	return n.DO.Delete(models)
}

func (n *nyhlMappingDo) withDO(do gen.Dao) *nyhlMappingDo {
	n.DO = *do.(*gen.DO)
	return n
}