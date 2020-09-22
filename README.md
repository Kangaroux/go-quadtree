This is an implementation of a point quadtree written in Go.

## What is a quadtree?
A quadtree is an intimidating sounding name for something that's very simple.

It's used for storing coordinates ("elements") in a 2D space, and then quickly looking up those elements at a later time. Specifically, the lookups are usually things like "find all elements in a particular region" or "find all elements that are near a certain point".

## How does it work?
A quadtree organizes elements that are close together into regions. The quadtree has a bucket you add elements into, but this bucket has a size limit. Once the bucket becomes full, the tree divides the region into 4 equal spaces/cells/quadrants, then distributes the elements among the quadrants.

In essence, a quadtree is using a grid to organize its elements. However, a key feature of the quadtree is that the grid doesn't need to be evenly divided up. A quadtree only subdivides a grid cell when it needs to. Some cells may be very large because there are few or no elements inside the cell. Other cells may contain many elements and have been subdivided several times.

## When would I use a quadtree?
If you need to perform lookups as mentioned in the first section, then a quadtree is probably a good choice.

A good example would be a 2D RPG game. Suppose there is a spellcaster with an ability that damages enemies in an area. You could create a quadtree containing the positions of all enemies, and then query the tree to find nearby enemies. In a game with only a few enemies on the screen a quadtree would be overkill, but if there are dozens or hundreds of enemies it could be a good option.
