package main

import (
	"fmt"
	"os"

	"github.com/newrelic/infra-integrations-sdk/log"
	"github.com/newrelic/infra-integrations-sdk/persist"
)

type Topping string
type Mass string

type Pizza struct {
	Toppings []Topping
	Mass     Mass
	Slices   int
}

func main() {
	// Opening persistent storage
	storer, err := persist.NewFileStore(persist.DefaultPath("pizza"), log.NewStdErr(false), persist.DefaultTTL)

	// Retrieving last night's pizza
	var pizza Pizza
	timestamp, err := storer.Get("last-dinner", &pizza)

	if err == nil {
		fmt.Printf("I found some pizza in the storage: %+v (stored at %v)\n", pizza, timestamp)

	} else if err == persist.ErrNotFound {
		fmt.Println("No pizza in the fridge. Ordering more...")
		pizza = Pizza{
			Toppings: []Topping{"Pepperoni", "Mozzarella", "Cheese"},
			Mass:     "thin",
			Slices:   4,
		}
	} else {
		fmt.Println("Unexpected error:", err)
		os.Exit(-1)
	}

	fmt.Println("eating a delicious slice of pizza")
	pizza.Slices--

	if pizza.Slices <= 0 {
		fmt.Println("No more pizza on the fridge. Deleting...")
		storer.Delete("last-dinner")
	} else {
		storer.Set("last-dinner", pizza)
	}

	storer.Save()
}
