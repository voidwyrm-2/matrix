package remotemod

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/voidwyrm-2/matrix/api/internal"
)

type RemoteModVersionDependency struct {
	VersionId string `json:"version_id"`
	ProjectId string `json:"project_id"`
	Filename  string `json:"file_name"`
	Kind      string `json:"dependency_type"`
}

func (rmvd RemoteModVersionDependency) String() string {
	return fmt.Sprintf("{%s, %s, %s, %s}", rmvd.VersionId, rmvd.ProjectId, rmvd.Filename, rmvd.Kind)
}

type RemoteModVersionFile struct {
	Filename, Url string
}

func (rmvf RemoteModVersionFile) String() string {
	return fmt.Sprintf("{%s, '%s'}", rmvf.Filename, rmvf.Url)
}

type RemoteModVersion struct {
	Id            string
	VersionNumber string   `json:"version_number"`
	GameVersions  []string `json:"game_versions"`
	Loaders       []string
	Dependencies  []RemoteModVersionDependency
	Files         []RemoteModVersionFile
}

func (rmv RemoteModVersion) String() string {
	formattedFiles := []string{}
	for _, f := range rmv.Files {
		formattedFiles = append(formattedFiles, f.String())
	}

	formattedDependencies := []string{}
	for _, d := range rmv.Dependencies {
		formattedDependencies = append(formattedDependencies, d.String())
	}

	return fmt.Sprintf("id: %s\nversion: %s\nmodLoaders: %s\ngameVersions: %s\ndependancies: %s\nfiles: %s", rmv.Id, rmv.VersionNumber, strings.Join(rmv.Loaders, ", "), strings.Join(rmv.GameVersions, ", "), strings.Join(formattedDependencies, ", "), strings.Join(formattedFiles, ", "))
}

type RemoteMod struct {
	Id, Slug, Title, Description string
	GameVersions                 []string `json:"game_versions"`
	Categories, Loaders          []string
	Versions                     []RemoteModVersion `json:"-"`
}

func FromProject(idOrSlug string) (RemoteMod, error) {
	mod := RemoteMod{}
	versions := []RemoteModVersion{}

	modResp, err := internal.Download("https://api.modrinth.com/v2/project/" + idOrSlug)
	if err != nil {
		return RemoteMod{}, err
	}

	err = json.Unmarshal(modResp, &mod)
	if err != nil {
		return RemoteMod{}, err
	}

	versionsResp, err := internal.Download("https://api.modrinth.com/v2/project/" + idOrSlug + "/version")
	if err != nil {
		return RemoteMod{}, err
	}

	err = json.Unmarshal(versionsResp, &versions)
	if err != nil {
		return RemoteMod{}, err
	}

	mod.Versions = versions

	return mod, nil
}
