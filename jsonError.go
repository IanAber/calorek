package main

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"runtime"
)

type errorObj struct {
	Device string
	Err    string
}

type JSONError struct {
	Success bool        `json:"success"`
	Errors  []*errorObj `json:"errors"`
}

func (j *JSONError) AddErrorString(device string, err string) error {
	e := new(errorObj)
	e.Device = device
	e.Err = err
	j.Errors = append(j.Errors, e)
	return fmt.Errorf("device : %s | error %s", device, err)
}

func (j *JSONError) AddError(device string, err error) error {
	e := new(errorObj)
	e.Device = device
	e.Err = err.Error()
	j.Errors = append(j.Errors, e)
	return err
}

func (j *JSONError) String() string {
	if s, err := json.Marshal(j); err != nil {
		log.Print(err)
		return ""
	} else {
		return string(s)
	}
}

func (j *JSONError) ReturnError(ctx *gin.Context, retCode int) {
	ctx.JSON(retCode, j)
}

func ReturnJSONError(ctx *gin.Context, device string, err error, httpReturnCode int, bLog bool) {
	var jErr JSONError

	_ = jErr.AddError(device, err)
	jErr.Success = false
	jErr.ReturnError(ctx, httpReturnCode)
	if bLog {
		_, caller, line, _ := runtime.Caller(1)
		log.Printf("%s : %d : %v", caller, line, err)
	}
}

func ReturnJSONErrorString(ctx *gin.Context, device string, errStr string, httpReturnCode int, bLog bool) {
	var jErr JSONError

	err := jErr.AddErrorString(device, errStr)
	jErr.Success = false
	jErr.ReturnError(ctx, httpReturnCode)
	if bLog {
		log.Print(err)
	}
}

//func ReturnJSONSuccess(ctx *gin.Context) {
//	ctx.JSON(http.StatusOK, gin.H{"success": true})
//}
