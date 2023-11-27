package main

import (
	"context"
	"flag"
	"fmt"
	"github.com/oceanweave/dynamic-log-set/dynamiclog"
	"k8s.io/client-go/tools/cache"
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

	logprint := dynamiclog.NewWithSharedInformerFactory(context.TODO(), sharedInformerFactory, "my-config2", "default", "log", "info")

	fmt.Println(logprint.GetLogPartLevelMap())

	// 获取 ConfigMap 的 Informer
	cmInformer := sharedInformerFactory.Core().V1().ConfigMaps().Informer()

	// 启动 Informer
	stopCh := make(chan struct{})
	defer close(stopCh)
	// 等待缓存同步完成
	//go sharedInformerFactory.Start(stopCh)
	//go cmInformer.Run(stopCh)

	if !cache.WaitForCacheSync(stopCh, cmInformer.HasSynced) {
		fmt.Println("Timed out waiting for caches to sync")
		return
	}
	//检查 HasSynced 的值
	if !cmInformer.HasSynced() {
		fmt.Println("Informer has not synced yet")
		return
	}
	fmt.Println("上面同步完成")
	////cmlister2 := cmInformer.Lister()
	//cmlister := sharedInformerFactory.Core().V1().ConfigMaps().Lister()
	//cmlist, _ := cmlister.ConfigMaps("default").Get("my-config2")
	//fmt.Println("cmlist:", cmlist)
	//
	////configMap, err := clientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "my-config2", metav1.GetOptions{})
	////fmt.Println(configMap.Data)
	//
	//cmStore := cmInformer.GetStore()
	//// 手动检索和遍历 ConfigMap 资源
	//cmKey, _ := cache.MetaNamespaceKeyFunc(&corev1.ConfigMap{
	//	ObjectMeta: metav1.ObjectMeta{
	//		Name:      "my-config",
	//		Namespace: "default",
	//	},
	//})
	//cmObj, _, err := cmStore.GetByKey(cmKey)
	//if err != nil {
	//	fmt.Printf("Error getting ConfigMap from Store: %v\n", err)
	//	return
	//}
	//fmt.Println(cmObj.(*corev1.ConfigMap).Data)

	// 初始化 klog，将日志输出到标准错误流
	klog.InitFlags(nil)
	flag.Set("logtostderr", "true")

	// 解析命令行参数
	flag.Parse()

	klog.V(logprint.KlogEnableLogPrint("dynamicLogLevel", dynamiclog.LogDebugLevel)).Info("---> DEBUG动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("dynamicLogLevel", dynamiclog.LogInfoLevel)).Info("---> INFO动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("dynamicLogLevel", dynamiclog.LogWarnLevel)).Info("---> WARN动态打印日志成功")
	klog.V(logprint.KlogEnableLogPrint("dynamicLogLevel", dynamiclog.LogErrorLevel)).Info("---> ERROR动态打印日志成功")
	// 等待程序终止
	select {}
}
