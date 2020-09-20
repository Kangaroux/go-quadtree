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
