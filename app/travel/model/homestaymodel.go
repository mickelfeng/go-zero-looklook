package model

import (
	"database/sql"
	"fmt"
	"math"
	"strings"
	"time"

	"looklook/common/globalkey"

	sqlBuilder "github.com/didi/gendry/builder"
	"github.com/tal-tech/go-zero/core/stores/builder"
	"github.com/tal-tech/go-zero/core/stores/cache"
	"github.com/tal-tech/go-zero/core/stores/sqlc"
	"github.com/tal-tech/go-zero/core/stores/sqlx"
	"github.com/tal-tech/go-zero/core/stringx"
)

var (
	homestayFieldNames          = builder.RawFieldNames(&Homestay{})
	homestayRows                = strings.Join(homestayFieldNames, ",")
	homestayRowsWithPlaceHolder = strings.Join(stringx.Remove(homestayFieldNames, "`id`", "`create_time`", "`update_time`", "`version`"), "=?,") + "=?"

	cacheLooklookTravelHomestayIdPrefix = "cache:looklookTravel:homestay:id:"
)

type (
	HomestayModel interface {
		FindPageListByHomestayBusinessId(lastId, pageSize, homestayBusinessId int64) ([]*Homestay, error)
		FindPageList(lastId, pageSize int64) ([]*Homestay, error)
		FindOne(id int64) (*Homestay, error)
		Insert(session sqlx.Session, data *Homestay) (sql.Result, error)
		Update(session sqlx.Session, data *Homestay) error
		Delete(session sqlx.Session, data *Homestay) error
		Trans(fn func(session sqlx.Session) error) error
	}

	defaultHomestayModel struct {
		sqlc.CachedConn
		table string
	}

	Homestay struct {
		Id                  int64     `db:"id"`
		CreateTime          time.Time `db:"create_time"`
		UpdateTime          time.Time `db:"update_time"`
		DeleteTime          time.Time `db:"delete_time"`
		DelState            int64     `db:"del_state"`
		Title               string    `db:"title"`                 // 标题
		SubTitle            string    `db:"sub_title"`             // 副标题
		Banner              string    `db:"banner"`                //轮播图，第一张封面`
		Info                string    `db:"info"`                  // 介绍
		PeopleNum           int64     `db:"people_num"`            // 容纳人的数量
		HomestayBusinessId  int64     `db:"homestay_business_id"`  // 民宿店铺id
		UserId              int64     `db:"user_id"`               // 房东id，冗余字段
		RowState            int64     `db:"row_state"`             // 0:下架 1:上架
		RowType             int64     `db:"row_type"`              // 售卖类型0：按房间出售 1:按人次出售
		FoodInfo            string    `db:"food_info"`             // 餐食标准
		FoodPrice           int64     `db:"food_price"`            // 餐食价格(分)
		HomestayPrice       int64     `db:"homestay_price"`        // 民宿价格(分)
		MarketHomestayPrice int64     `db:"market_homestay_price"` // 民宿市场价格(分)
	}
)

func NewHomestayModel(conn sqlx.SqlConn, c cache.CacheConf) HomestayModel {
	return &defaultHomestayModel{
		CachedConn: sqlc.NewConn(conn, c),
		table:      "`homestay`",
	}
}

func (m *defaultHomestayModel) FindPageListByHomestayBusinessId(lastId, pageSize, homestayBusinessId int64) ([]*Homestay, error) {

	if lastId == 0 {
		lastId = math.MaxInt64
	}

	where := map[string]interface{}{
		"`homestay_business_id`": homestayBusinessId,
		"`del_state`":            globalkey.DelStateNo,
		"`id` <":                 lastId,
		"_orderby":               "id DESC",
		"_limit":                 []uint{0, uint(pageSize)},
	}
	query, values, err := sqlBuilder.BuildSelect(m.table, where, homestayFieldNames)
	if err != nil {
		return nil, err
	}

	var resp []*Homestay
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

func (m *defaultHomestayModel) FindPageList(lastId, pageSize int64) ([]*Homestay, error) {

	if lastId == 0 {
		lastId = math.MaxInt64
	}

	where := map[string]interface{}{
		"`del_state`": globalkey.DelStateNo,
		"`id` < ":     lastId,
		"_orderby":    "id DESC",
		"_limit":      []uint{0, uint(pageSize)},
	}
	query, values, err := sqlBuilder.BuildSelect(m.table, where, homestayFieldNames)
	if err != nil {
		return nil, err
	}

	var resp []*Homestay
	err = m.QueryRowsNoCache(&resp, query, values...)
	switch err {
	case nil:
		return resp, nil
	default:
		return nil, err
	}
}

func (m *defaultHomestayModel) Insert(session sqlx.Session, data *Homestay) (sql.Result, error) {

	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	query := fmt.Sprintf("insert into .... (%s) values ...", m.table)
	if session != nil {

		//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
		return session.Exec(query, data.DeleteTime, data.DelState, data.Title, data.SubTitle, data.Info,
			data.PeopleNum, data.HomestayBusinessId, data.UserId, data.RowState, data.RowType,
			data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice)
	}
	//@todo self edit  value , because change table field is trouble in here , so self fix field is easy
	return m.ExecNoCache(query, data.DeleteTime, data.DelState, data.Title, data.SubTitle, data.Info,
		data.PeopleNum, data.HomestayBusinessId, data.UserId, data.RowState, data.RowType,
		data.FoodInfo, data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice)

}

func (m *defaultHomestayModel) FindOne(id int64) (*Homestay, error) {

	looklookTravelHomestayIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayIdPrefix, id)
	var resp Homestay
	err := m.QueryRow(&resp, looklookTravelHomestayIdKey, func(conn sqlx.SqlConn, v interface{}) error {
		query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayRows, m.table)
		return conn.QueryRow(v, query, id)
	})
	switch err {
	case nil:
		if resp.DelState == globalkey.DelStateYes {
			return nil, ErrNotFound
		}
		return &resp, nil
	case sqlc.ErrNotFound:
		return nil, ErrNotFound
	default:
		return nil, err
	}
}

func (m *defaultHomestayModel) Update(session sqlx.Session, data *Homestay) error {
	looklookTravelHomestayIdKey := fmt.Sprintf("%s%v", cacheLooklookTravelHomestayIdPrefix, data.Id)
	_, err := m.Exec(func(conn sqlx.SqlConn) (result sql.Result, err error) {
		query := fmt.Sprintf("update %s set %s where `id` = ?", m.table, homestayRowsWithPlaceHolder)

		if session != nil {
			return session.Exec(query, data.DeleteTime, data.DelState, data.Title, data.SubTitle, data.Banner, data.Info,
				data.PeopleNum, data.HomestayBusinessId, data.UserId, data.RowState, data.RowType, data.FoodInfo,
				data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice, data.Id)
		}
		return conn.Exec(query, data.DeleteTime, data.DelState, data.Title, data.SubTitle, data.Banner, data.Info,
			data.PeopleNum, data.HomestayBusinessId, data.UserId, data.RowState, data.RowType, data.FoodInfo,
			data.FoodPrice, data.HomestayPrice, data.MarketHomestayPrice, data.Id)
	}, looklookTravelHomestayIdKey)
	return err
}

func (m *defaultHomestayModel) Delete(session sqlx.Session, data *Homestay) error {
	data.DelState = globalkey.DelStateYes
	return m.Update(session, data)
}

func (m *defaultHomestayModel) Trans(fn func(session sqlx.Session) error) error {

	err := m.Transact(func(session sqlx.Session) error {
		return fn(session)
	})
	return err

}

func (m *defaultHomestayModel) formatPrimary(primary interface{}) string {
	return fmt.Sprintf("%s%v", cacheLooklookTravelHomestayIdPrefix, primary)
}

func (m *defaultHomestayModel) queryPrimary(conn sqlx.SqlConn, v, primary interface{}) error {
	query := fmt.Sprintf("select %s from %s where `id` = ? limit 1", homestayRows, m.table)
	return conn.QueryRow(v, query, primary)
}
