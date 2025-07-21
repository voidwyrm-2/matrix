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

type Modpack struct {
	onlySyncEmpty, ignoreExternals bool
	name, desc, modloader          string
	version, gameVersion           version.Version
	mods                           struct {
		mdrth    []localmod.LocalMod
		external map[string]string
	}
}

func (mp *Modpack) Populate(logger *log.Logger, workingDir string) error {
	modsFolder := filepath.Join(workingDir, "mods")

	err := os.MkdirAll(modsFolder, os.ModeDir|os.ModePerm)
	if err != nil {
		return err
	}

	err = mp.downloadMods(logger, modsFolder, mp.mods.mdrth, map[string]struct{}{}, false)
	if err != nil {
		return err
	}

	if !mp.ignoreExternals {
		for name, url := range mp.mods.external {
			logger.Printf("downloading external mod '%s'...\n", name)

			if resp, err := internal.Download(url); err != nil {
				if strings.TrimSpace(err.Error()) == "status code 403, '403 Forbidden'" {
					logger.Printf("external mod could not be downloaded, please remove or download manually: '%s' at '%s')\n", name, url)
					continue
				}
				return err
			} else if err = internal.WriteFile(filepath.Join(modsFolder, name), resp); err != nil {
				return err
			} else {
				logger.Printf("downloaded external mod '%s'\n", name)
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
			panic("unimplemented: " + fmt.Sprint(m))
		}
	}

	mp.mods.mdrth = cleanedMods

	return nil
}

func (mp *Modpack) downloadMods(logger *log.Logger, outputPath string, mods []localmod.LocalMod, alreadyDownloaded map[string]struct{}, downloadingDependencies bool) error {
	kind := "mod"
	if downloadingDependencies {
		kind = "dependancy"
	}

	for i, m := range mods {
		logger.Printf("downloading %s '%s'...\n", kind, m.GetIdOrSlug())

		if mp.onlySyncEmpty && !m.IsEmpty() {
			logger.Printf("skipped '%s' because only empty mods are being synced\n", m.GetIdOrSlug())
			continue
		}

		skipMod := m.AlreadyDownloaded(&alreadyDownloaded)
		if skipMod {
			logger.Printf("skipped '%s' because it's already been downloaded\n", m.GetIdOrSlug())
			continue
		}

		if mbytes, mname, dependancies, err := m.Download(logger, mp.gameVersion.String(), mp.modloader); err != nil {
			return err
		} else if err = internal.WriteFile(filepath.Join(outputPath, mname), mbytes); err != nil {
			return err
		} else {
			logger.Printf("downloaded %s '%s'\n", kind, mname)
			alreadyDownloaded[m.ToPublic().Id], alreadyDownloaded[m.ToPublic().Slug] = struct{}{}, struct{}{}

			if downloadingDependencies {
				mp.mods.mdrth = append(mp.mods.mdrth, m)
			} else {
				mp.mods.mdrth[i] = m
			}

			dmods := []localmod.LocalMod{}

			for _, d := range dependancies {
				// Exception for Sinytra Connector
				if d.ProjectId == "P7dR8mSH" && mp.modloader != "fabric" && mp.modloader != "quilt" {
					continue
				}

				if _, ok := alreadyDownloaded[d.ProjectId]; !ok && d.Kind == "required" {
					dmods = append(dmods, localmod.NewWithoutVersion("", "", d.ProjectId, "", "", m.ToPublic().ForceLoader))
				}
			}

			if err := mp.downloadMods(logger, outputPath, dmods, alreadyDownloaded, true); err != nil {
				return err
			}
		}
	}

	return nil
}

func (mp Modpack) Mods() []localmod.LocalMod {
	return mp.mods.mdrth
}

func (mp Modpack) Name() string {
	return mp.name
}

func (mp Modpack) Version() string {
	return mp.version.String()
}

func (mp Modpack) GameVersion() string {
	return mp.gameVersion.String()
}

func (mp Modpack) Modloader() string {
	return mp.modloader
}

func (mp Modpack) ToToml(name string) error {
	pm := internal.PublicModpack{
		Name:           mp.name,
		ModpackVersion: mp.version.String(),
		GameVersion:    mp.gameVersion.String(),
		Mods: struct {
			External map[string]string
			Mdrth    []internal.PublicLocalMod
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
	st := internal.PublicModpack{}

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

	mp := Modpack{name: st.Name, version: mpv, gameVersion: mcv, modloader: st.Modloader, mods: struct {
		mdrth    []localmod.LocalMod
		external map[string]string
	}{external: st.Mods.External}, onlySyncEmpty: onlySyncEmpty, ignoreExternals: ignoreExternals}

	for _, m := range st.Mods.Mdrth {
		lm := localmod.NewWithoutVersion(m.Name, m.Desc, m.Id, m.Slug, m.ForceVersion, m.ForceLoader)

		if _lm, err := localmod.New(m.Name, m.Desc, m.Id, m.Slug, m.ForceVersion, m.ForceLoader, m.Version); err != nil && !errIs(err, `strconv.ParseUint: parsing "": invalid syntax`) {
			return Modpack{}, err
		} else if err == nil {
			lm = _lm
		}

		mp.mods.mdrth = append(mp.mods.mdrth, lm)
	}

	return mp, nil
}

func parseMatrixfileEntryFlags(rawFlags []string) map[string]string {
	m := map[string]string{}

	for _, f := range rawFlags {
		if strings.Contains(f, ":") {
			spl := strings.Split(f, ":")
			if len(spl) > 1 {
				m[strings.TrimSpace(spl[0])] = strings.TrimSpace(strings.Join(spl[1:], ":"))
			}
		}
	}

	return m
}

func configureLocalMod(plm *internal.PublicLocalMod, flags map[string]string) {
	if v, ok := flags["v"]; ok {
		plm.ForceVersion = v
	}

	if v, ok := flags["l"]; ok {
		plm.ForceLoader = v
	}
}

func FromMatrixfile(matrixfile, output string) error {
	content, err := os.ReadFile(matrixfile)
	if err != nil {
		return err
	}

	pm := internal.PublicModpack{
		Mods: struct {
			External map[string]string
			Mdrth    []internal.PublicLocalMod
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

(<slug> OR id <id>) [flags...]
...`)
	}

	realI := 0
	for i, l := range strings.Split(strings.TrimSpace(string(content)), "\n") {
		l = strings.TrimSpace(l)
		if l == "" {
			continue
		}

		if realI == 0 {
			pm.Name = l
		} else if realI == 1 {
			pm.ModpackVersion = l
		} else if realI == 2 {
			pm.GameVersion = l
		} else if realI == 3 {
			pm.Modloader = strings.ToLower(l)
		} else {
			m := internal.PublicLocalMod{}

			spl := strings.Split(l, " ")

			switch spl[0] {
			case "ext", "external":
				if len(spl) < 3 {
					return fmt.Errorf("line %d: expected 3 items, but found %d", i+1, len(spl))
				}

				pm.Mods.External[spl[1]] = spl[2]
			case "id":
				if len(spl) < 2 {
					return fmt.Errorf("line %d: expected 2 items, but found %d", i+1, len(spl))
				}

				configureLocalMod(&m, parseMatrixfileEntryFlags(spl[2:]))

				m.Id = spl[1]
			default:
				configureLocalMod(&m, parseMatrixfileEntryFlags(spl[1:]))

				m.Slug = spl[0]
			}

			if m.Slug != "" || m.Id != "" {
				pm.Mods.Mdrth = append(pm.Mods.Mdrth, m)
			}
		}

		realI++
	}

	result, err := toml.Marshal(pm)
	if err != nil {
		return err
	}

	return internal.WriteFile(filepath.Join(filepath.Dir(matrixfile), output), result)
}
