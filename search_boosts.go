package searchpick

import (
  "reflect"
  "log"
  "time"
)

func (sFilter *SearchFilter) BoostFilters(boostBy map[string]interface{}, modifier map[string]interface{}) []interface{} {
  boosts := []interface{}
  for k, v := range boostBy {
    field := k.(string)
    value := v.(map[string]interface{})
    factor := 1
    if IsInt(value["factor"]) || IsFloat(value["factor"]) {
      factor := value["factor"]
    }

    fieldFactor := map[string]interface{}{
      "field": field,
      "factor": factor,
      "modifier": modifier,
    }

    if IsInt(value["missing"]) || IsFloat(value["missing"]) {
      fieldFactor["missing"] = value["missing"]
      boosts = append(boosts, map[string]interface{}{ "field_value_factor": fieldFactor })
    } else {
      boosts = append(boosts, map[string]interface{}{ 
        "field_value_factor": fieldFactor, 
        "exists": map[string]interface{}{
          "field": field,
        },
      })
    }
  }
  return boosts
}

func (sFilter *SearchFilter) CustomFilter(field string, value interface{}, factor float64) map[string]interface{} {
  cFilter := &SearchFilter{
    Filters: []interface{}{},
    Where: map[string]interface{}{
      "field": value,
    }
  }
  cFilter.SetFilters()
  return map[string]interface{}{
    "filter": cFilter.Filters,
    "weight": factor
  }
}

func (sFilter *SearchFilter) FilterBoostMultiply(sOption *SearchOption) *SearchFilter {
  sFilter.CustomFilters = []interface{}{}
  sFilter.MultiplyFilters = []interface{}{}

  boostBy := map[string]interface{}{}
  multiplyBy := map[string]interface{}{}

  if IsSliceExist(sOption.BoostBy) {
    for _, f := range sOption.BoostBy {
      boostBy[f] = map[string]interface{}{"factor": 1}
    }
  } else if IsMapExist(sOption.BoostBy) {
    partitioned := MapPartition(func(key string, value interface{}) bool {
      vMap := value.(map[string]interface{})
      boostMode := vMap["boost_mode"].(string)
      return boostMode == "multiply"
    }, sOption.BoostBy)
    multiplyBy = partitioned[0]
    boostBy = partitioned[1]
  }

  if sOption.Boost != "" {
    boostBy[sOption.Boost] = map[string]interface{}{"factor": 1}
  }

  sFilter.CustomFilters = append(boostBy, sFilter.BoostFilters(boostBy, map[string]interface{}{"modifier": "ln2p"}))
  sFilter.multiplyBy = append(multiplyBy, sFilter.BoostFilters(multiplyBy, map[string]interface{}{}))

  return sFilter
}

func (sFilter *SearchFilter) SetBoostWhere(sOption *SearchOption) *SearchFilter {
  boostWhere := map[string]interface{}{}
  if IsSliceExist(sOption.BoostWhere) {
    boostWhere = sOption.BoostWhere
  }
  for key, value := range boostWhere {
    field := key.(string)
    if IsSliceExist(value) && IsMap(reflect.ValueOf(value).Index(0).Interface()) {
      for _, cValue := range value.([]interface{}) {
        valueFactor := cValue.(map[string]interface{})
        sFilter.CustomFilters = append(sFilter.CustomFilters, sFilter.CustomFilter(field, valueFactor["value"], valueFactor["factor"].(float64)))
      }
    } else if IsMap(value) {
      cValue := value.(map[string]interface{})
      sFilter.CustomFilters = append(sFilter.CustomFilters, sFilter.CustomFilter(field, cValue["value"], cValue["factor"].(float64)))
    } else {
      sFilter.CustomFilters = append(sFilter.CustomFilters, sFilter.CustomFilter(field, value, 1000.0))
    }
  }
  return sFilter
}

func (sFilter *SearchFilter) SetBoostByDistance(sOption *SearchOption) *SearchFilter {
  if !IsMapExist(sOption.BoostByDistance) {return sFilter}
  boostByDistance := sOption.BoostByDistance
  if !IsEmpty(boostByDistance["field"]) {
    boostField := boostByDistance["field"]
    boostByDistance = map[string]interface{}{
      boostField: MapReject(boostByDistance, boostField),
    }
  }

  for k, v := range boostByDistance {
    field := k.(string)
    attributes := v.(map[string]interface{})
    MapMerge(&attributes, map[string]interface{}{"function": "gauss", "scale": "5mi"})

    if IsEmpty(attributes["origin"]) {
      log.Panic("boost_by_distance requires :origin")
    }

    functionParams := MapReject(attributes, "factor", "function")
    functionParams["origin"] = sFilter.LocationValue(functionParams["origin"])
    weightAttr := 1

    if !IsEmpty(attributes["factor"]) {
      weightAttr = attributes["factor"]
    }

    filterAttr := map[string]interface{}{
      "weight": weightAttr,
      attributes["function"]: map[string]interface{}{
        field: functionParams,
      },
    }

    sFilter.CustomFilters = append(sFilter.CustomFilters, filterAttr)
  }
  return sFilter
}

func (sFilter *SearchFilter) SetBoostByRecency(sOption *SearchFilter) *SearchFilter{
  if !IsMapExist(sOption.BoostByRecency) {return sFilter}
  for k, v := range sOption.BoostByRecency {
    field := k.(string)
    attributes := v.(map[string]interface{})
    attributes = MapMerge(&attributes, map[string]interface{}{"function": "gauss", "origin": time.Now()})

    fieldParams := MapReject(attributes, "factor", "function")
    
    weightAttr := 1

    if !IsEmpty(attributes["factor"]) {
      weightAttr = attributes["factor"]
    }

    filterAttr := map[string]interface{}{
      "weight": weightAttr,
      attributes["function"]: map[string]interface{}{
        field: functionParams,
      },
    }

    sFilter.CustomFilters = append(sFilter.CustomFilters, filterAttr)
  }
  return sFilter
}