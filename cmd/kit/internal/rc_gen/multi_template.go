package rc

import (
	"strings"
)

var _multiGetTemplate = `
// NAME {{or .Comment "get data from redis cache"}} 
func (repo *{{.StructName}}) NAME(c context.Context, ids []KEY {{.ExtraArgsType}}) (res map[KEY]VALUE, err error) {
    l := len(ids)
	if l == 0 {
		return
	}
	{{if .EnableBatch}}
		mutex := sync.Mutex{}
		for i:=0;i < l; i+= GROUPSIZE * MAXGROUP {
			var subKeys []KEY
			{{if .BatchErrBreak}}
				group, ctx := errgroup.WithContext(c)
			{{else}}
				group := &errgroup.Group{}
				ctx := c
			{{end}}
			if (i + GROUPSIZE * MAXGROUP) > l {
				subKeys = ids[i:]
			} else {
				subKeys = ids[i : i+GROUPSIZE * MAXGROUP]
			}
			subLen := len(subKeys)
			for j:=0; j< subLen; j += GROUPSIZE {
				var ks []KEY
				if (j+GROUPSIZE) > subLen {
					ks = subKeys[j:]
				} else {
					ks = subKeys[j:j+GROUPSIZE]
				}
				group.Go(func() (err error) {
					keys := make([]string, 0, len(ks))
					for _, id := range ks {
						key := {{.KeyMethod}}(id{{.ExtraArgs}})
						keys = append(keys, key)
					}
					replies, err := repo.data.rc.MGet(c, keys...).Result()
					if err != nil {
						return
					}
					for idx, reply := range replies {
						{{if .GetSimpleValue}}
							v, ok := reply.(string)
							if !ok {
								continue
							}
							r, err := {{.ConvertBytes2Value}}
							if err != nil {
								return
							}
						{{else}}
							{{if .GetDirectValue}}
								{{if eq .ValueType "[]byte"}}
									v, ok := reply.(string)
									if !ok {
										continue
									}
									r := []byte(v)
								{{else}}
									r, ok := reply.(string)
									if !ok {
										continue
									}
								{{end}}
							{{else}}
								{{if .PointType}}
									r := &{{.OriginValueType}}{}
									err = json.Unmarshal(reply, r)
								{{else}}
									r := {{.OriginValueType}}{}
									err = json.Unmarshal(reply, &r)
								{{end}}
								if err != nil {
									return
								}
							{{end}}
						{{end}}
						mutex.Lock()
						if res == nil {
							res = make(map[KEY]VALUE, len(keys))
						}
						res[ids[idx]] = r
						mutex.Unlock()
					}
				return
				})
			}
			err1 := group.Wait()
			if err1 != nil {
				err = err1
			{{if .BatchErrBreak}}
				break
			{{end}}
			}
		}
	{{else}}
		keys := make([]string, 0, l)
		for _, id := range ids {
			key := {{.KeyMethod}}(id{{.ExtraArgs}})
			keys = append(keys, key)
		}
		replies, err := repo.data.rc.MGet(c, keys...).Result()
		if err != nil {
			return
		}
		for idx, reply := range replies {
			{{if .GetSimpleValue}}
				v, ok := reply.(string)
				if !ok {
					continue
				}
				r, err := {{.ConvertBytes2Value}}
				if err != nil {
					return
				}
			{{else}}
				{{if .GetDirectValue}}
					{{if eq .ValueType "[]byte"}}
						v, ok := reply.(string)
						if !ok {
							continue
						}
						r := []byte(v)
					{{else}}
						r, ok := reply.(string)
						if !ok {
							continue
						}
					{{end}}
				{{else}}
					{{if .PointType}}
						r := &{{.OriginValueType}}{}
						err = json.Unmarshal(reply, r)
					{{else}}
						r := {{.OriginValueType}}{}
						err = json.Unmarshal(reply, &r)
					{{end}}
					if err != nil {
						return
					}
				{{end}}
			{{end}}
			if res == nil {
				res = make(map[KEY]VALUE, len(keys))
			}
			res[ids[idx]] = r
		}
	{{end}}
	return
}
`

var _multiSetTemplate = `
// NAME {{or .Comment "Set data to redis cache"}} 
func (repo *{{.StructName}}) NAME(c context.Context, values map[KEY]VALUE {{.ExtraArgsType}}) (err error) {
	if len(values) == 0 {
		return
	}
	vs := make(map[KEY][]byte, len(values))
	for id, val := range values {
		key := {{.KeyMethod}}(id{{.ExtraArgs}})
		{{if .SimpleValue}}
			bs := {{.ConvertValue2Bytes}}
		{{else}}
			bs, err := json.Marshal(val)
			if err != nil {
				return
			}
		{{end}}
		vs[key] = bs
	}
	return repo.data.rc.MSet(c, vs).Err()
}
`

var _multiAddTemplate = strings.Replace(_multiSetTemplate, "MSet", "MSetNX", -1)

var _multiDelTemplate = `
// NAME {{or .Comment "delete data from redis cache"}} 
func (repo *{{.StructName}}) NAME(c context.Context, ids []KEY {{.ExtraArgsType}}) (err error) {
	if len(ids) == 0 {
		return
	}
	for _, id := range ids {
		key := {{.KeyMethod}}(id{{.ExtraArgs}})
		if err = repo.data.rc.Del(c, key).Err(); err != nil {
			if err == redis.nil {
				err = nil
				continue
			}
			return
		}
	}
	return
}
`
