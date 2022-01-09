package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"

	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"sigs.k8s.io/kustomize/api/hasher"
	"sigs.k8s.io/kustomize/api/resmap"
	"sigs.k8s.io/kustomize/api/resource"

	getter "github.com/hashicorp/go-getter"
	kyaml "sigs.k8s.io/kustomize/kyaml/yaml"
)

var DefaultResource = RemoteResource{
	Name:         "",
	Repo:         "",
	RepoCreds:    "",
	Update:       false,
	Template:     "",
	TemplateGlob: "*.jsonnet",
	//TemplateOpts:    "",
	Kinds:   []string{"!namespace"},
	destDir: "",
}

// RenderPlugin is a plugin to generate k8s resources
// from a remote or local go templates.
type RenderPlugin struct {
	PluginConfig

	sourcesDir string
	renderTemp string
	rf         *resmap.Factory
	rh         *resmap.PluginHelpers
}

// RemoteResource is specification for remote templates (git, s3, http...)
type RemoteResource struct {
	// local name for remote
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// go-getter compatible uri to remote
	Repo string `json:"repo" yaml:"repo"`
	// go-getter creds profile for private repos, s3, etc..
	RepoCreds string `json:"repoCreds" yaml:"repoCreds"`
	// whether to update existing source
	Update bool `json:"update,omitempty" yaml:"update,omitempty"`
	// template
	Template     string `json:"template,omitempty" yaml:"template,omitempty"`
	TemplateGlob string `json:"templateGlob,omitempty" yaml:"templateGlob,omitempty"`
	//TemplateOpts string `json:"templateOpts,omitempty" yaml:"templateOpts,omitempty"` // PLACEHOLDER

	// kinds
	Kinds []string `json:"kinds,omitempty" yaml:"kinds,omitempty"`

	// destDir is where the resource is cloned
	destDir string
}

type PluginConfig struct {
	Sources []RemoteResource `json:"sources,omitempty" yaml:"sources,omitempty"`
}

// Config uses the input plugin configurations `config` to setup the generator
func (p *RenderPlugin) Config(h *resmap.PluginHelpers, config []byte) error {
	p.rh = h
	err := kyaml.Unmarshal(config, p)
	if err != nil {
		return err
	}
	return nil
}

// Generate fetch, render and return manifests from remote sources
func (p *RenderPlugin) Generate() (resMap resmap.ResMap, err error) {

	//DEBUG
	debug, _ := strconv.ParseBool(os.Getenv("DEBUG"))

	// tempdir render
	p.renderTemp, err = ensureWorkDir(os.Getenv("RENDER_TEMP"))
	if err != nil {
		return nil, fmt.Errorf("failed to create render dir: %w", err)
	}
	if !debug {
		defer os.RemoveAll(p.renderTemp)
	}

	// tempdir sources
	p.sourcesDir, err = ensureWorkDir(os.Getenv("SOURCES_DIR"))
	if err != nil {
		return nil, fmt.Errorf("failed to create sources dir: %w", err)
	}

	// update,validate source spec
	err = p.EvalSources()
	if err != nil {
		return nil, fmt.Errorf("failed to validate source: %w", err)
	}

	// fetch dependencies
	err = p.FetchSources()
	if err != nil {
		return nil, fmt.Errorf("failed to get remote source: %w", err)
	}

	p.rf = NewResMapFactory()
	resMap, err = p.RenderSources()
	if err != nil {
		return nil, fmt.Errorf("template rendering failed: %v", err)
	}
	return resMap, err
}

// Eval sources to update/enrich/validate
func (p *RenderPlugin) EvalSources() (err error) {
	for idx, rs := range p.Sources {
		// update destDir
		p.Sources[idx].destDir = filepath.Join(p.sourcesDir, rs.Name)

		// normalize Values for rendering, ie: `server:port:111` to `.server_port: 111`
		// map[string]interface{} keys are flattened to single level key `_` delimited
		nv := make(map[string]interface{})
		if len(rs.FlattenValuesBy) > 0 {
			FlattenMap(rs.FlattenValuesBy, p.Values, nv)
		} else {
			FlattenMap(DefaultResource.FlattenValuesBy, p.Values, nv)
		}
		p.Values = nv
	}
	return nil
}

// FetchSources calls go-getter to fetch remote sources
func (p *RenderPlugin) FetchSources() (err error) {
	for _, rs := range p.Sources {

		// ensure fetch destination (ie: <sourcesDir>/<sourceName>)
		// if _, err := os.Stat(rs.destDir); os.IsNotExist(err) {
		// 	_ = os.MkdirAll(rs.destDir, 0770)
		// }

		// skip if update is not requested
		updateSource, err := strconv.ParseBool(os.Getenv("UPDATE_SOURCE"))
		if err != nil {
			updateSource = false
		}
		_, err = os.Stat(rs.destDir)
		if !os.IsNotExist(err) && (!rs.Update || !updateSource) {
			continue
		}

		//fetch
		pwd, err := os.Getwd()
		if err != nil {
			return err
		}
		opts := []getter.ClientOption{}

		//options ...
		//https://github.com/hashicorp/go-getter/blob/main/cmd/go-getter/main.go
		//https://github.com/hashicorp/go-getter#git-git
		gettercreds, err := getRepoCreds(rs.RepoCreds)
		if err != nil {
			return err
		}

		client := &getter.Client{
			Ctx:     context.TODO(),
			Src:     rs.Repo + gettercreds,
			Dst:     rs.destDir,
			Pwd:     pwd,
			Mode:    getter.ClientModeAny,
			Options: opts,
		}

		err = client.Get()
		if err != nil {
			return fmt.Errorf("failed to fetch source %w", err)
		}
	}
	return nil
}

// RenderSources render jsonnet manifests
func (p *RenderPlugin) RenderSources() (resMap resmap.ResMap, err error) {
	var out bytes.Buffer
	resMap = resmap.New()
	for _, rs := range p.Sources {

		if rs.TemplateGlob == "" {
			rs.TemplateGlob = DefaultResource.TemplateGlob
		}

		// find templates
		templates, err := TemplateFinder(rs.destDir, rs.TemplateGlob)
		if os.IsNotExist(err) {
			continue
		} else if err != nil {
			return nil, err
		}

		// actual render
		// TODO, add render engine gomplate
		for _, t := range templates {
			out.WriteString("\n---\n")
			err := p.JsonnetRenderBuf(t, &out)
			if err != nil {
				return nil, err
			}
		}

		// convert to resMap (bytes)
		resMapSrc, err := p.rf.NewResMapFromBytes(out.Bytes())
		if err != nil {
			return nil, err
		}

		// filter kinds
		if len(rs.Kinds) == 0 {
			rs.Kinds = DefaultResource.Kinds
		}
		err = p.FilterKinds(rs.Kinds, resMapSrc)
		if err != nil {
			return nil, fmt.Errorf("failed to filter kinds %s for source %s", strings.Join(rs.Kinds, ","), rs.Name)
		}

		// append single source to output
		err = resMap.AppendAll(resMapSrc)
		if err != nil {
			return nil, err
		}
	}

	// convert to kyaml resource map
	return resMap, nil
}

// JsonnetRenderBuf process templates to buffer
func (p *RenderPlugin) JsonnetRenderBuf(t string, out *bytes.Buffer) error {

	// read template
	tContent, err := ioutil.ReadFile(t)
	if err != nil {
		return fmt.Errorf("read template failed: %w", err)
	}

	// TODO, FIXME, implementation JsonnetRenderBuf

	return nil
}

// FilterKinds
// https://kubectl.docs.kubernetes.io/faq/kustomize/eschewedfeatures/#removal-directives
// Kustomize lacks resource removal and multiple namespace manifests from bases, causing
// `already registered id: ~G_v1_Namespace|~X|sre\`
func (p *RenderPlugin) FilterKinds(kinds []string, rm resmap.ResMap) error {

	// kinds -> lowercase
	//var kindsLcs []string
	//return fmt.Errorf("FilterKinds, Not Implemented")

	// per kinds item in soruce config
	// - !namespace,secrets    # to remove
	// - Deployment,ConfigMap  # only to keep, no glob
	for _, kindsItem := range kinds {
		negativeF := strings.Contains(kindsItem, "!")
		kindsItem := strings.ToLower(kindsItem)
		kindsItem = strings.ReplaceAll(kindsItem, "!", "")
		kindsList := strings.Split(kindsItem, ",")

		// across all resoures
		resources := rm.Resources()
		for r := range resources {
			k := strings.ToLower(resources[r].GetKind())
			if filterListFn(kindsList, negativeF, k) {
				rm.Remove(resources[r].CurId())
			}
		}
	}
	return nil
}

func NewResMapFactory() *resmap.Factory {
	resourceFactory := resource.NewFactory(&hasher.Hasher{})
	resourceFactory.IncludeLocalConfigs = true
	return resmap.NewFactory(resourceFactory)
}

// UTILS

//TemplateFinder returns list of files matching regex pattern
func TemplateFinder(root, pattern string) (found []string, err error) {
	err = filepath.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			if matched, err := filepath.Match(pattern, filepath.Base(path)); err != nil {
				return err
			} else if matched {
				found = append(found, path)
				return nil
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return found, nil
}

// removeFilter returns true matching kinds
func filterListFn(list []string, negativeFilter bool, k string) bool {
	k = strings.TrimSpace(k)
	if negativeFilter {
		if stringInSlice(k, list) {
			return true
		}
	} else {
		if !stringInSlice(k, list) {
			return true
		}
	}
	return false
}

// ensureWorkDir prepare working directory
func ensureWorkDir(dir string) (string, error) {
	var err error
	if dir == "" {
		dir, err = ioutil.TempDir("", "fnJsonnetRender_")
		if err != nil {
			return "", err
		}
	} else {
		// create if missing
		if _, err := os.Stat(dir); os.IsNotExist(err) {
			err := os.MkdirAll(dir, 0770)
			if err != nil {
				return "", err
			}
		}
	}
	return dir, nil
}

//getRepoCreds read reference to credentials and returns go-getter URI
func getRepoCreds(repoCreds string) (string, error) {
	var cr = ""
	if repoCreds != "" {
		for _, e := range strings.Split(repoCreds, ",") {
			pair := strings.SplitN(e, "=", 2)
			//sshkey - for private git repositories
			if pair[0] == "sshkey" {
				key, err := ioutil.ReadFile(pair[1])
				if err != nil {
					return cr, err
				}
				keyb64 := base64.StdEncoding.EncodeToString([]byte(strings.TrimSpace(string(key))))
				cr = fmt.Sprintf("%s?sshkey=%s", cr, string(keyb64))
			}
		}
	}
	return cr, nil
}
