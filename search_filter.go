package searchpick

import (
  "encoding/json"
  "reflect"
  "strings"
  "log"
  "time"
)


func (fr SearchFilter) String() string {
  j, _ := json.Marshal(fr)
  return string(j)
}

func (sFilter *SearchFilter) ToFiltersFormat(i interface{}) []interface{} {
  j, _ := json.Marshal(i)
  jsonQuery := []interface{}{}
  _ = json.Unmarshal([]byte(string(j)), &jsonQuery)
  return jsonQuery
}

func (sFilter *SearchFilter) OperatorFilters(field string, values interface{}) *SearchFilter{
  newFilters := &SearchFilter{}
  newFilters.Where = map[string]interface{}{}
  newFilters.Where[field] = values
  newFilters.TermFilters()
  return newFilters
}

func (sFilter *SearchFilter) SetFilters() *SearchFilter {

  for field, queries := range sFilter.Where {

    if sFilter.CheckQueryOr(field, queries) { continue }
    if sFilter.CheckQuery_Or(field, queries) { continue }
    if sFilter.CheckQuery_Not(field, queries) { continue }
    if sFilter.CheckQuery_And(field, queries) { continue }

    typeQueries := reflect.TypeOf(queries).Kind()

    if typeQueries == reflect.Slice {
      queries = map[string]interface{}{"in": queries}
      typeQueries = reflect.TypeOf(queries).Kind()
    }

    if typeQueries == reflect.Map {
      
      hashQueries := queries.(map[string]interface{})

      for op, opValues := range hashQueries {

        if op == "within" || op == "bottom_right" || op == "bottom_left" { continue }
        
        esQuery := &SearchQuery{
          Field: field,
          Query: hashQueries,
          Operator: op,
          OperatorQuery: opValues,
        }

        if sFilter.CheckQueryIn(esQuery) { continue }
        if sFilter.CheckQueryRange(esQuery) { continue }
        if sFilter.CheckQueryGeo(esQuery) { continue }
        if sFilter.CheckQueryLike(esQuery) { continue }
        if sFilter.CheckQueryPrefix(esQuery) { continue }
        if sFilter.CheckQueryRegex(esQuery) { continue }
        if sFilter.CheckQueryNot(esQuery) { continue }
        if sFilter.CheckQueryAll(esQuery) { continue }
        if sFilter.CheckQueryExists(esQuery) { continue }
      }
    } else {
      sFilter.Filters = append(sFilter.Filters, sFilter.OperatorFilters( field, queries ).Where)
    }
    
  }
  return sFilter
}

func (sFilter *SearchFilter) MetricField(aggOption map[string]interface{}) string {
  aggsMetrick := []string{"avg", "cardinality", "max", "min", "sum"}
  optionKeys := SliceKeys(aggOption)
  matchKey := ""
  for _, k := range optionKeys {
    if SliceContainsString(aggsMetrick, k.(string)) {
      matchKey = k.(string)
      break
    }
  }
  return matchKey
}

func (sFilter *SearchFilter) SetAggregations(sOption *SearchOption) *SearchFilter {
  fieldPayload := map[string]interface{}{}
  
  if IsEmpty(sOption.Aggs) { return sFilter }
  if reflect.TypeOf(sOption.Aggs).Kind() == reflect.Slice && len(sOption.Aggs) < 1 { return sFilter }
  
  postFilters = []interface{}

  aggs := map[string]interface{}{}
  aggsPayload := map[string]interface{}{}

  if reflect.TypeOf(sOption.Aggs).Kind() == reflect.Slice {
    for _, field := range sOption.Aggs {
      aggs[field] = map[string]interface{}{}
    }
  } else {
    aggs = sOption.Aggs.(map[string]interface{})
  }

  for k, v := range aggs {
    field := k.(string)
    aggOption := v.(map[string]interface{})
    sharedAggOption := map[string]interface{}{}

    size := 1000
    limitType := reflect.TypeOf(aggOption["limit"]).Kind()

    if limitType == reflect.Int {
      size = aggOption["limit"].(int)
    } else if limitType == reflect.float64 {
      size = aggOption["limit"].(float64)
    }

    alterField := field
    if IsString(aggOption["field"]) && aggOption["field"] != "" {
      alterField = aggOption["field"]
    }

    if IsSliceExist(aggOption["ranges"]) {
      payloadOption := map[string]interface{}{
        "field": alterField,
        "ranges": aggOption["ranges"],
      }
      MapMerge(&payloadOption, sharedAggOption)
      fieldPayload["range"] = payloadOption
      aggsPayload[field] = fieldPayload

    } else if  IsSliceExist(aggOption["date_ranges"]) {
      payloadOption := map[string]interface{}{
        "field": alterField,
        "ranges": aggOption["date_ranges"],
      }
      MapMerge(&payloadOption, sharedAggOption)
      fieldPayload["date_range"] = payloadOption
      aggsPayload[field] = fieldPayload

    } else if IsMapExist(aggOption["date_histogram"]) {
      payloadOption := map[string]interface{}{
        "date_histogram": aggOption["date_histogram"],
      }
      MapMerge(&payloadOption, sharedAggOption)
      aggsPayload[field] = payloadOption

    } else if metricField := sFilter.MetricField(aggOption); metricField != "" {
      fieldPayload["field"] = field
      metricOption := aggOption[metricField].(map[string]interface{})
      if !IsEmpty(metricOption) && metricOption["field"].(string) != "" {
        fieldPayload["field"] = metricOption["field"]
      }
      aggsPayload[field] = map[string]interface{}{
        metricField: metricOption,
      }
    } else {
      payloadOption := map[string]interface{}{
        "field": alterField,
        "size": size,
      }
      MapMerge(&payloadOption, sharedAggOption)
      fieldPayload["terms"] = payloadOption
      aggsPayload[field] = fieldPayload
    }

    where := map[string]interface{}{}

    if sOption.SmartAggs != "false" && IsMapExist(sOption.Where) {
      where = MapReject(sOption.Where, field)
    }

    aggFilter := &SearchFilter{ Filters: []interface{}{}, Where: where }
    aggFilter.SetFilters()

    trueFilters = []interface{}
    falseFilters = []interface{}

    for _, filter := range sFilter.Filters {
      if SliceContains(aggFilter.Filters, filter) {
        trueFilters = append(trueFilters, filter)
      } else {
        postFilters = append(postFilters, filter)
        falseFilters = append(falseFilters, filter)
      }
    }

    sFilter.Filters = trueFilters

    if len(aggFilter.Filters) > 0 {
      fieldPayload[field] = map[string]interface{}{
        "filter": map[string]interface{}{
          "bool": map[string]interface{}{
            "must": aggFilter.Filters
          },
          "aggs": map[string]interface{}{
            field: fieldPayload[field],
          }
        },
      }
    }

  }

  sFilter.Payloads = map[string]interface{}{
    "aggs": fieldPayload,
    "post_filter": map[string]interface{}{
      "bool": map[string]interface{}{
        "filter": postFilters,
      },
    },
  }

  return sFilter
}

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
