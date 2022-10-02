package ejdb2

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMemory(t *testing.T) {
	var db EJDB
	openError := db.Open("test.database")
	assert.NoError(t, openError)
	defer db.Close()
	counter := 0
	for i := 0; i < 1; i++ {
		err := db.EnsureCollection("users")
		assert.NoError(t, err)
		err = db.EnsureIndex("users", "/name", String)
		assert.NoError(t, err)

		collections := db.GetCollections()

		assert.Equal(t, 1, len(collections))
		assert.Equal(t, "users", collections[0])

		id, err := db.PutNew("users", `{"name": "John", "age": 30}`)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		assert.Equal(t, int64(1), id)

		id, err = db.PutNew("users", `{"name": "Marta", "age": 29}`)
		assert.NoError(t, err)
		assert.NotEmpty(t, id)
		assert.Equal(t, int64(2), id)

		db.Get("users", `/=[1, 2]`, func(id int64, entry string) {
			fmt.Println("Entry:", entry)
		})
		//fmt.Println("New record ID:", id)

		count := db.Count("users", `/[age > 20]`)
		assert.Equal(t, count, int64(2))
		//fmt.Println("Count:", count)
		err = db.GetWithArguments("users", `/[age > :age]`, J{"age": 20}, func(id int64, record string) {
			counter++
		})
		assert.NoError(t, err)
		db.GetByID("users", id)
		//fmt.Println("Entry from ID:", entryFromID)
		_, err = db.OnlineBackup("test.database.bak")
		//fmt.Println(db.GetMeta())
		assert.NoError(t, err)
		err = db.Del("users", id)
		assert.NoError(t, err)
		err = db.RemoveIndex("users", "/name", String)
		assert.NoError(t, err)
		err = db.RenameCollection("users", "people")
		assert.NoError(t, err)
		err = db.RemoveCollection("people")
		assert.NoError(t, err)
		//fmt.Println(db.GetMeta())
	}

}
