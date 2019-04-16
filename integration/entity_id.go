package integration

import (
	"fmt"
	"sort"
	"strings"
)

// EmptyKey empty entity key.
var EmptyKey = EntityKey("")

// EntityKey unique identifier for an entity within a New Relic customer account.
type EntityKey string

//IDAttributes list of identifier attributes used to provide uniqueness for an entity key.
type IDAttributes []IDAttribute

// IDAttribute identifier attribute key-value pair.
type IDAttribute struct {
	Key   string
	Value string
}

// NewIDAttribute creates new identifier attribute.
func NewIDAttribute(key, value string) IDAttribute {
	return IDAttribute{
		Key:   key,
		Value: value,
	}
}

// String stringer stuff
func (k EntityKey) String() string {
	return string(k)
}

// Key generates the entity key based on the entity metadata.
func (m *EntityMetadata) Key() (EntityKey, error) {
	if len(m.Name) == 0 {
		return EmptyKey, nil // Empty value means this agent's default entity identifier
	}
	if m.Namespace == "" {
		//invalid entity: it has name, but not type.
		return EmptyKey, fmt.Errorf("missing 'namespace' field for entity name '%v'", m.Name)
	}

	attrsStr := ""
	sort.Sort(m.IDAttrs)
	m.IDAttrs.removeEmptyAndDuplicates()
	for _, attr := range m.IDAttrs {
		attrsStr = fmt.Sprintf("%v:%v=%v", attrsStr, attr.Key, attr.Value)
	}

	return EntityKey(fmt.Sprintf("%v:%v%s", m.Namespace, m.Name, strings.ToLower(attrsStr))), nil
}

func idAttributes(idAttrs ...IDAttribute) IDAttributes {
	attrs := make(IDAttributes, len(idAttrs))
	if len(attrs) == 0 {
		return attrs
	}
	for i, attr := range idAttrs {
		attrs[i] = attr
	}

	return attrs
}

// Len is part of sort.Interface.
func (a IDAttributes) Len() int {
	return len(a)
}

// Swap is part of sort.Interface.
func (a IDAttributes) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

// Less is part of sort.Interface.
func (a IDAttributes) Less(i, j int) bool {
	return a[i].Key < a[j].Key
}

func (a *IDAttributes) removeEmptyAndDuplicates() {

	var uniques IDAttributes
	var prev IDAttribute
	for i, attr := range *a {
		if prev.Key != attr.Key && attr.Key != "" {
			uniques = append(uniques, attr)
		} else if uniques.Len() >= 1 {
			uniques[i-1].Value = attr.Value
		}
		prev = attr
	}

	*a = uniques
}
