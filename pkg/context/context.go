package context

type Context struct {
	Local  map[string]interface{}
	Global map[string]interface{}
}
