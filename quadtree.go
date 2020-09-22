package quadtree

import "image"

const (
	// DefaultBucketSize probably shouldn't be used unless you are working with small datasets.
	// See the comment for NewQuadTree.
	DefaultBucketSize = 4

	// DefaultMaxDepth is a reasonable default to use for smaller datasets. See the comment
	// for NewQuadTree.
	DefaultMaxDepth = 4
)

// QElement is an interface for an element in a QTree. Each element has coordinates and a value.
type QElement interface {
	// Point returns the element's coordinates.
	Point() image.Point

	// Value returns the element's value.
	Value() interface{}
}

// QTree is an interface for a quad tree.
type QTree interface {
	// Bounds returns the tree's boundaries.
	Bounds() image.Rectangle

	// InBounds returns whether the point could fit in the quadtree.
	InBounds(p image.Point) bool

	// Insert adds a point and value to the quad tree and returns true if it was successful.
	// A return value of false usually means the point is outside the bounds of the quad tree.
	Insert(p image.Point, val interface{}) bool

	// Select uses the given rect to search for any elements in the provided area. Returns nil
	// if no elements were found.
	Select(rect image.Rectangle) []QElement
}

type qElement struct {
	p   image.Point
	val interface{}
}

var _ QElement = (*qElement)(nil)

func (e *qElement) Point() image.Point {
	return e.p
}

func (e *qElement) Value() interface{} {
	return e.val
}

type qTree struct {
	// The maximum number of elements a tree can hold before it needs to be subdivided.
	bucketSize int

	// The tree's children, created by subdividing. This is nil prior to subdividing.
	children []*qTree

	// How much further we can dig down into the children before we hit the "bottom".
	// When a tree is subdivided, its children's depth is set to this value - 1.
	// A tree can no longer subdivide once depthRemaining is 0.
	depthRemaining int

	// A list of elements in this tree's bucket. Only leaves can contain elements. When a tree
	// is subdivided its elements are distributed among the subdivisions.
	elements []*qElement

	// The tree's bounds. Only elements which fit inside the bounds can be added to the tree.
	bounds image.Rectangle
}

var _ QTree = (*qTree)(nil)

// NewQuadTree returns a new quad tree.
//
// The bounds is the size of the quad tree space. The quad tree can only contain elements which
// exist within the bounds.
//
// The bucketSize is the maximum number of elements a tree can hold before it is subdivided.
// A larger value for the bucketSize uses less memory but can make fine grained selecting slow.
// Using a value that's too small will cause the tree to become imbalanced. Suffice to say, the
// correct value for the bucket size is application dependent, and you will probably need to test
// different bucket sizes before finding a good middleground.
//
// The maxDepth is the maximum number of times a tree can subdivide itself. This number should
// reflect the size of your dataset. Once a tree has hit its subdivision limit, it will continue
// to add elements beyond its bucketSize. If the maxDepth is too small, the elements will be
// contained in fewer lists, which will cause searching to act more like a linear search rather
// than a binary search. However, if the maxDepth is too large, it can cause problems if
// elements are very close together. For example, if you add several elements directly on top
// of each other, the tree will keep subdividing itself over and over again as it tries to make
// the bounds small enough. Of course, the tree will stop subdividing once it hits the maxDepth.
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
		elements:       make([]*qElement, 0, bucketSize),
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
	return t.insert(&qElement{p: p, val: val})
}

func (t *qTree) Select(rect image.Rectangle) []QElement {
	if !t.bounds.Overlaps(rect) {
		return nil
	}

	elements := []QElement{}

	if t.children == nil {
		if len(t.elements) > 0 {
			leafElements := make([]QElement, len(t.elements))

			for i, leaf := range t.elements {
				leafElements[i] = leaf
			}

			elements = append(elements, leafElements...)
		}
	} else {
		for _, child := range t.children {
			childElements := child.Select(rect)

			if len(childElements) > 0 {
				elements = append(elements, childElements...)
			}
		}
	}

	if len(elements) == 0 {
		return nil
	}

	return elements
}

func (t *qTree) insert(element *qElement) bool {
	if !t.InBounds(element.p) {
		return false
	}

	// If this tree has children that means the bucket was already filled, so insert
	// into one of the children.
	if t.children != nil {
		for _, child := range t.children {
			if child.insert(element) {
				break
			}
		}

		return true
	}

	// Add the element to this tree if there's room for it, or if we have hit the depth limit.
	if t.depthRemaining == 0 || len(t.elements) < t.bucketSize {
		t.elements = append(t.elements, element)
		return true
	}

	// This tree is now at capacity. Subdivide into quadrants and move the leaves into the children.
	t.children = t.subdivide()
	leaves := append(t.elements, element)

	for _, leaf := range leaves {
		for _, child := range t.children {
			if child.insert(leaf) {
				break
			}
		}
	}

	t.elements = nil

	return true
}

func (t *qTree) subdivide() []*qTree {
	trees := make([]*qTree, 4)
	min := t.bounds.Min
	max := t.bounds.Max
	center := min.Add(max).Div(2)

	// North west
	trees[0] = newQuadTree(
		image.Rectangle{
			Min: min,
			Max: center,
		},
		t.bucketSize,
		t.depthRemaining-1,
	)

	// North east
	trees[1] = newQuadTree(
		image.Rectangle{
			Min: image.Pt(center.X, min.Y),
			Max: image.Pt(max.X, center.Y),
		},
		t.bucketSize,
		t.depthRemaining-1,
	)

	// South west
	trees[2] = newQuadTree(
		image.Rectangle{
			Min: image.Pt(min.X, center.Y),
			Max: image.Pt(center.X, max.Y),
		},
		t.bucketSize,
		t.depthRemaining-1,
	)

	// South east
	trees[3] = newQuadTree(
		image.Rectangle{
			Min: center,
			Max: max,
		},
		t.bucketSize,
		t.depthRemaining-1,
	)

	return trees
}
