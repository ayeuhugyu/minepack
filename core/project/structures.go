package project

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"gopkg.in/yaml.v3"
)

// Project

type ModloaderVersion struct {
	Name    string
	Version string
}

type ProjectVersions struct {
	Game     string
	Loader   ModloaderVersion
	Minepack string
}

type Project struct {
	Name          string
	Description   string
	Author        string
	Root          string
	Versions      ProjectVersions
	DefaultSource string // "modrinth" or "curseforge"
}

func (p *Project) HasMod(idOrSlug string) bool {
	var sums *[]SummaryObject
	sums, err := ParseSum(p.Root)
	if err != nil {
		return false
	}
	for _, sum := range *sums {
		if sum.Id == idOrSlug || sum.Slug == idOrSlug {
			return true
		}
	}
	return false
}

// add the shorthand version to the sums file and the full file to root/content/slug.mp.yaml
func (p *Project) AddContent(content ContentData) error {
	var sums *[]SummaryObject
	sums, err := ParseSum(p.Root)
	if err != nil {
		return err
	}
	// add the shorthand version to the sums file
	for _, sum := range *sums {
		if sum.Slug == content.Slug {
			return fmt.Errorf("content with slug %s already exists", content.Slug)
		}
	}
	*sums = append(*sums, SummaryObject{
		Slug:        content.Slug,
		Id:          content.Id,
		ContentType: content.ContentType,
		Source:      content.Source,
	})
	// write the new sums file
	err = WriteSum(*sums, p.Root)
	if err != nil {
		return err
	}
	// add the full file to root/content/slug.mp.yaml
	fullPath := filepath.Join(p.Root, "content", fmt.Sprintf("%s.mp.yaml", content.Slug))

	contentFile, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer contentFile.Close()

	encoder := yaml.NewEncoder(contentFile)
	defer encoder.Close()

	return encoder.Encode(content)
}

func (p *Project) UpdateContent(content ContentData) error {
	var sums *[]SummaryObject
	sums, err := ParseSum(p.Root)
	if err != nil {
		return err
	}
	// update the shorthand version in the sums file
	for i, sum := range *sums {
		if sum.Slug == content.Slug {
			(*sums)[i] = SummaryObject{
				Slug:        content.Slug,
				Id:          content.Id,
				ContentType: content.ContentType,
				Source:      content.Source,
			}
			break
		}
	}
	err = WriteSum(*sums, p.Root)
	if err != nil {
		return err
	}
	// update the full file at root/content/slug.mp.yaml
	fullPath := filepath.Join(p.Root, "content", fmt.Sprintf("%s.mp.yaml", content.Slug))
	contentFile, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	defer contentFile.Close()

	encoder := yaml.NewEncoder(contentFile)
	defer encoder.Close()

	return encoder.Encode(content)
}

func (p *Project) RemoveContent(idOrSlug string) error {
	var sums *[]SummaryObject
	sums, err := ParseSum(p.Root)
	if err != nil {
		return err
	}
	// remove the shorthand version from the sums file
	for i, sum := range *sums {
		if sum.Slug == idOrSlug {
			*sums = append((*sums)[:i], (*sums)[i+1:]...)
			break
		}
	}
	err = WriteSum(*sums, p.Root)
	if err != nil {
		return err
	}
	// remove the full file from root/content/slug.mp.yaml
	fullPath := filepath.Join(p.Root, "content", fmt.Sprintf("%s.mp.yaml", idOrSlug))
	return os.Remove(fullPath)
}

func (p *Project) GetContent(idOrSlug string) (*ContentData, error) {
	fullPath := filepath.Join(p.Root, "content", fmt.Sprintf("%s.mp.yaml", idOrSlug))
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("content with slug %s does not exist", idOrSlug)
	}
	contentFile, err := os.Open(fullPath)
	if err != nil {
		return nil, err
	}
	defer contentFile.Close()

	var content ContentData
	decoder := yaml.NewDecoder(contentFile)
	if err := decoder.Decode(&content); err != nil {
		return nil, err
	}
	return &content, nil
}

func (p *Project) GetAllContent() ([]ContentData, error) {
	var sums *[]SummaryObject
	sums, err := ParseSum(p.Root)
	if err != nil {
		return nil, err
	}

	// if no content, return empty slice
	if len(*sums) == 0 {
		return []ContentData{}, nil
	}

	// use channels and goroutines for parallel processing
	type result struct {
		content *ContentData
		err     error
		index   int
	}

	resultsChan := make(chan result, len(*sums))
	var wg sync.WaitGroup

	// start a goroutine for each content item
	for i, sum := range *sums {
		wg.Add(1)
		go func(index int, slug string) {
			defer wg.Done()
			content, err := p.GetContent(slug)
			resultsChan <- result{content: content, err: err, index: index}
		}(i, sum.Slug)
	}

	// close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// collect results and maintain order
	contentList := make([]ContentData, len(*sums))
	var firstError error

	for res := range resultsChan {
		if res.err != nil && firstError == nil {
			firstError = res.err
		}
		if res.content != nil {
			contentList[res.index] = *res.content
		}
	}

	if firstError != nil {
		return nil, firstError
	}

	return contentList, nil
}

// summary file

type SummaryObject struct {
	Slug        string      `yaml:"slug"`
	Id          string      `yaml:"id"`
	ContentType ContentType `yaml:"content_type"`
	Source      Source      `yaml:"source"`
}

// Mods

type ContentType int

const (
	Mod ContentType = iota
	Resourcepack
	Shaderpack
	Datapack
	World
)

func ContentTypeToString(ct ContentType) string {
	switch ct {
	case Mod:
		return "mod"
	case Resourcepack:
		return "resourcepack"
	case Shaderpack:
		return "shaderpack"
	case Datapack:
		return "datapack"
	case World:
		return "world"
	default:
		return "unknown"
	}
}
func StringToContentType(s string) ContentType {
	switch s {
	case "mod":
		return Mod
	case "resourcepack":
		return Resourcepack
	case "shaderpack":
		return Shaderpack
	case "datapack":
		return Datapack
	case "world":
		return World
	default:
		return -1
	}
}

type HashFormat int

const (
	SHA1 HashFormat = iota
	SHA256
	SHA512
	MD5
)

func HashFormatToString(hf HashFormat) string {
	switch hf {
	case SHA1:
		return "sha1"
	case SHA256:
		return "sha256"
	case SHA512:
		return "sha512"
	case MD5:
		return "md5"
	default:
		return "unknown"
	}
}
func StringToHashFormat(s string) HashFormat {
	switch s {
	case "sha1":
		return SHA1
	case "sha256":
		return SHA256
	case "sha512":
		return SHA512
	case "md5":
		return MD5
	default:
		return -1
	}
}

type ModSide int

const (
	None ModSide = iota
	Client
	Server
	Both
)

func ModSideToString(ms ModSide) string {
	switch ms {
	case None:
		return "none"
	case Client:
		return "client"
	case Server:
		return "server"
	case Both:
		return "both"
	default:
		return "unknown"
	}
}
func StringToModSide(s string) ModSide {
	switch s {
	case "none":
		return None
	case "client":
		return Client
	case "server":
		return Server
	case "both":
		return Both
	default:
		return -1
	}
}

type Source int

const (
	Modrinth Source = iota
	Curseforge
	Custom
)

func SourceToString(s Source) string {
	switch s {
	case Modrinth:
		return "modrinth"
	case Curseforge:
		return "curseforge"
	case Custom:
		return "custom"
	default:
		return "unknown"
	}
}
func StringToSource(s string) Source {
	switch s {
	case "modrinth":
		return Modrinth
	case "curseforge":
		return Curseforge
	case "custom":
		return Custom
	default:
		return -1
	}
}

type Hashes struct {
	Sha1   string
	Sha256 string
	Sha512 string
	Md5    string
}

type FileData struct {
	Filename string
	Filesize int64
	Filepath string // relative to project root
	Hashes   Hashes
}

type DependencyType int

const (
	Required DependencyType = iota
	Optional
	Embedded
	Incompatible
)

func DependencyTypeToString(dep DependencyType) string {
	switch dep {
	case Required:
		return "required"
	case Optional:
		return "optional"
	case Incompatible:
		return "incompatible"
	case Embedded:
		return "embedded"
	default:
		return "unknown"
	}
}

type Dependency struct {
	Name           string
	Slug           string
	Id             string
	DependencyType DependencyType
}

type RequiredBy struct {
	Name string
	Slug string
	Id   string
}

type ContentData struct {
	// todo: finish
	ContentType       ContentType
	Name              string
	Id                string
	Slug              string
	Side              ModSide
	PageUrl           string
	DownloadUrl       string
	VersionId         string
	Source            Source
	File              FileData
	Dependencies      []Dependency
	RequiredBy        []RequiredBy
	AddedAsDependency bool
}

type Manifest struct {
	ContentDirectory string
	Content          []ContentData
}
