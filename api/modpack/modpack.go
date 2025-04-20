package modpack

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/BurntSushi/toml"
	"github.com/voidwyrm-2/matrix/api/internal"
	"github.com/voidwyrm-2/matrix/api/localmod"
	"github.com/voidwyrm-2/matrix/api/version"
)

func errIs(e error, s string) bool {
	if e == nil {
		return false
	}

	return e.Error() == s
}

type publicModpack struct {
	Name, ModpackVersion, GameVersion, Modloader string
	Mods                                         struct {
		External map[string]string
		Mdrth    []struct {
			Id, Slug, Name, Desc, Version, ForceVersion string `json:",omitempty"`
		}
	}
}

type Modpack struct {
	onlySyncEmpty, ignoreExternals bool
	name, desc                     string
	version, gameVersion           version.Version
	mods                           struct {
		mdrth    []localmod.LocalMod
		external map[string]string
	}
}

func (mp *Modpack) Populate() error {
	os.Mkdir("mods", os.ModeDir|os.ModePerm)

	initialD := map[string]struct{}{}
	for _, mod := range mp.mods.mdrth {
		initialD[mod.ToPublic().Id] = struct{}{}
	}

	err := mp.downloadMods(mp.mods.mdrth, initialD, false)
	if err != nil {
		return err
	}

	if !mp.ignoreExternals {
		for name, url := range mp.mods.external {
			log.Printf("\033[93mdownloading external mod '%s'...\033[0m\n", name)

			if resp, err := internal.Download(url); err != nil {
				if strings.TrimSpace(err.Error()) == "status code 403, '403 Forbidden'" {
					log.Printf("\033[91mexternal mod not downloaded, please remove or download manually: '%s' at '%s')\033[0m\n", name, url)
					continue
				}
				return err
			} else if err = internal.WriteFile(filepath.Join("mods", name), resp); err != nil {
				return err
			} else {
				log.Printf("\033[92mdownloaded external mod '%s'\033[0m\n", name)
			}
		}
	}

	existing := map[string]struct{}{}
	cleanedMods := []localmod.LocalMod{}

	for _, m := range mp.mods.mdrth {
		if _, ok := existing[m.GetIdOrSlug()]; !ok && !m.IsEmpty() {
			existing[m.GetIdOrSlug()] = struct{}{}
			cleanedMods = append(cleanedMods, m)
		} else {
		}
	}

	mp.mods.mdrth = cleanedMods

	return nil
}

func (mp *Modpack) downloadMods(mods []localmod.LocalMod, downloadedDependencies map[string]struct{}, downloadingDependencies bool) error {
	kind := "mod"
	if downloadingDependencies {
		kind = "dependacy"
	}

	for i, m := range mods {
		log.Printf("\033[93mdownloading %s '%s'...\033[0m\n", kind, m.GetIdOrSlug())

		if mp.onlySyncEmpty && !m.IsEmpty() {
			log.Printf("\033[94mskipped '%s' because only empty mods are being synced\033[0m\n", m.GetIdOrSlug())
			continue
		}

		if mbytes, mname, dependancies, err := m.Download(mp.gameVersion.String()); err != nil {
			return err
		} else if err = internal.WriteFile(filepath.Join("mods", mname), mbytes); err != nil {
			return err
		} else {
			log.Printf("\033[92mdownloaded %s '%s'\033[0m\n", kind, mname)
			if downloadingDependencies {
				mp.mods.mdrth = append(mp.mods.mdrth, m)
			} else {
				mp.mods.mdrth[i] = m
			}

			dmods := []localmod.LocalMod{}

			for _, d := range dependancies {
				if _, ok := downloadedDependencies[d.ProjectId]; !ok && d.Kind == "required" {
					downloadedDependencies[d.ProjectId] = struct{}{}
					dmods = append(dmods, localmod.NewWithoutVersion("", "", d.ProjectId, "", ""))
				}
			}

			if err := mp.downloadMods(dmods, downloadedDependencies, true); err != nil {
				return err
			}
		}
	}

	return nil
}

func (mp Modpack) Mods() []localmod.LocalMod {
	return mp.mods.mdrth
}

func (mp Modpack) ToToml(name string) error {
	pm := publicModpack{
		Name:           mp.name,
		ModpackVersion: mp.version.String(),
		GameVersion:    mp.gameVersion.String(),
		Mods: struct {
			External map[string]string
			Mdrth    []struct {
				Id, Slug, Name, Desc, Version, ForceVersion string `json:",omitempty"`
			}
		}{
			External: mp.mods.external,
		},
	}

	for _, m := range mp.mods.mdrth {
		pm.Mods.Mdrth = append(pm.Mods.Mdrth, m.ToPublic())
	}

	result, err := toml.Marshal(pm)
	if err != nil {
		return err
	}

	os.Remove(name)

	return internal.WriteFile(name, result)
}

func FromToml(name string, onlySyncEmpty, ignoreExternals bool) (Modpack, error) {
	st := publicModpack{}

	_, err := toml.DecodeFile(name, &st)
	if err != nil {
		return Modpack{}, err
	}

	mpv, err := version.FromString(st.ModpackVersion, ".", 10)
	if err != nil {
		return Modpack{}, err
	}

	mcv, err := version.FromString(st.GameVersion, ".", 10)
	if err != nil {
		return Modpack{}, err
	}

	mp := Modpack{name: st.Name, version: mpv, gameVersion: mcv, mods: struct {
		mdrth    []localmod.LocalMod
		external map[string]string
	}{external: st.Mods.External}, onlySyncEmpty: onlySyncEmpty, ignoreExternals: ignoreExternals}

	for _, m := range st.Mods.Mdrth {
		lm := localmod.NewWithoutVersion(m.Name, m.Desc, m.Id, m.Slug, m.ForceVersion)

		if _lm, err := localmod.New(m.Name, m.Desc, m.Id, m.Slug, m.ForceVersion, m.Version); err != nil && !errIs(err, `strconv.ParseUint: parsing "": invalid syntax`) {
			return Modpack{}, err
		} else if err == nil {
			lm = _lm
		}

		mp.mods.mdrth = append(mp.mods.mdrth, lm)
	}

	return mp, nil
}

func FromMatrixfile(name string) error {
	content, err := internal.ReadOptions("Matrixfile", "matrixfile", "Matrixfile.txt", "matrixfile.txt")
	if err != nil {
		return err
	}

	pm := publicModpack{
		Mods: struct {
			External map[string]string
			Mdrth    []struct {
				Id, Slug, Name, Desc, Version, ForceVersion string `json:",omitempty"`
			}
		}{
			External: map[string]string{},
		},
	}

	if len(strings.Split(strings.TrimSpace(string(content)), "\n")) < 4 {
		return errors.New(`invalid Matrixfile format, expected
<name>
<pack version>
<Minecraft version>
<modloader>

<slug> [forced version] OR id <id> [forced version]
...`)
	}

	for i, l := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}

		if i == 0 {
			pm.Name = l
		} else if i == 1 {
			pm.ModpackVersion = l
		} else if i == 2 {
			pm.GameVersion = l
		} else if i == 3 {
			pm.Modloader = strings.ToLower(l)
		} else {
			m := struct {
				Id, Slug, Name, Desc, Version, ForceVersion string `json:",omitempty"`
			}{}

			spl := strings.Split(l, " ")
			if spl[0] == "id" {
				if len(spl) < 2 {
					return fmt.Errorf("line %d: expected 2 items, but found 1", i+1)
				} else if len(spl) > 2 {
					m.ForceVersion = spl[2]
				}

				m.Id = spl[2]
			} else {
				if len(spl) > 1 {
					m.ForceVersion = spl[1]
				}

				m.Slug = spl[0]
			}

			pm.Mods.Mdrth = append(pm.Mods.Mdrth, m)
		}
	}

	result, err := toml.Marshal(pm)
	if err != nil {
		return err
	}

	os.Remove(name)

	return internal.WriteFile(name, result)
}
