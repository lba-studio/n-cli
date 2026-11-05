package config

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"gopkg.in/yaml.v3"
)

const (
	disabledKey = "disabled"
	systemKey   = "system"
	customKey   = "custom"
	customsKey  = "customs"
)

var customsIndexPattern = regexp.MustCompile(`^customs\[(\d+)\]$`)

// SetChannelDisabled sets the `disabled` flag of the channel identified by label
// in the config file at configPath to the given value.
//
// The file is edited in place: only the `disabled` value (or a newly inserted
// `disabled:` line) changes, so comments, key order, blank lines and formatting
// elsewhere are kept byte for byte.
//
// Accepted labels: "system", "discord", "slack", "custom" (legacy block),
// "customs[i]", or the name of a custom entry. It returns a human-readable
// description of the resolved channel and whether the file was changed.
func SetChannelDisabled(configPath, label string, disabled bool) (resolved string, changed bool, err error) {
	src, err := os.ReadFile(configPath)
	if err != nil && !os.IsNotExist(err) {
		return "", false, fmt.Errorf("read %s: %w", configPath, err)
	}

	doc, err := parseConfigDocument(src)
	if err != nil {
		return "", false, fmt.Errorf("parse %s: %w", configPath, err)
	}

	label = strings.TrimSpace(label)
	ch, err := resolveChannel(doc, label)
	if err != nil {
		return "", false, err
	}

	var out []byte
	if ch.node == nil {
		// Only reachable for "system", which is on by default and may have no block.
		if !disabled {
			return ch.label, false, nil
		}
		out, err = insertSystemBlock(src, doc, ch.key)
	} else {
		var current bool
		if current, err = mappingDisabled(ch.node); err != nil {
			return "", false, fmt.Errorf("%s: %w", ch.label, err)
		}
		if current == disabled {
			return ch.label, false, nil
		}
		out, err = setDisabledValue(src, ch.node, disabled)
	}
	if err != nil {
		return "", false, fmt.Errorf("%s: %w", ch.label, err)
	}

	if err := verifyEdit(src, out, ch.path, disabled); err != nil {
		return "", false, fmt.Errorf("%s: refusing to write %s: %w", ch.label, configPath, err)
	}
	if err := writeFilePreservingPerm(configPath, out); err != nil {
		return "", false, err
	}
	return ch.label, true, nil
}

// channel is a resolved toggle target.
type channel struct {
	label string        // human-readable description, e.g. "customs[0] (pagerduty)"
	path  []interface{} // location in the decoded config: string keys and int indices
	node  *yaml.Node    // the channel's mapping node; nil if the block does not exist
	key   *yaml.Node    // the channel's key node, when node is nil but the key exists
}

func parseConfigDocument(src []byte) (*yaml.Node, error) {
	var doc yaml.Node
	if len(bytes.TrimSpace(src)) == 0 {
		return &doc, nil
	}
	if err := yaml.Unmarshal(src, &doc); err != nil {
		return nil, err
	}
	return &doc, nil
}

// rootMapping returns the document's top-level mapping, or nil for an empty document.
func rootMapping(doc *yaml.Node) (*yaml.Node, error) {
	if doc.Kind != yaml.DocumentNode || len(doc.Content) == 0 {
		return nil, nil
	}
	root := doc.Content[0]
	if isNull(root) {
		return nil, nil
	}
	if root.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("config root must be a mapping")
	}
	return root, nil
}

func resolveChannel(doc *yaml.Node, label string) (channel, error) {
	if label == "" {
		return channel{}, fmt.Errorf("label must not be empty")
	}

	root, err := rootMapping(doc)
	if err != nil {
		return channel{}, err
	}

	switch label {
	case systemKey:
		key, value := mappingEntry(root, systemKey)
		ch := channel{label: systemKey, path: []interface{}{systemKey}, key: key}
		if value == nil || isNull(value) {
			return ch, nil
		}
		node, err := channelMapping(systemKey, value)
		if err != nil {
			return channel{}, err
		}
		ch.node, ch.key = node, nil
		return ch, nil
	case "discord", "slack", customKey:
		_, value := mappingEntry(root, label)
		if value == nil || isNull(value) {
			return channel{}, fmt.Errorf("%q is not configured", label)
		}
		node, err := channelMapping(label, value)
		if err != nil {
			return channel{}, err
		}
		return channel{label: describe(label, node), path: []interface{}{label}, node: node}, nil
	}

	customs, err := customsEntries(root)
	if err != nil {
		return channel{}, err
	}

	if match := customsIndexPattern.FindStringSubmatch(label); match != nil {
		i, err := strconv.Atoi(match[1])
		if err != nil || i >= len(customs) {
			return channel{}, fmt.Errorf("%s does not exist (%d customs entries configured)", label, len(customs))
		}
		return channel{label: describe(label, customs[i]), path: []interface{}{customsKey, i}, node: customs[i]}, nil
	}

	var matches []channel
	if _, value := mappingEntry(root, customKey); value != nil && value.Kind == yaml.MappingNode && entryName(value) == label {
		matches = append(matches, channel{label: customKey, path: []interface{}{customKey}, node: value})
	}
	for i, entry := range customs {
		if entryName(entry) == label {
			matches = append(matches, channel{label: fmt.Sprintf("customs[%d]", i), path: []interface{}{customsKey, i}, node: entry})
		}
	}

	switch len(matches) {
	case 0:
		return channel{}, fmt.Errorf("no channel matches %q; use system, discord, slack, custom, customs[i] or a custom entry's name", label)
	case 1:
		m := matches[0]
		m.label = describe(m.label, m.node)
		return m, nil
	default:
		labels := make([]string, len(matches))
		for i, m := range matches {
			labels[i] = m.label
		}
		return channel{}, fmt.Errorf("%q is ambiguous; use one of: %s", label, strings.Join(labels, ", "))
	}
}

// customsEntries returns the mapping nodes of root["customs"].
func customsEntries(root *yaml.Node) ([]*yaml.Node, error) {
	_, value := mappingEntry(root, customsKey)
	if value == nil || isNull(value) {
		return nil, nil
	}
	if value.Kind != yaml.SequenceNode {
		return nil, fmt.Errorf("%s must be a list", customsKey)
	}
	entries := make([]*yaml.Node, len(value.Content))
	for i, item := range value.Content {
		node, err := channelMapping(fmt.Sprintf("%s[%d]", customsKey, i), item)
		if err != nil {
			return nil, err
		}
		entries[i] = node
	}
	return entries, nil
}

// channelMapping checks that value is a mapping that can be edited in place.
func channelMapping(label string, value *yaml.Node) (*yaml.Node, error) {
	if value.Kind == yaml.AliasNode {
		return nil, fmt.Errorf("%s is a YAML alias; edit the file manually so the anchor's other users are not changed", label)
	}
	if value.Kind != yaml.MappingNode {
		return nil, fmt.Errorf("%s must be a mapping", label)
	}
	return value, nil
}

// mappingEntry returns the key and value nodes for key in m, or nils.
func mappingEntry(m *yaml.Node, key string) (*yaml.Node, *yaml.Node) {
	if m == nil || m.Kind != yaml.MappingNode {
		return nil, nil
	}
	for i := 0; i+1 < len(m.Content); i += 2 {
		k := m.Content[i]
		if k.Kind == yaml.ScalarNode && k.Value == key {
			return k, m.Content[i+1]
		}
	}
	return nil, nil
}

func isNull(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.ShortTag() == "!!null"
}

func entryName(m *yaml.Node) string {
	_, v := mappingEntry(m, "name")
	if v == nil || v.Kind != yaml.ScalarNode {
		return ""
	}
	return v.Value
}

// mappingDisabled reports the effective `disabled` value of m, including values
// inherited through merge keys.
func mappingDisabled(m *yaml.Node) (bool, error) {
	var decoded map[string]interface{}
	if err := m.Decode(&decoded); err != nil {
		return false, err
	}
	return decoded[disabledKey] == true, nil
}

func describe(label string, m *yaml.Node) string {
	if name := entryName(m); name != "" && name != label {
		return fmt.Sprintf("%s (%s)", label, name)
	}
	return label
}

// setDisabledValue rewrites the `disabled` value of mapping m in src, or inserts a
// `disabled:` entry before m's first key when m has no direct `disabled` key.
func setDisabledValue(src []byte, m *yaml.Node, disabled bool) ([]byte, error) {
	value := strconv.FormatBool(disabled)
	lines := newLineIndex(src)

	if key, v := mappingEntry(m, disabledKey); key != nil {
		if v.Kind != yaml.ScalarNode {
			return nil, fmt.Errorf("%s at line %d must be true or false", disabledKey, v.Line)
		}
		start, end, err := scalarSpan(lines, v)
		if err != nil {
			return nil, err
		}
		return splice(src, start, end, value), nil
	}

	entry := disabledKey + ": " + value
	if len(m.Content) == 0 {
		if m.Style&yaml.FlowStyle == 0 {
			return nil, fmt.Errorf("cannot edit the empty mapping at line %d; edit the file manually", m.Line)
		}
		// `{}`: insert between the braces.
		open, err := lines.offset(m.Line, m.Column)
		if err != nil || src[open] != '{' {
			return nil, fmt.Errorf("cannot locate the mapping at line %d; edit the file manually", m.Line)
		}
		return splice(src, open+1, open+1, entry), nil
	}

	first := m.Content[0]
	at, err := lines.offset(first.Line, first.Column)
	if err != nil {
		return nil, err
	}
	if m.Style&yaml.FlowStyle != 0 {
		return splice(src, at, at, entry+", "), nil
	}
	// Block mapping: put the new entry on its own line, aligned with the first key.
	// This also works when the first key follows a sequence dash ("- name: x").
	indent := strings.Repeat(" ", first.Column-1)
	return splice(src, at, at, entry+lines.newline+indent), nil
}

// insertSystemBlock adds `disabled: true` under `system`, which has no block yet.
// key is the existing `system:` key node when its value is null, else nil.
func insertSystemBlock(src []byte, doc *yaml.Node, key *yaml.Node) ([]byte, error) {
	lines := newLineIndex(src)
	nl := lines.newline
	entry := disabledKey + ": true"

	root, err := rootMapping(doc)
	if err != nil {
		return nil, err
	}
	if root != nil && (root.Style&yaml.FlowStyle != 0 || root.Column != 1) {
		return nil, fmt.Errorf("cannot add a %s block to this config layout; edit the file manually", systemKey)
	}

	if key != nil {
		// "system:" with an empty value: add the entry on the following line.
		if _, v := mappingEntry(root, systemKey); v.Value != "" {
			return nil, fmt.Errorf("cannot edit the %s value at line %d; edit the file manually", systemKey, key.Line)
		}
		eol := lines.lineEnd(key.Line)
		indent := strings.Repeat(" ", key.Column-1+2)
		return splice(src, eol, eol, nl+indent+entry), nil
	}

	// Append at the end of the file, where a new top-level key is valid.
	out := append([]byte{}, src...)
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out = append(out, nl...)
	}
	return append(out, systemKey+":"+nl+"  "+entry+nl...), nil
}

// scalarSpan returns the byte range of the single-line scalar v in src.
func scalarSpan(lines lineIndex, v *yaml.Node) (int, int, error) {
	start, err := lines.offset(v.Line, v.Column)
	if err != nil {
		return 0, 0, err
	}
	rest := lines.src[start:lines.lineEnd(v.Line)]

	var token string
	switch v.Style {
	case 0:
		token = v.Value
	case yaml.DoubleQuotedStyle:
		token = `"` + v.Value + `"`
	case yaml.SingleQuotedStyle:
		token = `'` + v.Value + `'`
	}
	if token == "" || !bytes.HasPrefix(rest, []byte(token)) {
		return 0, 0, fmt.Errorf("cannot edit the %s value at line %d; edit the file manually", disabledKey, v.Line)
	}
	return start, start + len(token), nil
}

func splice(src []byte, start, end int, insert string) []byte {
	out := make([]byte, 0, len(src)-(end-start)+len(insert))
	out = append(out, src[:start]...)
	out = append(out, insert...)
	return append(out, src[end:]...)
}

// lineIndex maps yaml.Node line/column positions (1-based, columns in characters)
// to byte offsets.
type lineIndex struct {
	src     []byte
	starts  []int
	newline string
}

func newLineIndex(src []byte) lineIndex {
	idx := lineIndex{src: src, starts: []int{0}, newline: "\n"}
	for i, b := range src {
		if b == '\n' {
			idx.starts = append(idx.starts, i+1)
		}
	}
	if bytes.Contains(src, []byte("\r\n")) {
		idx.newline = "\r\n"
	}
	return idx
}

// lineEnd returns the offset of the line terminator of line (1-based).
func (idx lineIndex) lineEnd(line int) int {
	end := len(idx.src)
	if line < len(idx.starts) {
		end = idx.starts[line] - 1
	}
	if end > idx.starts[line-1] && idx.src[end-1] == '\r' {
		end--
	}
	return end
}

func (idx lineIndex) offset(line, column int) (int, error) {
	if line < 1 || line > len(idx.starts) || column < 1 {
		return 0, fmt.Errorf("position %d:%d is outside the file", line, column)
	}
	off := idx.starts[line-1]
	end := idx.lineEnd(line)
	for c := 1; c < column; c++ {
		if off >= end {
			return 0, fmt.Errorf("position %d:%d is outside the file", line, column)
		}
		_, size := utf8.DecodeRune(idx.src[off:end])
		off += size
	}
	return off, nil
}

// verifyEdit checks that out decodes to the same data as src except for the
// `disabled` value of the channel at path.
func verifyEdit(src, out []byte, path []interface{}, disabled bool) error {
	var want, got interface{}
	if err := yaml.Unmarshal(src, &want); err != nil {
		return err
	}
	if err := yaml.Unmarshal(out, &got); err != nil {
		return fmt.Errorf("edited config does not parse: %w", err)
	}
	if want == nil {
		want = map[string]interface{}{}
	}

	parent := want
	for i, step := range path {
		last := i == len(path)-1
		switch p := parent.(type) {
		case map[string]interface{}:
			key := step.(string)
			if last && p[key] == nil {
				p[key] = map[string]interface{}{}
			}
			parent = p[key]
		case []interface{}:
			parent = p[step.(int)]
		default:
			return fmt.Errorf("unexpected config shape")
		}
	}
	target, ok := parent.(map[string]interface{})
	if !ok {
		return fmt.Errorf("unexpected config shape")
	}
	target[disabledKey] = disabled

	if !reflect.DeepEqual(want, got) {
		return fmt.Errorf("edit changed more than %s", disabledKey)
	}
	return nil
}

func writeFilePreservingPerm(path string, data []byte) error {
	perm := os.FileMode(0644)
	if info, err := os.Stat(path); err == nil {
		perm = info.Mode().Perm()
	} else if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return fmt.Errorf("create config directory: %w", err)
	}
	if err := os.WriteFile(path, data, perm); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}
