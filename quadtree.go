package quadtree

import "image"

type direction int

const (
	nw direction = iota
	ne
	sw
	se
)

const (
	DefaultBucketSize = 4
)

type QuadTree interface {
	// InBounds returns whether the point could fit in the quadtree.
	InBounds(p image.Point) bool

	// Insert adds a point and value to the quad tree and returns true if it was successful.
	// A return value of false usually means the point is outside the bounds of the quad tree.
	Insert(p image.Point, val interface{}) bool

	// Select uses the given rect to search for any entries in the provided area. Returns nil
	// if no entries were found.
	Select(rect image.Rectangle) []*QEntry
}

type qTree struct {
	bucketSize int
	children   []*qTree
	leaves     []*QEntry
	rect       image.Rectangle
}

type QEntry struct {
	pos image.Point
	val interface{}
}

func NewQuadTree(bounds image.Rectangle, bucketSize int) QuadTree {
	return newQuadTree(bounds, bucketSize)
}

func newQuadTree(bounds image.Rectangle, bucketSize int) *qTree {
	if bucketSize < 1 {
		panic("bucketSize must be greater than 0")
	}

	bounds = bounds.Canon()

	if bounds.Empty() {
		panic("bounds must have a positive length for both width and height")
	}

	return &qTree{
		bucketSize: bucketSize,
		leaves:     make([]*QEntry, bucketSize),
		rect:       bounds,
	}
}

func (t *qTree) InBounds(p image.Point) bool {
	return p.In(t.rect)
}

func (t *qTree) Insert(p image.Point, val interface{}) bool {
	return t.insert(&QEntry{pos: p, val: val})
}

func (t *qTree) Select(rect image.Rectangle) []*QEntry {
	if !t.rect.Overlaps(rect) {
		return nil
	}

	entries := []*QEntry{}

	if t.children == nil {
		entries = append(entries, t.leaves...)
	} else {
		for _, child := range t.children {
			childEntries := child.Select(rect)

			if childEntries != nil {
				entries = append(entries, t.leaves...)
			}
		}
	}

	if len(entries) == 0 {
		return nil
	}

	return entries
}

func (t *qTree) insert(leaf *QEntry) bool {
	if !t.InBounds(leaf.pos) {
		return false
	}

	// Insert to a child tree if we can.
	if t.children != nil {
		for _, child := range t.children {
			if child.insert(leaf) {
				break
			}
		}

		return true
	}

	// Insert into this tree if there's room.
	if len(t.leaves) < t.bucketSize {
		t.leaves = append(t.leaves, leaf)
		return true
	}

	// This tree is now at capacity. Subdivide into quadrants and move the leaves into the children.
	t.children = t.subdivide()

	for _, leaf := range t.leaves {
		for _, child := range t.children {
			if child.insert(leaf) {
				break
			}
		}
	}

	t.leaves = nil

	return true
}

func (t *qTree) subdivide() []*qTree {
	trees := make([]*qTree, 4)
	min := t.rect.Min
	max := t.rect.Max
	center := min.Add(max).Div(2)

	// North west
	trees[nw] = newQuadTree(
		image.Rectangle{
			Min: min,
			Max: center,
		},
		t.bucketSize,
	)

	// North east
	trees[ne] = newQuadTree(
		image.Rectangle{
			Min: image.Pt(center.X, min.Y),
			Max: image.Pt(max.X, center.Y),
		},
		t.bucketSize,
	)

	// South west
	trees[sw] = newQuadTree(
		image.Rectangle{
			Min: image.Pt(min.X, center.Y),
			Max: image.Pt(center.X, max.Y),
		},
		t.bucketSize,
	)

	// South east
	trees[se] = newQuadTree(
		image.Rectangle{
			Min: center,
			Max: max,
		},
		t.bucketSize,
	)

	return trees
}
