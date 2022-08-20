package simpletracker

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"log"
	"net/http"
	"strings"
	"sync"
)

type Tracker struct {
	Map             *sync.Map
	Router          *mux.Router
	ReqKeyFunc      func(r *http.Request) (string, error)
	PopulateKeyFunc func(...string) (string, error)
}

type Counter struct {
	Count uint64
}

func New(syncMap *sync.Map, router *mux.Router, reqKeyFunc func(r *http.Request) (string, error), populateKeyFunc func(...string) (string, error)) *Tracker {
	if syncMap == nil {
		syncMap = &sync.Map{}
	}
	if router == nil {
		router = &mux.Router{}
	}

	if reqKeyFunc == nil {
		reqKeyFunc = func(r *http.Request) (string, error) {
			curRoute := mux.CurrentRoute(r)
			curRoutePath, err := curRoute.GetPathTemplate()
			if err != nil {
				return "", err
			}
			curRouteMethod := r.Method
			mapKey := curRoutePath + "_" + curRouteMethod
			return mapKey, nil
		}
	}

	if populateKeyFunc == nil {
		populateKeyFunc = func(parts ...string) (string, error) {
			var holder []string
			for _, n := range parts {
				holder = append(holder, n)
			}
			return strings.Join(holder, "_"), nil
		}
	}

	return &Tracker{
		Map:             syncMap,
		Router:          router,
		ReqKeyFunc:      reqKeyFunc,
		PopulateKeyFunc: populateKeyFunc,
	}
}

func (t *Tracker) TrackerToS() (string, error) {
	m := map[string]interface{}{}
	t.Map.Range(func(key, value interface{}) bool {
		m[fmt.Sprint(key)] = value
		return true
	})

	b, err := json.MarshalIndent(m, "", " ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}

func (t *Tracker) PopulateRoutes() error {
	err := t.Router.Walk(func(route *mux.Route, router *mux.Router, ancestors []*mux.Route) error {
		pathTemplate, err := route.GetPathTemplate()
		if err != nil {
			return err
		}
		methods, err := route.GetMethods()
		if err != nil {
			return err
		}
		mapKeyPart := strings.Join(methods, "_")
		mapKey, err := t.PopulateKeyFunc(pathTemplate, mapKeyPart)
		if err != nil {
			return err
		}
		t.Map.Store(mapKey, Counter{uint64(0)})
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (t *Tracker) SimpleTracker(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mapKey, err := t.ReqKeyFunc(r)
		if err != nil {
			log.Printf("unable to create mapKey, continuing down the chain: %s", err)
			next.ServeHTTP(w, r)
		}
		entry, ok := t.Map.Load(mapKey)
		if !ok {
			log.Printf("Found a route we didn't have a key for... route: %s", mapKey)
			t.Map.Store(mapKey, Counter{uint64(1)})
		} else {
			curCount := entry.(Counter)
			curCount.Count++
			t.Map.Store(mapKey, curCount)
		}

		next.ServeHTTP(w, r)
	})
}
