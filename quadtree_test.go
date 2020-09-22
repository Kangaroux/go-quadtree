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
	t.Run("invalid bucketSize", func(t *testing.T) {
		inputs := []int{-500, -1, 0}

		for _, x := range inputs {
			require.Panics(t, func() {
				quadtree.NewQuadTree(Rect(), x)
			})
		}
	})

	t.Run("invalid rect", func(t *testing.T) {
		inputs := []image.Rectangle{
			image.Rect(0, 0, 100, 0),
			image.Rect(0, 0, 0, 100),
			image.Rect(0, 0, 0, 0),
		}

		for _, x := range inputs {
			require.Panics(t, func() {
				quadtree.NewQuadTree(x, 1)
			})
		}
	})

	t.Run("ok", func(t *testing.T) {
		quadtree.NewQuadTree(Rect(), quadtree.DefaultBucketSize)
	})
}

func Test_InBounds(t *testing.T) {
	type testInput struct {
		in bool
		p  image.Point
	}

	bounds := Rect()
	tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize)

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
		tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize)

		require.False(t, tree.Insert(bounds.Min.Sub(image.Pt(1, 1)), nil))
		require.True(t, tree.Insert(bounds.Min, nil))
	})
}

func Test_Select(t *testing.T) {
	t.Run("empty tree", func(t *testing.T) {
		bounds := Rect()
		tree := quadtree.NewQuadTree(bounds, 1)

		require.Empty(t, tree.Select(bounds))
	})

	t.Run("select single from flat tree", func(t *testing.T) {
		bounds := Rect()
		tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize)

		tree.Insert(bounds.Min, nil)

		entries := tree.Select(bounds)

		require.Equal(t, 1, len(entries))
		require.True(t, bounds.Min.Eq(entries[0].Point()))
	})

	t.Run("select multiple from flat tree", func(t *testing.T) {
		bounds := Rect()
		tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize)
		points := []image.Point{
			bounds.Min,
			bounds.Max.Sub(image.Pt(1, 1)),
		}

		for _, p := range points {
			require.True(t, tree.Insert(p, nil))
		}

		entries := tree.Select(bounds)

		require.Equal(t, len(points), len(entries))
	})

	t.Run("select multiple from deep tree", func(t *testing.T) {
		bounds := image.Rect(0, 0, 500, 500)
		tree := quadtree.NewQuadTree(bounds, quadtree.DefaultBucketSize)
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

		entries := tree.Select(bounds)

		require.Equal(t, len(points), len(entries))
	})
}
