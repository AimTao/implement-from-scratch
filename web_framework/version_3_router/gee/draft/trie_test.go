package draft

import (
	"fmt"
	"testing"
)

func Test(t *testing.T) {
	trie := Trie{
		root: &Node{next: make(map[rune]*Node)},
	}

	trie.InsertMore("AB", "ABC", "DF", "DH", "XY")
	Print(trie.root)
}

func Print(node *Node) {
	fmt.Printf("Node{isWord:%t, next:map[rune]*Node[", node.isWord)
	n := len(node.next)
	i := 0
	for k, v := range node.next {
		fmt.Printf("%c: %p", k, v)
		i++
		if i < n {
			fmt.Printf(",")
		}
	}
	fmt.Println("]}")
	for _, v := range node.next {
		Print(v)
	}
}
