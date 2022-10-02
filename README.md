## Go Wrapper / Bindings for EJDB

I found this neat little embeddable DB and wanted to use it in go.


## Requirements

Make sure you have the dev headers of ejdb.

  - https://ejdb.org/
  - https://github.com/Softmotions/ejdb

cgo will be looking for these header files during compilation:

  - stdlib.h
  - ejdb2/ejdb2.h
  - ejdb2/iowow/iwkv.h
  - ejdb2/iowow/iwlog.h

## Add to your project

    go get github.com/memmaker/go-ejdb2/v2

## Example usage

```go
package main

import (
    "fmt"
    "github.com/memmaker/go-ejdb2/v2"
)

func main() {
    db := ejdb2.EJDB{}
    db.Open("test.database")
    defer db.Close()
    db.EnsureCollection("users")
    id := db.PutNew("users", `{"name": "John", "age": 30}`)
    fmt.Println("New record ID:", id)
    user := db.GetByID("users", id)
    fmt.Println(user)
}
```