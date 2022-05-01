/**
@author ChenZhiYin 1981330085@qq.com
@date 2021/11/23
*/

package bmiddle

import (
	"encoding/json"
	"fmt"
	"github.com/go-kratos/kratos/v2/encoding"
	"github.com/go-kratos/kratos/v2/errors"
	stdhttp "net/http"
	"strconv"
	"time"
)

func ResponseSuccess(writer stdhttp.ResponseWriter, request *stdhttp.Request, i interface{}) error {
	codec := encoding.GetCodec("json")
	data, err := codec.Marshal(i)
	if err != nil {
		writer.WriteHeader(stdhttp.StatusInternalServerError)
		return nil
	}
	body := fmt.Sprintf(`{"code":1,"msg":"Success","time":%d,"data":%s}`, time.Now().Unix(), data)

	writer.Header().Set("Content-Type", "application/json")
	// 设置HTTP Status Code
	writer.WriteHeader(stdhttp.StatusOK)
	_, _ = writer.Write([]byte(body))
	return nil
}

func ResponseError(writer stdhttp.ResponseWriter, request *stdhttp.Request, err error) {
	type resp struct {
		Code int         `json:"code"`
		Msg  string      `json:"msg"`
		Data interface{} `json:"data"`
		Time int         `json:"time"`
	}

	code := 0
	message := err.Error()
	var toerr *errors.Error

	if se := errors.FromError(err); se != nil {
		code, _ = strconv.Atoi(se.GetReason())
		message = se.GetMessage()
		writer.WriteHeader(int(se.GetCode()))
	} else if errors.As(err, &toerr) {
		code, _ = strconv.Atoi(toerr.GetReason())
		message = toerr.GetMessage()
		writer.WriteHeader(int(toerr.GetCode()))
	}

	if code == 0 && message == "" {
		message = "未知错误"
	}

	reply, _ := json.Marshal(resp{
		Code: code,
		Msg:  message,
		Time: int(time.Now().Unix()),
	})

	writer.Header().Set("Content-Type", "application/json")
	_, _ = writer.Write(reply)
	return
}
