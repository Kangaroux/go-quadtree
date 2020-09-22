package quadtree_test

import (
	"image"
	"testing"

	"github.com/Kangaroux/go-quadtree"
	"github.com/stretchr/testify/require"
)

func Rect() image.Rectangle {
	return image.Rect(0, 0, 500, 500)
}

func Test_NewQuadTree(t *testing.T) {
	t.Run("bucketSize", func(t *testing.T) {
		type testInput struct {
			ok  bool
			val int
		}

		inputs := []testInput{
			{false, -1},
			{false, 0},
			{true, 1},
			{true, quadtree.DefaultBucketSize},
		}

		for _, input := range inputs {
			fn := func() {
				quadtree.NewQuadTree(Rect(), input.val, quadtree.DefaultMaxDepth)
			}

			if input.ok {
				require.NotPanics(t, fn)
			} else {
				require.Panics(t, fn)
			}
		}
	})

	t.Run("maxDepth", func(t *testing.T) {
		type testInput struct {
			ok  bool
			val int
		}

		inputs := []testInput{
			{false, -1},
			{true, 0},
			{true, quadtree.DefaultMaxDepth},
		}

		for _, input := range inputs {
			fn := func() {
				quadtree.NewQuadTree(Rect(), quadtree.DefaultBucketSize, input.val)
			}

			if input.ok {
				require.NotPanics(t, fn)
			} else {
				require.Panics(t, fn)
			}
		}
	})

	t.Run("bounds", func(t *testing.T) {
		type testInput struct {
			ok  bool
			val image.Rectangle
		}

		inputs := []testInput{
			{false, image.Rect(0, 0, 0, 0)},
			{false, image.Rect(0, 0, 100, 0)},
			{false, image.Rect(0, 0, 0, 100)},
			{true, image.Rect(0, 0, 100, 100)},
			{true, image.Rect(100, 100, 0, 0)},
		}

		for _, input := range inputs {
			fn := func() {
				quadtree.NewQuadTree(input.val, quadtree.DefaultBucketSize, quadtree.DefaultMaxDepth)
			}

			if input.ok {
				require.NotPanics(t, fn)
			} else {
				require.Panics(t, fn)
			}
		}
	})

	t.Run("ok", func(t *testing.T) {
		quadtree.NewQuadTree(Rect(), quadtree.DefaultBucketSize, quadtree.DefaultMaxDepth)
	})
}

func Test_InBounds(t *testing.T) {
	type testInput struct {
		in bool
		p  image.Point
	}

	bounds := Rect()
	tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize, quadtree.DefaultMaxDepth)

	inputs := []testInput{
		{false, bounds.Min.Sub(image.Pt(1, 1))},
		{true, bounds.Min},
		{true, bounds.Max.Sub(image.Pt(1, 1))},
		{false, bounds.Max},
	}

	for _, input := range inputs {
		require.Equal(t, input.in, tree.InBounds(input.p), input)
	}
}

func Test_Insert(t *testing.T) {
	t.Run("checks boundary", func(t *testing.T) {
		bounds := Rect()
		tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize, quadtree.DefaultMaxDepth)

		require.False(t, tree.Insert(bounds.Min.Sub(image.Pt(1, 1)), nil))
		require.True(t, tree.Insert(bounds.Min, nil))
	})
}

func Test_Select(t *testing.T) {
	t.Run("empty tree", func(t *testing.T) {
		bounds := Rect()
		tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize, quadtree.DefaultMaxDepth)

		require.Empty(t, tree.Select(bounds))
	})

	t.Run("select single from flat tree", func(t *testing.T) {
		bounds := Rect()
		tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize, quadtree.DefaultMaxDepth)

		tree.Insert(bounds.Min, nil)

		elements := tree.Select(bounds)

		require.Equal(t, 1, len(elements))
		require.True(t, bounds.Min.Eq(elements[0].Point()))
	})

	t.Run("select multiple from flat tree", func(t *testing.T) {
		bounds := Rect()
		tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize, quadtree.DefaultMaxDepth)
		points := []image.Point{
			bounds.Min,
			bounds.Max.Sub(image.Pt(1, 1)),
		}

		for _, p := range points {
			require.True(t, tree.Insert(p, nil))
		}

		elements := tree.Select(bounds)

		require.Equal(t, len(points), len(elements))
	})

	t.Run("select multiple from deep tree", func(t *testing.T) {
		bounds := image.Rect(0, 0, 500, 500)
		tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize, quadtree.DefaultMaxDepth)
		points := []image.Point{
			image.Pt(0, 0),
			image.Pt(0, 50),
			image.Pt(10, 0),
			image.Pt(460, 123),
			image.Pt(400, 20),
			image.Pt(100, 350),
			image.Pt(150, 200),
		}

		for _, p := range points {
			require.True(t, tree.Insert(p, nil))
		}

		elements := tree.Select(bounds)

		require.Equal(t, len(points), len(elements))
	})
}
