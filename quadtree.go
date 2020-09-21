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

type QEntry interface {
	Point() image.Point
	Value() interface{}
}

type QTree interface {
	// Bounds returns the tree's boundaries.
	Bounds() image.Rectangle

	// InBounds returns whether the point could fit in the quadtree.
	InBounds(p image.Point) bool

	// Insert adds a point and value to the quad tree and returns true if it was successful.
	// A return value of false usually means the point is outside the bounds of the quad tree.
	Insert(p image.Point, val interface{}) bool

	// Select uses the given rect to search for any entries in the provided area. Returns nil
	// if no entries were found.
	Select(rect image.Rectangle) []QEntry
}

type qEntry struct {
	p   image.Point
	val interface{}
}

var _ QEntry = (*qEntry)(nil)

func (e *qEntry) Point() image.Point {
	return e.p
}

func (e *qEntry) Value() interface{} {
	return e.val
}

type qTree struct {
	bucketSize int
	children   []*qTree
	leaves     []*qEntry
	rect       image.Rectangle
}

var _ QTree = (*qTree)(nil)

func NewQuadTree(bounds image.Rectangle, bucketSize int) QTree {
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
		leaves:     make([]*qEntry, 0, bucketSize),
		rect:       bounds,
	}
}

func (t *qTree) Bounds() image.Rectangle {
	return t.rect
}

func (t *qTree) InBounds(p image.Point) bool {
	return p.In(t.rect)
}

func (t *qTree) Insert(p image.Point, val interface{}) bool {
	return t.insert(&qEntry{p: p, val: val})
}

func (t *qTree) Select(rect image.Rectangle) []QEntry {
	if !t.rect.Overlaps(rect) {
		return nil
	}

	entries := []QEntry{}

	if t.children != nil {
		if len(t.leaves) > 0 {
			leafEntries := make([]QEntry, len(t.leaves))

			for i, leaf := range t.leaves {
				leafEntries[i] = leaf
			}

			entries = append(entries, leafEntries...)
		}
	} else {
		for _, child := range t.children {
			childEntries := child.Select(rect)

			if len(childEntries) > 0 {
				entries = append(entries, childEntries...)
			}
		}
	}

	if len(entries) == 0 {
		return nil
	}

	return entries
}

func (t *qTree) insert(entry *qEntry) bool {
	if !t.InBounds(entry.p) {
		return false
	}

	// If this tree has children that means the bucket was already filled, so insert
	// into one of the children.
	if t.children != nil {
		for _, child := range t.children {
			if child.insert(entry) {
				break
			}
		}

		return true
	}

	// If there's no children try and add it to this bucket.
	if len(t.leaves) < t.bucketSize {
		t.leaves = append(t.leaves, entry)
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
