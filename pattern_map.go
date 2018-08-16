package iafon

// tPatternMapInterface is just for comment, not used in code
type PatternMapInterface interface {
	Set(pattern string, value interface{})
	Get(pattern string) interface{}
	Match(str string) (value interface{}, params map[string]string, redirect bool, substr string)
	Len() int
}
