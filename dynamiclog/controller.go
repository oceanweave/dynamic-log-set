package dynamiclog

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	clientv1 "k8s.io/client-go/kubernetes/typed/core/v1"
	v1 "k8s.io/client-go/listers/core/v1"
	"k8s.io/client-go/tools/cache"
	"k8s.io/klog/v2"
	"strings"
)

var LogLevelMap = map[string]int{
	"DEBUG": 1,
	"INFO":  2,
	"WARN":  3,
	"ERROR": 4,
}

const (
	LogEnable        = 0
	LogDisable       = 10
	LogDebugLevel    = 1
	LogInfoLevel     = 2
	LogWarnLevel     = 3
	LogErrorLevel    = 4
	DefaultInfoLevel = "Info"
)

type LogInterface interface {
	EnableLogPrint(string, int) int
	KlogEnableLogPrint(string, int) klog.Level
	GetLogPartLevelMap() map[string]string
}

type LogController struct {
	client    clientv1.ConfigMapInterface // Used for clientset mode.
	cmLister  v1.ConfigMapLister
	cmInfomer cache.SharedIndexInformer // Used for informer mode.
	ctx       context.Context           // Context.
	cmInfo    *ConfigMapInfo
	cmChan    chan *corev1.ConfigMap // Used for informer mode to buffer ConfigMap.
}

type ConfigMapInfo struct {
	name          string // ConfigMap name.
	namespace     string // ConfigMap namespace.
	logKey        string
	defalultLevel string
	partLevelMap  map[string]string
	partList      []string
	rev           string // ConfigMap recent revision.
	cm            *corev1.ConfigMap
}

// nowLevel 为用户此处设置日志级别
// dynamicLevel 是 configmap 中 partName 对应的字段
// 此处根据 dynamicLevel 和 nowLevel 判断是否打印， nowLevel 高于或等于 dynamicLevel 时，此处日志才会打印
// 如 nowLevel = warn， dynamic = debug， 此处日志会打印
func (c *LogController) EnableLogPrint(partName string, nowLevel int) int {
	var dynamicLevel string
	var ok bool

	// 使用 configmap 中为设置的 partName， 就设置为 Info 日志级别
	if dynamicLevel, ok = c.cmInfo.partLevelMap[partName]; !ok {
		dynamicLevel = c.cmInfo.defalultLevel
	}

	if nowLevel >= LogLevelMap[strings.ToUpper(dynamicLevel)] {
		return LogEnable
	}
	return LogDisable
}

func (c *LogController) KlogEnableLogPrint(partName string, nowLevel int) klog.Level {
	var dynamicLevel string
	var ok bool

	// 使用 configmap 中为设置的 partName， 就设置为 Info 日志级别
	if dynamicLevel, ok = c.cmInfo.partLevelMap[partName]; !ok {
		fmt.Println(dynamicLevel)
		dynamicLevel = c.cmInfo.defalultLevel
	}

	fmt.Println(dynamicLevel)
	if nowLevel >= LogLevelMap[strings.ToUpper(dynamicLevel)] {
		return LogEnable
	}
	return LogDisable
}

func (c *LogController) GetLogPartLevelMap() map[string]string {
	return c.cmInfo.partLevelMap
}

func (c *LogController) GetLogPartList() []string {
	return c.cmInfo.partList
}

// update handle ConfigMap add event.
func (c *LogController) add(obj interface{}) {
	if cm, ok := obj.(*corev1.ConfigMap); ok && cm.Name == c.cmInfo.name && cm.Namespace == c.cmInfo.namespace {
		c.cmChan <- cm
	}
}

// update handle ConfigMap delete event.
func (c *LogController) delete(obj interface{}) {
	if cm, ok := obj.(*corev1.ConfigMap); ok && cm.Name == c.cmInfo.name && cm.Namespace == c.cmInfo.namespace {
		c.cmInfo.rev = ""
		// 当检测到 configmap 删除时，自动将所有字段设置为 默认级别
		for key, _ := range c.cmInfo.partLevelMap {
			c.cmInfo.partLevelMap[key] = c.cmInfo.defalultLevel
		}
		c.cmInfo.cm = &corev1.ConfigMap{}
	}
}

// update handle ConfigMap update event.
func (c *LogController) update(oldObj, newObj interface{}) {
	// 检查是否为同一类型的资源
	if oldConfigMap, ok := oldObj.(*corev1.ConfigMap); ok {
		if newConfigMap, ok := newObj.(*corev1.ConfigMap); ok {
			// 在这里，你可以比较两个 ConfigMap 对象的特定字段来判断是否更新
			if oldConfigMap.ResourceVersion != newConfigMap.ResourceVersion {
				// 资源已更新，可以执行相应的处理逻辑
				ns := newConfigMap.Namespace
				name := newConfigMap.Name
				if ns == c.cmInfo.namespace && name == c.cmInfo.name {
					c.cmChan <- newConfigMap
				}
			}
		}
	}
}

// runWithInformer handle ConfigMap changes send by informer.
func (c *LogController) runWithInformer() {
	for {
		select {
		case cm := <-c.cmChan:
			c.parse(cm)
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *LogController) runInit(stopCh <-chan struct{}) {
	// 获取已存在的 ConfigMap 资源列表
	go c.cmInfomer.Run(stopCh)
	fmt.Println("开始初始化")
	if !cache.WaitForCacheSync(stopCh, c.cmInfomer.HasSynced) {
		fmt.Println("Timed out waiting for caches to sync")
		return
	}
	// 检查 HasSynced 的值
	if !c.cmInfomer.HasSynced() {
		fmt.Println("Informer has not synced yet")
		return
	}
	existingConfigs, _ := c.cmLister.List(labels.Everything())
	//fmt.Println("Existing ConfigMaps:", existingConfigs)
	//fmt.Println("Existing ConfigMaps:")
	fmt.Println("cminfo:", c.cmInfo.name, c.cmInfo.namespace)
	for _, cm := range existingConfigs {
		//fmt.Println("allcm:", cm.Name, cm.Namespace)
		if cm.Namespace == c.cmInfo.namespace && cm.Name == c.cmInfo.name {
			fmt.Println(cm)
			c.parse(cm)
		}
		//if !ok {
		//	fmt.Println("Error casting object to ConfigMap")
		//	continue
		//}
		//fmt.Printf("ConfigMap: %s in namespace %s\n", cm.Name, cm.Namespace)
		//if
		//c.parse(cm)
	}
}

func (c *LogController) parse(cm *corev1.ConfigMap) {
	c.cmInfo.rev = cm.ResourceVersion
	c.cmInfo.cm = cm
	c.cmInfo.parseConfigLogData()
	//fmt.Println("parse:", c.cminfo.cm)
}

func (cmi *ConfigMapInfo) parseConfigLogData() {
	cmi.partLevelMap = make(map[string]string)

	// 获取该 configmap 中指定 key 的内容
	lines := strings.Split(cmi.cm.Data[cmi.logKey], "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			cmi.partLevelMap[key] = value
			cmi.partList = append(cmi.partList, key)
		}
	}
}
