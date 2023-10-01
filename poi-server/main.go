package main

import (
	"context"
	"flag"
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

type Feature struct {
	Type       string `json:"type"`
	Properties struct {
		Names struct {
			Common []struct {
				Value    string `json:"value"`
				Language string `json:"language"`
			} `json:"common"`
		} `json:"names"`
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
		println("initFile took", elapsed.String())
	}()
	jsonFile, err := os.ReadFile(p)
	must(err)
	f := &PlacesFile{}
	err = sonic.Unmarshal(jsonFile, f)
	must(err)
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

func bindAndValidate(ctx *app.RequestContext) (*NearbyRequest, error) {
	q := &NearbyRequest{}
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

// NearbyResponse is a GeoJSON format result.
type NearbyResponse struct {
	Type     string     `json:"type"`
	Name     string     `json:"name"`
	Features []*Feature `json:"features"`
}

type Searcher interface {
	Nearby(lng float64, lat float64, count int) []*Feature
	Name() string
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

func newServer(s Searcher, cfg ...config.Option) *server.Hertz {
	h := server.New(cfg...)
	h.GET("/nearby", func(c context.Context, ctx *app.RequestContext) {
		q, err := bindAndValidate(ctx)
		if err != nil {
			ctx.JSON(http.StatusUnprocessableEntity, utils.H{
				"error": err.Error(),
				"query": string(ctx.Request.QueryString()),
			})
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
