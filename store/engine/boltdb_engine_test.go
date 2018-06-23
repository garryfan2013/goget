package engine

import (
	"encoding/json"
	"fmt"
	"testing"
)

const (
	bucketUser    = "user"
	bucketCompany = "company"
)

type Company struct {
	Name   string
	Phone  string
	TaxNum string
	Users  []*User
}

type User struct {
	Name   string
	Age    int
	Addr   string
	Mobile string
}

var (
	bs, _ = new(BoltDBStoreBuilder).Build()

	c1 = &Company{
		Name:   "IBM",
		Phone:  "82886709",
		TaxNum: "tax000001",
		Users: []*User{
			&User{
				Name:   "gary",
				Addr:   "LA 5th street",
				Age:    35,
				Mobile: "13912345678",
			},
			&User{
				Name:   "jason",
				Addr:   "Seatle beverly park 6-308",
				Age:    42,
				Mobile: "18866663333",
			},
		},
	}

	c2 = &Company{
		Name:   "Apple",
		Phone:  "82886709",
		TaxNum: "tax000001",
		Users: []*User{
			&User{
				Name:   "gary",
				Addr:   "LA 5th street",
				Age:    35,
				Mobile: "13912345678",
			},
			&User{
				Name:   "jason",
				Addr:   "Seatle beverly park 6-308",
				Age:    42,
				Mobile: "18866663333",
			},
		},
	}
)

func TestPut(t *testing.T) {
	company := c1
	data, err := json.Marshal(company)
	if err != nil {
		t.Error(err)
	}

	err = bs.Put(bucketCompany, company.Name, data)
	if err != nil {
		t.Error(err)
	}
}

func TestPut2(t *testing.T) {
	company := c2
	data, err := json.Marshal(company)
	if err != nil {
		t.Error(err)
	}

	err = bs.Put(bucketCompany, company.Name, data)
	if err != nil {
		t.Error(err)
	}
}

func TestGet(t *testing.T) {
	data, err := bs.Get(bucketCompany, "IBM")
	if err != nil {
		t.Error(err)
	}

	if data == nil {
		t.Errorf("Should get a non-nil value")
	}

	fmt.Printf("%s\n", string(data))

	var c Company
	err = json.Unmarshal(data, &c)
	if err != nil {
		t.Error(err)
	}

	if c.Name != "IBM" {
		t.Errorf("Bad company name %s\n", c.Name)
	}

	if len(c.Users) != 2 {
		t.Errorf("Bad users count %d\n", len(c.Users))
	}

	fmt.Printf("%s,%s\n", c.Users[0].Name, c.Users[1].Name)
}

func TestForEach(t *testing.T) {
	err1 := bs.ForEach(bucketCompany, func(key string, val []byte) error {
		var c Company
		err2 := json.Unmarshal(val, &c)
		if err2 != nil {
			return err2
		}

		fmt.Printf("%s\n", string(val))
		return nil
	})
	if err1 != nil {
		t.Error(err1)
		return
	}
}

func TestMultiOpen(t *testing.T) {
	fmt.Printf("bs2.Open\n")
	bs2, err := new(BoltDBStoreBuilder).Build()
	if err == nil {
		fmt.Printf("bs2.Close\n")
		bs2.Close()
		t.Errorf("The 2nd Open should timeout, but it didnt\n")
		return
	}

	if err.Error() != "timeout" {
		t.Error(err)
		return
	}

	ch := make(chan error)
	go func() {
		fmt.Printf("bs3.Open\n")
		bs3, err1 := new(BoltDBStoreBuilder).Build()
		if err1 != nil {
			ch <- err1
			return
		}
		fmt.Printf("bs3.Close\n")
		bs3.Close()
		ch <- nil
	}()

	err = <-ch
	if err == nil {
		t.Errorf("The 3rd Open should timeout, but it didnt\n")
		return
	}

	if err.Error() != "timeout" {
		t.Error(err)
		return
	}
}

func TestClose(t *testing.T) {
	fmt.Printf("bs.Close\n")
	err := bs.Close()
	if err != nil {
		t.Error(err)
	}
}
