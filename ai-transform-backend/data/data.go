package data

import "database/sql"

type IData interface {
	NewTransformRecordsData() ITransformRecordsData
}
type data struct {
	db *sql.DB
}

func NewData(db *sql.DB) IData {
	return &data{
		db: db,
	}
}
func (d *data) NewTransformRecordsData() ITransformRecordsData {
	return &transformRecordsData{
		table: TBL_TRANSFORM_RECORDS,
		db:    d.db,
	}
}
