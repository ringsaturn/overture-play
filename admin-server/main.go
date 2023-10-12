package main

import (
	"context"
	"flag"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/bytedance/sonic"
	"github.com/cloudwego/hertz/pkg/app"
	"github.com/cloudwego/hertz/pkg/app/server"
	"github.com/cloudwego/hertz/pkg/common/config"
	"github.com/cloudwego/hertz/pkg/common/utils"
	"github.com/ringsaturn/polyf"
	"github.com/ringsaturn/polyf/integration/featurecollection"
)

type Property = map[string]any

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

type LocationRequest struct {
	Lng float64 `query:"lng" vd:"$>=-180 && $<=180"`
	Lat float64 `query:"lat" vd:"$>=-90 && $<=90"`
}

func newServer(f *polyf.F[Property], cfg ...config.Option) *server.Hertz {
	h := server.New(cfg...)
	h.GET("/admin", func(c context.Context, ctx *app.RequestContext) {
		q, err := bindAndValidate[LocationRequest](c, ctx)
		if err != nil {
			handlerUnprocessableEntity(c, ctx, err)
			ctx.Abort()
			return
		}
		res, _ := f.FindAll(q.Lng, q.Lat)
		ctx.JSON(http.StatusOK, utils.H{"data": res})
	})
	return h
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}

func newFile(p string) *featurecollection.BoundaryFile[Property] {
	start := time.Now()
	log.Println("initFile")
	defer func() {
		elapsed := time.Since(start)
		log.Println("initFile took", elapsed.String())
	}()
	jsonFile, err := os.ReadFile(p)
	must(err)
	f := &featurecollection.BoundaryFile[Property]{}
	must(sonic.Unmarshal(jsonFile, f))
	runtime.GC()
	return f
}

func newFinder(f *featurecollection.BoundaryFile[Property]) *polyf.F[Property] {
	finder, err := featurecollection.Do(f)
	must(err)
	runtime.GC()
	return finder
}

func main() {
	// fp := "themes/admins/countries.geojson"
	adminFilePath := flag.String("admin-file", "", "GeoJSON file of places")
	flag.Parse()
	f := newFile(*adminFilePath)
	finder := newFinder(f)
	s := newServer(finder)
	s.Run()
}
