package searchpick

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "category": []string{"shoe", "fashion"},
//   },
// })
func (sFilter *SearchFilter) CheckQueryIn(sq *SearchQuery) bool {
  if sq.Operator != "in" { return false }
  
  sFilter.Filters = append( sFilter.Filters,  sFilter.OperatorFilters(sq.Field, sq.OperatorQuery).Where )

  return true
}
