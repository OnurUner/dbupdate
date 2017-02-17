package challenge

import (
	"sync"
)

// START DO NOT EDIT
type DB interface {
	Get(ids []string) ([]int, error)
	Set(ids []string, vals []int) error
}

type update struct {
	ids  []string
	vals []int
}

func (u *update) apply(dbvalues []int) []int {
	if len(dbvalues) != len(u.vals) {
		panic("slice lengths do not match")
	}

	for i := range u.vals {
		dbvalues[i] += u.vals[i]
	}
	return dbvalues
}

// END DO NOT EDIT

func containsAny(items []string, list []string) bool {
	for _, x := range items {
		for _, y := range list {
			if x == y {
				return true
			}
		}
	}
	return false
}

func addRange(from []string, to []string) []string {
	for _, a := range from {
		to = append(to, a)
	}
	return to
}

func removeItems(remove []string, list []string) []string {
	for i := 0; i < len(list); i++ {
		item := list[i]
		for _, rem := range remove {
			if item == rem {
				list = append(list[:i], list[i+1:]...)
				i--
				break
			}
		}
	}
	return list
}

var dbKeys []string
var m = &sync.Mutex{}
var c = sync.NewCond(m)

func updateConcurrent(wg *sync.WaitGroup, database DB, item update) {
	defer wg.Done()

	c.L.Lock()
	if len(dbKeys) != 0 {
		for {
			if containsAny(item.ids, dbKeys) {
				//fmt.Println("WAIT")
				c.Wait()
			} else {
				break
			}
		}
	}

	dbKeys = addRange(item.ids, dbKeys)
	c.L.Unlock()

	//fmt.Println(item.ids, dbKeys)
	dbvals, err := database.Get(item.ids)
	if err != nil {
		panic(err)
	}

	item.apply(dbvals)
	if err := database.Set(item.ids, dbvals); err != nil {
		panic(err)
	}

	c.L.Lock()
	dbKeys = removeItems(item.ids, dbKeys)
	c.Broadcast()
	c.L.Unlock()

}

// modify below to get runUpdates to run faster without breaking the tests
func runUpdates(db DB, updates []update) error {
	var wg sync.WaitGroup

	for _, u := range updates {
		wg.Add(1)
		go updateConcurrent(&wg, db, u)
	}
	wg.Wait()

	return nil
}
