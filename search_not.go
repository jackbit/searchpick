package searchpick

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//       "_not": map[string]interface{}{ "store_id": 1 },
//     },
//   })
func (sFilter *SearchFilter) CheckQuery_Not(field string, queries interface{}) bool {
  if field != "_not" { return false }
  frQuery := &SearchFilter{
    Where: queries.(map[string]interface{}),
    Filters: []interface{}{},
  }
  
  frQuery.SetFilters()

  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{
    "bool": map[string]interface{}{
      "must_not": frQuery.Filters,
    },
  })
  return true
}
