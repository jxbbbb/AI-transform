package data

import (
	"database/sql"
	"fmt"
	"log"
	"strings"
)

const TBL_TRANSFORM_RECORDS = "transform_records"

type TransformRecords struct {
	ID                 int64
	UserID             int64
	ProjectName        string
	OriginalLanguage   string
	TranslatedLanguage string
	OriginalVideoUrl   string
	OriginalSrtUrl     string
	TranslatedSrtUrl   string
	TranslatedVideoUrl string
	ExpirationAt       int64
	CreateAt           int64
	UpdateAt           int64
}

type ITransformRecordsData interface {
	GetByID(id int64) (*TransformRecords, error)
	GetByUserID(userID int64) ([]*TransformRecords, error)
	Add(entity *TransformRecords) error
	Update(entity *TransformRecords) error
}

type transformRecordsData struct {
	table string
	db    *sql.DB
}

func (d *transformRecordsData) GetByID(id int64) (*TransformRecords, error) {
	sqlStr := fmt.Sprintf("select id,user_id,project_name,original_language,translated_language,original_video_url,original_srt_url,translated_srt_url,translated_video_url,expiration_at,create_at,update_at from %s where id = ?", d.table)
	row := d.db.QueryRow(sqlStr, id)
	entity := &TransformRecords{}
	err := row.Scan(&entity.ID, &entity.UserID, &entity.ProjectName, &entity.OriginalLanguage, &entity.TranslatedLanguage, &entity.OriginalVideoUrl, &entity.OriginalSrtUrl, &entity.TranslatedSrtUrl, &entity.TranslatedVideoUrl, &entity.ExpirationAt, &entity.CreateAt, &entity.UpdateAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return entity, nil
}

//func (d *transformRecordsData) GetByUserID(userID int64) ([]*TransformRecords, error) {
//  sqlStr := fmt.Sprintf("select id,user_id,project_name,original_language,translated_language,original_video_url,original_srt_url,translated_srt_url,translated_video_url,expiration_at,create_at,update_at from %s where user_id = ? order by id desc", d.table)
//  rows, err := d.db.Query(sqlStr, userID)
//  if err == sql.ErrNoRows {
//    return nil, nil
//  }
//  defer rows.Close()
//  list := make([]*TransformRecords, 0, 10)
//  for rows.Next() {
//    entity := &TransformRecords{}
//    err = rows.Scan(&entity.ID, &entity.UserID, &entity.ProjectName, &entity.OriginalLanguage, &entity.TranslatedLanguage, &entity.OriginalVideoUrl, &entity.OriginalSrtUrl, &entity.TranslatedSrtUrl, &entity.TranslatedVideoUrl, &entity.ExpirationAt, &entity.CreateAt, &entity.UpdateAt)
//    if err != nil {
//      return nil, err
//    }
//    list = append(list, entity)
//  }
//  return list, err
//}

func (d *transformRecordsData) GetByUserID(userID int64) ([]*TransformRecords, error) {
	sqlStr := fmt.Sprintf(`
        select id, user_id, project_name, original_language, translated_language,
               original_video_url, original_srt_url, translated_srt_url,
               translated_video_url, expiration_at, create_at, update_at
        from %s
        where user_id = ?
        order by id desc
    `, d.table)

	rows, err := d.db.Query(sqlStr, userID)
	if err != nil {
		// 打印 SQL 和错误，方便排查
		log.Printf("[GetByUserID] query error, sql=%s, userID=%d, err=%v", sqlStr, userID, err)
		return nil, err
	}
	defer rows.Close()

	list := make([]*TransformRecords, 0, 10)
	for rows.Next() {
		entity := &TransformRecords{}
		err = rows.Scan(
			&entity.ID,
			&entity.UserID,
			&entity.ProjectName,
			&entity.OriginalLanguage,
			&entity.TranslatedLanguage,
			&entity.OriginalVideoUrl,
			&entity.OriginalSrtUrl,
			&entity.TranslatedSrtUrl,
			&entity.TranslatedVideoUrl,
			&entity.ExpirationAt,
			&entity.CreateAt,
			&entity.UpdateAt,
		)
		if err != nil {
			log.Printf("[GetByUserID] scan error, err=%v", err)
			return nil, err
		}
		list = append(list, entity)
	}

	if err = rows.Err(); err != nil {
		log.Printf("[GetByUserID] rows iteration error, err=%v", err)
		return nil, err
	}

	return list, nil
}

func (d *transformRecordsData) Add(entity *TransformRecords) error {
	sqlStr := fmt.Sprintf("insert into %s (user_id,project_name,original_language,translated_language,original_video_url,original_srt_url,translated_srt_url,translated_video_url,expiration_at,create_at,update_at)values(?,?,?,?,?,?,?,?,?,?,?)", d.table)
	res, err := d.db.Exec(sqlStr, &entity.UserID, &entity.ProjectName, &entity.OriginalLanguage, &entity.TranslatedLanguage, &entity.OriginalVideoUrl, &entity.OriginalSrtUrl, &entity.TranslatedSrtUrl, &entity.TranslatedVideoUrl, &entity.ExpirationAt, &entity.CreateAt, &entity.UpdateAt)
	if err != nil {
		return err
	}
	entity.ID, err = res.LastInsertId()
	if err != nil {
		return err
	}
	return nil
}

func (d *transformRecordsData) Update(entity *TransformRecords) error {
	updateFields := make([]string, 0, 10)
	params := make([]any, 0, 10)
	if entity.OriginalSrtUrl != "" {
		updateFields = append(updateFields, "original_srt_url = ?")
		params = append(params, entity.OriginalSrtUrl)
	}
	if entity.TranslatedSrtUrl != "" {
		updateFields = append(updateFields, "translated_srt_url = ?")
		params = append(params, entity.TranslatedSrtUrl)
	}
	if entity.TranslatedVideoUrl != "" {
		updateFields = append(updateFields, "translated_video_url = ?")
		params = append(params, entity.TranslatedVideoUrl)
	}
	if entity.ExpirationAt != 0 {
		updateFields = append(updateFields, "expiration_at = ?")
		params = append(params, entity.ExpirationAt)
	}
	if len(params) == 0 {
		return nil
	}
	updateFields = append(updateFields, "update_at = ?")
	params = append(params, entity.UpdateAt)
	params = append(params, entity.ID)

	sqlStr := fmt.Sprintf("update %s set %s where id = ?", d.table, strings.Join(updateFields, ","))
	_, err := d.db.Exec(sqlStr, params...)
	if err != nil {
		return err
	}
	return nil
}
