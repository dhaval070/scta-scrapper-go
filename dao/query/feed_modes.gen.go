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

func newFeedMode(db *gorm.DB, opts ...gen.DOOption) feedMode {
	_feedMode := feedMode{}

	_feedMode.feedModeDo.UseDB(db, opts...)
	_feedMode.feedModeDo.UseModel(&model.FeedMode{})

	tableName := _feedMode.feedModeDo.TableName()
	_feedMode.ALL = field.NewAsterisk(tableName)
	_feedMode.ID = field.NewInt32(tableName, "id")
	_feedMode.FeedMode = field.NewString(tableName, "feed_mode")

	_feedMode.fillFieldMap()

	return _feedMode
}

type feedMode struct {
	feedModeDo feedModeDo

	ALL      field.Asterisk
	ID       field.Int32
	FeedMode field.String

	fieldMap map[string]field.Expr
}

func (f feedMode) Table(newTableName string) *feedMode {
	f.feedModeDo.UseTable(newTableName)
	return f.updateTableName(newTableName)
}

func (f feedMode) As(alias string) *feedMode {
	f.feedModeDo.DO = *(f.feedModeDo.As(alias).(*gen.DO))
	return f.updateTableName(alias)
}

func (f *feedMode) updateTableName(table string) *feedMode {
	f.ALL = field.NewAsterisk(table)
	f.ID = field.NewInt32(table, "id")
	f.FeedMode = field.NewString(table, "feed_mode")

	f.fillFieldMap()

	return f
}

func (f *feedMode) WithContext(ctx context.Context) *feedModeDo { return f.feedModeDo.WithContext(ctx) }

func (f feedMode) TableName() string { return f.feedModeDo.TableName() }

func (f feedMode) Alias() string { return f.feedModeDo.Alias() }

func (f feedMode) Columns(cols ...field.Expr) gen.Columns { return f.feedModeDo.Columns(cols...) }

func (f *feedMode) GetFieldByName(fieldName string) (field.OrderExpr, bool) {
	_f, ok := f.fieldMap[fieldName]
	if !ok || _f == nil {
		return nil, false
	}
	_oe, ok := _f.(field.OrderExpr)
	return _oe, ok
}

func (f *feedMode) fillFieldMap() {
	f.fieldMap = make(map[string]field.Expr, 2)
	f.fieldMap["id"] = f.ID
	f.fieldMap["feed_mode"] = f.FeedMode
}

func (f feedMode) clone(db *gorm.DB) feedMode {
	f.feedModeDo.ReplaceConnPool(db.Statement.ConnPool)
	return f
}

func (f feedMode) replaceDB(db *gorm.DB) feedMode {
	f.feedModeDo.ReplaceDB(db)
	return f
}

type feedModeDo struct{ gen.DO }

func (f feedModeDo) Debug() *feedModeDo {
	return f.withDO(f.DO.Debug())
}

func (f feedModeDo) WithContext(ctx context.Context) *feedModeDo {
	return f.withDO(f.DO.WithContext(ctx))
}

func (f feedModeDo) ReadDB() *feedModeDo {
	return f.Clauses(dbresolver.Read)
}

func (f feedModeDo) WriteDB() *feedModeDo {
	return f.Clauses(dbresolver.Write)
}

func (f feedModeDo) Session(config *gorm.Session) *feedModeDo {
	return f.withDO(f.DO.Session(config))
}

func (f feedModeDo) Clauses(conds ...clause.Expression) *feedModeDo {
	return f.withDO(f.DO.Clauses(conds...))
}

func (f feedModeDo) Returning(value interface{}, columns ...string) *feedModeDo {
	return f.withDO(f.DO.Returning(value, columns...))
}

func (f feedModeDo) Not(conds ...gen.Condition) *feedModeDo {
	return f.withDO(f.DO.Not(conds...))
}

func (f feedModeDo) Or(conds ...gen.Condition) *feedModeDo {
	return f.withDO(f.DO.Or(conds...))
}

func (f feedModeDo) Select(conds ...field.Expr) *feedModeDo {
	return f.withDO(f.DO.Select(conds...))
}

func (f feedModeDo) Where(conds ...gen.Condition) *feedModeDo {
	return f.withDO(f.DO.Where(conds...))
}

func (f feedModeDo) Order(conds ...field.Expr) *feedModeDo {
	return f.withDO(f.DO.Order(conds...))
}

func (f feedModeDo) Distinct(cols ...field.Expr) *feedModeDo {
	return f.withDO(f.DO.Distinct(cols...))
}

func (f feedModeDo) Omit(cols ...field.Expr) *feedModeDo {
	return f.withDO(f.DO.Omit(cols...))
}

func (f feedModeDo) Join(table schema.Tabler, on ...field.Expr) *feedModeDo {
	return f.withDO(f.DO.Join(table, on...))
}

func (f feedModeDo) LeftJoin(table schema.Tabler, on ...field.Expr) *feedModeDo {
	return f.withDO(f.DO.LeftJoin(table, on...))
}

func (f feedModeDo) RightJoin(table schema.Tabler, on ...field.Expr) *feedModeDo {
	return f.withDO(f.DO.RightJoin(table, on...))
}

func (f feedModeDo) Group(cols ...field.Expr) *feedModeDo {
	return f.withDO(f.DO.Group(cols...))
}

func (f feedModeDo) Having(conds ...gen.Condition) *feedModeDo {
	return f.withDO(f.DO.Having(conds...))
}

func (f feedModeDo) Limit(limit int) *feedModeDo {
	return f.withDO(f.DO.Limit(limit))
}

func (f feedModeDo) Offset(offset int) *feedModeDo {
	return f.withDO(f.DO.Offset(offset))
}

func (f feedModeDo) Scopes(funcs ...func(gen.Dao) gen.Dao) *feedModeDo {
	return f.withDO(f.DO.Scopes(funcs...))
}

func (f feedModeDo) Unscoped() *feedModeDo {
	return f.withDO(f.DO.Unscoped())
}

func (f feedModeDo) Create(values ...*model.FeedMode) error {
	if len(values) == 0 {
		return nil
	}
	return f.DO.Create(values)
}

func (f feedModeDo) CreateInBatches(values []*model.FeedMode, batchSize int) error {
	return f.DO.CreateInBatches(values, batchSize)
}

// Save : !!! underlying implementation is different with GORM
// The method is equivalent to executing the statement: db.Clauses(clause.OnConflict{UpdateAll: true}).Create(values)
func (f feedModeDo) Save(values ...*model.FeedMode) error {
	if len(values) == 0 {
		return nil
	}
	return f.DO.Save(values)
}

func (f feedModeDo) First() (*model.FeedMode, error) {
	if result, err := f.DO.First(); err != nil {
		return nil, err
	} else {
		return result.(*model.FeedMode), nil
	}
}

func (f feedModeDo) Take() (*model.FeedMode, error) {
	if result, err := f.DO.Take(); err != nil {
		return nil, err
	} else {
		return result.(*model.FeedMode), nil
	}
}

func (f feedModeDo) Last() (*model.FeedMode, error) {
	if result, err := f.DO.Last(); err != nil {
		return nil, err
	} else {
		return result.(*model.FeedMode), nil
	}
}

func (f feedModeDo) Find() ([]*model.FeedMode, error) {
	result, err := f.DO.Find()
	return result.([]*model.FeedMode), err
}

func (f feedModeDo) FindInBatch(batchSize int, fc func(tx gen.Dao, batch int) error) (results []*model.FeedMode, err error) {
	buf := make([]*model.FeedMode, 0, batchSize)
	err = f.DO.FindInBatches(&buf, batchSize, func(tx gen.Dao, batch int) error {
		defer func() { results = append(results, buf...) }()
		return fc(tx, batch)
	})
	return results, err
}

func (f feedModeDo) FindInBatches(result *[]*model.FeedMode, batchSize int, fc func(tx gen.Dao, batch int) error) error {
	return f.DO.FindInBatches(result, batchSize, fc)
}

func (f feedModeDo) Attrs(attrs ...field.AssignExpr) *feedModeDo {
	return f.withDO(f.DO.Attrs(attrs...))
}

func (f feedModeDo) Assign(attrs ...field.AssignExpr) *feedModeDo {
	return f.withDO(f.DO.Assign(attrs...))
}

func (f feedModeDo) Joins(fields ...field.RelationField) *feedModeDo {
	for _, _f := range fields {
		f = *f.withDO(f.DO.Joins(_f))
	}
	return &f
}

func (f feedModeDo) Preload(fields ...field.RelationField) *feedModeDo {
	for _, _f := range fields {
		f = *f.withDO(f.DO.Preload(_f))
	}
	return &f
}

func (f feedModeDo) FirstOrInit() (*model.FeedMode, error) {
	if result, err := f.DO.FirstOrInit(); err != nil {
		return nil, err
	} else {
		return result.(*model.FeedMode), nil
	}
}

func (f feedModeDo) FirstOrCreate() (*model.FeedMode, error) {
	if result, err := f.DO.FirstOrCreate(); err != nil {
		return nil, err
	} else {
		return result.(*model.FeedMode), nil
	}
}

func (f feedModeDo) FindByPage(offset int, limit int) (result []*model.FeedMode, count int64, err error) {
	result, err = f.Offset(offset).Limit(limit).Find()
	if err != nil {
		return
	}

	if size := len(result); 0 < limit && 0 < size && size < limit {
		count = int64(size + offset)
		return
	}

	count, err = f.Offset(-1).Limit(-1).Count()
	return
}

func (f feedModeDo) ScanByPage(result interface{}, offset int, limit int) (count int64, err error) {
	count, err = f.Count()
	if err != nil {
		return
	}

	err = f.Offset(offset).Limit(limit).Scan(result)
	return
}

func (f feedModeDo) Scan(result interface{}) (err error) {
	return f.DO.Scan(result)
}

func (f feedModeDo) Delete(models ...*model.FeedMode) (result gen.ResultInfo, err error) {
	return f.DO.Delete(models)
}

func (f *feedModeDo) withDO(do gen.Dao) *feedModeDo {
	f.DO = *do.(*gen.DO)
	return f
}