package searchpick

import (
  "log"
  "strconv"
  "encoding/json"
  // "errors"
  "strings"
  "reflect"
  "fmt"
  // "context"
)

type SearchOption struct {
  Term     string
  Fields   []string
  Operator string
  Page     int16
  PerPage  int16
  Limit    int16
  Padding  int16
  Offset   int16
  Order    map[string]interface{}
  Where    map[string]interface{}
  Similar  bool
  Match    string
}


type BoostField struct {
  Fields []string
  Boosts map[string]interface{}
}

type SearchQuery struct {
  Query         map[string]interface{}
  Field         string
  Operator      string
  OperatorQuery interface{}
}

type SearchFilter struct {
  Filters []interface{}
  Where   map[string]interface{}
  Field   string
}

type SearchResult struct {
  Params   *SearchOption
  Results  []interface{}
}

func (es SearchOption) String() string {
  j, _ := json.Marshal(es)
  return string(j)
}

func (sOption *SearchOption) SetPagination() {
  if sOption.Operator == "" { sOption.Operator = "and" }
  if sOption.Page <= 0 { sOption.Page = 1 }
  if sOption.PerPage <= 0 {
    if sOption.Limit > 0 {
      sOption.PerPage = sOption.Limit
    } else {
      sOption.PerPage = 10000
    }
  }
  if sOption.Padding <= 0 { sOption.Padding = 0 }
  if sOption.Offset <= 0 {
    sOption.Offset = (sOption.Page - 1) * sOption.PerPage + sOption.Padding
  }
}

func (sf BoostField) String() string {
  j, _ := json.Marshal(sf)
  return string(j)
}

func (sOption *SearchOption) SetFields() *BoostField {
  boostType := sOption.Match
  fields := []string{}
  boosts := map[string]interface{}{}

  if sOption.Match == "word" { boostType = "analyzed" }

  for i := range sOption.Fields {
    field := sOption.Fields[i]
    boost := reflect.ValueOf(strings.Split(field, "^"))
    field = boost.Index(0).String()
    newField := fmt.Sprintf("%s.%s", field, boostType)
    if boost.Len() > 1 {
      boosts[newField], _ = strconv.ParseFloat(boost.Index(1).String(), 64)
    }
    fields = append(fields, newField)
  }

  if len(sOption.Fields) < 1 {
    if boostType == "word" {
      fields = []string{ "*.analyzed" }
    } else {
      fields = []string{ "*."+boostType }
    }
  }

  return &BoostField{ 
    Fields: fields, 
    Boosts: boosts, 
  }
}

func (eq SearchQuery) String() string {
  j, _ := json.Marshal(eq)
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

func (fr SearchFilter) String() string {
  j, _ := json.Marshal(fr)
  return string(j)
}

func (s *Searchpick) Search(sOption *SearchOption) SearchResult {
  sQuery := &SearchQuery{ Query: map[string]interface{}{} } 

  if sOption.Term == "" { sOption.Term = "*" }
  if sOption.Match == "" { sOption.Match = "word" }
  sOption.SetPagination()
  
  boosField := sOption.SetFields()
  
  if sOption.Similar {
    sQuery.SetSimilar(sOption.Term, boosField)
  } else {
    sQuery.Query["match_all"] = map[string]interface{}{}
  }

  sFilter := &SearchFilter{ Filters: []interface{}{}, Where: sOption.Where }
  sFilter.SetFilters()
  
  // log.Println(sQuery.String())
  // log.Println(boosField.String())
  // log.Println(s.String())
  j, _ := json.Marshal(sFilter.Filters)
  log.Println(string(j))
  esResult := SearchResult{}
  return esResult
}

