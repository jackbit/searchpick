# Searchpick {Work in Progress}

Inspired by [Searchkick](https://github.com/ankane/searchkick), make elasticsearch easy in golang. 

Interface for searching is using SearchOption:
```
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
```
#### Connecting
use .env file or local env variable
```
ES_URL=http://localhost:9200
ES_USERNAME=
ES_PASSWORD=
ES_NAMESPACE="project_dev"
```


#### Model Integration

```
import (
  sp "github.com/jackbit/searchpick"
  "fmt"
)

type User struct {
  ID                uuid.UUID  `json:"id" db:"id"`
  Name              string     `json:"name" db:"name"`
  Email             string     `json:"email" db:"email"`
  Phone             string     `json:"phone" db:"phone"`
  Lat               float64    `json:"lat" db:"lat"`
  Lon               float64    `json:"lon" db:"lon"`
  Polygons          string     `json:"polygons" db:"polygons"`
  CreatedAt         time.Time  `json:"created_at,omitempty" db:"created_at"`
  UpdatedAt         time.Time  `json:"updated_at,omitempty" db:"updated_at"`
}

func (u *User) Searchpick() *sp.Searchpick {
  return &sp.Searchpick{
    Name: "users",
    Locations: []string{ "coordinate" }, 
    GeoShape: []string{ "area" }
  }
}

func (u *User) SearchData() *sp.Searchpick {
  data := map[string]interface{} {
    "keyword": fmt.Sprintf("%s %s %s", u.Name, u.Email, u.Phone),
    "created_at": u.CreatedAt,
    "updated_at": u.UpdatedAt,
    "coordinate": map[string]interface{}{
      "lat": u.Lat,
      "lon": u.Lon,
    },
    "area": map[string]interface{}{
      "type": "polygon",
      "coordinates": u.PolygonSlice()
    },
  }
  searchPick := u.Searchpick()
  searchPick.SearchData = searchPick.BuildSearchData(u.ID.String(), data)
  return searchPick
}

func (u *User) PolygonSlice() []interface{} {
  var polygons []interface{}
  _ = json.Unmarshal([]byte(u.Polygons), &polygons)
  return polygons

}

```

#### Indexing

Reindex or new index a record

```
user := &User{ 
  ID: 1,
  Name: "Github"
  Email: "github@local.host",
  Phone: "0987654321",
}

user.SearchData().Reindex()
```

Remove index from a record (not all index)
```
user := &User{ ID: 1 }
user.SearchData().IndexDelete()
```

Check if current record has index
```
user := &User{ ID: 1 }
user.SearchData().IndexExists()
```

Check index id in current model
```
user := &User{}
user.Searchpick().IndexId(2)
```


#### Querying

Query like SQL

```go
options := SearchOption{
    Term: "*",
    Where: map[string]interface{}{
        "expired_at": map[string]interface{}{"gt": time.Now()},
        "category_id": []interface{}{ 25, 30 },
        "size": map[string]interface{}{
            "all": []interface{}{ 
                "s", 
                "m", 
                "l", 
                "xl",
            },
        },
        "title": map[string]interface{}{
            "like": "%converse%",
        },
        "tag": map[string]interface{}{
            "regexp": "/shoe .+/",
        },
        "supplier_id": map[string]interface{}{
            "exists": true,
        },
        "payment_method": map[string]interface{}{
            "not": []interface{}{ 
                "cod", 
                "cash",
            },
        },
        "_or": []interface{}{
          map[string]interface{}{
              "in_stock": true,
              "backordered": true,
          },
        },
        "_and": []interface{}{
          map[string]interface{}{
              "city_id": []interface{}{ 1, 2, 3, 4},
          },
          map[string]interface{}{
              "shipping_method_id": []interface{}{ 10, 20, 30, 40},
          },
        },
        "_not": map[string]interface{}{
            "store_id": 1001,
        },
    },
    Page: 1,
    PerPage: 20,
    Order: map[string]interface{}{
        "price": "asc",
    }
}
response, err := product.Search(options)
if err != nil {
    log.Println(err.Error())
}
log.Println(response)
```

Query like JSON

```go
query := `{
    "expired_at": {"gt": "2020-12-24"},
    "category_id": [25, 30],
    "size": {"all": ["s", "m", "l", "xl"]},
    "title": {"like": "%converse%"},
    "tag": {"regexp": "/shoe .+/"},
    "supplier_id": {"exists": true},
    "payment_method": {"not": ["cod", "cash"]},
    "_or": [ {"in_stock": true}, {"backordered": true} ],
    "_and": [ {"city_id": [1, 2, 3, 4]}, {"shipping_method_id": [10, 20, 30, 40]} ],
    "_not": {"store_id": 1}
}`
jsonQuery, _ := json.Marshal(query)
whereQuery := jsonQuery.( map[string]interface{} )
options := SearchOption{
    Keyword: "*",
    Where: whereQuery,
    Page: 1,
    PerPage: 20,
    Order: map[string]interface{}{ "price": "asc" }
}
response, err := product.Search(options)
if err != nil {
    log.Println(err.Error())
}
log.Println(response)
```
