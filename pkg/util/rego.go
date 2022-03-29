/**
@author ChenZhiYin 1981330085@qq.com
@date 2021/11/23
*/

package util

import (
	"github.com/go-kratos/kratos/v2/log"
	"runtime"
)

func Go(fun func()) {
	go func() {
		defer func() {
			if rerr := recover(); rerr != nil {
				buf := make([]byte, 64<<10)
				n := runtime.Stack(buf, false)
				buf = buf[:n]

				logger := log.NewHelper(log.DefaultLogger)
				logger.Errorf("recover %v:\n%s\n", rerr, buf)
			}
		}()
		fun()
	}()
}
