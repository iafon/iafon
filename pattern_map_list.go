package iafon

import (
	"strings"
)

type PatternMapByList []*tPatternListItem

type tPatternListItem struct {
	pattern string
	value   interface{}

	parts []tSubPattern
}

type tSubPattern struct {
	pType tSubPatternType
	text  string
}

type tSubPatternType byte

const (
	cSubPatternStatic tSubPatternType = iota
	cSubPatternParam
)

func (m *PatternMapByList) Set(pattern string, value interface{}) {
	item := newMapItem(pattern, value)

	modified := false

	for i, p := range *m {
		if p.pattern == pattern {
			(*m)[i] = item
			modified = true
			break
		}
	}

	if !modified {
		*m = append(*m, item)
	}
}

func (m *PatternMapByList) Get(pattern string) interface{} {
	for _, p := range *m {
		if p.pattern == pattern {
			return p.value
		}
	}
	return nil
}

func (m *PatternMapByList) Match(path string) (value interface{}, params map[string]string, redirect bool, substr string) {
	var pre_p *tPatternListItem
	var pre_substr string
	var pre_params map[string]string
	var pre_redirect bool = false

	for _, p := range *m {
		if matched, substr, params, redirect := p.match(path); matched {
			if pre_p == nil ||
				len(pre_substr) < len(substr) ||
				(len(pre_substr) == len(substr) && pre_redirect && !redirect) ||
				(len(pre_substr) == len(substr) && len(pre_params) > len(params)) {
				pre_p = p
				pre_substr = substr
				pre_params = params
				pre_redirect = redirect
			}
		}
	}

	if pre_p == nil {
		return nil, nil, false, ""
	} else {
		return pre_p.value, pre_params, pre_redirect, pre_substr
	}
}

func (m *PatternMapByList) Len() int {
	return len(*m)
}

func newMapItem(pattern string, value interface{}) *tPatternListItem {
	p := &tPatternListItem{pattern: pattern, value: value}
	if strings.Contains(pattern, "?") {
		panic("route pattern error. '" + pattern + "' should not contain '?'")
	}
	for pos := 0; pos >= 0; {
		if pos_param := pos + strings.IndexByte(pattern[pos:], ':'); pos_param < pos {
			p.parts = append(p.parts, tSubPattern{pType: cSubPatternStatic, text: pattern[pos:]})
			pos = -1
		} else {
			p.parts = append(p.parts, tSubPattern{pType: cSubPatternStatic, text: pattern[pos:pos_param]})
			if pos = pos_param + strings.IndexByte(pattern[pos_param:], '/'); pos >= pos_param {
				p.parts = append(p.parts, tSubPattern{pType: cSubPatternParam, text: pattern[pos_param+1 : pos]})
			} else {
				p.parts = append(p.parts, tSubPattern{pType: cSubPatternParam, text: pattern[pos_param+1:]})
				pos = -1
			}
		}
	}
	return p
}

func (p *tPatternListItem) match(path string) (matched bool, substr string, params map[string]string, redirect bool) {
	matchedCount := 0
	substr_end := 0
	str := path

	end_i := len(p.parts) - 1

	for i, part := range p.parts {
		if part.pType == cSubPatternParam {
			if str == "" {
				// mismatch
				break
			}

			pos := strings.IndexByte(str, '/')

			if pos == 0 {
				// param value could not be empty
				// mismatch
				break
			}

			matchedCount++

			if params == nil {
				// make map only when param exists
				params = make(map[string]string)
			}

			if pos > 0 {
				params[part.text] = str[:pos]
				substr_end += pos
				str = str[pos:]
			} else {
				params[part.text] = str
				substr_end = len(path)
				// may match redirect
				str = ""
			}
		} else { // part.pType == cSubPatternStatic
			// part should be str prefix
			len_part := len(part.text)
			len_str := len(str)
			if len_part <= len_str && part.text == str[:len_part] {
				if len_part == len_str {
					substr_end += len_part
					matchedCount++
					// may match redirect
					str = ""
				} else if str[len_part-1] == '/' || str[len_part] == '/' {
					substr_end += len_part
					str = str[len_part:]
					matchedCount++
				} else {
					// mismatch
					break
				}
			} else if len_part > len_str {
				if i == end_i && len_part == len_str+1 && part.text[len_part-1] == '/' && str == part.text[:len_part-1] {
					matchedCount++
					redirect = true
					// match finished
					break
				} else {
					// mismatch
					break
				}
			} else {
				// mismatch
				break
			}
		}
	}

	if matchedCount == len(p.parts) {
		return true, path[:substr_end], params, redirect
	} else {
		return false, "", nil, false
	}
}
