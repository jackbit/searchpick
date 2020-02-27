package searchpick

func (sFilter *SearchFilter) SubQueries(clause string, subQueries []interface{}) *SearchFilter {
  filters := []interface{}{}

  for _, subQuery := range subQueries {
    frQuery := &SearchFilter{
      Where: subQuery.(map[string]interface{}),
      Filters: []interface{}{},
    }
    
    frQuery.SetFilters()

    query := map[string]interface{}{
      "bool": map[string]interface{}{
        "filter": frQuery.Filters,
      },
    }

    filters = append(filters, query)
  }

  clauseQuery := map[string]interface{}{}
  clauseQuery[clause] = filters

  result := map[string]interface{}{ "bool": clauseQuery }

  return &SearchFilter{ Where: result }
}

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "or": []interface{}{  
//        []interface{}{ map[string]interface{}{ "in_stock": true } },
//        []interface{}{ map[string]interface{}{ "backordered": true } },
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQueryOr(field string, queries interface{}) bool {
  if field != "or" { return false }
  qry := queries.([]interface{})

  for _, subQueries := range qry {
    frSubQueries := sFilter.SubQueries("should", subQueries.([]interface{}))
    sFilter.Filters = append(sFilter.Filters, frSubQueries.Where)
  }

  return true
}

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "_or": []interface{}{  
//        map[string]interface{}{ "in_stock": true },
//        map[string]interface{}{ "backordered": true },
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQuery_Or(field string, queries interface{}) bool {
  if field != "_or" { return false }

  frQueries := sFilter.SubQueries("should", queries.([]interface{}))
  sFilter.Filters = append(sFilter.Filters, frQueries.Where)

  return true
}
