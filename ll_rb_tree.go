// Copyright 2010 -- Peter Williams, all rights reserved
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The heteroset package implements heterogeneou sets
package heteroset

import "reflect"

// Implement 2-3 left Leaning Red Black Trees for for internal representation.
// It is based on the Java implementation described by Robert Sedgewick
// in his paper entitled "left-leaning Red-Black Trees"
// available at: <www.cs.princeton.edu/~rs/talks/LLRB/LLRB.pdf>.
// The principal difference (other than the conversion to Go) is that the items
// being inserted combine the roles of both key and value

// Prospective set items must implement this interface and must satisfy the
// following formal requirements:
//	 a.Compare(b) < 0 implies b.Compare(a) > 0
//	 a.Compare(b) > 0 implies b.Compare(a) < 0
//	 a.Compare(b) == 0 implies b.Compare(a) == 0
type Item interface {
	Compare(other Item) int
}

// LLRB tree node
type ll_rb_node struct {
	item Item
	left, right *ll_rb_node
	red bool
}

func new_ll_rb_node(item Item) *ll_rb_node {
	node := new(ll_rb_node)
	node.item = item
	node.red = true
	return node
}

func (this *ll_rb_node) compare_item(item Item) int {
	thistp := reflect.Typeof(this.item).PkgPath()
	itemtp := reflect.Typeof(item).PkgPath()
	for i := 0; ; i++ {
		if i >= len(thistp) {
			if len(thistp) == len(itemtp) {
				break
			} else {
				return -1
			}
		} else if i >= len(itemtp) {
			return 1
		} else if thistp[i] < itemtp[i] {
			return -1
		} else if thistp[i] > itemtp[i] {
			return 1
		}
	}
	return this.item.Compare(item)
}

func is_red(node *ll_rb_node) bool { return node != nil && node.red }

func flip_colours(node *ll_rb_node) {
	node.red = !node.red
	node.left.red = !node.left.red
	node.right.red = !node.right.red
}

func rotate_left(node *ll_rb_node) *ll_rb_node {
	tmp := node.right
	node.right = tmp.left
	tmp.left = node
	tmp.red = node.red
	node.red = true
	return tmp
}

func rotate_right(node *ll_rb_node) *ll_rb_node {
	tmp := node.left
	node.left = tmp.right
	tmp.right = node
	tmp.red = node.red
	node.red = true
	return tmp
}

func fix_up(node *ll_rb_node) *ll_rb_node {
	if is_red(node.right) && !is_red(node.left) {
		node = rotate_left(node)
	}
	if is_red(node.left) && is_red(node.left.left) {
		node = rotate_right(node)
	}
	if is_red(node.left) && is_red(node.right) {
		flip_colours(node)
	}
	return node
}

func insert(node *ll_rb_node, item Item) (*ll_rb_node, bool) {
	if node == nil {
		return new_ll_rb_node(item), true
	}
	inserted := false
	switch cmp := node.compare_item(item); {
	case cmp < 0:
		node, inserted = insert(node.left, item)
	case cmp > 0:
		node, inserted = insert(node.right, item)
	default:
	}
	return fix_up(node), inserted
}

func move_red_left(node *ll_rb_node) *ll_rb_node {
	flip_colours(node)
	if (is_red(node.right.left)) {
		node.right = rotate_right(node.right)
		node = rotate_left(node)
		flip_colours(node)
	}
	return node
}

func move_red_right(node *ll_rb_node) *ll_rb_node {
	flip_colours(node)
	if (is_red(node.left.left)) {
		node = rotate_right(node)
		flip_colours(node)
	}
	return node
}

func delete_left_most(node *ll_rb_node) *ll_rb_node {
	if node.left == nil {
		return nil
	}
	if !is_red(node.left) && !is_red(node.left.left) {
		node = move_red_left(node)
	}
	node.left = delete_left_most(node.left)
	return fix_up(node)
}

func delete(node *ll_rb_node, item Item) (*ll_rb_node, bool) {
	var deleted bool
	if node.compare_item(item) < 0 {
		if !is_red(node.left) && !is_red(node.left.left) {
			node = move_red_left(node)
		}
		node.left, deleted = delete(node.left, item)
	} else {
		if is_red(node.left) {
			node = rotate_right(node)
		}
		if node.compare_item(item) == 0 && node.right == nil {
			return nil, true
		}
		if !is_red(node.right) && !is_red(node.right.left) {
			node = move_red_right(node)
		}
		if node.compare_item(item) == 0 {
			left_most := node.right
			for left_most.left != nil {
				left_most = left_most.left
			}
			node.item = left_most.item
			node.right = delete_left_most(node.right)
			deleted = true
		} else {
			node.right, deleted = delete(node.right, item)
		}
	}
	return fix_up(node), deleted
}

// A stack to facilitate iteration
type node_stack struct {
	node *ll_rb_node
	stack *node_stack
}

func is_empty(stack *node_stack) bool {
	return stack != nil
}

func push(stack *node_stack, node *ll_rb_node) *node_stack {
	return &node_stack{node, stack}
}

func pop(stack *node_stack) (*node_stack, *ll_rb_node) {
	return stack.stack, stack.node
}

func iterate(node *ll_rb_node, c chan<- Item) {
	for stack := push(nil, node); !is_empty(stack); {
		var current *ll_rb_node
		stack, current = pop(stack)
		if current.right != nil {
			stack = push(stack, current.right)
		}
		if current.left != nil {
			stack = push(stack, current.left)
		}
		c <- current.item
	}
	close(c)
}

type ll_rb_tree struct {
	root *ll_rb_node
	count uint64
}

func (this ll_rb_tree) find(item Item) (found bool, iterations uint) {
	if this.count == 0 {
		return
	}
	for node := this.root; node != nil && !found; {
		iterations++
		switch cmp := node.compare_item(item); {
		case cmp < 0:
			node = node.left
		case cmp > 0:
			node = node.right
		default:
			found = true
		}
	}
	return
}

func (this ll_rb_tree) insert(item Item) {
	var inserted bool
	this.root, inserted = insert(this.root, item)
	if inserted {
		this.count++
	}
	this.root.red = false
}

func (this ll_rb_tree) delete(item Item) {
	var deleted bool
	this.root, deleted = delete(this.root, item)
	if deleted {
		this.count--
	}
	this.root.red = false
}

func (this ll_rb_tree) iterator() <-chan Item {
	c := make(chan Item)
	go iterate(this.root, c)
	return c
}

