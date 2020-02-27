package searchpick

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "_and": []interface{}{  
//        map[string]interface{}{ "in_stock": true },
//        map[string]interface{}{ "backordered": true },
//     },
//   },
// })
func (sFilter *SearchFilter) CheckQuery_And(field string, queries interface{}) bool {
  if field != "_and" { return false }
  frQueries := sFilter.SubQueries("must", queries.([]interface{}))
  sFilter.Filters = append(sFilter.Filters, frQueries.Where)
  return true
}
