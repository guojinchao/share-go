# Pongo2 模版引擎
这里的Pongo2是指Golang语言里模板引擎。[Pongo2](https://github.com/flosch/pongo2) 模板语法基本模仿Dango。

## 基本语法

### `extends` 继承

`Frame.html`
```html
<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <meta http-equiv="X-UA-Compatible" content="ie=edge">
    <meta content="{%block keywords %}{%endblock%}" name="keywords">
    <meta content="{%block description %}{%endblock%}" name="description">
    {%block links %}{%endblock%}
    <title>{%block title %}{%endblock%}</title>
    {%include "./analytic_google.html"%}
</head>
<body>
    {%block header  %}{%endblock%}
    {%block body    %}{%endblock%}
    {%block footer  %}{%endblock%}
    <script type="text/javascript" src="https://static.smm.cn/common.smm.cn/jquery1.11.3/dist/jquery.min.js"></script>
    {%block scripts %}{%endblock%}
    {%include "./analytiv_baidu.html"%}
</body>
</html>

```

继承于`Frame` 的 `index.html`
```html
{% extends "../../components/smm_components/src/frame/frame.html"%}

{%block description%}
上海有xxxxxxxxx之前列.
{%endblock%}

{%block keywords%}
 上海xxxxxxxxxxxxx情
{%endblock%}

{%block title%}
 上海xxxxxxxxxxxxxxxx门户
{%endblock%}

{%block links%}
  <link rel="shortcut icon" type="image/x-icon" href="{{static.UrlPrefix}}{{static.Version}}/image/favicon.ico" />
  <link rel="stylesheet" href="{{static.UrlPrefix}}{{static.Version}}/release/css/index.min.css">
{%endblock%}

{%block header%}
  {% import "../../components/header/index.html" header%}
  {{header()}}
{%endblock%}

{%block body%}
    {% import "../../components/back_top/index.html" back_to_top %}
    {% import "../../components/spot/index.html" spot%}
    {% import "../../components/focus/index.html" focus%}
    {% import "../../components/hot_goods/index.html" hotGoods %}
    {% import "../../components/ad/index.html" ad_right %}

    <div class="content-wrapper">
      <div class="content-wrap clear layout">
        <div class="left">
          <div class="content-top clear">
            {{spot()}}
            <div class="right">
                {{focus()}}
            </div>
          </div>
        </div>
        <div class="right">
          {{ad_right()}}
        </div>
      </div>
    </div>


    {{back_to_top()}}
{%endblock%}

{%block footer%}
  {% import "../../components/footer/index.html" footer %}
  {{footer()}}
{%endblock%}

{%block scripts%}
    <script type="text/javascript" src='{{config.GetSource("library")}}/jquery-qrcode/jquery.qrcode.min.js'></script>
    <script type="text/javascript" src='{{config.GetSource("library")}}/slick-carousel/slick/slick.min.js'></script>
    <script type="text/javascript" src="{{static.UrlPrefix}}{{static.Version}}/release/js/index.js"></script>
{%endblock%}

```


### `macro` 组合

```html
{% macro footer() export %}
{% import "./components/link/index.html" link %}
{% import "./components/copyright/index.html" copyright %}
<div class="components-footer">
  {{link()}}
  {{copyright()}}
</div>
{% endmacro %}

```

#### `import` 语句

```go
  {% import "../../components/header/index.html" header%}
  {{header()}}
```

### `filter` 过滤器

形式如下： 

```
{{ spot.IdAddZeroL | setFont | truncatechars "2"}}

```

使用“|” 对数据做多次的处理

`pongo2` 内置的`filter`

```golang
    RegisterFilter("escape", filterEscape)
	RegisterFilter("safe", filterSafe)
	RegisterFilter("escapejs", filterEscapejs)

	RegisterFilter("add", filterAdd)
	RegisterFilter("addslashes", filterAddslashes)
	RegisterFilter("capfirst", filterCapfirst)
	RegisterFilter("center", filterCenter)
	RegisterFilter("cut", filterCut)
	RegisterFilter("date", filterDate)
	RegisterFilter("default", filterDefault)
	RegisterFilter("default_if_none", filterDefaultIfNone)
	RegisterFilter("divisibleby", filterDivisibleby)
	RegisterFilter("first", filterFirst)
	RegisterFilter("floatformat", filterFloatformat)
	RegisterFilter("get_digit", filterGetdigit)
	RegisterFilter("iriencode", filterIriencode)
	RegisterFilter("join", filterJoin)
	RegisterFilter("last", filterLast)
	RegisterFilter("length", filterLength)
	RegisterFilter("length_is", filterLengthis)
	RegisterFilter("linebreaks", filterLinebreaks)
	RegisterFilter("linebreaksbr", filterLinebreaksbr)
	RegisterFilter("linenumbers", filterLinenumbers)
	RegisterFilter("ljust", filterLjust)
	RegisterFilter("lower", filterLower)
	RegisterFilter("make_list", filterMakelist)
	RegisterFilter("phone2numeric", filterPhone2numeric)
	RegisterFilter("pluralize", filterPluralize)
	RegisterFilter("random", filterRandom)
	RegisterFilter("removetags", filterRemovetags)
	RegisterFilter("rjust", filterRjust)
	RegisterFilter("slice", filterSlice)
	RegisterFilter("split", filterSplit)
	RegisterFilter("stringformat", filterStringformat)
	RegisterFilter("striptags", filterStriptags)
	RegisterFilter("time", filterDate) // time uses filterDate (same golang-format)
	RegisterFilter("title", filterTitle)
	RegisterFilter("truncatechars", filterTruncatechars)
	RegisterFilter("truncatechars_html", filterTruncatecharsHTML)
	RegisterFilter("truncatewords", filterTruncatewords)
	RegisterFilter("truncatewords_html", filterTruncatewordsHTML)
	RegisterFilter("upper", filterUpper)
	RegisterFilter("urlencode", filterUrlencode)
	RegisterFilter("urlize", filterUrlize)
	RegisterFilter("urlizetrunc", filterUrlizetrunc)
	RegisterFilter("wordcount", filterWordcount)
	RegisterFilter("wordwrap", filterWordwrap)
	RegisterFilter("yesno", filterYesno)

	RegisterFilter("float", filterFloat)     // pongo-specific
	RegisterFilter("integer", filterInteger) // pongo-specific
```

向`Pongo2` 注入一个`filter`非常简单
```golang
	pongo2.RegisterFilter("setFont", func(in *pongo2.Value, parame *pongo2.Value) (*pongo2.Value, *pongo2.Error) {
		var str string
		if in.IsNumber() {
			n := in.Integer()
			str = fmt.Sprint(n)
		} else {
			str = in.String()
		}
		return pongo2.AsValue(v2_utils.SetFontface2peopleUp(v2_utils.ChangeLoadFontface, str)), nil
	})
```

#### 流程控制

##### `for` 语句 
```
{% for comment in complex.comments %}
{{ forloop.Counter }} {{ forloop.Counter0 }} {{ forloop.First }} {{ forloop.Last }} 
{{ forloop.Revcounter }} {{ forloop.Revcounter0 }} {{ comment.Author.Name }}


{% endfor %}

reversed
'{% for item in simple.multiple_item_list reversed %}{{ item }} {% endfor %}'

sorted string map
'{% for key in simple.strmap sorted %}{{ key }} {% endfor %}'

sorted int map
'{% for key in simple.intmap sorted %}{{ key }} {% endfor %}'

sorted int list
'{% for key in simple.unsorted_int_list sorted %}{{ key }} {% endfor %}'

reversed sorted int list
'{% for key in simple.unsorted_int_list reversed sorted %}{{ key }} {% endfor %}'

reversed sorted string map
'{% for key in simple.strmap reversed sorted %}{{ key }} {% endfor %}'

reversed sorted int map
'{% for key in simple.intmap reversed sorted %}{{ key }} {% endfor %}'
```

##### `if` 语句

```go
{% if nothing %}false{% else %}true{% endif %}
{% if simple %}simple != nil{% endif %}
{% if simple.uint %}uint != 0{% endif %}
{% if simple.float %}float != 0.0{% endif %}
{% if !simple %}false{% else %}!simple{% endif %}
{% if !simple.uint %}false{% else %}!simple.uint{% endif %}
{% if !simple.float %}false{% else %}!simple.float{% endif %}
{% if "Text" in complex.post %}text field in complex.post{% endif %}
{% if 5 in simple.intmap %}5 in simple.intmap{% endif %}
{% if !0.0 %}!0.0{% endif %}
{% if !0 %}!0{% endif %}
{% if not complex.post %}true{% else %}false{% endif %}
{% if simple.number == 43 %}no{% else %}42{% endif %}
{% if simple.number < 42 %}false{% elif simple.number > 42 %}no{% elif simple.number >= 42 %}yes{% else %}no{% endif %}
{% if simple.number < 42 %}false{% elif simple.number > 42 %}no{% elif simple.number != 42 %}no{% else %}yes{% endif %}
{% if 0 %}!0{% elif nothing %}nothing{% else %}true{% endif %}
{% if 0 %}!0{% elif simple.float %}simple.float{% else %}false{% endif %}
{% if 0 %}!0{% elif !simple.float %}false{% elif "Text" in complex.post%}Elseif with no else{% endif %}
```

#####  `value`,`function` 表示

```go
Variables
{{ "hello" }}
{{ 'hello' }}
{{ "hell'o" }}

Filters
{{ 'Test'|slice:'1:3' }}
{{ '<div class=\"foo\"><ul class=\"foo\"><li class=\"foo\"><p class=\"foo\">This is a long test which will be cutted after some chars.</p></li></ul></div>'|truncatechars_html:25 }}
{{ '<a name="link"><p>This </a>is a long test which will be cutted after some chars.</p>'|truncatechars_html:25 }}

Tags
{% if 'Text' in complex.post %}text field in complex.post{% endif %}

Functions
{{ simple.func_variadic('hello') }}
```

##### `set` 语句

```go
{% set new_var = "hello" %}{{ new_var }}
```

##### `with` 语句

```go
new style
Start '{% with what_am_i=simple.name %}I'm {{what_am_i}}{% endwith %}' End
Start '{% with what_am_i=simple.name %}I'm {{what_am_i}}11{% endwith %}' End
Start '{% with number=7 what_am_i="guest" %}I'm {{what_am_i}}{{number}}{% endwith %}' End
Start '{% include "with.helper" with what_am_i=simple.name number=10 %}' End

old style - still supported by Django
Start '{% with simple.name as what_am_i %}I'm {{what_am_i}}{% endwith %}' End
Start '{% with simple.name as what_am_i %}I'm {{what_am_i}}11{% endwith %}' End
Start '{% with 7 as number "guest" as what_am_i %}I'm {{what_am_i}}{{number}}{% endwith %}' End
Start '{% include "with.helper" with what_am_i=simple.name number=10 %}' End

more with tests
{% with first_comment=complex.comments|first %}{{ first_comment.Author }}{% endwith %}
{% with first_comment=complex.comments|first %}{{ first_comment.Author.Name }}{% endwith %}
{% with first_comment=complex.comments|last %}{{ first_comment.Author.Name }}{% endwith %}
```

##### `include` 包含

```go
Start '{% include "includes.helper" %}' End
Start '{% include "includes.helper" if_exists %}' End
Start '{% include "includes.helper" with what_am_i=simple.name only %}' End
Start '{% include "includes.helper" with what_am_i=simple.name %}' End
Start '{% include simple.included_file|lower with number=7 what_am_i="guest" %}' End
Start '{% include "includes.helper.not_exists" if_exists %}' End
Start '{% include simple.included_file_not_exists if_exists with number=7 what_am_i="guest" %}' End
```



