package bts

var _noneTemplate = `
// NAME {{or .Comment "get data from cache if miss will call source method, then add to cache."}} 
func (repo *{{.StructName}}) NAME(c context.Context) (res VALUE, err error) {
	addCache := true
	res, err = CACHEFUNC(c)
	if err != nil {
		if err == redis.Nil {
			err = nil
		} else {
			{{if .CacheErrContinue}}
			addCache = false
			repo.log.WithContext(c).Errorf("CACHEFUNC: %+v", err)
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
	{{if .EnableSingleFlight}}
		fetchRaw := false
		var rr interface{}
		sf := cacheSFNAME()
		rr, err, _ = cacheSingleFlights[SFNUM].Do(sf, func() (r interface{}, e error) {
			repo.data.m.CacheMisses.With("bts:NAME").Inc()
			r, e = RAWFUNC(c)
			fetchRaw = true
			return
		})
		res = rr.(VALUE)
	{{else}}
		repo.data.m.CacheMisses.With("bts:NAME").Inc()
		res, err = RAWFUNC(c)
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
	var miss = res
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
		ace := ADDCACHEFUNC(c, miss)
		if ace != nil {
			repo.log.WithContext(c).Errorf("ADDCACHEFUNC: %+v, miss: %+v", ace, miss)
		}
	{{else}}
		cfe := repo.data.cf.Do(c, func(c context.Context) {
			ace := ADDCACHEFUNC(c, miss)
			if ace != nil {
				repo.log.WithContext(c).Errorf("ADDCACHEFUNC: %+v, miss: %+v", ace, miss)
			}
		})
		if cfe != nil {
			repo.log.WithContext(c).Errorf("cache fanout ADDCACHEFUNC: %+v", cfe)
		}
	{{end}}
	return
}
`
