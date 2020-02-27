// Code generated by $GOPATH/src/go-common/app/tool/cache/gen. DO NOT EDIT.

/*
  Package dao is a generated cache proxy package.
  It is generated from:
  type _cache interface {
		// cache: -nullcache=&model.Item{ID:-1} -check_null_code=$!=nil&&$.ID==-1
		Items(c context.Context, pid []int64) (info map[int64]*model.Item, err error)
		// cache: -nullcache=&model.ItemDetail{ProjectID:-1} -check_null_code=$!=nil&&$.ProjectID==-1
		ItemDetails(c context.Context, pid []int64) (details map[int64]*model.ItemDetail, err error)
		// cache: -nullcache=[]*model.TicketInfo{{TicketPrice:model.TicketPrice{ProjectID:-1}}} -check_null_code=len($)==1&&$[0].ProjectID==-1
		TkListByItem(c context.Context, pid []int64) (info map[int64][]*model.TicketInfo, err error)
		// cache: -nullcache=&model.Venue{ID:-1} -check_null_code=$!=nil&&$.ID==-1
		Venues(c context.Context, id []int64) (venues map[int64]*model.Venue, err error)
		// cache: -nullcache=&model.Place{ID:-1} -check_null_code=$!=nil&&$.ID==-1
		Place(c context.Context, id int64) (place *model.Place, err error)
		// cache: -nullcache=[]*model.Screen{{ProjectID:-1}} -check_null_code=len($)==1&&$[0].ProjectID==-1
		ScListByItem(c context.Context, pid []int64) (info map[int64][]*model.Screen, err error)
		// cache: -nullcache=&model.Screen{ProjectID:-1} -check_null_code=$!=nil&&$.ProjectID==-1
		ScList(c context.Context, sids []int64) (info map[int64]*model.Screen, err error)
		// cache: -nullcache=&model.TicketInfo{TicketPrice:model.TicketPrice{ProjectID:-1}} -check_null_code=$!=nil&&$.ProjectID==-1
		TkList(c context.Context, tids []int64) (info map[int64]*model.TicketInfo, err error)
	}
*/

package dao

import (
	"context"

	"go-common/app/service/openplatform/ticket-item/model"
	"go-common/library/stat/prom"
)

var _ _cache

// Items get data from cache if miss will call source method, then add to cache.
func (d *Dao) Items(c context.Context, keys []int64) (res map[int64]*model.Item, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheItems(c, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	prom.CacheHit.Add("Items", int64(len(keys)-len(miss)))
	for k, v := range res {
		if v != nil && v.ID == -1 {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*model.Item
	prom.CacheMiss.Add("Items", int64(len(miss)))
	missData, err = d.RawItems(c, miss)
	if res == nil {
		res = make(map[int64]*model.Item, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	for _, key := range miss {
		if res[key] == nil {
			missData[key] = &model.Item{ID: -1}
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheItems(c, missData)
	})
	return
}

// ItemDetails get data from cache if miss will call source method, then add to cache.
func (d *Dao) ItemDetails(c context.Context, keys []int64) (res map[int64]*model.ItemDetail, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheItemDetails(c, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	prom.CacheHit.Add("ItemDetails", int64(len(keys)-len(miss)))
	for k, v := range res {
		if v != nil && v.ProjectID == -1 {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*model.ItemDetail
	prom.CacheMiss.Add("ItemDetails", int64(len(miss)))
	missData, err = d.RawItemDetails(c, miss)
	if res == nil {
		res = make(map[int64]*model.ItemDetail, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	for _, key := range miss {
		if res[key] == nil {
			missData[key] = &model.ItemDetail{ProjectID: -1}
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheItemDetails(c, missData)
	})
	return
}

// TkListByItem get data from cache if miss will call source method, then add to cache.
func (d *Dao) TkListByItem(c context.Context, keys []int64) (res map[int64][]*model.TicketInfo, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheTkListByItem(c, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (len(res[key]) == 0) {
			miss = append(miss, key)
		}
	}
	prom.CacheHit.Add("TkListByItem", int64(len(keys)-len(miss)))
	for k, v := range res {
		if len(v) == 1 && v[0].ProjectID == -1 {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64][]*model.TicketInfo
	prom.CacheMiss.Add("TkListByItem", int64(len(miss)))
	missData, err = d.RawTkListByItem(c, miss)
	if res == nil {
		res = make(map[int64][]*model.TicketInfo, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	for _, key := range miss {
		if len(res[key]) == 0 {
			missData[key] = []*model.TicketInfo{{TicketPrice: model.TicketPrice{ProjectID: -1}}}
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheTkListByItem(c, missData)
	})
	return
}

// Venues get data from cache if miss will call source method, then add to cache.
func (d *Dao) Venues(c context.Context, keys []int64) (res map[int64]*model.Venue, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheVenues(c, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	prom.CacheHit.Add("Venues", int64(len(keys)-len(miss)))
	for k, v := range res {
		if v != nil && v.ID == -1 {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*model.Venue
	prom.CacheMiss.Add("Venues", int64(len(miss)))
	missData, err = d.RawVenues(c, miss)
	if res == nil {
		res = make(map[int64]*model.Venue, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	for _, key := range miss {
		if res[key] == nil {
			missData[key] = &model.Venue{ID: -1}
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheVenues(c, missData)
	})
	return
}

// Place get data from cache if miss will call source method, then add to cache.
func (d *Dao) Place(c context.Context, id int64) (res *model.Place, err error) {
	addCache := true
	res, err = d.CachePlace(c, id)
	if err != nil {
		addCache = false
		err = nil
	}
	defer func() {
		if res != nil && res.ID == -1 {
			res = nil
		}
	}()
	if res != nil {
		prom.CacheHit.Incr("Place")
		return
	}
	prom.CacheMiss.Incr("Place")
	res, err = d.RawPlace(c, id)
	if err != nil {
		return
	}
	miss := res
	if miss == nil {
		miss = &model.Place{ID: -1}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCachePlace(c, id, miss)
	})
	return
}

// ScListByItem get data from cache if miss will call source method, then add to cache.
func (d *Dao) ScListByItem(c context.Context, keys []int64) (res map[int64][]*model.Screen, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheScListByItem(c, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (len(res[key]) == 0) {
			miss = append(miss, key)
		}
	}
	prom.CacheHit.Add("ScListByItem", int64(len(keys)-len(miss)))
	for k, v := range res {
		if len(v) == 1 && v[0].ProjectID == -1 {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64][]*model.Screen
	prom.CacheMiss.Add("ScListByItem", int64(len(miss)))
	missData, err = d.RawScListByItem(c, miss)
	if res == nil {
		res = make(map[int64][]*model.Screen, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	for _, key := range miss {
		if len(res[key]) == 0 {
			missData[key] = []*model.Screen{{ProjectID: -1}}
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheScListByItem(c, missData)
	})
	return
}

// ScList get data from cache if miss will call source method, then add to cache.
func (d *Dao) ScList(c context.Context, keys []int64) (res map[int64]*model.Screen, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheScList(c, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	prom.CacheHit.Add("ScList", int64(len(keys)-len(miss)))
	for k, v := range res {
		if v != nil && v.ProjectID == -1 {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*model.Screen
	prom.CacheMiss.Add("ScList", int64(len(miss)))
	missData, err = d.RawScList(c, miss)
	if res == nil {
		res = make(map[int64]*model.Screen, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	for _, key := range miss {
		if res[key] == nil {
			missData[key] = &model.Screen{ProjectID: -1}
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheScList(c, missData)
	})
	return
}

// TkList get data from cache if miss will call source method, then add to cache.
func (d *Dao) TkList(c context.Context, keys []int64) (res map[int64]*model.TicketInfo, err error) {
	if len(keys) == 0 {
		return
	}
	addCache := true
	if res, err = d.CacheTkList(c, keys); err != nil {
		addCache = false
		res = nil
		err = nil
	}
	var miss []int64
	for _, key := range keys {
		if (res == nil) || (res[key] == nil) {
			miss = append(miss, key)
		}
	}
	prom.CacheHit.Add("TkList", int64(len(keys)-len(miss)))
	for k, v := range res {
		if v != nil && v.ProjectID == -1 {
			delete(res, k)
		}
	}
	missLen := len(miss)
	if missLen == 0 {
		return
	}
	var missData map[int64]*model.TicketInfo
	prom.CacheMiss.Add("TkList", int64(len(miss)))
	missData, err = d.RawTkList(c, miss)
	if res == nil {
		res = make(map[int64]*model.TicketInfo, len(keys))
	}
	for k, v := range missData {
		res[k] = v
	}
	if err != nil {
		return
	}
	for _, key := range miss {
		if res[key] == nil {
			missData[key] = &model.TicketInfo{TicketPrice: model.TicketPrice{ProjectID: -1}}
		}
	}
	if !addCache {
		return
	}
	d.cache.Do(c, func(c context.Context) {
		d.AddCacheTkList(c, missData)
	})
	return
}
