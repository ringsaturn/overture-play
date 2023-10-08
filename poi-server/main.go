package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/tidwall/rtree"
)

const (
	featureCollectionType = "FeatureCollection"
)

// Wait upstream update all enums.
//
// See:
//   - https://github.com/OvertureMaps/schema/blob/v0.4.0/schema/places/place.yaml#L26
//   - https://github.com/OvertureMaps/schema/blob/main/task-force-docs/places/overture_categories.csv
type Category string

type Feature struct {
	Type       string `json:"type"`
	Properties struct {
		ID         string `json:"id"`
		UpdateTime string `json:"updatetime"`
		Version    int    `json:"version"`
		Names      struct {
			Common []struct {
				Value    string `json:"value"`
				Language string `json:"language"`
			} `json:"common"`
		} `json:"names"`
		Categories struct {
			Main      Category   `json:"main"`
			Alternate []Category `json:"alternate"`
		} `json:"categories"`
		Confidence float64  `json:"confidence"`
		Websites   []string `json:"websites"`
		Socials    []string `json:"socials"`
		Emails     any      `json:"emails"`
		Phones     []string `json:"phones"`
		Brand      struct {
			Names struct {
				BrandNamesCommon []struct {
					Value    string `json:"value"`
					Language string `json:"language"`
				} `json:"brand_names_common"`
			} `json:"names"`
			Wikidata any `json:"wikidata"`
		} `json:"brand"`
		Addresses []struct {
			Locality string `json:"locality"`
			Postcode string `json:"postcode"`
			Freeform string `json:"freeform"`
			Country  string `json:"country"`
		} `json:"addresses"`
		Sources []struct {
			Dataset  string `json:"dataset"`
			Property string `json:"property"`
			RecordID string `json:"recordid"`
		} `json:"sources"`
	} `json:"properties"`
	Geometry struct {
		Type        string     `json:"type"`
		Coordinates [2]float64 `json:"coordinates"`
	} `json:"geometry"`
}

type PlacesFile struct {
	Type     string     `json:"type"`
	Name     string     `json:"name"`
	Features []*Feature `json:"features"`
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func newFile(p string) *PlacesFile {
	start := time.Now()
	defer func() {
		elapsed := time.Since(start)
		log.Println("initFile took", elapsed.String())
	}()
	jsonFile, err := os.ReadFile(p)
	must(err)
	f := &PlacesFile{}
	must(sonic.Unmarshal(jsonFile, f))
	return f
}

func newRTree(f *PlacesFile) *rtree.RTreeGN[float64, *Feature] {
	tr := &rtree.RTreeGN[float64, *Feature]{}
	for _, feature := range f.Features {
		tr.Insert(feature.Geometry.Coordinates, feature.Geometry.Coordinates, feature)
	}
	return tr
}

type NearbyRequest struct {
	Lng   float64 `query:"lng" vd:"$>=-180 && $<=180"`
	Lat   float64 `query:"lat" vd:"$>=-90 && $<=90"`
	Count int     `query:"count" vd:"$>0 && $<=50"`
}

// NearbyResponse is a GeoJSON format result.
type NearbyResponse struct {
	Type     string     `json:"type"`
	Name     string     `json:"name"`
	Features []*Feature `json:"features"`
}

// Searcher is a interface for searching nearby places.
type Searcher interface {
	Name() string
	Nearby(lng float64, lat float64, count int) []*Feature
}

type searcher struct {
	f  *PlacesFile
	tr *rtree.RTreeGN[float64, *Feature]
}

func newSearcher(f *PlacesFile, tr *rtree.RTreeGN[float64, *Feature]) Searcher {
	return &searcher{f: f, tr: tr}
}

func (s *searcher) Name() string {
	return s.f.Name
}

func (s *searcher) Nearby(lng float64, lat float64, count int) []*Feature {
	res := []*Feature{}
	p := [2]float64{lng, lat}
	s.tr.Nearby(
		rtree.BoxDist(
			p,
			p,
			func(min, max [2]float64, data *Feature) float64 {
				lngDiff := lng - data.Geometry.Coordinates[0]
				latDiff := lat - data.Geometry.Coordinates[1]
				return lngDiff*lngDiff + latDiff*latDiff
			}),
		func(min, max [2]float64, data *Feature, dist float64) bool {
			res = append(res, data)
			return len(res) < count
		},
	)
	return res
}

func bindAndValidate[T any](_ context.Context, ctx *app.RequestContext) (*T, error) {
	q := new(T)
	err := ctx.BindQuery(q)
	if err != nil {
		return nil, err
	}
	err = ctx.Validate(q)
	if err != nil {
		return nil, err
	}
	return q, nil
}

func handlerUnprocessableEntity(c context.Context, ctx *app.RequestContext, err error) {
	ctx.JSON(http.StatusUnprocessableEntity, utils.H{
		"error": err.Error(),
		"query": string(ctx.Request.QueryString()),
	})
}

func newServer(s Searcher, cfg ...config.Option) *server.Hertz {
	h := server.New(cfg...)
	h.GET("/nearby", func(c context.Context, ctx *app.RequestContext) {
		q, err := bindAndValidate[NearbyRequest](c, ctx)
		if err != nil {
			handlerUnprocessableEntity(c, ctx, err)
			ctx.Abort()
			return
		}
		ctx.JSON(http.StatusOK, &NearbyResponse{
			Type:     featureCollectionType,
			Name:     s.Name(),
			Features: s.Nearby(q.Lng, q.Lat, q.Count),
		})
	})
	return h
}

func main() {
	placesFilePath := flag.String("places-file", "", "GeoJSON file of places")
	flag.Parse()
	f := newFile(*placesFilePath)
	tr := newRTree(f)
	s := newSearcher(f, tr)
	h := newServer(s)
	h.Run()
}
