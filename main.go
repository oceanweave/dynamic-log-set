package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/oceanweave/dynamic-log-set/dynamiclog"
	"k8s.io/klog/v2"
	"os"
	"path/filepath"
	"time"

	"k8s.io/client-go/informers"
	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
)

func main() {
	kubeconfig := flag.String("kubeconfig", filepath.Join(homedir.HomeDir(), ".kube", "config"), "kubeconfig file")
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		fmt.Printf("Error building kubeconfig: %v\n", err)
		os.Exit(1)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		fmt.Printf("Error creating Kubernetes client: %v\n", err)
		os.Exit(1)
	}

	// 创建 SharedInformerFactory
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)

	// 动态日志的 Configmap 配置信息
	cmNamespace := "default"
	cmName := "log-demo-set"
	cmLogKey := "log"
	logDefaultLevel := "info"

	// 此处返回的是指针
	logprint := dynamiclog.NewWithSharedInformerFactory(context.TODO(), sharedInformerFactory, cmNamespace, cmName, cmLogKey, logDefaultLevel)

	fmt.Printf("Namespace: %s, ConfigMap: %s  --> Exist Confimap's Part-Key: %s\n", cmNamespace, cmName, logprint.GetLogPartNameList())
	fmt.Println("Now log level setting:", logprint.GetLogPartLevelMap())

	// 解析命令行参数
	flag.Parse()

	// 创建一个定时器，每5秒触发一次
	ticker := time.Tick(10 * time.Second)

	// 无限循环，定期执行 demo 函数
	for {
		select {
		case <-ticker:
			// 每次定时器触发时执行 demo 函数
			demo(logprint)
		}
	}

	// 阻塞，防止程序终止
	// select {}
}

func demo(logprint dynamiclog.LogInterface) {
	fmt.Println("====================================")
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogDebugLevel)).Info("---> Part-1-DEBUG动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogInfoLevel)).Info("---> Part-1-INFO动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogWarnLevel)).Info("---> Part-1-WARN动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogErrorLevel)).Info("---> Part-1-ERROR动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part1", dynamiclog.LogFatalLevel)).Info("---> Part-1-FATAL动态打印日志成功")

	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogDebugLevel)).Info("---> Part-2-DEBUG动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogInfoLevel)).Info("---> Part-2-INFO动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogWarnLevel)).Info("---> Part-2-WARN动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogErrorLevel)).Info("---> Part-2-ERROR动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part2", dynamiclog.LogFatalLevel)).Info("---> Part-2-FATAL动态打印日志成功")

	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogDebugLevel)).Info("---> Part-3-DEBUG动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogInfoLevel)).Info("---> Part-3-INFO动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogWarnLevel)).Info("---> Part-3-WARN动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogErrorLevel)).Info("---> Part-3-ERROR动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("part3", dynamiclog.LogFatalLevel)).Info("---> Part-3-FATAL动态打印日志成功")
}
