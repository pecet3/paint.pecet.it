package ward

type Group struct {
	ward                *Ward
	path                string
	middlewares         []Middleware
	authorizeMiddleware Middleware
}

func newGroup(ward *Ward, basePath string) *Group {
	return &Group{
		ward: ward,
		path: basePath,
	}
}

func (g *Group) Use(mws ...Middleware) *Group {
	g.middlewares = append(g.middlewares, mws...)
	return g
}

func (g *Group) combineMiddleware(mws []Middleware) []Middleware {
	combined := make([]Middleware, 0, len(g.middlewares)+len(mws))
	combined = append(combined, g.middlewares...)
	combined = append(combined, mws...)
	return combined
}
func (g *Group) Get(pattern string, handler func(wreq *Request), mws ...Middleware) {
	g.ward.Get(g.path+pattern, handler, g.combineMiddleware(mws)...)
}

func (g *Group) Put(pattern string, handler func(wreq *Request), mws ...Middleware) {
	g.ward.Put(g.path+pattern, handler, g.combineMiddleware(mws)...)
}

func (g *Group) Post(pattern string, handler func(wreq *Request), mws ...Middleware) {
	g.ward.Post(g.path+pattern, handler, g.combineMiddleware(mws)...)
}

func (g *Group) Delete(pattern string, handler func(wreq *Request), mws ...Middleware) {
	g.ward.Delete(g.path+pattern, handler, g.combineMiddleware(mws)...)
}
