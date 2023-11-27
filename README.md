# dynamic-log-set

## 动态日志配置 Configmap 创建示例
```txt
# 文件名称为 log， 这也是  NewWithSharedInformerFactory 函数需要指定的 cmLogKey 名称
part1: debug
part2: warn
```

``` shell
# 单文件
kubectl -n default create configmap log-demo-set --from-file=log
# 多文件，若需要在 Configmap 中配置其他信息，other-conf 表示用户需要的其他配置信息，此处我们只需要 log 文件的配置信息
kubectl -n default create configmap log-demo-set --from-file=other-conf --from-file=log

# 删除
kubectl -n default delete configmap log-demo-set
```

## 函数讲解
1. 日志信息存放在哪个 Configmap 中（ cmNameSpace， cmName）
2. 在该 Configmap 应该读取哪个 key 对应的信息（cmLogKey）
3. 没有配置 Configmap 时，或被误删除，应该打印什么级别的日志（logDefaultLevel）
4. KlogEnableLogPrint 函数，第一个参数读取 Configmap 中配置的“动态”日志级别，第二个参数设置此处“当前”的日志级别，若“当前日志级别”>=“动态日志级别”, 此处的日志就会打印
``` shell
-> % kubectl -n default create configmap log-demo-set --from-file=log
configmap/log-demo-set created

-> % kubectl describe cm log-demo-set 
Name:         log-demo-set
Namespace:    default
Labels:       <none>
Annotations:  <none>

Data
====
log:  # 此处是 cmLogKey， 也可以认为就是上面的 --from-file 指定的文件名称
----
part1: debug   # 动态日志级别配置
part2: warn    # 动态日志级别配置
# part3 此处没有配置动态日志级别，因此代码中若配置此字段，就会采用默认日志级别 logDefaultLevel 进行判断

BinaryData
====

```

``` go

import "github.com/oceanweave/dynamic-log-set/dynamiclog"

	// 创建 SharedInformerFactory
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	
	// 动态日志的 Configmap 配置信息
	cmNamespace := "default"
	cmName := "log-demo-set"
	cmLogKey := "log"
	logDefaultLevel := "info"

	// 此处返回的是指针
	logprint := dynamiclog.NewWithSharedInformerFactory(context.TODO(), sharedInformerFactory, cmNamespace, cmName, cmLogKey, logDefaultLevel)
    
        // GetLogPartNameList 获取 part 名称， GetLogPartLevelMap 获取 part 对应的 Level 的映射关系
	fmt.Printf("Namespace: %s, ConfigMap: %s  --> Exist Confimap's Part-Key: %s\n", cmNamespace, cmName, logprint.GetLogPartNameList())
	fmt.Println("Now log level setting:", logprint.GetLogPartLevelMap())
	
	// part1 的动态日志级别是"debug", 因此所有高于等于此 debug 级别都可以打印，就是下面所有都打印了
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogDebugLevel)).Info("---> Part-1-DEBUG动态打印日志成功") // 打印
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogInfoLevel)).Info("---> Part-1-INFO动态打印日志成功")   // 打印
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogWarnLevel)).Info("---> Part-1-WARN动态打印日志成功")   // 打印
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogErrorLevel)).Info("---> Part-1-ERROR动态打印日志成功") // 打印
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogFatalLevel)).Info("---> Part-1-FATAL动态打印日志成功") // 打印
    
        // part2 的动态日志级别是"warn", 因此所有高于等于此级别 warn 都可以打印
	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogDebugLevel)).Info("---> Part-2-DEBUG动态打印日志成功") // 不打印
	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogInfoLevel)).Info("---> Part-2-INFO动态打印日志成功")   // 不打印
	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogWarnLevel)).Info("---> Part-2-WARN动态打印日志成功")   // 打印
	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogErrorLevel)).Info("---> Part-2-ERROR动态打印日志成功") // 打印
	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogFatalLevel)).Info("---> Part-2-FATAL动态打印日志成功") // 打印
	// part3 没有设置动态日志级别，因此就会设置 logDefaultLevel 为动态日志级别，此处配置为 “info”，因此高于等于 info 级别日志都会打印
	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogDebugLevel)).Info("---> Part-3-DEBUG动态打印日志成功") // 不打印
	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogInfoLevel)).Info("---> Part-3-INFO动态打印日志成功")   // 打印
	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogWarnLevel)).Info("---> Part-3-WARN动态打印日志成功")   // 打印
	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogErrorLevel)).Info("---> Part-3-ERROR动态打印日志成功") // 打印
	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogFatalLevel)).Info("---> Part-3-FATAL动态打印日志成功") // 打印
```



## 测试
### 1. 创建 Configmap
``` shell
-> % kubectl -n default create configmap log-demo-set --from-file=log
configmap/log-demo-set created

-> % kubectl describe cm log-demo-set 
Name:         log-demo-set
Namespace:    default
Labels:       <none>
Annotations:  <none>

Data
====
log:
----
part1: debug
part2: warn


BinaryData
====

```

### 2. 运行函数

```
# 运行
go run main.go
# 获取集群中已存在的 Configmap 信息
Dynamic-log-set: Initing(load exist configmap) ...
Namespace: default, ConfigMap: log-demo-set  --> Exist Confimap's Part-Key: [part1 part2]
Now log level setting: map[part1:debug part2:warn]
====================================  # 没有设置 part3 字段，会默认配置为“默认等级”，由NewWithSharedInformerFactory函数传参指定
I1127 16:32:00.966732   67528 main.go:66] ---> Part-1-DEBUG动态打印日志成功
I1127 16:32:00.966908   67528 main.go:67] ---> Part-1-INFO动态打印日志成功
I1127 16:32:00.966915   67528 main.go:68] ---> Part-1-WARN动态打印日志成功
I1127 16:32:00.966921   67528 main.go:69] ---> Part-1-ERROR动态打印日志成功
I1127 16:32:00.966926   67528 main.go:70] ---> Part-1-FATAL动态打印日志成功
I1127 16:32:00.966935   67528 main.go:74] ---> Part-2-WARN动态打印日志成功
I1127 16:32:00.966942   67528 main.go:75] ---> Part-2-ERROR动态打印日志成功
I1127 16:32:00.966950   67528 main.go:76] ---> Part-2-FATAL动态打印日志成功
Dynamic-log-set: Not found “part3” log level set！Set the default “info” log level.
I1127 16:32:00.966972   67528 main.go:79] ---> Part-3-INFO动态打印日志成功
I1127 16:32:00.966981   67528 main.go:80] ---> Part-3-WARN动态打印日志成功
I1127 16:32:00.966990   67528 main.go:81] ---> Part-3-ERROR动态打印日志成功
I1127 16:32:00.967000   67528 main.go:82] ---> Part-3-FATAL动态打印日志成功
====================================  # 此处将 Configmap edit（kubectl edit cm log-demo-set），增加 part3: fatal 级别
I1127 16:32:20.966553   67528 main.go:66] ---> Part-1-DEBUG动态打印日志成功
I1127 16:32:20.966582   67528 main.go:67] ---> Part-1-INFO动态打印日志成功
I1127 16:32:20.966595   67528 main.go:68] ---> Part-1-WARN动态打印日志成功
I1127 16:32:20.966608   67528 main.go:69] ---> Part-1-ERROR动态打印日志成功
I1127 16:32:20.966619   67528 main.go:70] ---> Part-1-FATAL动态打印日志成功
I1127 16:32:20.966635   67528 main.go:74] ---> Part-2-WARN动态打印日志成功
I1127 16:32:20.966650   67528 main.go:75] ---> Part-2-ERROR动态打印日志成功
I1127 16:32:20.966666   67528 main.go:76] ---> Part-2-FATAL动态打印日志成功
I1127 16:32:20.966687   67528 main.go:82] ---> Part-3-FATAL动态打印日志成功
==================================== # 此处将 Configmap edit（kubectl edit cm log-demo-set），将 part3 由 fatal 改为 error 级别
I1127 16:32:50.966233   67528 main.go:66] ---> Part-1-DEBUG动态打印日志成功
I1127 16:32:50.966255   67528 main.go:67] ---> Part-1-INFO动态打印日志成功
I1127 16:32:50.966262   67528 main.go:68] ---> Part-1-WARN动态打印日志成功
I1127 16:32:50.966269   67528 main.go:69] ---> Part-1-ERROR动态打印日志成功
I1127 16:32:50.966274   67528 main.go:70] ---> Part-1-FATAL动态打印日志成功
I1127 16:32:50.966283   67528 main.go:74] ---> Part-2-WARN动态打印日志成功
I1127 16:32:50.966291   67528 main.go:75] ---> Part-2-ERROR动态打印日志成功
I1127 16:32:50.966300   67528 main.go:76] ---> Part-2-FATAL动态打印日志成功
I1127 16:32:50.966323   67528 main.go:81] ---> Part-3-ERROR动态打印日志成功
I1127 16:32:50.966334   67528 main.go:82] ---> Part-3-FATAL动态打印日志成功
^C

```