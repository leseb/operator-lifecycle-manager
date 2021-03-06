//go:generate counterfeiter -o fakes/fake_registry_client.go ../../../../vendor/github.com/operator-framework/operator-registry/pkg/client/client.go Interface
package resolver

import (
	"context"
	"fmt"

	"github.com/operator-framework/operator-registry/pkg/client"
	opregistry "github.com/operator-framework/operator-registry/pkg/registry"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type SourceRef struct {
	Address     string
	Client      client.Interface
	LastConnect metav1.Time
	LastHealthy metav1.Time
}

type SourceQuerier interface {
	FindProvider(api opregistry.APIKey) (*opregistry.Bundle, *CatalogKey, error)
	FindBundle(pkgName, channelName, bundleName string, initialSource CatalogKey) (*opregistry.Bundle, *CatalogKey, error)
	FindLatestBundle(pkgName, channelName string, initialSource CatalogKey) (*opregistry.Bundle, *CatalogKey, error)
	FindReplacement(bundleName, pkgName, channelName string, initialSource CatalogKey) (*opregistry.Bundle, *CatalogKey, error)
	Queryable() error
}

type NamespaceSourceQuerier struct {
	sources map[CatalogKey]client.Interface
}

var _ SourceQuerier = &NamespaceSourceQuerier{}

func NewNamespaceSourceQuerier(sources map[CatalogKey]client.Interface) *NamespaceSourceQuerier {
	return &NamespaceSourceQuerier{
		sources: sources,
	}
}

func (q *NamespaceSourceQuerier) Queryable() error {
	if len(q.sources) == 0 {
		return fmt.Errorf("no catalog sources available")
	}
	return nil
}

func (q *NamespaceSourceQuerier) FindProvider(api opregistry.APIKey) (*opregistry.Bundle, *CatalogKey, error) {
	for key, source := range q.sources {
		if bundle, err := source.GetBundleThatProvides(context.TODO(), api.Group, api.Version, api.Kind); err == nil {
			return bundle, &key, nil
		}
		if bundle, err := source.GetBundleThatProvides(context.TODO(), api.Plural+"."+api.Group, api.Version, api.Kind); err == nil {
			return bundle, &key, nil
		}
	}
	return nil, nil, fmt.Errorf("%s not provided by a package in any CatalogSource", api)
}

func (q *NamespaceSourceQuerier) FindPackage(pkgName, channelName, csvName string, initialSource CatalogKey) (*opregistry.Bundle, *CatalogKey, error) {
	if initialSource.Name != "" && initialSource.Namespace != "" {
		source, ok := q.sources[initialSource]
		if !ok {
			return nil, nil, fmt.Errorf("CatalogSource %s not found", initialSource)
		}

		var bundle *opregistry.Bundle
		var err error
		if csvName != "" {
			bundle, err = source.GetBundle(context.TODO(), pkgName, channelName, csvName)
		} else {
			bundle, err = source.GetBundleInPackageChannel(context.TODO(), pkgName, channelName)
		}

		if err != nil {
			return nil, nil, err
		}
		return bundle, &initialSource, nil
	}

	for key, source := range q.sources {
		var bundle *opregistry.Bundle
		var err error
		if csvName != "" {
			bundle, err = source.GetBundle(context.TODO(), pkgName, channelName, csvName)
		} else {
			bundle, err = source.GetBundleInPackageChannel(context.TODO(), pkgName, channelName)
		}
		if err == nil {
			return bundle, &key, nil
		}
	}
	return nil, nil, fmt.Errorf("%s/%s not found in any available CatalogSource", pkgName, channelName)
}

func (q *NamespaceSourceQuerier) FindBundle(pkgName, channelName, bundleName string, initialSource CatalogKey) (*opregistry.Bundle, *CatalogKey, error) {
	if initialSource.Name != "" && initialSource.Namespace != "" {
		source, ok := q.sources[initialSource]
		if !ok {
			return nil, nil, fmt.Errorf("CatalogSource %s not found", initialSource)
		}

		bundle, err := source.GetBundle(context.TODO(), pkgName, channelName, bundleName)
		if err != nil {
			return nil, nil, err
		}
		return bundle, &initialSource, nil
	}

	for key, source := range q.sources {
		bundle, err := source.GetBundle(context.TODO(), pkgName, channelName, bundleName)
		if err == nil {
			return bundle, &key, nil
		}
	}
	return nil, nil, fmt.Errorf("%s/%s/%s not found in any available CatalogSource", pkgName, channelName, bundleName)
}

func (q *NamespaceSourceQuerier) FindLatestBundle(pkgName, channelName string, initialSource CatalogKey) (*opregistry.Bundle, *CatalogKey, error) {
	if initialSource.Name != "" && initialSource.Namespace != "" {
		source, ok := q.sources[initialSource]
		if !ok {
			return nil, nil, fmt.Errorf("CatalogSource %s not found", initialSource)
		}

		bundle, err := source.GetBundleInPackageChannel(context.TODO(), pkgName, channelName)
		if err != nil {
			return nil, nil, err
		}
		return bundle, &initialSource, nil
	}

	for key, source := range q.sources {
		bundle, err := source.GetBundleInPackageChannel(context.TODO(), pkgName, channelName)
		if err == nil {
			return bundle, &key, nil
		}
	}
	return nil, nil, fmt.Errorf("%s/%s not found in any available CatalogSource", pkgName, channelName)
}

func (q *NamespaceSourceQuerier) FindReplacement(bundleName, pkgName, channelName string, initialSource CatalogKey) (*opregistry.Bundle, *CatalogKey, error) {
	if initialSource.Name != "" && initialSource.Namespace != "" {
		source, ok := q.sources[initialSource]
		if !ok {
			return nil, nil, fmt.Errorf("CatalogSource %s not found", initialSource.Name)
		}
		bundle, err := source.GetReplacementBundleInPackageChannel(context.TODO(), bundleName, pkgName, channelName)
		if err != nil {
			return nil, nil, err
		}
		return bundle, &initialSource, nil
	}

	for key, source := range q.sources {
		bundle, err := source.GetReplacementBundleInPackageChannel(context.TODO(), bundleName, pkgName, channelName)
		if err == nil {
			return bundle, &key, nil
		}
	}
	return nil, nil, fmt.Errorf("%s/%s not found in any available CatalogSource", pkgName, channelName)
}
