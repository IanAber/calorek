package main

import (
	"database/sql"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
	"time"
)

const DataBySecond = `select UNIX_TIMESTAMP(logged) as logged
                       ,(dischargePressure) as dischargePressure
                       ,(suctionPressure) as suctionPressure
					   ,(sourceOutTemp) as sourceOutTemp
					   ,(sourceInTemp) as sourceInTemp
     				   ,(loadOutTemp) as loadOutTemp
					   ,(loadInTemp) as loadInTemp
					   ,(compressorSpeed) as compressorSpeed
					   ,(eev_pos) as eevPos
					   ,(demand) as demand
                   from heatpump.values
                  where logged between ? and ?`

const DataByMinute = `select min(UNIX_TIMESTAMP(logged)) as logged
                       ,AVG(dischargePressure) as dischargePressure
                       ,AVG(suctionPressure) as suctionPressure
					   ,AVG(sourceOutTemp) as sourceOutTemp
					   ,AVG(sourceInTemp) as sourceInTemp
     				   ,AVG(loadOutTemp) as loadOutTemp
					   ,AVG(loadInTemp) as loadInTemp
					   ,AVG(compressorSpeed) as compressorSpeed
					   ,AVG(eev_pos) as eevPos
					   ,AVG(demand) as demand
                   from heatpump.values
                  where logged between ? and ?
	              group by UNIX_TIMESTAMP(logged) div 60`

/*
*
GetTimeRange returns the start and end times passed as query parameters.
*/
func GetTimeRange(ctx *gin.Context) (start time.Time, end time.Time, err error) {
	startParam := ctx.Query("start")
	if len(startParam) == 0 {
		err = fmt.Errorf("Exactly one 'start=' value must be supplied for start time")
		return
	}
	timeVal, err := time.Parse("2006-1-2 15:4", startParam)
	if err != nil {
		return
	} else {
		start = timeVal
	}

	endParam := ctx.Query("end")
	if len(endParam) == 0 {
		err = fmt.Errorf("Exactly one 'start=' value must be supplied for start time")
		return
	}
	timeVal, err = time.Parse("2006-1-2 15:4", endParam)
	if err != nil {
		return
	} else {
		end = timeVal
	}
	if end.Before(start) {
		err = fmt.Errorf("End Time (%s) must be after Start Time (%s)", end.String(), start.String())
		return
	}
	return
}

//func getDatabaseRowsAsJSON(pdb *sql.DB, qry string, args ...any) ([]interface{}, error) {
//	if pDB == nil {
//		return nil, fmt.Errorf("the database is not connected")
//	}
//	if Rows, err := pdb.Query(qry, args...); err != nil {
//		return nil, err
//	} else {
//		result := make([]interface{}, 0)
//		for Rows.Next() {
//			result = append(result, Rows.Scan())
//		}
//		return result, err
//	}
//}

type row map[string]interface{}

func Jsonify(rows *sql.Rows) []row {
	columns, err := rows.Columns()
	if err != nil {
		panic(err.Error())
	}

	values := make([]interface{}, len(columns))

	scanArgs := make([]interface{}, len(values))
	for i := range values {
		scanArgs[i] = &values[i]
	}

	c := 0
	var data []row

	for rows.Next() {
		results := make(map[string]interface{})

		err = rows.Scan(scanArgs...)
		if err != nil {
			panic(err.Error())
		}

		for i, value := range values {
			switch value.(type) {
			case nil:
				results[columns[i]] = nil

			case []byte:
				s := string(value.([]byte))
				x, err := strconv.Atoi(s)

				if err != nil {
					results[columns[i]] = s
				} else {
					results[columns[i]] = x
				}

			default:
				results[columns[i]] = value
			}
		}

		data = append(data, results)
		c++
	}
	return data
}

func getDatabaseRowsAsJSON(qry string, args ...any) ([]row, error, int) {
	if pDB == nil {
		return nil, fmt.Errorf("the database is not connected"), http.StatusInternalServerError
	}
	if Rows, err := pDB.Query(qry, args...); err != nil {
		return nil, err, http.StatusInternalServerError
	} else {
		return Jsonify(Rows), err, 0
	}
}

func SendDataAsJSON(ctx *gin.Context, function string, sqlQry string, args ...any) {
	if data, err, httpError := getDatabaseRowsAsJSON(sqlQry, args...); err != nil {
		ReturnJSONError(ctx, function, err, httpError, true)
	} else {
		ctx.JSON(http.StatusOK, data)
	}
}
