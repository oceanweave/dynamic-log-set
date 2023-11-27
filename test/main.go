package main

import (
	"k8s.io/klog/v2"
)

func main() {
	// 初始化 klog，将日志输出到标准错误流
	klog.InitFlags(nil)
	//flag.Set("logtostderr", "true")

	// 解析命令行参数
	//flag.Parse()

	// 设置日志级别，可选的级别有 V(0-9)，0 最低，9 最高
	klog.V(2).Info("This is a V(2) log message")

	// 普通日志记录
	klog.Info("This is an info log message")

	// 警告日志记录
	klog.Warning("This is a warning log message")

	// 错误日志记录
	klog.Error("This is an error log message")

	// 打印调试信息
	klog.V(3).Info("This is a V(3) log message")

	// 在程序结束时调用，确保日志刷新
	//defer klog.Flush()
}
