package ejdb2

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemory(t *testing.T) {
	var db EJDB
	db.Open("test.database")
	defer db.Close()
	counter := 0
	for i := 0; i < 100; i++ {
		db.EnsureCollection("users")
		db.EnsureIndex("users", "/name", String)

		id := db.PutNew("users", `{"name": "John", "age": 30}`)
		assert.NotEmpty(t, id)
		assert.Equal(t, int64(1), id)
		//fmt.Println("New record ID:", id)

		count := db.Count("users", `/[age > :age]`, J{"age": 20})
		assert.Equal(t, count, int64(1))
		//fmt.Println("Count:", count)
		db.Get("users", `/[age > :age]`, J{"age": 20}, func(record string) {
			counter++
		})
		db.GetByID("users", id)
		//fmt.Println("Entry from ID:", entryFromID)

		//fmt.Println(db.GetMeta())

		db.Del("users", id)

		db.RemoveIndex("users", "/name", String)

		db.RenameCollection("users", "people")

		db.RemoveCollection("people")

		//fmt.Println(db.GetMeta())
	}

}
