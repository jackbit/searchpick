package searchpick

import (
  "strconv"
  "encoding/json"
  "reflect"
  "fmt"
  "strings"
)

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

func (sOption *SearchOption) SetExclude(field string, analyzer string) []map[string]interface{} {
  excludes := []map[string]interface{}{}
  for _, phrase := range sOption.Exclude {
    excludes = append(excludes, map[string]interface{}{
      "multi_match": map[string]interface{}{
        "fields": []string{ field },
        "query": phrase,
        "analyzer": analyzer,
        "type": "phrase",
      }
    })
  }
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