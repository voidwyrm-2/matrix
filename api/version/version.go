package version

import (
	"fmt"
	"strconv"
	"strings"
)

type Version struct {
	sections []uint16
}

func New(sections []uint16) Version {
	return Version{sections: sections}
}

func FromString(version, delimiter string, base int) (Version, error) {
	sections := []uint16{}

	for _, s := range strings.Split(version, delimiter) {
		if n, err := strconv.ParseUint(s, base, 16); err != nil {
			return Version{}, err
		} else {
			sections = append(sections, uint16(n))
		}
	}

	return Version{sections: sections}, nil
}

func (v Version) Cmp(other Version) int {
	if len(v.sections) < len(other.sections) {
		return -1
	} else if len(v.sections) > len(other.sections) {
		return 1
	}

	for i, s := range v.sections {
		if s < other.sections[i] {
			return -1
		} else if s > other.sections[i] {
			return 1
		}
	}

	return 0
}

func (v Version) Eq(other Version) bool {
	return v.Cmp(other) == 0
}

func (v Version) String() string {
	formatted := []string{}

	for _, s := range v.sections {
		formatted = append(formatted, fmt.Sprint(s))
	}

	return strings.Join(formatted, ".")
}
