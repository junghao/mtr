package main

import (
	"net/http"
	"strconv"
	"time"
	"database/sql"
)

type dataLatency struct{}

func (d *dataLatency) save(r *http.Request) *result {
	if res := checkQuery(r, []string{"siteID", "typeID", "time", "mean"}, []string{"min", "max", "fifty", "ninety"}); !res.ok {
		return res
	}

	var err error

	var t time.Time
	var mean, min, max, fifty, ninety int

	if mean, err = strconv.Atoi(r.URL.Query().Get("mean")); err != nil {
		return badRequest("invalid value for mean")
	}

	if r.URL.Query().Get("min") != "" {
		if min, err = strconv.Atoi(r.URL.Query().Get("min")); err != nil {
			return badRequest("invalid value for min")
		}
	}

	if r.URL.Query().Get("max") != "" {
		if max, err = strconv.Atoi(r.URL.Query().Get("max")); err != nil {
			return badRequest("invalid value for max")
		}
	}

	if r.URL.Query().Get("fifty") != "" {
		if fifty, err = strconv.Atoi(r.URL.Query().Get("fifty")); err != nil {
			return badRequest("invalid value for fifty")
		}
	}

	if r.URL.Query().Get("ninety") != "" {
		if ninety, err = strconv.Atoi(r.URL.Query().Get("ninety")); err != nil {
			return badRequest("invalid value for ninety")
		}
	}

	if t, err = time.Parse(time.RFC3339, r.URL.Query().Get("time")); err != nil {
		return badRequest("invalid time")
	}

	var res *result
	var dt dataType

	if dt, res = loadDataType(r.URL.Query().Get("typeID")); !res.ok {
		return res
	}

	var sitePK int

	if sitePK, res = dataSitePK(r.URL.Query().Get("siteID")); !res.ok {
		return res
	}

	// Update or save the latest value.  Not rate limited.
	// TODO switch to Postgres 9.5 and use upsert.
	if _, err = db.Exec(`INSERT INTO data.latency_latest(sitePK, typePK, time, mean, min, max, fifty, ninety) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
		sitePK, dt.typePK, t, int32(mean), int32(min), int32(max), int32(fifty), int32(ninety)); err != nil {
		if _, err = db.Exec(`UPDATE data.latency_latest SET time = $3, mean = $4, min = $5, max = $6, fifty = $7, ninety = $8
				WHERE sitePK = $1
				AND typePK = $2
				AND time <= $3`,
			sitePK, dt.typePK, t, int32(mean), int32(min), int32(max), int32(fifty), int32(ninety)); err != nil {
			return internalServerError(err)
		}
	}

	// Rate limit the stored data to 1 per minute
	var count int
	if err = db.QueryRow(`SELECT count(*) FROM data.latency
				WHERE sitePK = $1
				AND typePK = $2
				AND date_trunc('minute', time) = $3`, sitePK, dt.typePK, t.Truncate(time.Minute)).Scan(&count); err != nil {
		if err != nil {
			return internalServerError(err)
		}
	}

	if count != 0 {
		return &statusTooManyRequests
	}

	if _, err = db.Exec(`INSERT INTO data.latency(sitePK, typePK, time, mean, min, max, fifty, ninety) VALUES($1, $2, $3, $4, $5, $6, $7, $8)`,
		sitePK, dt.typePK, t, int32(mean), int32(min), int32(max), int32(fifty), int32(ninety)); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}

func (d *dataLatency) delete(r *http.Request) *result {
	if res := checkQuery(r, []string{"siteID", "typeID"}, []string{}); !res.ok {
		return res
	}

	var err error

	var res *result
	var dt dataType

	if dt, res = loadDataType(r.URL.Query().Get("typeID")); !res.ok {
		return res
	}

	var sitePK int

	if sitePK, res = dataSitePK(r.URL.Query().Get("siteID")); !res.ok {
		return res
	}

	var txn *sql.Tx

	if txn, err = db.Begin(); err != nil {
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM data.latency WHERE sitePK = $1 AND typePK = $2`,
		sitePK, dt.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	if _, err = txn.Exec(`DELETE FROM data.latency_latest WHERE sitePK = $1 AND typePK = $2`,
		sitePK, dt.typePK); err != nil {
		txn.Rollback()
		return internalServerError(err)
	}

	// TODO when add latency thresholds look at this.
	//if _, err = txn.Exec(`DELETE FROM data.threshold WHERE devicePK = $1 AND typePK = $2`,
	//	f.devicePK, f.fieldType.typePK); err != nil {
	//	txn.Rollback()
	//	return internalServerError(err)
	//}

	if err = txn.Commit(); err != nil {
		return internalServerError(err)
	}

	return &statusOK
}
