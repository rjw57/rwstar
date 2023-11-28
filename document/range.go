package document

type Range struct {
	start *Point
	end   *Point
}

func NewRange(start *Point, end *Point) *Range {
	if start.d != end.d {
		panic("Start and end points in range from different documents.")
	}
	return &Range{start: start, end: end}
}

func (r *Range) Start() *Point {
	return r.start
}

func (r *Range) End() *Point {
	return r.end
}

func (r *Range) Document() *Document {
	return r.start.d
}

func (r *Range) Contains(p *Point) bool {
	if p.d != r.start.d {
		return false
	}
	if p.paraIndex < r.start.paraIndex || p.paraIndex > r.end.paraIndex {
		return false
	}
	if p.paraIndex == r.start.paraIndex && p.textOffset < r.start.textOffset {
		return false
	}
	if p.paraIndex == r.end.paraIndex && p.textOffset >= r.end.textOffset {
		return false
	}
	return true
}
