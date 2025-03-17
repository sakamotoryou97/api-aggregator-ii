package helper

import "reflect"

func IsStringEmpty(s string) bool {
  return s == ""
}

func IsString(s any) bool {
  val := reflect.ValueOf(s)

  if val.Kind() == reflect.Pointer {
    val = val.Elem()
  }

  return val.Kind() == reflect.String
}
