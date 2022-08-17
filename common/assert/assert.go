package assert

import (
	"fmt"
	"reflect"
)

func Assert(b bool, msg string) {
	if !b {
		panic(msg)
	}
}

func NotNil[V any](obj V, name string) V {
	Assert(any(obj) != nil && !reflect.ValueOf(obj).IsNil(), fmt.Sprintf("%s must be not nil", name))
	return obj
}

func IsNil[V any](obj V, name string) V {
	Assert(any(obj) == nil || reflect.ValueOf(obj).IsNil(), fmt.Sprintf("%s must nil", name))
	return obj
}

func NotEmpty[V any](obj V, name string) V {
	Assert(reflect.ValueOf(obj).Len() != 0, fmt.Sprintf("%s must be not empty", name))
	return obj
}

func MustEmpty[V any](obj V, name string) V {
	Assert(reflect.ValueOf(obj).Len() == 0, fmt.Sprintf("%s must be not empty", name))
	return obj
}

func NotZero[V any](obj V, name string) V {
	Assert(!reflect.ValueOf(obj).IsZero(), fmt.Sprintf("%s must be not zero", name))
	return obj
}

func IsZero[V any](obj V, name string) V {
	Assert(reflect.ValueOf(obj).IsZero(), fmt.Sprintf("%s must be not zero", name))
	return obj
}
