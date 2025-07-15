package proc

import (
	"errors"
	"fmt"
	"reflect"
	"slices"
	"strings"
	"unicode"
)

var opMap = map[rune]struct {
	f func(st *stack) error
	i []reflect.Kind
}{
	'+': {
		func(st *stack) error {
			b, a := st.pop().(float32), st.pop().(float32)

			st.push(a + b)

			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.Float32},
	},
	'-': {
		func(st *stack) error {
			b, a := st.pop().(float32), st.pop().(float32)

			st.push(a - b)

			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.Float32},
	},
	'×': { // MULTIPLICATION SIGN
		func(st *stack) error {
			b, a := st.pop().(float32), st.pop().(float32)

			st.push(a * b)

			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.Float32},
	},
	'÷': { // DIVISION SIGN
		func(st *stack) error {
			b, a := st.pop().(float32), st.pop().(float32)

			st.push(a / b)

			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.Float32},
	},
	'|': {
		func(st *stack) error {
			sep := st.pop().(string)
			st.push(strings.Split(st.pop().(string), sep))
			return nil
		},
		[]reflect.Kind{reflect.String, reflect.String},
	},
	'[': {
		func(st *stack) error {
			n := int(st.pop().(float32))

			st.push(st.pop().([]string)[n])
			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.Slice},
	},
	']': {
		func(st *stack) error {
			n := int(st.pop().(float32))

			st.push(float32(st.pop().(string)[n]))
			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.String},
	},
	'⌈': { // LEFT CEILING
		func(st *stack) error {
			n := int(st.pop().(float32))

			st.push(st.pop().([]string)[n:])

			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.Slice},
	},
	'⌊': { // LEFT FLOOR
		func(st *stack) error {
			n := int(st.pop().(float32))

			st.push(st.pop().(string)[n:])

			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.String},
	},
	'⌉': { // RIGHT CEILING
		func(st *stack) error {
			n := int(st.pop().(float32))

			st.push(st.pop().([]string)[:n])

			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.Slice},
	},
	'⌋': { // RIGHT FLOOR
		func(st *stack) error {
			n := int(st.pop().(float32))

			st.push(st.pop().(string)[:n])

			return nil
		},
		[]reflect.Kind{reflect.Float32, reflect.String},
	},
	'.': {
		func(st *stack) error {
			val := st.pop()

			st.push(val, val)

			return nil
		},
		[]reflect.Kind{reflect.Invalid},
	},
	':': {
		func(st *stack) error {
			b, a := st.pop(), st.pop()

			st.push(b, a)

			return nil
		},
		[]reflect.Kind{reflect.Invalid, reflect.Invalid},
	},
	'↦': { // RIGHTWARDS ARROW FROM BAR
		func(st *stack) error {
			b, a := st.pop(), st.pop()

			st.push(b, a)

			return nil
		},
		[]reflect.Kind{reflect.Invalid, reflect.Invalid},
	},
	'⑂': { // OCR FORK
		func(st *stack) error {
			st.push(len(st.pop().([]string)))
			return nil
		},
		[]reflect.Kind{reflect.Slice},
	},
	'⑃': { // OCR INVERTED FORK
		func(st *stack) error {
			st.push(len(st.pop().(string)))
			return nil
		},
		[]reflect.Kind{reflect.String},
	},
	'⎲': { // SUMMATION TOP
		func(st *stack) error {
			sl := st.pop().([]string)
			st.push(sl[len(sl)-1])

			return nil
		},
		[]reflect.Kind{reflect.Slice},
	},
	'⎳': { // SUMMATION BOTTOM
		func(st *stack) error {
			sl := st.pop().(string)
			st.push(float32(sl[len(sl)-1]))

			return nil
		},
		[]reflect.Kind{reflect.String},
	},
	'd': { // debug
		func(st *stack) error {
			fmt.Printf("debug: %v\n", st.inner)

			return nil
		},
		[]reflect.Kind{},
	},
	'j': { // join
		func(st *stack) error {
			sep := st.pop().(string)
			st.push(strings.Join(st.pop().([]string), sep))
			return nil
		},
		[]reflect.Kind{reflect.String, reflect.Slice},
	},
	's': { // tostring
		func(st *stack) error {
			st.push(string(rune(int(st.pop().(float32)))))
			return nil
		},
		[]reflect.Kind{reflect.Float32},
	},
	'␣': { // UP DOWN ARROW WITH BASE
		func(st *stack) error {
			arr := st.pop().([]string)

			str := strings.Builder{}

			for _, s := range arr {
				str.WriteString(s)
			}

			st.push(str.String())

			return nil
		},
		[]reflect.Kind{reflect.Slice},
	},
	'f': { // format
		func(st *stack) error {
			st.push(fmt.Sprint(st.pop()))
			return nil
		},
		[]reflect.Kind{reflect.Invalid},
	},
	'm': { // make list
		func(st *stack) error {
			if st.hasFloat() {
				return errors.New("stack must contain only strings to be convertable to a list")
			}

			strs := []string{}

			for _, a := range st.inner {
				strs = append(strs, a.(string))
			}

			st.inner = []any{strs}

			return nil
		},
		[]reflect.Kind{},
	},
	'=': {
		func(st *stack) error {
			st.push(st.pop() == st.pop())
			return nil
		},
		[]reflect.Kind{reflect.Invalid, reflect.Invalid},
	},
	'≠': { // NOT EQUALS
		func(st *stack) error {
			st.push(st.pop() != st.pop())
			return nil
		},
		[]reflect.Kind{reflect.Invalid, reflect.Invalid},
	},
}

type stack struct {
	inner []any
}

func (s stack) expect(types ...reflect.Kind) error {
	newTypes := make([]reflect.Kind, 0, len(types))
	copy(newTypes, types)
	slices.Reverse(newTypes)

	if len(types) > s.len() {
		return fmt.Errorf("expected %d items on the stack but found %d items instead", len(types), s.len())
	}

	for i, t := range newTypes {
		if ac := reflect.TypeOf(s.inner[i]).Kind(); t != ac && t != reflect.Invalid {
			return fmt.Errorf("expected '%v' in position %d of the stack, but found '%v' instead", t, i, ac)
		}
	}

	return nil
}

func (s stack) len() int {
	return len(s.inner)
}

func (s stack) empty() bool {
	return s.len() == 0
}

func (s stack) hasFloat() bool {
	for _, a := range s.inner {
		if _, ok := a.(float32); ok {
			return true
		}
	}

	return false
}

func (s *stack) push(a ...any) {
	s.inner = append(s.inner, a...)
}

func (s *stack) pop() any {
	a := s.inner[s.len()-1]
	s.inner = s.inner[:s.len()-1]
	return a
}

func Apply(slug, input, ops string) (string, error) {
	st := stack{[]any{input}}

	runes := []rune(ops)
	acc := ""
	pushC, pushS := false, false

	for i, r := range runes {
		errf := func(str string, a ...any) error {
			return fmt.Errorf("op %d ('%c') of '%s': %s", i+1, r, slug, fmt.Sprintf(str, a...))
		}

		if pushS {
			if r == '"' {
				pushS = false
				st.push(acc)
				acc = ""
			} else {
				acc += string(r)
			}
		} else if pushC {
			st.push(string(r))
			pushC = false
		} else if unicode.IsSpace(r) {
			continue
		} else if r == '\'' {
			pushC = true
		} else if r == '"' {
			pushS = true
		} else if r >= '0' && r <= '9' {
			st.push(float32(r - 48))
		} else if e, ok := opMap[r]; !ok {
			return "", errf("invalid op '%c'", r)
		} else if err := st.expect(e.i...); err != nil {
			return "", errf(err.Error())
		} else if err = e.f(&st); err != nil {
			return "", errf(err.Error())
		}
	}

	if st.empty() {
		return "", fmt.Errorf("end of eval of '%s': stack underflow", slug)
	}

	a := st.pop()
	if str, ok := a.(string); !ok {
		return "", fmt.Errorf("end of eval of '%s': expected string, but found '%v' instead", slug, reflect.TypeOf(a))
	} else {
		return str, nil
	}
}
