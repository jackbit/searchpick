package searchpick

import (
  "encoding/json"
  "reflect"
  "github.com/imdario/mergo"
  "github.com/thoas/go-funk"
)


func (fr SearchFilter) String() string {
  j, _ := json.Marshal(fr)
  return string(j)
}

func (sFilter *SearchFilter) IsSlice(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.Slice
}

func (sFilter *SearchFilter) IsMap(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.Map
}

func (sFilter *SearchFilter) IsString(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.String
}

func (sFilter *SearchFilter) IsInt(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.Int
}

func (sFilter *SearchFilter) IsFloat(i interface{}) bool {
  return reflect.TypeOf(i).Kind() == reflect.Float64
}

func (sFilter *SearchFilter) IsEmpty(i interface{}) bool {
  return reflect.ValueOf(i).IsZero()
}

func (sFilter *SearchFilter) IsMapExist(i interface{}) bool {
  return sFilter.IsMap(i) && !sFilter.IsEmpty(i)
}

func (sFilter *SearchFilter) IsSliceExist(i interface{}) bool {
  return sFilter.IsSlice(i) && len(i) > 0
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
  optionKeys := funk.Keys(aggOption)
  matchKey := ""
  for _, k := range optionKeys {
    if funk.ContainsString(aggsMetrick, k.(string)) {
      matchKey = k.(string)
      break
    }
  }
  return matchKey
}

func (sFilter *SearchFilter) SetAggregations(sOption *SearchOption) *SearchFilter {
  fieldPayload := map[string]interface{}{}
  
  if reflect.ValueOf(sOption.Aggs).IsZero() { return sFilter }
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
    if sFilter.IsString(aggOption["field"]) && aggOption["field"] != "" {
      alterField = aggOption["field"]
    }

    if sFilter.IsSliceExist(aggOption["ranges"]) {
      payloadOption := map[string]interface{}{
        "field": alterField,
        "ranges": aggOption["ranges"],
      }
      mergo.Merge(&payloadOption, sharedAggOption)
      fieldPayload["range"] = payloadOption
      aggsPayload[field] = fieldPayload

    } else if  sFilter.IsSliceExist(aggOption["date_ranges"]) {
      payloadOption := map[string]interface{}{
        "field": alterField,
        "ranges": aggOption["date_ranges"],
      }
      mergo.Merge(&payloadOption, sharedAggOption)
      fieldPayload["date_range"] = payloadOption
      aggsPayload[field] = fieldPayload

    } else if sFilter.IsMapExist(aggOption["date_histogram"]) {
      payloadOption := map[string]interface{}{
        "date_histogram": aggOption["date_histogram"],
      }
      mergo.Merge(&payloadOption, sharedAggOption)
      aggsPayload[field] = payloadOption

    } else if metricField := sFilter.MetricField(aggOption); metricField != "" {
      fieldPayload["field"] = field
      metricOption := aggOption[metricField].(map[string]interface{})
      if !sFilter.IsEmpty(metricOption) && metricOption["field"].(string) != "" {
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
      mergo.Merge(&payloadOption, sharedAggOption)
      fieldPayload["terms"] = payloadOption
      aggsPayload[field] = fieldPayload
    }

    where := map[string]interface{}{}

    if sOption.SmartAggs != "false" && sFilter.IsMapExist(sOption.Where) {
      smartWhere := map[string]interface{}{}
      for k, v := range sOption.Where {
        if k.(string) != field {
          smartWhere[k.(string)] = v
        }
      }
      where = smartWhere
    }

    aggFilter := &SearchFilter{ Filters: []interface{}{}, Where: where }
    aggFilter.SetFilters()

    trueFilters = []interface{}
    falseFilters = []interface{}

    for _, filter := range sFilter.Filters {
      if funk.Contains(aggFilter.Filters, filter) {
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

}
