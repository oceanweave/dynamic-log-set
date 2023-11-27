package dynamiclog

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
	"os"
	"strings"
	"time"
)

// NewWithConfigPath  create konfig with shared informer factory.
func NewWithConfigPath(ctx context.Context, configPath string, name, namespace, logKey, defaultLevel string) LogInterface {
	config, err := clientcmd.BuildConfigFromFlags("", configPath)
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
	c := NewWithSharedInformerFactory(ctx, sharedInformerFactory, namespace, name, logKey, defaultLevel)
	return c
}

// NewWithSharedInformerFactory create konfig with shared informer factory.
// 返回值为接口形式，那么用户使用返回的结构体，只能调用该接口规定的方法
// args:
// cmNamespace --> log-configmap 所在的 namespace，
// cmName --> log-configmap 的名称，
// cmLogKey --> log-configmap 中 log 配置字段的 key 值（可以理解是文件名，就是下面命令中的 log； kubectl -n default create configmap log-demo-set --from-file=log），
// logDefaultLevel --> 若没有配置字段，或误删除，会配置此 log 级别
func NewWithSharedInformerFactory(ctx context.Context, factory informers.SharedInformerFactory, cmNamespace, cmName, cmLogKey, logDefaultLevel string) LogInterface {
	c := &LogController{
		ctx:       ctx,
		cmLister:  factory.Core().V1().ConfigMaps().Lister(),
		cmInfomer: factory.Core().V1().ConfigMaps().Informer(),
		cmChan:    make(chan *corev1.ConfigMap, 10),
		cmInfo: &ConfigMapInfo{
			name:          cmName,
			namespace:     cmNamespace,
			logKey:        cmLogKey,
			cm:            &corev1.ConfigMap{},
			defalultLevel: logDefaultLevel,
			partLevelMap:  make(map[string]string),
		},
	}

	if _, ok := LogLevelMap[strings.ToUpper(logDefaultLevel)]; !ok {
		c.cmInfo.defalultLevel = DefaultInfoLevel
	}

	// Add ConfigMap event handler.
	c.cmInfomer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.add,
		UpdateFunc: c.update,
		DeleteFunc: c.delete,
	})
	stopCh := make(chan struct{})
	c.runInit(stopCh)
	go c.runWithInformer()
	return c
}
