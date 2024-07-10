package bts

var _singleTemplate = `
// NAME {{or .Comment "get data from cache if miss will call source method, then add to cache."}} 
func (repo *{{.StructName}}) NAME(c context.Context, {{.IDName}} KEY{{.ExtraArgsType}}) (res VALUE, err error) {
	addCache := true
	res, err = CACHEFUNC(c, {{.IDName}} {{.ExtraCacheArgs}})
	if err != nil {
		if err == redis.Nil {
			err = nil
		} else {
			{{if .CacheErrContinue}}
			addCache = false
			repo.log.WithContext(c).Errorf("CACHEFUNC: %+v, {{.IDName}}: %+v", err, {{.IDName}})
			err = nil
			{{else}}
			return
			{{end}}
		}
	}
	{{if .EnableNullCache}}
	defer func() {
		{{if .SimpleValue}} if res == {{.NullCache}} { {{else}} if {{.CheckNullCode}} { {{end}}
			res = {{.ZeroValue}}
		}
	}()
	{{end}}
	{{if .GoValue}}
	if len(res) != 0 {
	{{else}}
	if res != {{.ZeroValue}} {
	{{end}}
	repo.data.m.CacheHits.With("bts:NAME").Inc()
		return
	}
	{{if .EnablePaging}}
	var miss VALUE
	{{end}}
	{{if .EnableSingleFlight}}
		fetchRaw := false
		var rr interface{}
		sf := cacheSFNAME({{.IDName}} {{.ExtraArgs}})
		rr, err, _ = cacheSingleFlights[SFNUM].Do(sf, func() (r interface{}, e error) {
			repo.data.m.CacheMisses.With("bts:NAME").Inc()
			{{if .EnablePaging}}
				var rrs [2]interface{}
				rrs[0], rrs[1], e = RAWFUNC(c, {{.IDName}} {{.ExtraRawArgs}})
				r = rrs
			{{else}}
				r, e = RAWFUNC(c, {{.IDName}} {{.ExtraRawArgs}})
			{{end}}
			fetchRaw = true
			return
		})
		{{if .EnablePaging}}
			res = rr.([2]interface{})[0].(VALUE)
			miss = rr.([2]interface{})[1].(VALUE)
		{{else}}
			res = rr.(VALUE)
		{{end}}
	{{else}}
		repo.data.m.CacheMisses.With("bts:NAME").Inc()
		{{if .EnablePaging}}
		res, miss, err = RAWFUNC(c, {{.IDName}} {{.ExtraRawArgs}})
		{{else}}
		res, err = RAWFUNC(c, {{.IDName}} {{.ExtraRawArgs}})
		{{end}}
	{{end}}
	if err != nil {
		return
	}
	if !addCache {
		return
	}
	{{if .EnableSingleFlight}}
		if !fetchRaw {
			return
		}
	{{end}}
	{{if .EnablePaging}}
	{{else}}
		miss := res
	{{end}}
	{{if .EnableNullCache}}
		{{if .GoValue}}
		if len(miss) == 0 {
		{{else}}
		if miss == {{.ZeroValue}} {
		{{end}}
		miss = {{.NullCache}}
	}
	{{end}}
	{{if .Sync}}
		ace := ADDCACHEFUNC(c, {{.IDName}}, miss {{.ExtraAddCacheArgs}})
		if ace != nil {
			repo.log.WithContext(c).Errorf("ADDCACHEFUNC: %+v, {{.IDName}}: %+v, miss: %+v", ace, {{.IDName}}, miss)
		}
	{{else}}
		cfe := repo.data.cf.Do(c, func(c context.Context) {
			ace := ADDCACHEFUNC(c, {{.IDName}}, miss {{.ExtraAddCacheArgs}})
			if ace != nil {
				repo.log.WithContext(c).Errorf("ADDCACHEFUNC: %+v, {{.IDName}}: %+v, miss: %+v", ace, {{.IDName}}, miss)
			}
		})
		if cfe != nil {
			repo.log.WithContext(c).Errorf("cache fanout ADDCACHEFUNC: %+v", cfe)
		}
	{{end}}
	return
}
`
