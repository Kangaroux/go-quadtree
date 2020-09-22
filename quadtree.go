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
	DefaultMaxDepth   = 8
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
	// The maximum number of entries a tree can hold before it needs to be subdivided.
	bucketSize int

	// The tree's children, created by subdividing. This is nil prior to subdividing.
	children []*qTree

	// How much further we can dig down into the children before we hit the "bottom".
	// When a tree is subdivided, its children's depth is set to this value - 1.
	// A tree can no longer subdivide once depthRemaining is 0.
	depthRemaining int

	// A list of entries in this tree's bucket. Only leaves can contain entries. When a tree
	// is subdivided its entries are distributed among the subdivisions.
	entries []*qEntry

	// The tree's bounds. Only entries which fit inside the bounds can be added to the tree.
	bounds image.Rectangle
}

var _ QTree = (*qTree)(nil)

func NewQuadTree(bounds image.Rectangle, bucketSize int, maxDepth int) QTree {
	return newQuadTree(bounds, bucketSize, maxDepth)
}

func newQuadTree(bounds image.Rectangle, bucketSize int, maxDepth int) *qTree {
	if bucketSize < 1 {
		panic("bucketSize must be greater than 0")
	}

	if maxDepth < 0 {
		panic("maxDepth cannot be negative")
	}

	bounds = bounds.Canon()

	if bounds.Empty() {
		panic("bounds must have a positive length for both width and height")
	}

	return &qTree{
		bucketSize:     bucketSize,
		depthRemaining: maxDepth,
		entries:        make([]*qEntry, 0, bucketSize),
		bounds:         bounds,
	}
}

func (t *qTree) Bounds() image.Rectangle {
	return t.bounds
}

func (t *qTree) InBounds(p image.Point) bool {
	return p.In(t.bounds)
}

func (t *qTree) Insert(p image.Point, val interface{}) bool {
	return t.insert(&qEntry{p: p, val: val})
}

func (t *qTree) Select(rect image.Rectangle) []QEntry {
	if !t.bounds.Overlaps(rect) {
		return nil
	}

	entries := []QEntry{}

	if t.children == nil {
		if len(t.entries) > 0 {
			leafEntries := make([]QEntry, len(t.entries))

			for i, leaf := range t.entries {
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

	// Add the entry to this tree if there's room for it, or if we have hit the depth limit.
	if t.depthRemaining == 0 || len(t.entries) < t.bucketSize {
		t.entries = append(t.entries, entry)
		return true
	}

	// This tree is now at capacity. Subdivide into quadrants and move the leaves into the children.
	t.children = t.subdivide()
	leaves := append(t.entries, entry)

	for _, leaf := range leaves {
		for _, child := range t.children {
			if child.insert(leaf) {
				break
			}
		}
	}

	t.entries = nil

	return true
}

func (t *qTree) subdivide() []*qTree {
	trees := make([]*qTree, 4)
	min := t.bounds.Min
	max := t.bounds.Max
	center := min.Add(max).Div(2)

	// North west
	trees[nw] = newQuadTree(
		image.Rectangle{
			Min: min,
			Max: center,
		},
		t.bucketSize,
		t.depthRemaining-1,
	)

	// North east
	trees[ne] = newQuadTree(
		image.Rectangle{
			Min: image.Pt(center.X, min.Y),
			Max: image.Pt(max.X, center.Y),
		},
		t.bucketSize,
		t.depthRemaining-1,
	)

	// South west
	trees[sw] = newQuadTree(
		image.Rectangle{
			Min: image.Pt(min.X, center.Y),
			Max: image.Pt(center.X, max.Y),
		},
		t.bucketSize,
		t.depthRemaining-1,
	)

	// South east
	trees[se] = newQuadTree(
		image.Rectangle{
			Min: center,
			Max: max,
		},
		t.bucketSize,
		t.depthRemaining-1,
	)

	return trees
}
