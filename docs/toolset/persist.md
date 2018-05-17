# Persistence

The GoSDK v3 provides the [persist.Storer](https://godoc.org/github.com/newrelic/infra-integrations-sdk/persist#Storer)
interface, which allow any integration to access a simple key-value storage.

Document structure:

* [Basic functionality](#basic-functionality)
* [Restrictions](#restrictions)
* [Code example](#example)

## Basic functionality

The `persist` package consist of the `Storer` interface plus the `NewFileStore`
function:

* `NewFileStore` returns a disk-backed `Storer` using the provided file path.
    - Arguments:
        - `storagePath string`: the path to the file where the persisted data
          will be stored.
        - `ilog log.Logger`: [internal logger](log.md) where some debug/error
          messages will be shown.
        - `ttl time.Duration`: _time to live_. Files older than this duration
          will be discarded and won't be loaded.
    - Returns:
        - The instantiated `Storer`.
        - An error, if any error happen during the creation.

For the `Storer` interface:

* `Set` stores a value in the storage, associated to a given key.
    - Arguments:
        - `key string`: the key to be associated to the given value.
        - `value interface{}`: any value to store (struct, array, string,
          primitive type...)
    - Return:
        - A Unix timestamp indicating the moment the data has been stored.
        - An error, if the store operation has not been possible.
        
* `Get` reads the value associated to a given key and stores in the argument
  passed as reference.
    - Arguments:
        - `key string`: the key associated to the given value.
        - `valuePtr interface{}`: a **pointer to a value** (argument passing by
          reference) where the read data will be stored.
    - Return:
        - A Unix timestamp indicating the moment the data was stored.
        - An error, if the read operation has not been possible.
* `Delete` removes the cached data for the given key. If the data does not
  exist, the system does not return any error.
    - Arguments:
        - `key string`: the key associated to the value to be removed.
    - Return:
        - An error if the deletion has not been possible (f the data does not
          exist, the system does not return any error).
* `Save` persists all the data in the disk.
    - Return:
        - An error if the save operation has not been possible.

## Restrictions

Because of the way the data is serialized on disk (we rely on Go standard
library JSON encoding/decoding), the next restrictions may apply:

* Any stored number must be deserialized into an `int64` primitive type.
* You can store any type of map, but the destination map for the `Get`
  function must be of type `map[string]interface{}`.
* The private fields of structs won't be neither stored nor retrieved, so
  you must be sure all the persisted fields are public (their name start
  with capital letter).

## Example

The [Pizza ordering example](examples/persist.go) shows the basic operation of the `persist.Storer` interface:

* Creation through the `NewFileStore` function.
* Normal operation through the `Get`/`Set`/`Delete` functions.
* Persistence of data by means of `Save` function.

The [Pizza ordering example](examples/persist.go) uses persistence functionality with a complete structure:

```go
type Topping string
type Mass string

type Pizza struct {
	Toppings []Topping
	Mass Mass
	Slices int
}
```

1. The example program first instantiates the file-backed `Storer` implementation:
   ```go
   storer, err := persist.NewFileStore(
                            persist.DefaultPath("pizza"),
                            log.NewStdErr(false),
                            persist.DefaultTTL)   
   ```

2. After that, it looks for pizza into the `Storer`.
   ```go
   var pizza Pizza
   timestamp, err := storer.Get("last-dinner", &pizza)   
   ```

3. If there is no pizza, it creates a new one.
   ```go
   if err == persist.ErrNotFound {
		pizza = Pizza {
			Toppings : []Topping{"Pepperoni", "Mozzarella", "Cheese"},
			Mass: "thin",
			Slices: 4,
		}   
   ```

4. Once it has a pizza, it consumes a `Slices` unit and stores the updated `Pizza`...
   ```go
   pizza.Slices--
   /* (...) */
   storer.Set("last-dinner", pizza)
   ```

5. ... or deletes it when no more slices are available.
   ```go
   storer.Delete("last-dinner")   
   ```

6. At the end, the changes are persisted into the disk.
   ```go
   storer.Save()
   ```

Executing repeatedly the [Pizza ordering go file](examples/persist.go) would show the next output:

```
examples $ go run persist.go
No pizza in the fridge. Ordering more...
eating a delicious slice of pizza

examples $ go run persist.go
I found some pizza in the storage: {Toppings:[Pepperoni Mozzarella Cheese] Mass:thin Slices:3} (stored at 1527082597)
eating a delicious slice of pizza

examples $ go run persist.go
I found some pizza in the storage: {Toppings:[Pepperoni Mozzarella Cheese] Mass:thin Slices:2} (stored at 1527082600)
eating a delicious slice of pizza

examples $ go run persist.go
I found some pizza in the storage: {Toppings:[Pepperoni Mozzarella Cheese] Mass:thin Slices:1} (stored at 1527082602)
eating a delicious slice of pizza
No more pizza on the fridge. Deleting...

examples $ go run persist.go
No pizza in the fridge. Ordering more...
eating a delicious slice of pizza
```


