package internal

type PublicLocalMod struct {
	Id, Slug, Name, Desc, Version, ForceVersion, ForceLoader string `json:",omitempty"`
}

type PublicModpack struct {
	Name, ModpackVersion, GameVersion, Modloader string
	Mods                                         struct {
		External map[string]string
		Mdrth    []PublicLocalMod
	}
}
