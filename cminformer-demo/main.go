package main

import (
	"context"
	"flag"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
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

	// 问题1 -- 获取集群中在 Informer 启动前已存在的资源，如 default namespace 下，名称为 my-config2 的 configmap
	// 方法1：通过 clientset 获取，最简单
	configMap, err := clientset.CoreV1().ConfigMaps("default").Get(context.TODO(), "my-config2", metav1.GetOptions{})
	// 打印 configmap 中的内容
	fmt.Println(configMap.Data)

	// 方法2：通过 Informer 获取
	// 创建 SharedInformerFactory
	sharedInformerFactory := informers.NewSharedInformerFactory(clientset, time.Second*30)
	// 监听 ConfigMap 的 Informer
	cmInformer := sharedInformerFactory.Core().V1().ConfigMaps().Informer()

	// 2.1 启动 Informer, 两种方法
	stopCh := make(chan struct{})
	defer close(stopCh)
	// 注意：
	// a. 以下两种方法都需要 go 协程运行，若没有开启 go 协程，将不会顺利启动（目前不清楚原因）
	// b. 若只启动 cmInformer，下面两种方法任选其一即可；都执行也可以的
	// c. 若需要启动多个 Informer，则可采用 sharedInformerFactory 方式，该方法会启动所有 Informer
	// 如又新建个监听 Node 资源的 Informer
	//nodeInformer := sharedInformerFactory.Core().V1().Nodes().Informer()
	//nodelister := sharedInformerFactory.Core().V1().Nodes().Lister()
	// 注意下面缓存同步函数，替换 2.2 部分的函数
	//if !cache.WaitForCacheSync(stopCh, cmInformer.HasSynced, nodeInformer.HasSynced) {
	//existingNodes, _ := nodelister.List(labels.Everything())
	//fmt.Println("Existing ConfigMaps:", existingNodes)

	// 2.1.1 方法1：通过 sharedInformerFactory 启动，此处会将所有通过 sharedInformerFactory 创建的 Informer 都启动
	go sharedInformerFactory.Start(stopCh)
	// 2.1.2 方法2：指定 Informer 启动
	go cmInformer.Run(stopCh)

	// 2.2 等待缓存同步完成，就是确保该 Informer 可以获取到目前集群种已存在的资源
	if !cache.WaitForCacheSync(stopCh, cmInformer.HasSynced) {
		fmt.Println("Timed out waiting for caches to sync")
		return
	}
	// 检查 HasSynced 的值，true 表示 Informer 的缓存已同步已存在的资源
	if !cmInformer.HasSynced() {
		fmt.Println("Informer has not synced yet")
		return
	}
	fmt.Println("上面同步完成")
	// 2.3 利用 Informer 的缓存，获取当前集群中已存在的资源，两种方法
	// 2.3.1 方法1：通过 lister 获取，当前已存在的 configmap 资源（ default namespace 下名为 my-config2 的 configmap）
	cmlister := sharedInformerFactory.Core().V1().ConfigMaps().Lister()
	// 筛选指定的 configmap
	configMap2, _ := cmlister.ConfigMaps("default").Get("my-config2")
	fmt.Println("Find specfic configMap:", configMap2)
	// 获取全部已存在的 configmap 资源
	existingConfigs, _ := cmlister.List(labels.Everything())
	fmt.Println("Existing ConfigMaps:", existingConfigs)

	// 2.3.2 方法2：通过 Informer 的缓存，手动构建想要查找的 configmap 的 key，并进行查找
	cmStore := cmInformer.GetStore()
	// 利用 configmap name 和 namespace，手动构建查询的 key
	cmKey, _ := cache.MetaNamespaceKeyFunc(&corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "my-config2",
			Namespace: "default",
		},
	})
	// 在 Informer 缓存中进行查找
	cmObj, _, err := cmStore.GetByKey(cmKey)
	if err != nil {
		fmt.Printf("Error getting ConfigMap from Store: %v\n", err)
		return
	}
	fmt.Println(cmObj.(*corev1.ConfigMap).Data)

	// 问题2 -- Informer 如何实现监听
	// Add ConfigMap event handler.
	c := ControllerSample{
		cmchan: make(chan *corev1.ConfigMap, 10),
		// 此处举例子，比如我关注如下信息的 Configmap，通过该 Informer 进行监控
		cmNamespace: "default",
		cmName:      "new-config",
	}
	cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.add,
		UpdateFunc: c.update,
		DeleteFunc: c.delete,
	})

	// 阻塞，用于 Informer 监控 Configmap 资源
	select {}
}

type ControllerSample struct {
	// 此处是你指定的要关注的 Configmap 的资源信息
	cmName      string
	cmNamespace string
	// 此处只是个举例子，若事件更新太快，就需要 workqueue 限速队列了
	cmchan chan *corev1.ConfigMap // Used for informer mode to buffer ConfigMap.
	ctx    context.Context
}

// runWithInformer handle ConfigMap changes send by informer.
func (c *ControllerSample) runWithInformer() {
	for {
		select {
		case cm := <-c.cmchan:
			c.parse(cm)
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *ControllerSample) parse(cm *corev1.ConfigMap) {
	fmt.Println("进行add 和 update 处理")
	fmt.Println("如打印新 Configmap 的值", cm.Data)
}

// update handle ConfigMap add event.
func (c *ControllerSample) add(obj interface{}) {
	// 判断是否是 Configmap 类型，以及是否是自己需要的资源
	if cm, ok := obj.(*corev1.ConfigMap); ok && cm.Name == c.cmName && cm.Namespace == c.cmNamespace {
		// 放入 channel 中等待处理
		c.cmchan <- cm
	}
}

// update handle ConfigMap delete event.
func (c *ControllerSample) delete(obj interface{}) {
	fmt.Println("delete event --> to do something")
}

// update handle ConfigMap update event.
func (c *ControllerSample) update(oldObj, newObj interface{}) {
	// 检查是否为同一类型的资源
	if oldConfigMap, ok := oldObj.(*corev1.ConfigMap); ok {
		if newConfigMap, ok := newObj.(*corev1.ConfigMap); ok {
			// 在这里，你可以比较两个 ConfigMap 对象的特定字段来判断是否更新
			if oldConfigMap.ResourceVersion != newConfigMap.ResourceVersion {
				// 资源已更新，可以执行相应的处理逻辑
				// 判断是否是自己需要的资源
				if newConfigMap.Name == c.cmName && newConfigMap.Namespace == c.cmNamespace {
					// 放入 channel 中等待处理，因此若是更新过快，就会有很多事件，所以就是需要限速队列了
					c.cmchan <- newConfigMap
				}
			}
		}
	}
}
