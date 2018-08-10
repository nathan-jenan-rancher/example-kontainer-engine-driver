package v3

import (
	"context"

	"github.com/rancher/norman/controller"
	"github.com/rancher/norman/objectclient"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/tools/cache"
)

var (
	KontainerDriverGroupVersionKind = schema.GroupVersionKind{
		Version: Version,
		Group:   GroupName,
		Kind:    "KontainerDriver",
	}
	KontainerDriverResource = metav1.APIResource{
		Name:         "kontainerdrivers",
		SingularName: "kontainerdriver",
		Namespaced:   false,
		Kind:         KontainerDriverGroupVersionKind.Kind,
	}
)

type KontainerDriverList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []KontainerDriver
}

type KontainerDriverHandlerFunc func(key string, obj *KontainerDriver) error

type KontainerDriverLister interface {
	List(namespace string, selector labels.Selector) (ret []*KontainerDriver, err error)
	Get(namespace, name string) (*KontainerDriver, error)
}

type KontainerDriverController interface {
	Informer() cache.SharedIndexInformer
	Lister() KontainerDriverLister
	AddHandler(name string, handler KontainerDriverHandlerFunc)
	AddClusterScopedHandler(name, clusterName string, handler KontainerDriverHandlerFunc)
	Enqueue(namespace, name string)
	Sync(ctx context.Context) error
	Start(ctx context.Context, threadiness int) error
}

type KontainerDriverInterface interface {
	ObjectClient() *objectclient.ObjectClient
	Create(*KontainerDriver) (*KontainerDriver, error)
	GetNamespaced(namespace, name string, opts metav1.GetOptions) (*KontainerDriver, error)
	Get(name string, opts metav1.GetOptions) (*KontainerDriver, error)
	Update(*KontainerDriver) (*KontainerDriver, error)
	Delete(name string, options *metav1.DeleteOptions) error
	DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error
	List(opts metav1.ListOptions) (*KontainerDriverList, error)
	Watch(opts metav1.ListOptions) (watch.Interface, error)
	DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error
	Controller() KontainerDriverController
	AddHandler(name string, sync KontainerDriverHandlerFunc)
	AddLifecycle(name string, lifecycle KontainerDriverLifecycle)
	AddClusterScopedHandler(name, clusterName string, sync KontainerDriverHandlerFunc)
	AddClusterScopedLifecycle(name, clusterName string, lifecycle KontainerDriverLifecycle)
}

type kontainerDriverLister struct {
	controller *kontainerDriverController
}

func (l *kontainerDriverLister) List(namespace string, selector labels.Selector) (ret []*KontainerDriver, err error) {
	err = cache.ListAllByNamespace(l.controller.Informer().GetIndexer(), namespace, selector, func(obj interface{}) {
		ret = append(ret, obj.(*KontainerDriver))
	})
	return
}

func (l *kontainerDriverLister) Get(namespace, name string) (*KontainerDriver, error) {
	var key string
	if namespace != "" {
		key = namespace + "/" + name
	} else {
		key = name
	}
	obj, exists, err := l.controller.Informer().GetIndexer().GetByKey(key)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(schema.GroupResource{
			Group:    KontainerDriverGroupVersionKind.Group,
			Resource: "kontainerDriver",
		}, key)
	}
	return obj.(*KontainerDriver), nil
}

type kontainerDriverController struct {
	controller.GenericController
}

func (c *kontainerDriverController) Lister() KontainerDriverLister {
	return &kontainerDriverLister{
		controller: c,
	}
}

func (c *kontainerDriverController) AddHandler(name string, handler KontainerDriverHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}
		return handler(key, obj.(*KontainerDriver))
	})
}

func (c *kontainerDriverController) AddClusterScopedHandler(name, cluster string, handler KontainerDriverHandlerFunc) {
	c.GenericController.AddHandler(name, func(key string) error {
		obj, exists, err := c.Informer().GetStore().GetByKey(key)
		if err != nil {
			return err
		}
		if !exists {
			return handler(key, nil)
		}

		if !controller.ObjectInCluster(cluster, obj) {
			return nil
		}

		return handler(key, obj.(*KontainerDriver))
	})
}

type kontainerDriverFactory struct {
}

func (c kontainerDriverFactory) Object() runtime.Object {
	return &KontainerDriver{}
}

func (c kontainerDriverFactory) List() runtime.Object {
	return &KontainerDriverList{}
}

func (s *kontainerDriverClient) Controller() KontainerDriverController {
	s.client.Lock()
	defer s.client.Unlock()

	c, ok := s.client.kontainerDriverControllers[s.ns]
	if ok {
		return c
	}

	genericController := controller.NewGenericController(KontainerDriverGroupVersionKind.Kind+"Controller",
		s.objectClient)

	c = &kontainerDriverController{
		GenericController: genericController,
	}

	s.client.kontainerDriverControllers[s.ns] = c
	s.client.starters = append(s.client.starters, c)

	return c
}

type kontainerDriverClient struct {
	client       *Client
	ns           string
	objectClient *objectclient.ObjectClient
	controller   KontainerDriverController
}

func (s *kontainerDriverClient) ObjectClient() *objectclient.ObjectClient {
	return s.objectClient
}

func (s *kontainerDriverClient) Create(o *KontainerDriver) (*KontainerDriver, error) {
	obj, err := s.objectClient.Create(o)
	return obj.(*KontainerDriver), err
}

func (s *kontainerDriverClient) Get(name string, opts metav1.GetOptions) (*KontainerDriver, error) {
	obj, err := s.objectClient.Get(name, opts)
	return obj.(*KontainerDriver), err
}

func (s *kontainerDriverClient) GetNamespaced(namespace, name string, opts metav1.GetOptions) (*KontainerDriver, error) {
	obj, err := s.objectClient.GetNamespaced(namespace, name, opts)
	return obj.(*KontainerDriver), err
}

func (s *kontainerDriverClient) Update(o *KontainerDriver) (*KontainerDriver, error) {
	obj, err := s.objectClient.Update(o.Name, o)
	return obj.(*KontainerDriver), err
}

func (s *kontainerDriverClient) Delete(name string, options *metav1.DeleteOptions) error {
	return s.objectClient.Delete(name, options)
}

func (s *kontainerDriverClient) DeleteNamespaced(namespace, name string, options *metav1.DeleteOptions) error {
	return s.objectClient.DeleteNamespaced(namespace, name, options)
}

func (s *kontainerDriverClient) List(opts metav1.ListOptions) (*KontainerDriverList, error) {
	obj, err := s.objectClient.List(opts)
	return obj.(*KontainerDriverList), err
}

func (s *kontainerDriverClient) Watch(opts metav1.ListOptions) (watch.Interface, error) {
	return s.objectClient.Watch(opts)
}

// Patch applies the patch and returns the patched deployment.
func (s *kontainerDriverClient) Patch(o *KontainerDriver, data []byte, subresources ...string) (*KontainerDriver, error) {
	obj, err := s.objectClient.Patch(o.Name, o, data, subresources...)
	return obj.(*KontainerDriver), err
}

func (s *kontainerDriverClient) DeleteCollection(deleteOpts *metav1.DeleteOptions, listOpts metav1.ListOptions) error {
	return s.objectClient.DeleteCollection(deleteOpts, listOpts)
}

func (s *kontainerDriverClient) AddHandler(name string, sync KontainerDriverHandlerFunc) {
	s.Controller().AddHandler(name, sync)
}

func (s *kontainerDriverClient) AddLifecycle(name string, lifecycle KontainerDriverLifecycle) {
	sync := NewKontainerDriverLifecycleAdapter(name, false, s, lifecycle)
	s.AddHandler(name, sync)
}

func (s *kontainerDriverClient) AddClusterScopedHandler(name, clusterName string, sync KontainerDriverHandlerFunc) {
	s.Controller().AddClusterScopedHandler(name, clusterName, sync)
}

func (s *kontainerDriverClient) AddClusterScopedLifecycle(name, clusterName string, lifecycle KontainerDriverLifecycle) {
	sync := NewKontainerDriverLifecycleAdapter(name+"_"+clusterName, true, s, lifecycle)
	s.AddClusterScopedHandler(name, clusterName, sync)
}
