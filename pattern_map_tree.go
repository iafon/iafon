package iafon

import (
	"fmt"
	"strconv"
	"strings"
)

type PatternMapByTree struct {
	RouteTree
	values map[string]interface{}
}

type RouteTree struct {
	RouteTreeNode
	trees []*RouteTree
}

type RouteTreeNode struct {
	nType RouteTreeNodeType
	text  string
	value interface{}
}

type RouteTreeNodeType byte

const (
	cStatic RouteTreeNodeType = iota
	cParam
)

func (m *PatternMapByTree) Set(pattern string, value interface{}) {
	m.RouteTree.Set(pattern, value)
	if m.values == nil {
		m.values = make(map[string]interface{})
	}
	m.values[pattern] = value
}

func (m *PatternMapByTree) Get(pattern string) interface{} {
	return m.values[pattern]
}

func (m *PatternMapByTree) Len() int {
	return len(m.values)
}

func (t *RouteTree) Set(pattern string, value interface{}) {
	if pattern == "" {
		panic("route: pattern should not be empty")
	}

	if pattern[0] != '/' {
		panic("route: pattern should start with /")
	}

	// we only match path in url, not parameters
	if strings.Contains(pattern, "?") {
		panic("route: pattern should not contain ?")
	}

	if value == nil {
		panic("route: value should not be nil")
	}

	if !t.mergePath(newRoutePath(pattern, value)) {
		panic("route path pattern should start with /")
	}
}

// TODO: do not recursive call
func (t *RouteTree) Match(path string) (value interface{}, params map[string]string, redirect bool, substr string) {
	if t.nType == cStatic {
		len_n := len(t.text)
		len_p := len(path)

		if len_n-len_p < 1 {
			if t.text == path[:len_n] {
				value, params, redirect, substr = t.matchSubTrees(path[len_n:])
				if value != nil {
					if substr != "" {
						substr = t.text + substr
					} else if t.value != nil {
						value = t.value
						redirect = false
						substr = t.text
					}
				} else if len_n == len_p || path[len_n] == '/' {
					value = t.value
					substr = t.text
				}
			}
		} else if len_n-len_p == 1 {
			if t.text[len_n-1] == '/' && t.text[:len_n-1] == path {
				value = t.value
				redirect = true
				substr = path
			}
		}
	} else if t.nType == cParam {
		if path != "" {
			if param_end := strings.IndexByte(path, '/'); param_end != 0 {
				param := ""
				sub_path := ""

				if param_end > 0 {
					param = path[:param_end]
					sub_path = path[param_end:]
				} else {
					param = path
					sub_path = ""
				}

				value, params, redirect, substr = t.matchSubTrees(sub_path)
				if value != nil {
					if redirect && substr == "" && t.value != nil {
						value = t.value
						redirect = false
					}
					substr = param + substr
				} else {
					value = t.value
					substr = param
				}
				if params == nil {
					params = make(map[string]string)
				}
				params[t.text] = param
			}
		}
	}

	return
}

func (t *RouteTree) matchSubTrees(path string) (value interface{}, params map[string]string, redirect bool, substr string) {
	len_substr, len_curr_substr := 0, 0
	static_matched := false
	for _, st := range t.trees {
		if st.nType == cStatic && static_matched {
			continue
		}
		if curr_value, curr_params, curr_redirect, curr_substr := st.Match(path); curr_value != nil {
			if st.nType == cStatic {
				static_matched = true
			}
			len_substr, len_curr_substr = len(substr), len(curr_substr)
			if value == nil ||
				len_substr < len_curr_substr ||
				(len_substr == len_curr_substr && redirect && !curr_redirect) ||
				(len_substr == len_curr_substr && len(params) > len(curr_params)) {
				value = curr_value
				params = curr_params
				redirect = curr_redirect
				substr = curr_substr
			}
		}
	}
	return
}

func (t *RouteTree) mergePath(p *RouteTree) bool {
	// if t is empty tree, assign *p to *t
	if t.text == "" && len(t.trees) == 0 {
		*t = *p
		return true
	}

	var mergeSameNode = func(t, p *RouteTree) {
		// tree node is the same, merge p's subtree to t's subtrees
		if len(p.trees) > 0 {
			if len(t.trees) == 0 {
				t.trees = p.trees
			} else {
				merged := false
				for _, st := range t.trees {
					if m := st.mergePath(p.trees[0]); m {
						merged = m
						break
					}
				}
				if !merged {
					t.trees = append(t.trees, p.trees[0])
				}
			}
		} else {
			// leaf node, set value
			t.value = p.value
		}
	}

	// if all static text, find prefix as a tree node
	if t.nType == cStatic && p.nType == cStatic {
		prefix, flag := commonPrefix(t.text, p.text)

		switch flag {
		case 0:
			// no common prefix, merge failed
			return false
		case 1:
			// tree node is the same
		case 2:
			t.split(prefix)
			p.split(prefix)
			// after split, tree node is the same
		case 3:
			p.split(prefix)
			// after split, tree node is the same
		case 4:
			t.split(prefix)
			// after split, tree node is the same
		default:
			panic("commonPrefix flag error")
		}

		mergeSameNode(t, p)
		return true
	} else if t.nType == cParam && p.nType == cParam {
		if t.text == p.text {
			mergeSameNode(t, p)
			return true
		} else {
			// param name not equal, merge failed
			return false
		}
	} else {
		// cStatic tree node can not merge with cParam tree node, merge failed
		return false
	}
}

func (t *RouteTree) split(prefix string) {
	suffixNode := newRoutePath(t.text[len(prefix):], t.value)
	suffixNode.trees = t.trees
	t.text = prefix
	t.value = nil
	t.trees = []*RouteTree{suffixNode}
}

func (t *RouteTree) Print(_indent ...int) {
	indent := 0
	if len(_indent) > 0 {
		indent = _indent[0]
	}

	text := t.text
	if t.nType == cParam {
		text = ":" + text
	}

	fmt.Printf("%"+strconv.Itoa(indent)+"s%s : %t\n", "", text, t.value != nil)

	for _, st := range t.trees {
		st.Print(indent + len(text))
	}
}

// path is a tree
func newRoutePath(pattern string, value interface{}) *RouteTree {
	// root node
	root := &RouteTree{}

	// current node which new node should add to
	curr := root

	pos := 0

	for {
		if pos_param := pos + strings.IndexByte(pattern[pos:], ':'); pos_param < pos {
			curr.nType = cStatic
			curr.text = pattern[pos:]
			break
		} else {
			curr.nType = cStatic
			curr.text = pattern[pos:pos_param]
			curr.trees = []*RouteTree{&RouteTree{}}

			curr = curr.trees[0]
			curr.nType = cParam

			if pos = pos_param + strings.IndexByte(pattern[pos_param:], '/'); pos >= pos_param {
				if pos_param+1 >= pos {
					panic("route: param name should not be empty")
				}
				curr.text = pattern[pos_param+1 : pos]
				curr.trees = []*RouteTree{&RouteTree{}}

				curr = curr.trees[0]
			} else {
				curr.text = pattern[pos_param+1:]
				break
			}
		}
	}

	curr.value = value

	return root
}

// flag meaning:
// 0: no prefix
// 1: a == b
// 2: prefix short than both
// 3: a is prefix
// 4: b is prefix
func commonPrefix(a, b string) (prefix string, flag int) {
	a_len := len(a)
	b_len := len(b)

	min_len := a_len
	if min_len > b_len {
		min_len = b_len
	}

	i := 0

	for ; i < min_len; i++ {
		if a[i] != b[i] {
			break
		}
	}

	prefix = a[:i]

	if i == 0 {
		flag = 0
	} else if i < min_len {
		flag = 2
	} else if a_len == b_len {
		flag = 1
	} else if i == a_len {
		flag = 3
	} else {
		flag = 4
	}

	return prefix, flag
}
