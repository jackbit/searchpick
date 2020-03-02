package searchpick

import (
  "reflect"
  "github.com/imdario/mergo"
)

func IsSlice(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.Slice
}

func IsMap(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.Map
}

func IsString(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.String
}

func IsInt(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.Int
}

func IsFloat(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.Float64
}

func IsEmpty(i interface{}) bool {
  return reflect.ValueOf(i).IsZero()
}

func IsMapExist(i interface{}) bool {
  return sFilter.IsMap(i) && !sFilter.IsEmpty(i)
}

func IsSliceExist(i interface{}) bool {
  return sFilter.IsSlice(i) && len(i) > 0
}

func MapMerge(dest, src interface{}) {
  mergo.Merge(dest, src)
}

func MapReject(input map[string]interface{}, keys ...string) map[string]interface{} {
  return MapSelectORreject(true, input, keys...)
}

func MapSelect(input map[string]interface{}, keys ...string) map[string]interface{} {
  return MapSelectORreject(false, input, keys...)
}

func MapSelectORreject(reject bool, input map[string]interface{}, keys ...string) (output map[string]interface{}) {
  size := len(input)
  keysSize := len(keys)
  if size <= 0 {
    return nil
  }
  if keysSize <= 0 {
    return input
  }
  if size >= keysSize {
    size = size - keysSize
  }
  output = make(map[string]interface{}, size)
  for key, value := range input {
    if reject {
      if !MapIncludes(key, keys...) {
        output[key] = value
      }
    } else {
      if MapIncludes(key, keys...) {
        output[key] = value
      }
    }
  }
  return output
}

func MapIncludes(k string, keys ...string) bool {
  for _, key := range keys {
    if key == k {
      return true
    }
  }
  return false
}

func MapPartition(f func(string, interface{}) bool, input map[string]interface{}) (partition []map[string]interface{}) {
  partition = make([]map[string]interface{}, 2)
  size := len(input)
  if size == 0 {
    partition[0] = input
    partition[1] = nil
    return partition
  }
  // Assuming half of key values will be partitioned
  partition[0] = make(map[string]interface{}, size/2)
  partition[1] = make(map[string]interface{}, size/2)
  for key, value := range input {
    if f(key, value) {
      partition[0][key] = value
    } else {
      partition[1][key] = value
    }
  }
  return partition
}


func SliceContainsString(s []string, v string) bool {
  for _, vv := range s {
    if vv == v {
      return true
    }
  }
  return false
}

func SliceUniqString(a []string) []string {
  length := len(a)

  seen := make(map[string]struct{}, length)
  j := 0

  for i := 0; i < length; i++ {
    v := a[i]

    if _, ok := seen[v]; ok {
      continue
    }

    seen[v] = struct{}{}
    a[j] = v
    j++
  }

  return a[0:j]
}

func SliceContains(in interface{}, elem interface{}) bool {
  inValue := reflect.ValueOf(in)
  elemValue := reflect.ValueOf(elem)
  inType := inValue.Type()
  for _, key := range inValue.MapKeys() {
    if equal(key.Interface(), elem) {
      return true
    }
  }
  return false
}

func SliceKeys(out interface{}) interface{} {
  value := redirectValue(reflect.ValueOf(out))
  valueType := value.Type()

  if value.Kind() == reflect.Map {
    keys := value.MapKeys()

    length := len(keys)

    resultSlice := reflect.MakeSlice(reflect.SliceOf(valueType.Key()), length, length)

    for i, key := range keys {
      resultSlice.Index(i).Set(key)
    }

    return resultSlice.Interface()
  }

  if value.Kind() == reflect.Struct {
    length := value.NumField()

    resultSlice := make([]string, length)

    for i := 0; i < length; i++ {
      resultSlice[i] = valueType.Field(i).Name
    }

    return resultSlice
  }

  panic(fmt.Sprintf("Type %s is not supported by Keys", valueType.String()))
}

func redirectValue(value reflect.Value) reflect.Value {
  for {
    if !value.IsValid() || value.Kind() != reflect.Ptr {
      return value
    }

    res := reflect.Indirect(value)

    if res.Kind() == reflect.Ptr && value.Pointer() == res.Pointer() {
      return value
    }

    value = res
  }
}

