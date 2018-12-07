# Gin 分析
这里的Gin是指Golang语言里Web框架。[Gin](https://github.com/gin-gonic/gin)的使用简单方便，与node的express非常像。


## 如何开始?
![how to start it ?](../assets/BB920F6D-7F4A-4E58-8C03-9D8D3BF3C18F.png)

起点是 调用`gin.Default()`函数生成默认引擎，也就是*Engine结构体。

签名如下：![Engine签名](../assets/how_to_start_it.png)

简单概括里面功能：
- 根据当前环境选择打印warnning信息
- 生成 `*Engine`
- 默认使用 `log`、 `recovery` 中间件
- 返回 `*Engine`

然后使用`Engine`注册一个`/ping`路径 并绑定一个`HandlerFunc`.

最后 `Run(:addr)`启动整个引擎

## 那故事应该从`Engine`说起了！
![Engine](../assets/engine_struct.png)

以下几个比较重要：
- `RouterGroup` 管理路由组，实现`IRoutes`以及`IRouter`接口
- `HTMLRender` 模版引擎，默认使用golang下`template`作为模版引擎
- `trees`      存储`路径`以及对应`HandlerChain`详细信息
- `noRoute`、`noMethod` 默认404，405处理handler 可以通过`engine.NoRoute(...HandlerChain)``engine.Method(...HandlerChain)`设置

## 先说 `RouterGroup`
`RouterGroup`在`Engine` 里是内嵌结构体，它是这个样子的：
![RouterGroup](../assets/RouterGroup.png)

里面比较重要是`HandlerChain`:

![HandlerChain](../assets/HandlerChain.png)

`RouterGroup`实现了一下接口：

![IRoute](../assets/IRoute.png)

因为`RouterGroup`是内嵌在`Engine`里的所以 `Engine`具备了`Use`,`Get`等方法。

这些方法本质都是在生成`HandlerChain`和对应路径 保存在 trees里面
看下面例子：

### Example
```golang
    r := gin.New()
	r.Use(HandlerFn1, HandlerFn2)
	r.Get("/",HandlerFn3,HandlerFn4,HandlerIndex)
    admin := r.Group("/admin",HandlerAdminLimit1,HandlerAdminLimit2)
        .Get("/money",HandlerMoneyLimit,HandlerMoney)
        .Get("/vote",HanlderVoteLimit,HandlerVote)
        .Get("/email",HanlderEmailLimit,HandlerEmail)
```

#### `Use 使用`
```golang
func (group *RouterGroup) Use(middleware ...HandlerFunc) IRoutes {
	group.Handlers = append(group.Handlers, middleware...)
	return group.returnObj()
}
```
最初`engine.Handlers`是空的,添加了`HandlerFn1`,`HandlerFn2`后=>[HandlerFn1,HandlerFn2].`basePath`为‘/’


#### 当执行了`r.Get("/",HandlerFn3,HandlerFn4,HandlerIndex)`
```golang
// GET is a shortcut for router.Handle("GET", path, handle).
func (group *RouterGroup) GET(relativePath string, handlers ...HandlerFunc) IRoutes {
	return group.handle("GET", relativePath, handlers)
}

func (group *RouterGroup) handle(httpMethod, relativePath string, handlers HandlersChain) IRoutes {
	absolutePath := group.calculateAbsolutePath(relativePath)
	handlers = group.combineHandlers(handlers)
	group.engine.addRoute(httpMethod, absolutePath, handlers)
	return group.returnObj()
}

func (group *RouterGroup) combineHandlers(handlers HandlersChain) HandlersChain {
	finalSize := len(group.Handlers) + len(handlers)
	if finalSize >= int(abortIndex) {
		panic("too many handlers")
	}
	mergedHandlers := make(HandlersChain, finalSize)
	copy(mergedHandlers, group.Handlers)
	copy(mergedHandlers[len(group.Handlers):], handlers)
	return mergedHandlers
}

```
拼接`RouterGroup`路径和HandlerChain,存在`trees` 里面


```javascript
{
    path:"/",
    handlers:[HandlerFn1,HandlerFn2,HandlerFn3,HandlerFn4,HandlerIndex]
}
```
将这样的信息（还有其他附加信息）添加到`engine.trees`里。

#### 当执行`   admin := r.Group("/admin",HandlerAdminLimit1,HandlerAdminLimit2)`

`admin` 这个新的`RouterGroup`的信息是这样

```javascript
{
    basePath:"/admin",
    Handlers:[HandlerFn1,HandlerFn2,HandlerAdminLimit1,HandlerAdminLimit2]
}
```

#### 执行了`.Get("/money",HandlerMoneyLimit,HandlerMoney)`


将`admin`里面的`basePath`拼接当前路径
将`admin`里面的`Handlers`以及当前传入的`HandlerChain`copy到新的`HandlerChain`下


```javascript
{
    path:"/admin/money",
    handlers:[HandlerFn1,HandlerFn2,HandlerAdminLimit1,HandlerAdminLimit2,HandlerMoneyLimit,HandlerMoney]
}
```
将这样的信息（还有其他附加信息）添加到`engine.trees`里,下面以此类推。

### 这些信息如何使用呢？

 当客户端请求“/admin/money”,`Engine`在`tree`里搜索返回`[HandlerFn1,HandlerFn2,HandlerAdminLimit1,HandlerAdminLimit2,HandlerMoneyLimit,HandlerMoney]`所有这些`handler`,使用一个`*Context`上下文
 ```golang
 // ServeHTTP conforms to the http.Handler interface.
func (engine *Engine) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	c := engine.pool.Get().(*Context)
	c.writermem.reset(w)
	c.Request = req
	c.reset()

	engine.handleHTTPRequest(c)

	engine.pool.Put(c)
}

func (c *Context) Next() {
	c.index++
	s := int8(len(c.handlers))
	for ; c.index < s; c.index++ {
		c.handlers[c.index](c)
	}
}
 ```

 `Context`可以中断`HandlerChain`的执行。比如，用户权限未认证。

 做到极致，我们可以这样：[gin 代码](../src/gin/demo01/main.go)

 ```golang
package main

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
)

const (
	ERROR_AUTH_LIMIT = "用户权限认证失败！"

	STORE = "fns"
)

// UnitFunc 单位函数
type UnitFunc func(m *sync.Map)

// UnitFuncs 单位函数集合
type UnitFuncs []UnitFunc

func (u *UnitFuncs) Push(item UnitFunc) *UnitFuncs {
	*u = append(*u, item)
	return u
}

// Protocol 协议层
type Protocol struct {
	Code int         `json:"code"`
	Msg  string      `json:"msg"`
	Data interface{} `json:"data"`
}

// SetData 设置内容
func (p *Protocol) SetData(data interface{}) *Protocol {
	p.Data = data
	return p
}

func (p *Protocol) Error(code int, msg string) *Protocol {
	p.Code = code
	p.Msg = msg
	return p
}

type Auth struct {
	Auth string `json:"auth"`
}

func (a *Auth) Valide() bool {
	return a.Auth != ""
}

// AuthLimit 认证
func AuthLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		//  这里处理业务需求
		fmt.Println("this is in AuthLimit")
		if c.Query("auth") == "" {
			c.AbortWithStatusJSON(http.StatusOK, (&Protocol{}).Error(http.StatusOK, ERROR_AUTH_LIMIT))
		}
		fmt.Println("Abort 也会执行到这里Run here")
	}
}

// News 获取新闻
func News() gin.HandlerFunc {
	return func(c *gin.Context) {
		//  这里处理业务需求
		fmt.Println("this is in News")
		setUnitFn(c, func(m *sync.Map) {
			fmt.Println("this is News UnitFn")
		})
	}
}

// Spot 获取期货
func Spot() gin.HandlerFunc {
	return func(c *gin.Context) {
		//  这里处理业务需求
		fmt.Println("this is in Spot")
		setUnitFn(c, func(m *sync.Map) {
			fmt.Println("this is Spot UnitFn")
		})
	}
}

// Future 获取期货
func Future() gin.HandlerFunc {
	return func(c *gin.Context) {
		//  这里处理业务需求
		fmt.Println("this is in Future")
		setUnitFn(c, func(m *sync.Map) {
			fmt.Println("this is Future UnitFn")
		})
	}
}

// Vote 获取投票
func Vote() gin.HandlerFunc {
	return func(c *gin.Context) {
		//  这里处理业务需求
		fmt.Println("this is in Vote")
		setUnitFn(c, func(m *sync.Map) {
			fmt.Println("this is Vote UnitFn")
		})
	}
}

// Parallel  并行处理
func Parallel(handlers ...gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		fmt.Println("this is in Parallel")
		for _, handler := range handlers {
			handler(c)
		}
		var tasks = getUnitFn(c)
		caption := len(tasks) - len(handlers) + 1

		parallel := tasks[len(tasks)-len(handlers):]
		parallelTask := func(m *sync.Map) {
			var group = &sync.WaitGroup{}
			for i := 0; i < len(parallel); i++ {
				group.Add(1)
				go func(i int) {
					parallel[i](m)
					group.Done()
				}(i)
			}
			group.Wait()
			fmt.Println("Parallel is Done")
		}

		var handlerchain = make(UnitFuncs, caption)
		copy(handlerchain, tasks[:len(tasks)-len(handlers)])
		copy(handlerchain[len(tasks)-len(handlers):], *(&UnitFuncs{}).Push(parallelTask))
		c.Set(STORE, handlerchain)
	}
}

// String  使用HTML 展示数据
func String(text string) gin.HandlerFunc {
	return func(c *gin.Context) {
		//  这里处理业务需求
		fmt.Println("this is in HTML. String is  :", text)
		handlerChain := getUnitFn(c)
		fmt.Println(handlerChain)
		// parallel(handlerChain, &sync.Map{})
		for i := 0; i < len(handlerChain); i++ {
			handlerChain[i](&sync.Map{})
		}
		c.String(http.StatusOK, text)
	}
}

// JSON 使用JSON展示数据
func JSON() gin.HandlerFunc {
	return func(c *gin.Context) {
		//  这里处理业务需求
		fmt.Println("this is in JSON")
		handlerChain := getUnitFn(c)
		// parallel(handlerChain, &sync.Map{})
		for i := 0; i < len(handlerChain); i++ {
			handlerChain[i](&sync.Map{})
		}
		c.JSON(http.StatusOK, (&Protocol{}).SetData("Hello testing."))
	}
}

func main() {
	engine := gin.Default()

	admin := engine.Group("/admin", AuthLimit())
	{
		admin.GET("/index", News(), Spot(), Future(), Vote(), String("index is ok"))
		admin.GET("/parallel", News(), Parallel(Spot(), Future()), Vote(), String("index is ok"))
		admin.GET("/vote", News(), Spot(), Vote(), String("vote is ok alse"))
	}

	api := engine.Group("/api", AuthLimit())
	{
		api.GET("/vote", Vote(), JSON())
	}

	engine.Run(":8080")
}

// utils  工具

func setUnitFn(c *gin.Context, unit UnitFunc) {
	var handlerChain UnitFuncs
	if fns, exites := c.Get(STORE); exites {
		handlerChain = fns.(UnitFuncs)
		handlerChain.Push(unit)
	} else {
		handlerChain = make(UnitFuncs, 0)
		handlerChain.Push(unit)
	}
	c.Set(STORE, handlerChain)
}

func getUnitFn(c *gin.Context) UnitFuncs {
	var handlerChain UnitFuncs
	if fns, exites := c.Get(STORE); exites {
		handlerChain = fns.(UnitFuncs)
	} else {
		handlerChain = make(UnitFuncs, 0)
	}
	return handlerChain
}

// 简单 parallel
func parallel(tasks UnitFuncs, m *sync.Map) {
	var group = &sync.WaitGroup{}
	for i := 0; i < len(tasks); i++ {
		group.Add(1)
		go func(i int) {
			fmt.Println("this index is ", i)
			tasks[i](m)
			group.Done()
		}(i)
	}
	group.Wait()
}

 ```

##  `tree`

 ```golang
type methodTree struct {
	method string
	root   *node
}

type methodTrees []methodTree

type node struct {
	path      string
	wildChild bool
	nType     nodeType
	maxParams uint8
	indices   string
	children  []*node
	handlers  HandlersChain
	priority  uint32
}
 ```

 树形结构  当接受到“/admin/index” 根据node节点一步一步向下匹配，找到对应的HandlerChain执行。

##  HTMLRender

```golang
type HTMLRender interface {
	Instance(string, interface{}) Render
}

type Render interface {
	Render(http.ResponseWriter) error
	WriteContentType(w http.ResponseWriter)
}
```

我们公司使用`pongo2`模版引擎。使用`pongo2gin`来适配粘合两个引擎

```golang
// Package pongo2gin is a template renderer that can be used with the Gin
// web framework https://github.com/gin-gonic/gin it uses the Pongo2 template
// library https://github.com/flosch/pongo2
package pongo2gin

import (
	"net/http"
	"path"

	"github.com/flosch/pongo2"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/render"
)

// RenderOptions is used to configure the renderer.
type RenderOptions struct {
	TemplateDir string
	ContentType string
}

// Pongo2Render is a custom Gin template renderer using Pongo2.
type Pongo2Render struct {
	Options  *RenderOptions
	Template *pongo2.Template
	Context  pongo2.Context
}

// New creates a new Pongo2Render instance with custom Options.
func New(options RenderOptions) *Pongo2Render {
	return &Pongo2Render{
		Options: &options,
	}
}

// Default creates a Pongo2Render instance with default options.
func Default() *Pongo2Render {
	return New(RenderOptions{
		TemplateDir: "templates",
		ContentType: "text/html; charset=utf-8",
	})
}

// Instance should return a new Pongo2Render struct per request and prepare
// the template by either loading it from disk or using pongo2's cache.
func (p Pongo2Render) Instance(name string, data interface{}) render.Render {
	var template *pongo2.Template
	filename := path.Join(p.Options.TemplateDir, name)

	// always read template files from disk if in debug mode, use cache otherwise.
	if gin.Mode() == "debug" {
		template = pongo2.Must(pongo2.FromFile(filename))
	} else {
		template = pongo2.Must(pongo2.FromCache(filename))
	}

	return Pongo2Render{
		Template: template,
		Context:  data.(pongo2.Context),
		Options:  p.Options,
	}
}

// Render should render the template to the response.
func (p Pongo2Render) Render(w http.ResponseWriter) error {
	p.WriteContentType(w)
	err := p.Template.ExecuteWriter(p.Context, w)
	return err
}

// WriteContentType should add the Content-Type header to the response
// when not set yet.
func (p Pongo2Render) WriteContentType(w http.ResponseWriter) {
	header := w.Header()
	if val := header["Content-Type"]; len(val) == 0 {
		header["Content-Type"] = []string{p.Options.ContentType}
	}
}
```


  














