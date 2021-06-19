/*
Copyright 2018 The Kubernetes Authors.

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

package generators

import (
	"fmt"
	"path/filepath"

	"k8s.io/gengo/args"
	"k8s.io/gengo/generator"
	"k8s.io/gengo/namer"
	"k8s.io/gengo/types"
	"k8s.io/klog/v2"

	generatorargs "k8s.io/kube-openapi/cmd/openapi-gen/args"
)

type identityNamer struct{}

func (_ identityNamer) Name(t *types.Type) string {
	return t.Name.String()
}

var _ namer.Namer = identityNamer{}

// NameSystems returns the name system used by the generators in this package.
func NameSystems() namer.NameSystems {
	return namer.NameSystems{
		"raw":           namer.NewRawNamer("", nil),
		"sorting_namer": identityNamer{},
	}
}

// DefaultNameSystem returns the default name system for ordering the types to be
// processed by the generators in this package.
func DefaultNameSystem() string {
	return "sorting_namer"
}

func Packages(context *generator.Context, arguments *args.GeneratorArgs) generator.Packages {
	boilerplate, err := arguments.LoadGoBoilerplate()
	if err != nil {
		klog.Fatalf("Failed loading boilerplate: %v", err)
	}
	header := append([]byte(fmt.Sprintf("// +build !%s\n\n", arguments.GeneratedBuildTag)), boilerplate...)
	header = append(header, []byte(
		`
// This file was autogenerated by openapi-gen. Do not edit it manually!

`)...)

	reportPath := "-"
	var featureGateFileNames []string
	if customArgs, ok := arguments.CustomArgs.(*generatorargs.CustomArgs); ok {
		reportPath = customArgs.ReportFilename
		featureGateFileNames = customArgs.FeatureGateFileNames
	}
	context.FileTypes[apiViolationFileType] = apiViolationFile{
		unmangledPath: reportPath,
	}

	return generator.Packages{
		&generator.DefaultPackage{
			PackageName: filepath.Base(arguments.OutputPackagePath),
			PackagePath: arguments.OutputPackagePath,
			HeaderText:  header,
			GeneratorFunc: func(c *generator.Context) (generators []generator.Generator) {
				return []generator.Generator{
					newOpenAPIGen(
						arguments.OutputFileBaseName,
						arguments.OutputPackagePath,
					),
					newAPIViolationGen(featureGateFileNames),
				}
			},
			FilterFunc: apiTypeFilterFunc,
		},
	}
}
