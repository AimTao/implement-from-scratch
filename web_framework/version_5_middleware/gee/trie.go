// gee/trie.go

package gee

import (
    "strings"
)

type node struct {
    path     string           // 匹配上的完整的路由地址，只有最后节点才能保存 path
    part     string           // 当前节点的 URL 片段
    children map[string]*node // 储存后续片段的节点
    isWild   bool             // 是否是通配符节点
}

// 路由注册，将路由地址插入到前缀树中
func (root *node) insert(pattern string) {
    cur := root
    patterns := parsePath(pattern) // 提前将路由地址，分割成片段保存在数组中

    // 依次遍历路由地址的片段，不存在则创建保存该片段的节点。
    for _, part := range patterns {
        if _, ok := cur.children[part]; !ok {
            cur.children[part] = &node{
                part:     part,
                children: make(map[string]*node),
                isWild:   part[0] == ':' || part[0] == '*',
            }
        }
        cur = cur.children[part] // cur 指向保存当前片段的节点，后面的片段保存在 cur 当前节点的子节点中
    }

    cur.path = pattern // 当遍历完路由地址，在最后一个节点中，保存完整的路由地址，其他节点不保存完整的路由地址。
}

// search 获取路由树节点以及请求地址中的变量
func (root *node) search(pattern string) (*node, map[string]string) {
    params := make(map[string]string)

    cur := root
    patterns := parsePath(pattern)

    // 依次遍历路由地址的片段，无法匹配上则退出
    for _, part := range patterns {
        if cur.children[part] == nil { // 无法匹配上，开始尝试通配符匹配
            for k, v := range cur.children { // 遍历当前片段下的所有子节点，排查有通配符的节点。
                if v.isWild == true && k[0] == '*' { // 找到*，保存该片段及该片段后的所有内容做参数
                    params[k[1:]] = pattern[strings.Index(pattern, part):]
                    return v, params
                } else if v.isWild == true && k[0] == ':' { // 找到：，保存该片段做参数
                    params[k[1:]] = part
                    cur = v
                    break
                } else { // 没有通配符节点
                    return nil, nil
                }
            }
        } else { // 当前片段准确匹配上，继续匹配后面的片段
            cur = cur.children[part]
        }
    }

    // 所有请求路经片段均匹配完毕，检查当前节点是否有完整的路由地址。比如,路由注册了 /a/b，请求路经是 /a，虽然也匹配上了，但 a 这个节点未保存完整的路经，只有最后的节点 b 节点会保存。
    if cur.path != "" {
        return cur, params
    }
    return nil, nil
}

func parsePath(pattern string) []string {
    patterns := strings.Split(pattern, "/")
    /* 注意：
       pattern 是 "/hello"，前面会被分出两个空字符串，需要删除
       pattern 是 "hello/", 后面会被分出一个空字符串，需要删除
       pattern 是 "/"，前后会被分出两个空字符串，需要删除
       所以，直接删除前面的空格和后面的空格。
    */
    if len(patterns) > 0 && patterns[0] == "" { // 删除前面的空格
        patterns = patterns[1:]
    }
    if len(patterns) > 0 && patterns[len(patterns)-1] == "" { // 删除后面的空格
        patterns = patterns[:len(patterns)-1]
    }
    return patterns
}