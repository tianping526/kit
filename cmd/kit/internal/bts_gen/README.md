#### bts

> cache aside代码生成

##### 项目简介

从缓存中获取数据 如果miss则调用回源函数从数据源获取 然后塞入缓存

支持以下功能:

- 单飞限制回源并发 防止打爆数据源
- 空缓存 防止缓存穿透
- 分批获取数据 降低延时
- 默认异步加缓存 可选同步加缓存
- prometheus回源比监控
- 多行注释生成代码
- 支持分页(限单key模板)
- 自定义注释
- 支持忽略参数

##### 使用方式:
1. 在data package中 增加注解 //go:generate tool bts 定义bts接口 声明需要的方法
2. 在data 文件夹中执行 go generate命令 将会生成相应的缓存代码
3. 调用生成的XXX方法
4. 示例:
```go
//go:generate kit bts
type _bts interface {
	// Demos bts: -bts_batch=2 -bts_max_group=20 -bts_batch_err=break -null_cache=&Demo{ID:-1} -bts_check_null_code=$.ID==-1
	Demos(c context.Context, keys []int64) (map[int64]*Demo, error)
	// Demos1 bts: -bts_batch=2 -bts_max_group=20 -bts_batch_err=continue -null_cache=&Demo{ID:-1} -bts_check_null_code=$.ID==-1
	Demos1(c context.Context, keys []int64) (map[int64]*Demo, error)
	// Demo bts: -sync=true -null_cache=&Demo{ID:-1} -bts_check_null_code=$.ID==-1
	Demo(c context.Context, key int64) (*Demo, error)
	// Demo1 bts: -paging=true
	Demo1(c context.Context, key int64, pn int, ps int) (*Demo, error)
	// None bts: -null_cache=&Demo{ID:-1} -bts_check_null_code=$.ID==-1
	None(c context.Context) (*Demo, error)
}
```

要求:
repo里面需要有cache对象 代码会调用repo.cache来新增缓存
需要实现代码中所需的方法 每一个缓存方法都需要实现以下方法:
从缓存中获取数据 名称为Cache+方法名 函数定义和声明一致
从数据源(db/api/...)获取数据 名称为Raw+方法 函数定义和声明一致
存入缓存方法  名称为AddCache+方法名 函数定义为 func AddCache方法名(c context.Context, ...) (error)

##### 注解参数:
| 参数名称                 | 默认值      | 说明                                                                       | 示例                                                |
|----------------------|----------|--------------------------------------------------------------------------|---------------------------------------------------|
| -null_cache          |          | 空指针对象(存正常业务不会出现的内容 id的话像是-1这样的)                                          | &Demo{ID:-1} 或-1 或"null"                          |
| -bts_check_null_code |          | 开启空缓存并且value为指针对象时必填 用于判断是否是空缓存 $来指代对象名                                  | `-bts_check_null_code=$!=nil&&$.ID==-1  或  $ == -1` |
| -cache_err           | continue | 缓存出错的时候的行为 continue: 继续执行 break: 抛出错误 方法返回                               | break                                             |
| -bts_batch           |          | (限多key模板) 批量获取数据 每组大小                                                    | 100                                               |
| -bts_max_group       |          | (限多key模板)批量获取数据 最大组数量                                                    | 10                                                |
| -bts_batch_err       | break    | (限多key模板)批量获取数据回源错误的时候 降级继续请求(continue)还是直接返回(break)                     | break 或 continue                                  |
| -singleflight        | false    | 是否开启单飞（开启后生成函数会多一个单飞名称参数 生成的代码会调用cacheSFNAME方法获取单飞的key）开启单飞后，批量相关功能会被禁用。 | true                                              |
| -sync                | false    | 是否同步增加缓存                                                                 | false                                             |
| -paging              | false    | (限单key模板)分页 数据源应返回2个值 第一个为对外数据 第二个为全量数据 用于新增缓存                           | false                                             |
| -ignores             |          | 用于依赖的三个方法参数和主方法参数不一致的情况. 忽略方法的某些参数 用 &#124; 分隔方法逗号分隔参数                   | pn,ps&#124;pn&#124;origin 表示"缓存获取"方法忽略pn,ps两个参数 回源方法忽略pn参数 加缓存方法忽略origin参数 |
| -custom_method       | false    | 自定义方法名 &#124; 分隔 缓存获取方法名&#124;回源方法名&#124;增加缓存方法名                         | repo.rc.Demo&#124;repo.db.Demo&#124;repo.rc.AddDemo         |
| -bts_struct_name     | repo     | 所属结构体名称                                                                  | Repo                                              |