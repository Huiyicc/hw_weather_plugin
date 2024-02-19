## 天气插件
---
[瀚文扩展](https://github.com/Huiyicc/HelloWord_HY)的天气插件

## 构建方法

### Windows
```shell
go build -buildmode=c-shared -o bin/life.win -tags stdc -ldflags="-s -w" main.go plugin.go
```
### Linux
```shell
go build -buildmode=c-shared -o bin/life.linux -tags stdc -ldflags="-s -w" main.go plugin.go
```

### Mac
```shell
go build -buildmode=c-shared -o bin/life.mac -tags stdc -ldflags="-s -w" main.go plugin.go
```

---

## 编写方法与扩展参见文档

[GitHub](https://github.com/Huiyicc/HelloWord_HY/blob/master/plugindev.md)  
[Gitee](https://gitee.com/LoveA/HelloWord_HY/blob/master/plugindev.md)
