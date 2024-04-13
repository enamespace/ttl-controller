/*
Copyright The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by lister-gen. DO NOT EDIT.

package v1alpha1

import (
	v1alpha1 "github.com/enamespace/ttl-controller/pkg/apis/ttlcontroller/v1alpha1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/client-go/tools/cache"
)

// TTLLister helps list TTLs.
// All objects returned here must be treated as read-only.
type TTLLister interface {
	// List lists all TTLs in the indexer.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.TTL, err error)
	// TTLs returns an object that can list and get TTLs.
	TTLs(namespace string) TTLNamespaceLister
	TTLListerExpansion
}

// tTLLister implements the TTLLister interface.
type tTLLister struct {
	indexer cache.Indexer
}

// NewTTLLister returns a new TTLLister.
func NewTTLLister(indexer cache.Indexer) TTLLister {
	return &tTLLister{indexer: indexer}
}

// List lists all TTLs in the indexer.
func (s *tTLLister) List(selector labels.Selector) (ret []*v1alpha1.TTL, err error) {
	err = cache.ListAll(s.indexer, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.TTL))
	})
	return ret, err
}

// TTLs returns an object that can list and get TTLs.
func (s *tTLLister) TTLs(namespace string) TTLNamespaceLister {
	return tTLNamespaceLister{indexer: s.indexer, namespace: namespace}
}

// TTLNamespaceLister helps list and get TTLs.
// All objects returned here must be treated as read-only.
type TTLNamespaceLister interface {
	// List lists all TTLs in the indexer for a given namespace.
	// Objects returned here must be treated as read-only.
	List(selector labels.Selector) (ret []*v1alpha1.TTL, err error)
	// Get retrieves the TTL from the indexer for a given namespace and name.
	// Objects returned here must be treated as read-only.
	Get(name string) (*v1alpha1.TTL, error)
	TTLNamespaceListerExpansion
}

// tTLNamespaceLister implements the TTLNamespaceLister
// interface.
type tTLNamespaceLister struct {
	indexer   cache.Indexer
	namespace string
}

// List lists all TTLs in the indexer for a given namespace.
func (s tTLNamespaceLister) List(selector labels.Selector) (ret []*v1alpha1.TTL, err error) {
	err = cache.ListAllByNamespace(s.indexer, s.namespace, selector, func(m interface{}) {
		ret = append(ret, m.(*v1alpha1.TTL))
	})
	return ret, err
}

// Get retrieves the TTL from the indexer for a given namespace and name.
func (s tTLNamespaceLister) Get(name string) (*v1alpha1.TTL, error) {
	obj, exists, err := s.indexer.GetByKey(s.namespace + "/" + name)
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.NewNotFound(v1alpha1.Resource("ttl"), name)
	}
	return obj.(*v1alpha1.TTL), nil
}