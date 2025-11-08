package htmlutil

import "golang.org/x/net/html"

// GetAttr returns the value of the specified attribute from an HTML node.
// Returns empty string if the attribute is not found.
func GetAttr(node *html.Node, key string) string {
	for _, attr := range node.Attr {
		if attr.Key == key {
			return attr.Val
		}
	}
	return ""
}
