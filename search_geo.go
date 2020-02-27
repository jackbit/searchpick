package searchpick

import "reflect"

func (sFilter *SearchFilter) CoodinateArray(value interface{}) interface{} {
  typeValue := reflect.TypeOf(value).Kind()
  if typeValue == reflect.Map {
    valueSlice := value.(map[string]interface{})
    return []interface{}{ valueSlice["lon"].(float64), valueSlice["lat"].(float64) }
  } else {
    if typeValue == reflect.Slice {
      valueArray := value.([]interface{})
      typeFirtValue := reflect.TypeOf(valueArray[0]).Kind()
      if typeFirtValue == reflect.Float64 ||  typeFirtValue == reflect.Int {
        return value
      } else {
        newValues := []interface{}{}
        for _, val := range valueArray {
          newValues = append(newValues, sFilter.CoodinateArray(val))
        }
        return newValues
      }
    } else {
      return value
    }
  }
}

func (sFilter *SearchFilter) LocationValue(opQuery interface{}) interface{} {
  typeQuery := reflect.TypeOf(opQuery).Kind()
  if typeQuery == reflect.Slice {
    coordinates := sFilter.ToFiltersFormat(opQuery)
    for i, j := 0, len(coordinates)-1; i < j; i, j = i+1, j-1 {
      coordinates[i], coordinates[j] = coordinates[j].(float64), coordinates[i].(float64)
    }
    return coordinates
  }
  return opQuery
}

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "location": map[string]interface{}{
//       "near": []interface{}{ 30.0, 100.0},
//       "within": "1km",
//     },
//   },
// })
// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "location": map[string]interface{}{
//       "near": map[string]interface{}{"lat": 30.0, "lon": 100.0},
//       "within": "1km",
//     },
//   },
// })
func (sFilter *SearchFilter) GeoNearQuery(sq *SearchQuery) {
  geoDistance := map[string]interface{}{}
  geoDistance[sq.Field] = sFilter.LocationValue(sq.OperatorQuery)
  geoDistance["within"] = "50mi"
  if sq.Query["within"] != nil {
    geoDistance["within"] = sq.Query["within"]
  }

  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{ "geo_distance": geoDistance})
}

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "location": map[string]interface{}{
//       "geo_polygon": map[string]interface{}{
//         "points": []interface{}{
//           map[string]interface{}{"lat": 30, "lon": -120},
//           map[string]interface{}{"lat": 33, "lon": -123},
//           map[string]interface{}{"lat": 40, "lon": -130},
//         },
//       },
//     },
//   },
// })
// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "location": map[string]interface{}{
//       "geo_polygon": []interface{}{ []interface{}{30.0, 100.0}, []interface{}{20.0, 120.0} },
//     },
//   },
// })
func (sFilter *SearchFilter) GeoPolygonQuery(sq *SearchQuery) {
  geoQuery := map[string]interface{}{}
  geoQuery[sq.Field] = sq.OperatorQuery
  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{ "geo_polygon": geoQuery})
}

  // user.Searchpick().Search(&sp.SearchOption{
  //   Where: map[string]interface{}{
  //     "area": map[string]interface{}{
  //       "geo_shape": map[string]interface{}{
  //         "type": "point",
  //         "coordinates": []interface{}{20.0, 120.0},
  //       },
  //     },
  //   },
  // })
func (sFilter *SearchFilter) GeoShapeQuery(sq *SearchQuery) {
  geoQuery := map[string]interface{}{}
  opQuery := sq.OperatorQuery.(map[string]interface{})
  shape := map[string]interface{}{}
  relation := "intersects"

  for k, v := range opQuery {
    if k == "relation" {
      if v != "" {
        relation = v.(string)
      }
    } else {
      shape[k] = v
    }
  }

  if shape["coordinates"] != nil {
    shape["coordinates"] = sFilter.CoodinateArray(shape["coordinates"]) 
  }

  geoQuery[sq.Field] = map[string]interface{}{ "relation": relation, "shape": shape }
  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{ "geo_shape": geoQuery})
}

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "location": map[string]interface{}{
//       "top_left": map[string]interface{}{
//         "lat": 38,
//         "lon": -123,
//       },
//       "bottom_right": map[string]interface{}{
//         "lat": 37,
//         "lon": -122,
//       },
//     },
//   },
// })
func (sFilter *SearchFilter) GeoTopLeftQuery(sq *SearchQuery) {
  geoQuery := map[string]interface{}{}
  geoQuery[sq.Field] = map[string]interface{}{
    "top_left": sFilter.LocationValue(sq.OperatorQuery),
    "bottom_right": sFilter.LocationValue(sq.Query["bottom_right"]),
  }
  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{ "geo_bounding_box": geoQuery})
}

// user.Searchpick().Search(&sp.SearchOption{
//   Where: map[string]interface{}{
//     "location": map[string]interface{}{
//       "top_right": map[string]interface{}{
//         "lat": 38,
//         "lon": -123,
//       },
//       "bottom_left": map[string]interface{}{
//         "lat": 37,
//         "lon": -122,
//       },
//     },
//   },
// })
func (sFilter *SearchFilter) GeoTopRightQuery(sq *SearchQuery) {
  geoQuery := map[string]interface{}{}
  geoQuery[sq.Field] = map[string]interface{}{
    "top_right": sFilter.LocationValue(sq.OperatorQuery),
    "bottom_left": sFilter.LocationValue(sq.Query["bottom_left"]),
  }
  sFilter.Filters = append(sFilter.Filters, map[string]interface{}{ "geo_bounding_box": geoQuery})
}

func (sFilter *SearchFilter) CheckQueryGeo(sq *SearchQuery) bool {
  isGeo := true

  switch sq.Operator {
    case "near":
      sFilter.GeoNearQuery(sq)
    case "geo_polygon":
      sFilter.GeoPolygonQuery(sq)
    case "geo_shape":
      sFilter.GeoShapeQuery(sq)
    case "top_left":
      sFilter.GeoTopLeftQuery(sq)
    case "top_right":
      sFilter.GeoTopRightQuery(sq)
    default:
      isGeo = false
  }

  return isGeo
}