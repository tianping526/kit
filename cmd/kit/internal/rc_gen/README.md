
#### rc

> redis 缓存代码生成

##### 项目简介

自动生成redis缓存代码 和缓存回源工具bts配合使用体验更佳
支持以下功能:
- 常用redis命令(get/set/add/delete)
- 多种数据存储格式(json/pb)
- 常用值类型自动转换(int/bool/float...)
- 自定义缓存名称和过期时间
- 记录pkg/error错误栈
- 记录日志trace id
- prometheus错误监控
- 自定义参数个数
- 自定义注释

##### 使用方式:
1. dao.go文件中新增 _rc interface
2. 在data 文件夹中执行 go generate命令 将会生成相应的缓存代码
3. 示例:
```go
//go:generate kit rc
type _rc interface {
	// CacheDemos rc: -key=demoKey
	CacheDemos(c context.Context, keys []int64) (map[int64]*Demo, error)
	// CacheDemo rc: -key=demoKey
	CacheDemo(c context.Context, key int64) (*Demo, error)
	// CacheDemo1 rc: -key=keyMid
	CacheDemo1(c context.Context, key int64, mid int64) (*Demo, error)
	// CacheNone rc: -key=noneKey
	CacheNone(c context.Context) (*Demo, error)
	// CacheString rc: -key=demoKey
	CacheString(c context.Context, key int64) (string, error)

	// AddCacheDemos rc: -key=demoKey -expire=repo.demoExpire -encode=json
	AddCacheDemos(c context.Context, values map[int64]*Demo) error
	// AddCacheDemos2 rc: -key=demo2Key -expire=repo.demoExpire -encode=json
	AddCacheDemos2(c context.Context, values map[int64]*Demo, tp int64) error
	// AddCacheDemo 这里也支持自定义注释 会替换默认的注释
	// rc: -key=demoKey -expire=repo.demoExpire -encode=json|gzip
	AddCacheDemo(c context.Context, key int64, value *Demo) error
	// AddCacheDemo1 rc: -key=keyMid -expire=repo.demoExpire -encode=gob
	AddCacheDemo1(c context.Context, key int64, value *Demo, mid int64) error
	// AddCacheNone rc: -key=noneKey
	AddCacheNone(c context.Context, value *Demo) error
	// AddCacheString rc: -key=demoKey -expire=repo.demoExpire
	AddCacheString(c context.Context, key int64, value string) error

	// DelCacheDemos rc: -key=demoKey
	DelCacheDemos(c context.Context, keys []int64) error
	// DelCacheDemo rc: -key=demoKey
	DelCacheDemo(c context.Context, key int64) error
	// DelCacheDemo1 rc: -key=keyMid
	DelCacheDemo1(c context.Context, key int64, mid int64) error
	// DelCacheNone rc: -key=noneKey
	DelCacheNone(c context.Context) error
}
```

##### 注意:
类型会根据前缀进行猜测
set / add 对应redis方法Set
del 对应redis方法 Delete
get / cache对应redis方法Get
redis Add方法需要用注解 -type=only_add单独指定

#### 注解参数:
| 名称                 | 默认值                 | 可用范围    | 说明                                           | 可选值                 | 示例   |
|--------------------|---------------------| ----------- |----------------------------------------------|---------------------|------|
| encode             | 根据值类型raw或json       | set/add | 数据存储的格式                                      | json/pb             | json |
| type               | 前缀推断                | 全部        | redis方法 set/get/delete...                    | get/set/del/only_add | get 或 set 等 |
| key                | 根据方法名称生成            | 全部        | 缓存key名称                                      | -                   | demoKey |
| expire             | 根据方法名称生成            | 全部        | 缓存过期时间                               | -                   | repo.demoExpire |
| rc_batch           |                     | get(限多key模板) | 批量获取数据 每组大小                            | -                   | 100  |
| rc_max_group       |                     | get(限多key模板) | 批量获取数据 最大组数量                           | -                   | 10   |
| rc_batch_err       | break               | get(限多key模板) | 批量获取数据回源错误的时候 降级继续请求(continue)还是直接返回(break)  | break 或 continue    | continue |
| rc_struct_name     | repo                | 全部        | 用户自定义Repo结构体名称                               |                     | xxxRepo |
| rc_check_null_code |                     | add/set       | (和null_expire配套使用)判断是否是空缓存的代码 用于为空缓存独立设定过期时间 |               | $.ID==-1 或者 $=="-1"等 |
| null_expire        | 300(5分钟)            |add/set| (和check_null_code配套使用)空缓存的过期时间               |                     | repo.nullExpire      |
| cas_code           | ""                  |set| 使用redis lua cas更新值                           |               | "cas"       |
| cas_version_code   | "Key{#name}Version" |set| 使用redis lua cas更新值时当前版本        |                   | "schemaVersion"          |