package document

type Range struct {
	start *Point
	end   *Point
}

func (r *Range) Start() *Point {
	return r.start
}

func (r *Range) End() *Point {
	return r.end
}

func (r *Range) Document() *Document {
	if r.start.d != r.end.d {
		panic("Start and end points in range from different documents.")
	}
	return r.start.d
}
