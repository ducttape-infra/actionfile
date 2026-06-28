package actionfile

import (
	"fmt"
	"os"
	"strings"
)

// FindFile locates an Actionfile in the given directory.
func FindFile(dir string) (string, error) {
	for _, name := range CandidateFiles {
		path := dir + "/" + name
		if fi, err := os.Stat(path); err == nil && !fi.IsDir() {
			return path, nil
		}
	}
	return "", fmt.Errorf("no Actionfile found in %s", dir)
}

// ExtractSections parses a markdown file and returns all ###-level sections.
func ExtractSections(file string) (sections []Section, config, vars, shared string, err error) {
	data, err := os.ReadFile(file)
	if err != nil {
		return nil, "", "", "", err
	}
	content := string(data)

	config = extractBlock(content, "config", "ini")
	vars = extractBlock(content, "vars", "sh")
	sections = extractAllSections(content)
	shared = takeShared(&sections)

	return sections, config, vars, shared, nil
}

func extractBlock(content, header, fence string) string {
	lines := strings.Split(content, "\n")
	inHeader := false
	inCode := false
	var body []string
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "### "+header) {
			inHeader = true
			continue
		}
		if !inHeader {
			continue
		}
		if inCode {
			if t == "```" {
				break
			}
			body = append(body, line)
		} else if t == "```"+fence || t == "```" {
			inCode = true
		}
	}
	return strings.TrimSpace(strings.Join(body, "\n"))
}

func extractAllSections(content string) []Section {
	var sections []Section
	lines := strings.Split(content, "\n")
	var cur *Section
	inCode := false

	flush := func() {
		if cur != nil && cur.Body != "" {
			cur.Body = strings.TrimSuffix(cur.Body, "\n")
			sections = append(sections, *cur)
		}
		cur = nil
		inCode = false
	}

	for _, line := range lines {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "### ") {
			flush()
			rest := strings.TrimSpace(t[4:])
			key := ""
			if strings.HasPrefix(rest, "`") && strings.HasSuffix(rest, "`") {
				key = rest[1 : len(rest)-1]
			} else if idx := strings.Index(rest, " "); idx > 0 {
				key = rest[:idx]
			} else {
				key = rest
			}
			if key != "" && key != "config" && key != "vars" {
				cur = &Section{Key: key}
			}
			continue
		}
		if cur != nil {
			if strings.HasPrefix(t, "```sh") {
				inCode = true
				modeStr := strings.TrimSpace(t[5:])
				if modeStr != "" {
					cur.Mode = modeStr
				}
				continue
			}
			if inCode {
				if t == "```" {
					inCode = false
					continue
				}
				if cur.Body != "" {
					cur.Body += "\n"
				}
				cur.Body += line
			}
		}
	}
	flush()
	return sections
}

func takeShared(sections *[]Section) string {
	for i, s := range *sections {
		if s.Key == "shared" {
			*sections = append((*sections)[:i], (*sections)[i+1:]...)
			return s.Body
		}
	}
	return ""
}

// ParseIni converts INI-style config into shell export statements.
func ParseIni(config string) string {
	lines := strings.Split(config, "\n")
	var out []string
	section := ""
	for _, line := range lines {
		t := strings.TrimSpace(line)
		if t == "" {
			continue
		}
		if strings.HasPrefix(t, "[") && strings.HasSuffix(t, "]") {
			section = strings.ToUpper(t[1 : len(t)-1])
			continue
		}
		if idx := strings.Index(t, "="); idx > 0 {
			key := strings.ToUpper(strings.TrimSpace(t[:idx]))
			val := strings.TrimSpace(t[idx+1:])
			val = strings.Trim(val, "\"")
			prefix := ""
			if section != "" {
				prefix = section + "_"
			}
			out = append(out, fmt.Sprintf("export %s%s=\"%s\"", prefix, key, val))
		}
	}
	return strings.Join(out, "\n")
}

// ListKeys returns all section keys.
func ListKeys(sections []Section) []string {
	var keys []string
	for _, s := range sections {
		keys = append(keys, s.Key)
	}
	return keys
}

// ListActions returns keys formatted as "action context" pairs.
func ListActions(sections []Section) []string {
	var out []string
	for _, s := range sections {
		key := s.Key
		if idx := strings.LastIndex(key, "-"); idx > 0 && idx < len(key)-1 {
			out = append(out, key[idx+1:]+" "+key[:idx])
		} else {
			out = append(out, key)
		}
	}
	return out
}
