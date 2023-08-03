package draft

type Node struct {
    isWord bool           // 是否是单词结尾
    next   map[rune]*Node // 子节点
}
type Trie struct {
    root *Node
}

// Insert 将单词插入前缀树
func (trie *Trie) Insert(word string) {
    cur := trie.root
    for _, char := range []rune(word) {
        if _, ok := cur.next[char]; !ok {
            cur.next[char] = &Node{next: make(map[rune]*Node)}
        }
        cur = cur.next[char]
    }
    cur.isWord = true
}

func (trie *Trie) InsertMore(words ...string) {
    for _, word := range words {
        trie.Insert(word)
    }
}

// Search 在前缀树中搜索单词是否存在
func (trie *Trie) Search(word string) bool {
    cur := trie.root
    for _, char := range []rune(word) {
        if _, ok := cur.next[char]; !ok {
            return false
        }
        cur = cur.next[char]
    }
    return cur.isWord
}