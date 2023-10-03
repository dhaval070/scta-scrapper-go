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

func newZone(db *gorm.DB, opts ...gen.DOOption) zone {
	_zone := zone{}

	_zone.zoneDo.UseDB(db, opts...)
	_zone.zoneDo.UseModel(&model.Zone{})

	tableName := _zone.zoneDo.TableName()
	_zone.ALL = field.NewAsterisk(tableName)
	_zone.ID = field.NewInt32(tableName, "id")
	_zone.ZoneName = field.NewString(tableName, "zone_name")

	_zone.fillFieldMap()

	return _zone
}

type zone struct {
	zoneDo zoneDo

	ALL      field.Asterisk
	ID       field.Int32
	ZoneName field.String

	fieldMap map[string]field.Expr
}

func (z zone) Table(newTableName string) *zone {
	z.zoneDo.UseTable(newTableName)
	return z.updateTableName(newTableName)
}

func (z zone) As(alias string) *zone {
	z.zoneDo.DO = *(z.zoneDo.As(alias).(*gen.DO))
	return z.updateTableName(alias)
}

func (z *zone) updateTableName(table string) *zone {
	z.ALL = field.NewAsterisk(table)
	z.ID = field.NewInt32(table, "id")
	z.ZoneName = field.NewString(table, "zone_name")

	z.fillFieldMap()

	return z
}

func (z *zone) WithContext(ctx context.Context) *zoneDo { return z.zoneDo.WithContext(ctx) }

func (z zone) TableName() string { return z.zoneDo.TableName() }

func (z zone) Alias() string { return z.zoneDo.Alias() }

func (z zone) Columns(cols ...field.Expr) gen.Columns { return z.zoneDo.Columns(cols...) }

func (z *zone) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := z.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (z *zone) fillFieldMap() {
	z.fieldMap = make(map[string]field.Expr, 2)
	z.fieldMap["id"] = z.ID
	z.fieldMap["zone_name"] = z.ZoneName
}

func (z zone) clone(db *gorm.DB) zone {
	z.zoneDo.ReplaceConnPool(db.Statement.ConnPool)
	return z
}

func (z zone) replaceDB(db *gorm.DB) zone {
	z.zoneDo.ReplaceDB(db)
	return z
}

type zoneDo struct{ gen.DO }

func (z zoneDo) Debug() *zoneDo {
	return z.withDO(z.DO.Debug())
}

func (z zoneDo) WithContext(ctx context.Context) *zoneDo {
	return z.withDO(z.DO.WithContext(ctx))
}

func (z zoneDo) ReadDB() *zoneDo {
	return z.Clauses(dbresolver.Read)
}

func (z zoneDo) WriteDB() *zoneDo {
	return z.Clauses(dbresolver.Write)
}

func (z zoneDo) Session(config *gorm.Session) *zoneDo {
	return z.withDO(z.DO.Session(config))
}

func (z zoneDo) Clauses(conds ...clause.Expression) *zoneDo {
	return z.withDO(z.DO.Clauses(conds...))
}

func (z zoneDo) Returning(value interface{}, columns ...string) *zoneDo {
	return z.withDO(z.DO.Returning(value, columns...))
}

func (z zoneDo) Not(conds ...gen.Condition) *zoneDo {
	return z.withDO(z.DO.Not(conds...))
}

func (z zoneDo) Or(conds ...gen.Condition) *zoneDo {
	return z.withDO(z.DO.Or(conds...))
}

func (z zoneDo) Select(conds ...field.Expr) *zoneDo {
	return z.withDO(z.DO.Select(conds...))
}

func (z zoneDo) Where(conds ...gen.Condition) *zoneDo {
	return z.withDO(z.DO.Where(conds...))
}

func (z zoneDo) Order(conds ...field.Expr) *zoneDo {
	return z.withDO(z.DO.Order(conds...))
}

func (z zoneDo) Distinct(cols ...field.Expr) *zoneDo {
	return z.withDO(z.DO.Distinct(cols...))
}

func (z zoneDo) Omit(cols ...field.Expr) *zoneDo {
	return z.withDO(z.DO.Omit(cols...))
}

func (z zoneDo) Join(table schema.Tabler, on ...field.Expr) *zoneDo {
	return z.withDO(z.DO.Join(table, on...))
}

func (z zoneDo) LeftJoin(table schema.Tabler, on ...field.Expr) *zoneDo {
	return z.withDO(z.DO.LeftJoin(table, on...))
}

func (z zoneDo) RightJoin(table schema.Tabler, on ...field.Expr) *zoneDo {
	return z.withDO(z.DO.RightJoin(table, on...))
}

func (z zoneDo) Group(cols ...field.Expr) *zoneDo {
	return z.withDO(z.DO.Group(cols...))
}

func (z zoneDo) Having(conds ...gen.Condition) *zoneDo {
	return z.withDO(z.DO.Having(conds...))
}

func (z zoneDo) Limit(limit int) *zoneDo {
	return z.withDO(z.DO.Limit(limit))
}

func (z zoneDo) Offset(offset int) *zoneDo {
	return z.withDO(z.DO.Offset(offset))
}

func (z zoneDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *zoneDo {
	return z.withDO(z.DO.Scopes(funcs...))
}

func (z zoneDo) Unscoped() *zoneDo {
	return z.withDO(z.DO.Unscoped())
}

func (z zoneDo) Create(values ...*model.Zone) error {
	if len(values) == 0 {
		return nil
	}
	return z.DO.Create(values)
}

func (z zoneDo) CreateInBatches(values []*model.Zone, batchSize int) error {
	return z.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (z zoneDo) Save(values ...*model.Zone) error {
	if len(values) == 0 {
		return nil
	}
	return z.DO.Save(values)
}

func (z zoneDo) First() (*model.Zone, error) {
	if result, err := z.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.Zone), nil
	}
}

func (z zoneDo) Take() (*model.Zone, error) {
	if result, err := z.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.Zone), nil
	}
}

func (z zoneDo) Last() (*model.Zone, error) {
	if result, err := z.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.Zone), nil
	}
}

func (z zoneDo) Find() ([]*model.Zone, error) {
	result, err := z.DO.Find()
	return result.([]*model.Zone), err
}

func (z zoneDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.Zone, err error) {
	buf := make([]*model.Zone, 0, batchSize)
	err = z.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (z zoneDo) FindInBatches(result *[]*model.Zone, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return z.DO.FindInBatches(result, batchSize, fc)
}

func (z zoneDo) Attrs(attrs ...field.AssignExpr) *zoneDo {
	return z.withDO(z.DO.Attrs(attrs...))
}

func (z zoneDo) Assign(attrs ...field.AssignExpr) *zoneDo {
	return z.withDO(z.DO.Assign(attrs...))
}

func (z zoneDo) Joins(fields ...field.RelationField) *zoneDo {
	for _, _f := range fields {
		z = *z.withDO(z.DO.Joins(_f))
	}
	return &z
}

func (z zoneDo) Preload(fields ...field.RelationField) *zoneDo {
	for _, _f := range fields {
		z = *z.withDO(z.DO.Preload(_f))
	}
	return &z
}

func (z zoneDo) FirstOrInit() (*model.Zone, error) {
	if result, err := z.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.Zone), nil
	}
}

func (z zoneDo) FirstOrCreate() (*model.Zone, error) {
	if result, err := z.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.Zone), nil
	}
}

func (z zoneDo) FindByPage(offset int, limit int) (result []*model.Zone, count int64, err error) {
	result, err = z.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = z.Offset(-1).Limit(-1).Count()
	return
}

func (z zoneDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = z.Count()
	if err != nil {
		return
	}

	err = z.Offset(offset).Limit(limit).Scan(result)
	return
}

func (z zoneDo) Scan(result interface{}) (err error) {
	return z.DO.Scan(result)
}

func (z zoneDo) Delete(models ...*model.Zone) (result gen.ResultInfo, err error) {
	return z.DO.Delete(models)
}

func (z *zoneDo) withDO(do gen.Dao) *zoneDo {
	z.DO = *do.(*gen.DO)
	return z
}