package localmod

import (
	"encoding/json"
	"fmt"
	"log"
	"slices"
	"strings"

	"github.com/voidwyrm-2/matrix/api/internal"
	"github.com/voidwyrm-2/matrix/api/remotemod"
	"github.com/voidwyrm-2/matrix/api/version"
)

var customProcs = map[string]func(s string) string{
	"create": func(s string) string {
		return strings.Split(s, "-")[1]
	},

	"petrols-parts": func(s string) string {
		return strings.Split(s, "-")[1]
	},
	"create-mechanical-extruder": func(s string) string {
		return strings.Split(s, "-")[1]
	},
	"create-mechanical-spawner": func(s string) string {
		return strings.Split(s, "-")[1]
	},
	"create-dreams-and-desires": func(s string) string {
		if strings.Contains(s, "-") {
			s = strings.Split(s, "-")[1]
		}

		return strings.Join(strings.Split(s, ".")[:2], ".")
	},
}

type LocalMod struct {
	name, desc, id, slug, forceVersion string
	version                            version.Version
}

func New(name, desc, id, slug, forceVersion, mVersion string) (LocalMod, error) {
	v, err := version.FromString(mVersion, ".", 10)
	if err != nil {
		return LocalMod{}, err
	}

	return LocalMod{name: name, desc: desc, id: id, slug: slug, forceVersion: forceVersion, version: v}, nil
}

func NewWithoutVersion(name, desc, id, slug, forceVersion string) LocalMod {
	return LocalMod{name: name, desc: desc, id: id, slug: slug, forceVersion: forceVersion}
}

func (lm LocalMod) GetIdOrSlug() string {
	if lm.slug != "" {
		return lm.slug
	}

	return lm.id
}

func (lm LocalMod) Name() string {
	return lm.name
}

func (lm LocalMod) IsEmpty() bool {
	return (strings.TrimSpace(lm.id) == "" && strings.TrimSpace(lm.slug) == "") || strings.TrimSpace(lm.name) == ""
}

func (lm LocalMod) ToPublic() struct {
	Id, Slug, Name, Desc, Version, ForceVersion string `json:",omitempty"`
} {
	return struct {
		Id, Slug, Name, Desc, Version, ForceVersion string `json:",omitempty"`
	}{
		Id:           lm.id,
		Slug:         lm.slug,
		Name:         lm.name,
		Desc:         lm.desc,
		Version:      lm.version.String(),
		ForceVersion: lm.forceVersion,
	}
}

func (lm *LocalMod) Download(gameVersion string) ([]byte, string, []remotemod.RemoteModVersionDependency, error) {
	versionToUse := remotemod.RemoteModVersion{}

	if lm.forceVersion != "" {
		log.Printf("\033[94mmod '%s' has been forced to use version '%s'\033[0m\n", lm.GetIdOrSlug(), lm.forceVersion)

		if resp, err := internal.Download("https://api.modrinth.com/v2/version/" + lm.forceVersion); err != nil {
		} else if err = json.Unmarshal(resp, &versionToUse); err != nil {
			return []byte{}, "", []remotemod.RemoteModVersionDependency{}, err
		}
	} else {
		remote, err := remotemod.FromProject(lm.GetIdOrSlug())
		if err != nil {
			return []byte{}, "", []remotemod.RemoteModVersionDependency{}, err
		}

		if lm.IsEmpty() {
			lm.id = remote.Id
			lm.slug = remote.Slug
			lm.name = remote.Title
			lm.desc = remote.Description
		}

		if !slices.Contains(remote.GameVersions, gameVersion) {
			return []byte{}, "", []remotemod.RemoteModVersionDependency{}, fmt.Errorf("no mods found with version %s for '%s'('%s')\n", gameVersion, lm.slug, lm.id)
		}

		filteredVersions := []remotemod.RemoteModVersion{}

		for _, rm := range remote.Versions {
			if slices.Contains(rm.GameVersions, gameVersion) && slices.Contains(rm.Loaders, "neoforge") {
				filteredVersions = append(filteredVersions, rm)
			}
		}

		// why
		parseModVersion := func(s string) version.Version {
			if proc, ok := customProcs[lm.slug]; ok {
				s = proc(s)
			}

			spl := strings.Split(s, "+")

			if strings.HasPrefix(spl[0], "v") {
				spl[0] = spl[0][1:]
			}

			for _, spl := range strings.Split(strings.Join(strings.Split(s, "-"), "+"), "+") {
				if strings.HasPrefix(spl, "v") {
					spl = spl[1:]
				}

				if v, err := version.FromString(spl, ".", 20); err != nil {
					panic(err.Error())
				} else {
					return v
				}
			}

			panic(fmt.Sprintf("cannot parse '%s'(from '%s') as version", s, lm.GetIdOrSlug()))
		}

		slices.SortFunc(filteredVersions, func(a, b remotemod.RemoteModVersion) int {
			return parseModVersion(a.VersionNumber).Cmp(parseModVersion(b.VersionNumber))
		})

		if len(filteredVersions) == 0 {
			return []byte{}, "", []remotemod.RemoteModVersionDependency{}, fmt.Errorf("no mods found with version %s for '%s'('%s')\n", gameVersion, lm.slug, lm.id)
		}

		versionToUse = filteredVersions[len(filteredVersions)-1]
	}

	resp, err := internal.Download(versionToUse.Files[0].Url)

	return resp, versionToUse.Files[0].Filename, versionToUse.Dependencies, err
}
