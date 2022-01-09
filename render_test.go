// Copyright 2022 Petr Michalec
// SPDX-License-Identifier: Apache-2.0

// DRAFT, NOT FINISHED

package main_test


import (
	"testing"
	"sigs.k8s.io/kustomize/kyaml/fn/framework"
  "sigs.k8s.io/kustomize/kyaml/yaml"
)

var defaultResourceList = framework.ResourceList{
  Items []*yaml.RNode{}
  FunctionConfig *yaml.RNode{}
  Results framework.Results{}
}

func testRenderFunction(config *yaml.RNode) (string, error) {
  var rl framework.ResourceList
  rl = defaultResourceList
  rl.FunctionConfig = config

  err := run(*rl)
  if err != nil {
    return "", err
  }

  y, err := yaml.Marshal(rl.Results)
  if err != nil {
    return errors.Wrap(err)
  }
  return y
}

func TestRenderPlugin(t *testing.T) {

  testcases := []struct {
		name   string
		config yaml.RNode
		output string
	}{
		{ 
      name := "default"
      config :=  yaml.ReadFile("./example/fnRenderJsonnet_test.yml")
      output :=`
---
apiVersion: v1
kind: Namespace
metadata:
  name: media
---
apiVersion: v1
data:
  nginx-config.yaml: '# DUMMY'
kind: ConfigMap
metadata:
  labels:
    app: nginx
  name: nginx-config
  namespace: media
`)

  for _, tc := range testcases {
    out, err := testRenderFunction(tc.input)
    if out != tc.output) {
      t.Errorf("in testcase %q, expect: %#v, but got: %#v", tc.name, tc.output, out)
    }
  }
}
