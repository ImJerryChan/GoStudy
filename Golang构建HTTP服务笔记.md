# Golang-net/http模块

## 一、最简单的HTTP服务器

go语言的http服务是如何被搭建的，可以用几句话大概描述

* 对我们要访问的` URL `和用来处理请求的` 函数方法 `进行`路由注册`
* 新建一个` Server `示例，并根据配置开启服务` 监听 `
* 服务器接收到请求后，服务器根据用来处理请求的` 函数方法`处理请求，处理完毕后返回给客户端

那么在go的net/http库里面是如何实现的呢？



### 1. 基本概念

为了便于理解，本文约定了如下规则

* handler函数：具有`func(w http.ResponseWriter, r *http.Request)`签名的函数，如：

  ```go
  // 这是一个handler函数
  func IndexHandler(w http.ResponseWriter, r *http.Request) {
  	fmt.Fprintln(w, "hello world")
  }
  ```

* handler对象：实现了`ServeHTTP方法`的结构，如：

  ```go
  // 这是一个handler对象
  type textHandler struct {
  	responseText string
  }
  
  func (th *textHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
  	fmt.Fprintf(w, th.responseText)
  }
  
  type Handler interface {
      ServeHTTP(ResponseWriter, *Request)
  }
  ```


* handler处理器：经由`http.HandlerFunc`结构体封装后的`handler函数`，其具特点：由于它实现了`ServeHTTP方法`，因此它是一个`handler对象`，如：

  ```go
  func text(w http.ResponseWriter, req *http.Request) {
  	fmt.Fprintln(w, "hello world")
  }
  
  type HandlerFunc func(ResponseWriter, *Request)
  
  func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
  	f(w, r)
  }
  
  func main() {
      ...
      // 这里text函数被结构体 http.HandlerFunc封装，使得text也实现了ServeHTTP方法
      // 因此text函数也是一个handler，这个handler调用ServeHTTP方法就是调用text函数本身
      http.Handle("/", http.HandlerFunc(text))
  }
  ```

* ServeMux：通俗理解就是这是一个`路由`，默认我们在注册路由时使用的是` DefautServeMux `，其中最重要的是字段`m`： `key`是一些url模式，`value`是一个muxEntry结构，后者里定义存储了具体的`pattern`和`handler `

  ```go
  type ServeMux struct {
      mu    sync.RWMutex
      m     map[string]muxEntry
      hosts bool 
  }
  
  type muxEntry struct {
      explicit bool
      h        Handler // 实现了ServeHTTP的结构
      pattern  string  // 如：127.0.0.1:9000
  }
  ```

  需要注意的是，`ServeMux`也实现了`ServeHTTP`接口，也是一个`handler`，但是它是用来找路由注册的handler的，后面会作解释

* Server：结构如下：

  ```go
  type Server struct {
      Addr         string        
      Handler      Handler     // 这里的Handler暂定为都应该是ServeMux
      ReadTimeout  time.Duration 
      WriteTimeout time.Duration 
      TLSConfig    *tls.Config   
  
      MaxHeaderBytes int
  
      TLSNextProto map[string]func(*Server, *tls.Conn, Handler)
  
      ConnState func(net.Conn, ConnState)
      ErrorLog *log.Logger
      disableKeepAlives int32     nextProtoOnce     sync.Once 
      nextProtoErr      error     
  }
  ```

  server结构存储了服务器处理请求常见的字段 ，也包含了一个`Handler`，如果Server没有提供Handler对象，那么会默认使用`  DefautServeMux `做路由（这个在后面serverHandler.ServeHTTP方法中可以体现)



### 2. 源码解析

下面是一个最简单的http服务器，接下来会结合源码解析分析http服务器是如何工作的:

```go
// 1. 这是一个handler函数
// 2. 同时这也是一个handler对象（由于实现了ServeHTTP方法，请继续往下看）
func IndexHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hello world")
}

func main() {
    // 注册路由
    http.HandleFunc("/index", IndexHandler)
    // 实例化server对象并开始监听
    http.ListenAndServe("127.0.0.0:8000", nil)
}
```



#### ① 如何完成路由注册

路由注册

```go
http.HandleFunc("/index", IndexHandler)
```

首先http.HandleFunc方法调用 DefaultServeMux.HandleFunc(pattern, handler)

```go
func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
    // DefaultServeMux 是一个 ServeMux（即默认路由，见上文）
	DefaultServeMux.HandleFunc(pattern, handler)
}
```

其本质是调用ServeMux的HandleFunc方法

```go
func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	if handler == nil {
		panic("http: nil handler")
	}
	mux.Handle(pattern, HandlerFunc(handler))
}
```

在这里，我们在最开始传入的handler（即IndexHandler这个`handler方法`）会被封装成一个`handler处理器`，而ServeMux的Handle方法就是把我们的pattern和handler处理器封装成Key-Value形式

```go
func (mux *ServeMux) Handle(pattern string, handler Handler) {
	...

	if mux.m == nil {
		mux.m = make(map[string]muxEntry)
	}
    
    // 这里的muxEntry可以理解成是前面所说的key-value结构
	e := muxEntry{h: handler, pattern: pattern}
    
    // 举个例子：mux.m["/index"] = muxEntry{h: IndexHandler, pattern: "/index"}
	mux.m[pattern] = e
    
	if pattern[len(pattern)-1] == '/' {
		mux.es = appendSorted(mux.es, e)
	}

    ...
}
```



#### ② 启动服务并开启监听

```go
http.ListenAndServe("127.0.0.0:8000", nil)
```

服务监听需要新建一个Server对象，在这里其实我们先新建了一个Server对象，并调用它的ListenAndServe方法

```go
func ListenAndServe(addr string, handler Handler) error {
	server := &Server{Addr: addr, Handler: handler}
	return server.ListenAndServe()
}
```

根据这个ListenAndServe方法我们启动一个TCP监听，最后将监听的TCP对象传入Server的Serve方法

```go
func (srv *Server) ListenAndServe() error {
	...
	ln, err := net.Listen("tcp", addr)
	...
	return srv.Serve(ln)
}
```

开启监听后，go就开启一个协程处理请求，主要逻辑都在serve方法之中

```go
baseCtx := context.Background()
ctx := context.WithValue(baseCtx, ServerContextKey, srv)
ctx = context.WithValue(ctx, LocalAddrContextKey, l.Addr())
for {
    rw, e := l.Accept()
    ...
    c := srv.newConn(rw)
    c.setState(c.rwc, StateNew) // before Serve can return
    go c.serve(ctx)  // 主要逻辑
}
```

serve 中定义了函数退出时连接关闭相关的处理，然后就是读取连接的网络数据，并处理读取完毕时候的状态，随后调用`serverHandler{c.server}.ServeHTTP(w, w.req)`方法处理请求。最后就是请求处理完毕的逻辑

```go
func (c *conn) serve(ctx context.Context) {
    ...

	for {
		...
		serverHandler{c.server}.ServeHTTP(w, w.req)  // 重要
		...
	}
}
```

serverHandler里面存储了一个上面介绍的Server结构，同时它也实现了Handler接口方法ServeHTTP，并在该接口方法中做了一个重要的事情，初始化multiplexer路由多路复用器（如果server对象没有指定Handler，则使用默认的DefaultServeMux作为路由Multiplexer），并调用ServeMux的ServeHTTP方法

```go
type serverHandler struct {
	srv *Server
}

func (sh serverHandler) ServeHTTP(rw ResponseWriter, req *Request) {
	handler := sh.srv.Handler
	if handler == nil {
		handler = DefaultServeMux
	}
	if req.RequestURI == "*" && req.Method == "OPTIONS" {
		handler = globalOptionsHandler{}
	}
	handler.ServeHTTP(rw, req)
}
```

也就是说，如果我们没有自定义路由(ServeMux)，则会调用DefaultServeMux的ServeHTTP方法

```GO
func (mux *ServeMux) ServeHTTP(w ResponseWriter, r *Request) {
	if r.RequestURI == "*" {
		if r.ProtoAtLeast(1, 1) {
			w.Header().Set("Connection", "close")
		}
		w.WriteHeader(StatusBadRequest)
		return
	}
	h, _ := mux.Handler(r)
    
    // 这里对应本例的IndexHandler.ServeHTTP(w, r)，也就是自己调用自己（因为使用了http.HandlerFunc封装）
	h.ServeHTTP(w, r)
}
```

后续的处理过程是，ServeMux的ServeHTTP方法调用其Handler方法寻找注册到路由上的handler对象，并调用该函数的ServeHTTP方法(本例则是IndexHandler函数)，response写到http.RequestWirter对象返回给客户端。

上述函数运行结束即`serverHandler{c.server}.ServeHTTP(w, w.req)`运行结束。接下来就是对请求处理完毕和连接断开的相关逻辑。



#### ③ 稍微把例子复杂化

其实在上面的源码中模块帮我们省略了很多代码，稍微把它修改一下：

```go
func IndexHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hello world")
}

func main() {
    mux := http.NewServeMux()
    
    // 使用新建的ServeMux注册路由
	mux.Handle("index", http.HandlerFunc(IndexHandler))
	http.ListenAndServe("127.0.0.1:8000", mux)
}
```

在原来的代码基础上我们修改了部分代码，这里并没有使用http.HandleFunc注册路由，而是**直接使用了ServeMux注册路由**，那是因为我们http.HandleFunc其实就是使用默认的DefaultServeMux，并在此路由上进行注册

```go
func (mux *ServeMux) HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	if handler == nil {
		panic("http: nil handler")
	}
	mux.Handle(pattern, HandlerFunc(handler))
}
```

我们在启动监听的时候把mux当成参数传入了 http.ListenAndServe，这是因为ServeMux在模块中也实现了 ServeHTTP 方法，因此也是一个handler

还记得上面serverHandler的ServeHTTP 吗，我们的mux其实就是传入到了它的handler，因此如果我们没有新建一个ServeMux的话，它会帮我们使用默认的DefaultServeMux作为路由并从中查找我们对应的pattern和handler的对应关系

既然我们可以直接新建ServeMux注册路由，当然我们也可以**自定义一个Server对象**

```go
func IndexHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintln(w, "hello world")
}

func main(){
    http.HandleFunc("/index", IndexHandler)

    // http.ListenAndServe("127.0.0.1:8000", mux)
    // 本质上就是新建一个Server{}
    server := &http.Server{
        Addr: ":8000",
        ReadTimeout: 60 * time.Second,
        WriteTimeout: 60 * time.Second,
    }
    server.ListenAndServe()
}
```



#### ④ 小结

从上面例子我们可以看到：

* **路由注册由ServeMux的Handle方法完成**，简单来说是把pattern和handler处理器(本例的IndexHandler，其实还是handler对象)进行Key-Value映射
* 启动服务需要新建一个Server对象，serverHandler的 ServeHTTP方法会**调用路由**(默认是DefaultServeMux)的Handler方法**从Key-Value映射中找到request所对应的handler对象**（本例是IndexHandler），并调用其ServeHTTP方法



## 二、中间件Middleware

### 1. 如何实现中间件

所谓中间件，就是连接上下级不同功能的函数或组件，在这里通常指的是包裹函数行为，为被包裹的函数提供一些额外的功能。在上面介绍的http.HandlerFunc就是把签名为`func(w http.ResponseWriter, req *http.Request)`的函数包裹成一个handler

这里以HTTP请求的中间件例子为例，实现一个log中间件，能够打印出每个请求的log

```go
func index(w http.ResponseWriter, req *http.Request) {
    fmt.Println("Hello world!")
}

func logWrapperHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        // 执行handler之前的逻辑
        start := time.Now()
        fmt.Printf("Started %s %s", req.Method, req.URL.Path)
        
        next.ServeHTTP(w, req)
        
        // 执行完毕handler后的逻辑
        fmt.Printf("Complete %s %v", req.URL.Path, time.Since(start))
    })
}

func main() {
    // http.HandlerFunc把handler函数index包装成一个handler处理器
    http.Handle("/", logWrapperHandler(http.HandlerFunc(index)))
    http.ListenAndServe("127.0.0.1:8000", nil)
}
```

logWrapperHandler是一个中间件函数，如果熟悉python的话，这里说的中间函数和python的wrapper非常类似，这也是在这里我把函数名称写为wrapper的原因，方便理解。其中此函数的参数为handler对象(处理器)，作用是输出请求开始及完成时间

```go
2020/02/23 21:18:13 Started GET /
2020/02/23 21:18:13 Comleted / in 13.365µs
2020/02/23 21:18:20 Started GET /
2020/02/23 21:18:20 Comleted / in 17.541µs
```

由于上面的中间件函数的参数为一个handler对象(处理器)，函数返回结果也是一个handler对象(处理器)，因此很容易联想到我们可以在中间件函数的基础上再封装一层中间件，下面在原来代码基础上再添加一个hook函数

```go
func hookWrapperHandler(next http.Handler) http.Handler {
    return http.Handler(http.HandlerFunc(w http.ResponseWriter, *http.Request) {
        fmt.Println("before hook")
        next.ServeHTTP(w, req)
        fmt.Println("after hook")
    })
}

func main() {
    http.Handle("/", hookWrapperHandler(logWrapperHandler(http.HandlerFunc(index))))
    http.ListenAndServe("127.0.0.1:8000", nil)
}
```

这样我们在logWrapperHandler上又封装了一层hook，可以看到输出为：

```
2020/02/23 21:18:10 before hook
2020/02/23 21:18:13 Started GET /
2020/02/23 21:18:13 Comleted / in 13.365µs
2020/02/23 21:18:20 Started GET /
2020/02/23 21:18:20 Comleted / in 17.541µs
2020/02/23 21:18:21 after hook
```



### 2. cookie和session在HTTP中的应用

上一小节学习了如何在一个HTTP请求中附加一些格外功能，在这一小节我们将会讲述如何在HTTP中使用cookie和session，**下面代码中未定义的函数我们无需在意它的具体实现，只需要根据函数名称知道它做什么即可**

对于一个购物网站说，一般都会需要进行用户登录后才可以下单购物，在这里我们也不例外，我们提供了用户登录的功能：

```go
func main() {
    ...
    http.Handle("/user/login", http.HandlerFunc(Login))
    http.ListenAndServe("127.0.0.1:8000", nil)
}
```

大概的用户登录逻辑实现如下，首先我们确保用户输入的用户信息校验无误，确认无误后，我们生成一个用于保存token的cookie，**保存在浏览器上**，之后每一次用户请求都会带上这个名称为remember-me-token的cookie：

```go
// 这里的store是用于存储session的
store = sessions.NewCookieStore([]byte("OnNUU5RUr6Ii2HMI0d6E54bXTS52tCCL"))

func Login(w http.Response, req *http.Request) {
    ...
    // 1. 查找request里面是否存在用户输入的用户信息
    rsp, err := QueryUserByName(userName)
    
    response := map[string]interface{}{
		"ref": time.Now().UnixNano(),
	}
    response["data"] = rsp.User
    ...
    
    // 2. 用户存在且校验密码成功后，生成一个token
    rsp2, err := MakeAccessToken()
    response["token"] = rsp2.Token
    
    w.Header().Add("set-cookie", "application/json; charset=utf-8")
    
    // 3. 使用cookie存储token
    cookie := &http.Cookie{
        Name: "remember-me-token",
        Value: rsp.Token,
        Path: "/",
        Expires: time.Now().Add(30 * time.Minute),
        MaxAge: 90000,
    }
    http.SetCookie(w, &cookie)
    
    // 4. 获取session
    sess := session.GetSession(w, req)
    
    // 5. 我们可以在session里面存储用户登录信息
    sess.Values["userName"] = rsp.User.Name
    sess.Values["userId"] = rsp.User.Id
    _ := sess.Save(r, w)
    
    w.Header().Add("Content-Type", "application/json; charset=utf-8")
    
    if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
}
```

我们可以看到，在上面的Login函数中也是用了session，其中session的作用是**在服务器上记录用户登录信息**，这里我们的session保存了其登录时候使用的用户名和用户ID，其中生成session的方法如下：

```go
func GetSession(w http.Response, req *http.Request) *session.Session {
	var sID string
	for _, c := range r.Cookies() {
		if strings.Index(c.Name, sessionIdNamePrefix) == 0 {
			sID = c.Name
			break
		}
	}

	if sID == "" {
		sID = sessionIdNamePrefix + uuid.New().String()
	}

	ses, _ := store.Get(r, sID)
    
	// 没有id说明是新session
	if ses.ID == "" {
        // 在这个例子中，我们把session也保存在了cookie当中
		cookie := &http.Cookie{
			Name: sID,
			Value: sID,
			Path: "/",
			Expires: time.Now().Add(30 * time.Second),
			MaxAge: 0,
		}
		http.SetCookie(w, cookie)

		ses.ID = sID
		ses.Save(r, w)
	}
	return ses
}
```

当我们完成用户登录之后，为了让服务器知道我们接下来的如新增订单或支付的请求是来自同一个会话，我们会在它们的handler函数外面包裹一层中间件进行**请求鉴权**

```go
func main() {
    // PayOrder是用于订单支付的函数
    payOrderHandler := http.HandlerFunc(PayOrder)
    http.Handle("/payment/pay-order", AuthWrapperHandler(payOrderHandler))
}
```

下面我们来看下请求鉴权的实现示例：

```go
func AuthWrapperHandler(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        // 从req中获取cookie，里面存储着我们的token，根据这个cookie名称查找
        ck, _ := req.Cookie("remember-me-token")
        
        if ck == nil {
            log.Logf("token 不存在")
            http.Error(w, "非法请求", 400)
            return
        }
        
        sess := session.GetSession(w, req)
        
        if sess.ID != "" {
            if sess.Values["valid"] != nil {
                next.ServeHTTP(w, req)
                return
            } else {
                if sess.Values["userId"] == nil {
                    log.Logf("[AuthWrapperHandler], 用户未登录")
                    http.Error(w, "请求异常，请登陆后再试", 400)
                }
                userId := sess.Values["userId"].(int64)
                if userId != 0 {
                    // GetCachedAccessToken函数，会根据提供的用户ID，从redis缓存中
                    // 找到所对应的key的value，即token
                    rsp, err := GetCachedAccessToken(userId)
                    if err != nil {
						log.Logf("[AuthWrapper]，err：%s", err)
						http.Error(w, "非法请求", 400)
						return
					}
                    
                    if rsp.Token != ck.Value {
                        log.Logf("[AuthWrapper]，token不一致")
						http.Error(w, "非法请求", 400)
						return
                    }
                } else {
                    log.Logf("[AuthWrapper]，session不合法，无用户id")
					http.Error(w, "非法请求", 400)
					return
                }
            }
        } else {
            http.Error(w, "非法请求", 400)
			return
        }
        h.ServeHTTP(w, r)
    })
}
```

当我们客户端调用PayOrder方法对订单进行支付时，首先会在req中获取cookie提取出我们登录时存入的token，然后从session中提取出这个会话的登录用户的token，然后两者进行对比，如果两个token一致，则调用PayOrder方法进行订单支付



### 3. 小结

从上面示例代码可以总结出：

* session，可以类比为用户信息档案表，包含了用户的认证信息和登陆状态信息，在上文示例里面，session就存储了用户的userName和userId，存储在服务器中。通常只需要在客户端上保存一个id
* cookie是session的一种实现方案，在上文示例中，session不仅可以用来存储我们的token（用于校验登录是否合法），还可以存储session



## 三、参考链接

[1. Golang构建HTTP服务（一）--- net/http库源码笔记 ]( https://www.jianshu.com/p/be3d9cdc680b )

[2. Golang构建HTTP服务（二）--- net/http库源码笔记]( https://www.jianshu.com/p/16210100d43d )

[3. 彻底弄懂session、cookie、token]( https://segmentfault.com/a/1190000017831088?utm_source=tag-newest )



## 四、TO BE CONTINUE