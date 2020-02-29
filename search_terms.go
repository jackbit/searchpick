package searchpick

import "reflect"

func (sFilter *SearchFilter) TermFilterArray(field string) *SearchFilter {
  nonEmptyQueries := []interface{}{}
  totalNil := 0

  for _, item := range sFilter.Filters {
    if reflect.TypeOf(item) == nil {
      totalNil += 1
    } else {
      nonEmptyQueries = append(nonEmptyQueries, item)
    }
  }

  if totalNil > 0 {
    frQueryNil := sFilter.QueryNil(field)
    frQueryArray := sFilter.QueryArray(field, nonEmptyQueries)

    sFilter.Where = map[string]interface{}{
      "bool": map[string]interface{}{
        "should": []interface{}{
          frQueryNil.Where, 
          frQueryArray.Where,
        },
      },
    }
  } else {
    sFilter.Where = sFilter.QueryArray(field, nonEmptyQueries).Where
  }

  return sFilter
}


func (sFilter *SearchFilter) QueryNil(field string) *SearchFilter {
  fr := &SearchFilter{}
  fr.Where = map[string]interface{}{
    "bool": map[string]interface{}{
      "must_not": map[string]interface{}{
        "exists": field,
      },
    },
  }
  return fr
}

func (sFilter *SearchFilter) QueryArray(field string, cols []interface{}) *SearchFilter {
  fr := &SearchFilter{}
  terms := map[string]interface{}{}
  terms[field] = cols
  fr.Where = map[string]interface{}{
    "terms": terms,
  }
  return fr
}

//TODO : Add Terms for Regex : https://github.com/ankane/searchkick/blob/master/lib/searchkick/query.rb#L1023
func (sFilter *SearchFilter) TermFilters() *SearchFilter {
  newQuery := map[string]interface{}{}

  for field, queries := range sFilter.Where {

    queriesType := reflect.TypeOf(queries)

    if queriesType.Kind() == reflect.Slice {

      frArray := &SearchFilter{}

      if queriesType.String() == "[]interface {}" {
        frArray.Filters = queries.([]interface{})
      } else {
        frArray.Filters = frArray.ToFiltersFormat(queries)
      }

      sFilter.Where = frArray.TermFilterArray( field ).Where
    } else if queriesType == nil {
      sFilter.Where = sFilter.QueryNil(field).Where
    } else {
      newQuery[field] = queries
      sFilter.Where = map[string]interface{}{
        "term": newQuery,
      }
    }
  }
  return sFilter
}
