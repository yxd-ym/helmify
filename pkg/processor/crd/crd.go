package crd

import (
	"bytes"
	"fmt"
	"io"

	"github.com/arttor/helmify/pkg/helmify"
	yamlformat "github.com/arttor/helmify/pkg/yaml"
	"github.com/pkg/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
)

const crdTeml = `apiVersion: apiextensions.k8s.io/v1
kind: CustomResourceDefinition
metadata:
  name: %[1]s
  annotations:
    controller-gen.kubebuilder.io/version: v0.4.1
  labels:
  {{- include "%[2]s.labels" . | nindent 4 }}
spec:
%[3]s
status:
  acceptedNames:
    kind: ""
    plural: ""
  conditions: []
  storedVersions: []`

var crdGVC = schema.GroupVersionKind{
	Group:   "apiextensions.k8s.io",
	Version: "v1",
	Kind:    "CustomResourceDefinition",
}

// New creates processor for k8s CustomResourceDefinition resource.
func New() helmify.Processor {
	return &crd{}
}

type crd struct{}

// Process k8s CustomResourceDefinition object into template. Returns false if not capable of processing given resource type.
func (c crd) Process(appMeta helmify.AppMetadata, obj *unstructured.Unstructured) (bool, helmify.Template, error) {
	if obj.GroupVersionKind() != crdGVC {
		return false, nil, nil
	}
	specUnstr, ok, err := unstructured.NestedMap(obj.Object, "spec")
	if err != nil || !ok {
		return true, nil, errors.Wrap(err, "unable to create crd template")
	}
	versions, _ := yaml.Marshal(specUnstr)
	versions = yamlformat.Indent(versions, 2)
	versions = bytes.TrimRight(versions, "\n ")

	res := fmt.Sprintf(crdTeml, obj.GetName(), appMeta.ChartName(), string(versions))
	name, _, err := unstructured.NestedString(obj.Object, "spec", "names", "singular")
	if err != nil || !ok {
		return true, nil, errors.Wrap(err, "unable to create crd template")
	}
	return true, &result{
		name: name + "-crd.yaml",
		data: []byte(res),
	}, nil
}

type result struct {
	name string
	data []byte
}

func (r *result) Filename() string {
	return r.name
}

func (r *result) Values() helmify.Values {
	return helmify.Values{}
}

func (r *result) Write(writer io.Writer) error {
	_, err := writer.Write(r.data)
	return err
}
