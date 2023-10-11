package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/ringsaturn/polyf/integration/featurecollection"
)

type Property = map[string]any

func main() {
	fp := "countries.geojson"
	f, err := os.Open(fp)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	b, err := io.ReadAll(f)
	if err != nil {
		panic(err)
	}
	var bf = &featurecollection.BoundaryFile[Property]{}
	err = json.Unmarshal(b, &bf)
	if err != nil {
		panic(err)
	}
	finder, err := featurecollection.Do(bf)
	if err != nil {
		panic(err)
	}
	res, _ := finder.FindOne(-122, 39)
	resBytes, _ := json.Marshal(res)
	fmt.Println(string(resBytes))
}
