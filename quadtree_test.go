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
		p  image.Point
		in bool
	}

	bounds := Rect()
	tree := quadtree.NewQuadTree(bounds, 1)

	inputs := []testInput{
		{
			p:  bounds.Min.Sub(image.Pt(1, 1)),
			in: false,
		},
		{
			p:  bounds.Min,
			in: true,
		},
		{
			p:  bounds.Max.Sub(image.Pt(1, 1)),
			in: true,
		},
		{
			p:  bounds.Max,
			in: false,
		},
	}

	for _, input := range inputs {
		require.Equal(t, input.in, tree.InBounds(input.p), input)
	}
}
