package helper

import "reflect"

func IsInterfaceNil(data any) bool {
  return data == nil
}

func IsDeepNil(data any) bool {
  val := reflect.ValueOf(data)
  return val.IsNil()
}

func IsPointer(data any) bool {
  val := reflect.ValueOf(data)
  return val.Kind() == reflect.Pointer
}

