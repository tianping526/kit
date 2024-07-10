package rc

var _singleGetTemplate = `
// NAME {{or .Comment "get data from redis cache"}} 
func (repo *{{.StructName}}) NAME(c context.Context, id KEY {{.ExtraArgsType}}) (res VALUE, err error) {
	key := {{.KeyMethod}}(id{{.ExtraArgs}})
	{{if .GetSimpleValue}}
		err = repo.data.rc.Get(c, key).Scan(&res)
		return
	{{else}}
		{{if .GetDirectValue}}
			err = repo.data.rc.Get(c, key).Scan(&res)
			return
		{{else}}
			var val []byte
			val, err = repo.data.rc.Get(c, key).Bytes()
			if err != nil {
				return
			}
			{{if .PointType}}
				res = &{{.OriginValueType}}{}
				err = json.Unmarshal(val, res)
				if err != nil {
					res = nil
				}
			{{else}}
				err = json.Unmarshal(val, &res)
			{{end}}
			return
		{{end}}
	{{end}}
}
`

var _singleSetTemplate = `
// NAME {{or .Comment "Set data to redis cache"}} 
func (repo *{{.StructName}}) NAME(c context.Context, id KEY, val VALUE{{.ExtraArgsType}}) (err error) {
	{{if .PointType}}
      if val == nil {
        return 
      }
	{{end}}
	{{if .LenType}}
      if len(val) == 0 {
        return 
      }
	{{end}}
	key := {{.KeyMethod}}(id{{.ExtraArgs}})
	{{if ne .CasCode ""}}
		verKey := fmt.Sprintf("%s:version", key)
	{{end}}
	{{if .SimpleValue}}
		bs := {{.ConvertValue2Bytes}}
	{{else}}
		bs, err := json.Marshal(val)
		if err != nil {
			return
		}
	{{end}}
		expire := {{.ExpireCode}}()
	{{if .EnableNullCode}}
		if {{.CheckNullCode}} {
			expire = {{.ExpireNullCode}}()
		}
	{{end}}
	{{if ne .CasCode ""}}
		ver := {{.CasVersionCode}}(val)
		return {{.CasCode}}.Run(c, repo.data.rc, []string{key, verKey}, bs, ver, expire).Err()
	{{else}}
		expireDuration := time.Duration(expire) * time.Second
		return repo.data.rc.Set(c, key, bs, expireDuration).Err()
	{{end}}
}
`

var _singleAddTemplate = `
// NAME {{or .Comment "Set data to redis cache"}} 
func (repo *{{.StructName}}) NAME(c context.Context, id KEY, val VALUE {{.ExtraArgsType}}) (err error) {
	{{if .PointType}}
      if val == nil {
        return 
      }
	{{end}}
	{{if .LenType}}
      if len(val) == 0 {
        return 
      }
	{{end}}
	key := {{.KeyMethod}}(id{{.ExtraArgs}})
	{{if .SimpleValue}}
		bs := {{.ConvertValue2Bytes}}
	{{else}}
		bs, err := json.Marshal(val)
		if err != nil {
			return
		}
	{{end}}
		expire := {{.ExpireCode}}()
	{{if .EnableNullCode}}
		if {{.CheckNullCode}} {
			expire = {{.ExpireNullCode}}()
		}
	{{end}}
	expireDuration := time.Duration(expire) * time.Second
	return repo.data.rc.SetNX(c, key, bs, expire).Err()
}
`

var _singleDelTemplate = `
// NAME {{or .Comment "delete data from redis cache"}} 
func (repo *{{.StructName}}) NAME(c context.Context, id KEY {{.ExtraArgsType}}) (err error) {
	key := {{.KeyMethod}}(id{{.ExtraArgs}})
	return repo.data.rc.Del(c, key).Err()
}
`
