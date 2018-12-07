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
			c.AbortWithStatusJSON(http.StatusOK, (&Protocol{}).Error(http.StatusFound, ERROR_AUTH_LIMIT))
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
			//
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
		api.GET("/spot", Spot(), News(), JSON())
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
