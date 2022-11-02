// Copyright 2016-2022, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"

	"github.com/spf13/cobra"

	javagen "github.com/pulumi/pulumi-java/pkg/codegen/java"
	yamlgen "github.com/pulumi/pulumi-yaml/pkg/pulumiyaml/codegen"
	gogen "github.com/pulumi/pulumi/pkg/v3/codegen/go"
	"github.com/pulumi/pulumi/pkg/v3/codegen/nodejs"
	"github.com/pulumi/pulumi/pkg/v3/codegen/pcl"
	"github.com/pulumi/pulumi/pkg/v3/codegen/python"
	"github.com/pulumi/pulumi/pkg/v3/codegen/schema"
	"github.com/pulumi/pulumi/pkg/v3/engine"
	"github.com/pulumi/pulumi/sdk/v3/go/common/diag"
	"github.com/pulumi/pulumi/sdk/v3/go/common/encoding"
	"github.com/pulumi/pulumi/sdk/v3/go/common/resource/plugin"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/cmdutil"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/contract"
	"github.com/pulumi/pulumi/sdk/v3/go/common/util/result"
	"github.com/pulumi/pulumi/sdk/v3/go/common/workspace"
)

type projectGeneratorFunc func(directory string, project workspace.Project, p *pcl.Program) error

func newConvertCmd() *cobra.Command {
	var outDir string
	var language string
	var generateOnly bool

	cmd := &cobra.Command{
		Use:   "convert",
		Args:  cmdutil.MaximumNArgs(0),
		Short: "Convert Pulumi programs from YAML into other supported languages",
		Long: "Convert Pulumi programs from YAML into other supported languages.\n" +
			"\n" +
			"The YAML program to convert will default to the manifest in the current working directory.\n",
		Run: cmdutil.RunResultFunc(func(cmd *cobra.Command, args []string) result.Result {
			cwd, err := os.Getwd()
			if err != nil {
				return result.FromError(fmt.Errorf("could not resolve current working directory"))
			}

			return runConvert(cwd, language, outDir, generateOnly)
		}),
	}

	cmd.PersistentFlags().StringVar(
		//nolint:lll
		&language, "language", "", "Which language plugin to use to generate the pulumi project")
	if err := cmd.MarkPersistentFlagRequired("language"); err != nil {
		panic("failed to mark 'language' as a required flag")
	}

	cmd.PersistentFlags().StringVar(
		//nolint:lll
		&outDir, "out", ".", "The output directory to write the convert project to")

	cmd.PersistentFlags().BoolVar(
		//nolint:lll
		&generateOnly, "generate-only", false, "Generate the converted program(s) only; do not install dependencies")

	return cmd
}

// runConvert converts a Pulumi program from YAML into PCL without generating a full pcl.Program
func runConvertPcl(host plugin.Host, cwd string, outDir string) result.Result {
	loader := schema.NewPluginLoader(host)
	_, template, diags, err := yamlgen.LoadTemplate(cwd)
	if err != nil {
		return result.FromError(err)
	}

	if diags.HasErrors() {
		return result.FromError(diags)
	}

	programText, diags, err := yamlgen.ConvertTemplateIL(template, loader)
	if err != nil {
		return result.FromError(err)
	}

	if outDir != "." {
		err := os.MkdirAll(outDir, 0755)
		if err != nil {
			return result.FromError(fmt.Errorf("could not create output directory: %w", err))
		}
	}

	outputFile := path.Join(outDir, "main.pp")
	err = ioutil.WriteFile(outputFile, []byte(programText), 0600)
	if err != nil {
		return result.FromError(fmt.Errorf("could not write output program: %w", err))
	}

	return nil
}

func runConvert(cwd string, language string, outDir string, generateOnly bool) result.Result {
	host, err := newPluginHost()
	if err != nil {
		return result.FromError(fmt.Errorf("could not create plugin host: %w", err))
	}
	defer contract.IgnoreClose(host)

	// Translate well known languages to runtimes
	switch language {
	case "csharp", "c#":
		language = "dotnet" // nolint: goconst
	}

	var projectGenerator projectGeneratorFunc
	switch language {
	case "go":
		projectGenerator = gogen.GenerateProject
	case "typescript":
		projectGenerator = nodejs.GenerateProject
	case "python": // nolint: goconst
		projectGenerator = python.GenerateProject
	case "java": // nolint: goconst
		projectGenerator = javagen.GenerateProject
	case "yaml": // nolint: goconst
		projectGenerator = yamlgen.GenerateProject
	case "pulumi", "pcl":
		if cmdutil.IsTruthy(os.Getenv("PULUMI_DEV")) {
			// since we don't need Eject to get the full program,
			// we can just convert the YAML directly to PCL
			return runConvertPcl(host, cwd, outDir)
		}
		return result.Errorf("cannot generate programs for %q language", language)
	default:
		projectGenerator = func(directory string, project workspace.Project, program *pcl.Program) error {
			// TODO: There's probably a way to go from pcl.Program to string but for now I'm just half
			// converting again, once everything is moved over we'll just skip making a pcl.Program here
			// completely anyway.
			loader := schema.NewPluginLoader(host)
			_, template, diags, err := yamlgen.LoadTemplate(cwd)
			if err != nil {
				return err
			}

			if diags.HasErrors() {
				return diags
			}

			programText, diags, err := yamlgen.ConvertTemplateIL(template, loader)
			if err != nil {
				return err
			}
			if diags.HasErrors() {
				return diags
			}

			languagePlugin, err := host.LanguageRuntime(language)
			if err != nil {
				return err
			}

			projectBytes, err := encoding.JSON.Marshal(project)
			if err != nil {
				return err
			}
			projectJSON := string(projectBytes)

			err = languagePlugin.GenerateProject(directory, projectJSON, programText)
			if err != nil {
				return err
			}

			return nil
		}
	}

	if outDir != "." {
		err := os.MkdirAll(outDir, 0755)
		if err != nil {
			return result.FromError(fmt.Errorf("could not create output directory: %w", err))
		}
	}

	loader := schema.NewPluginLoader(host)
	proj, pclProgram, err := yamlgen.Eject(cwd, loader)
	if err != nil {
		return result.FromError(fmt.Errorf("could not load yaml program: %w", err))
	}

	err = projectGenerator(outDir, *proj, pclProgram)
	if err != nil {
		return result.FromError(fmt.Errorf("could not generate output program: %w", err))
	}

	// Project should now exist at outDir. Run installDependencies in that directory
	// Change the working directory to the specified directory.
	if err := os.Chdir(outDir); err != nil {
		return result.FromError(fmt.Errorf("changing the working directory: %w", err))
	}

	// Load the project, to
	proj, root, err := readProject()
	if err != nil {
		return result.FromError(err)
	}

	projinfo := &engine.Projinfo{Proj: proj, Root: root}
	pwd, _, ctx, err := engine.ProjectInfoContext(projinfo, nil, cmdutil.Diag(), cmdutil.Diag(), false, nil)
	if err != nil {
		return result.FromError(err)
	}

	defer ctx.Close()

	if !generateOnly {
		if err := installDependencies(ctx, &proj.Runtime, pwd); err != nil {
			return result.FromError(err)
		}
	}

	return nil
}

func newPluginHost() (plugin.Host, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	sink := diag.DefaultSink(os.Stderr, os.Stderr, diag.FormatOptions{
		Color: cmdutil.GetGlobalColorization(),
	})
	pluginCtx, err := plugin.NewContext(sink, sink, nil, nil, cwd, nil, true, nil)
	if err != nil {
		return nil, err
	}
	return pluginCtx.Host, nil
}
