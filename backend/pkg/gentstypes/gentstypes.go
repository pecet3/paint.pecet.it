package gentstypes

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"unicode"
)

type goField struct {
	goName  string // Uuid string `json:"uuid"` = Uuid
	goType  string // string, bool, int etc or raw name of struct if it's nested
	tag     string // Uuid string `json:"uuid"` = json:"uuid"
	comment string // if above field is comment put it here
}

type goStruct struct {
	alternativeName string
	ident           string // packageName + "." + structName
	structName      string // golang struct name
	file            string // path to file where is struct
	packageName     string // go package name from struct comes
	comment         string // if above struct is comment put it here
	fields          []goField
	rawFields       []string // whole lines Uuid string `json:"uuid"` = Uuid string `json:"uuid"` = Uuid
}

type tsField struct {
	tsName        string // comes from json tag
	tsType        string // string, boolean, number. if nested tsType.name.
	comment       string // if above field is comment put it hered
	inlineComment string
}
type tsType struct {
	name    string //packageName with first char toUpper + goStruct.structName with first char toUpper example - AuthUser
	fields  []tsField
	comment string
}

func parseFile(path string, structsMap map[string]*goStruct) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var packageName string
	var currentComment string
	var alternativeName string
	var currentStruct *goStruct
	inStruct := false
	braceCount := 0

	structs := []*goStruct{}
	for scanner.Scan() {
		err := scanner.Err()
		if err != nil {
			panic(err)
		}
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "//") {
			c := strings.TrimSpace(strings.TrimPrefix(trimmed, "//"))
			if strings.HasPrefix(c, "name:") {
				parts := strings.Fields(c)
				if len(parts) >= 2 {
					alternativeName = parts[1]
					continue
				}
			}
			if currentComment != "" {
				currentComment += "\n" + c
			} else {
				currentComment = c
			}
			continue
		}

		if strings.HasPrefix(trimmed, "package ") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 2 {
				packageName = parts[1]
			}
			currentComment = ""
			continue
		}

		if strings.HasPrefix(trimmed, "type ") &&
			strings.Contains(trimmed, " struct") &&
			strings.HasSuffix(trimmed, "{") {
			parts := strings.Fields(trimmed)
			if len(parts) >= 3 {
				structName := parts[1]
				inStruct = true
				braceCount = 1

				currentStruct = &goStruct{
					ident:           packageName + "." + structName,
					structName:      structName,
					file:            path,
					packageName:     packageName,
					comment:         currentComment,
					alternativeName: alternativeName,
				}
				structs = append(structs, currentStruct)
			}
			currentComment = ""
			alternativeName = ""
			continue
		}

		if inStruct {
			if strings.Contains(trimmed, "{") {
				braceCount += strings.Count(trimmed, "{")
			}
			if strings.Contains(trimmed, "}") {
				braceCount -= strings.Count(trimmed, "}")
				if braceCount == 0 {
					inStruct = false
					currentStruct = nil
					currentComment = ""
					continue
				}
			}
			if trimmed != "" {
				raw := trimmed
				if currentComment != "" {
					raw = "COMMENT:" + currentComment + "|||" + raw
				}
				currentStruct.rawFields = append(currentStruct.rawFields, raw)
			}
			currentComment = ""
		} else {
			if trimmed == "" {
				currentComment = ""
			}
		}
	}

	// add to map only structs with json tag
	for _, goStruct := range structs {
		for _, field := range goStruct.rawFields {
			if strings.Contains(field, "json:") {
				structsMap[goStruct.ident] = goStruct
				continue
			}
		}
	}
}

func processRawFields(structsMap map[string]*goStruct) {
	for _, t := range structsMap {
		for _, raw := range t.rawFields {
			if strings.Contains(raw, "json:") {
				field := parseRawField(raw)
				t.fields = append(t.fields, field)
			}
		}
	}
}

func parseRawField(raw string) goField {
	var field goField

	if strings.HasPrefix(raw, "COMMENT:") {
		parts := strings.SplitN(raw, "|||", 2)
		if len(parts) == 2 {
			field.comment = strings.TrimPrefix(parts[0], "COMMENT:")
			raw = parts[1]
		}
	}

	inlineParts := strings.SplitN(raw, "//", 2)
	if len(inlineParts) == 2 {
		c := strings.TrimSpace(inlineParts[1])
		if field.comment != "" {
			field.comment += "\n" + c
		} else {
			field.comment = c
		}
		raw = strings.TrimSpace(inlineParts[0])
	}

	tokens := strings.Fields(raw)
	if len(tokens) >= 1 {
		field.goName = tokens[0]
	}
	if len(tokens) >= 2 {
		field.goType = tokens[1]
	}
	if len(tokens) >= 3 {
		field.tag = strings.Join(tokens[2:], " ")
	}

	return field
}

func buildTSTypes(structsMap map[string]*goStruct) []tsType {
	var tsTypes []tsType

	for _, goStr := range structsMap {
		tType := tsType{
			name:    formatTSName(goStr.packageName, goStr.structName),
			comment: goStr.comment,
		}
		if goStr.alternativeName != "" {
			tType.name = goStr.alternativeName
		}
		for _, f := range goStr.fields {
			tsF := tsField{
				tsName:        extractJSONName(f.tag),
				comment:       f.comment,
				inlineComment: f.tag,
			}

			if tsF.tsName == "" || tsF.tsName == "-" {
				continue
			}

			isSlice := strings.Contains(f.goType, "[]")

			baseType := strings.ReplaceAll(f.goType, "[]", "")
			baseType = strings.ReplaceAll(baseType, "*", "")

			var lookupIdent string
			if strings.Contains(baseType, ".") {
				lookupIdent = baseType
			} else {
				lookupIdent = goStr.packageName + "." + baseType
			}

			if targetStruct, exists := structsMap[lookupIdent]; exists {
				if targetStruct.alternativeName != "" {
					tsF.tsType = targetStruct.alternativeName
				} else {
					tsF.tsType = formatTSName(targetStruct.packageName, targetStruct.structName)
				}

			} else {
				tsF.tsType = mapBasicType(baseType)
			}

			if isSlice {
				tsF.tsType += "[]"
			}

			tType.fields = append(tType.fields, tsF)
		}
		if len(goStr.fields) > 0 && len(tType.fields) > 0 {
			tsTypes = append(tsTypes, tType)
		}

	}

	return tsTypes
}

func generateTSFile(tsTypes []tsType, outputPath string) {
	var sb strings.Builder

	for _, t := range tsTypes {
		fmt.Printf(`%s %s\n`, t.name, t.comment)
		if t.comment != "" {
			sb.WriteString("// ")
			sb.WriteString(t.comment)
			sb.WriteString("\n")
		}
		sb.WriteString("export type ")
		sb.WriteString(t.name)
		sb.WriteString(" = {\n")

		for _, f := range t.fields {
			if f.comment != "" {
				lines := strings.Split(f.comment, "\n")
				for _, l := range lines {
					sb.WriteString("\t// ")
					sb.WriteString(strings.TrimSpace(l))
					sb.WriteString("\n")
				}
			}
			sb.WriteString("\t")
			sb.WriteString(f.tsName)
			sb.WriteString(": ")
			sb.WriteString(f.tsType)
			sb.WriteString(";")
			if f.inlineComment != "" {
				sb.WriteString(" // ")
				sb.WriteString(f.inlineComment)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("}\n\n")
	}

	err := os.MkdirAll(filepath.Dir(outputPath), 0755)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = os.WriteFile(outputPath, []byte(sb.String()), 0644)
	if err != nil {
		fmt.Println(err)
	}
}

func extractJSONName(tag string) string {
	cleanTag := strings.TrimSpace(strings.ReplaceAll(tag, "`", ""))
	st := reflect.StructTag(cleanTag)
	jsonTag := st.Get("json")
	if jsonTag == "" {
		return ""
	}
	return strings.Split(jsonTag, ",")[0]
}

func formatTSName(pkg string, str string) string {
	return capitalize(pkg) + capitalize(str)
}

func capitalize(s string) string {
	if len(s) == 0 {
		return ""
	}
	r := []rune(s)
	r[0] = unicode.ToUpper(r[0])
	return string(r)
}

func mapBasicType(goType string) string {
	switch goType {
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64":
		return "number"
	case "bool":
		return "boolean"
	case "string":
		return "string"
	case "time.Time":
		return "string"
	case "interface{}":
		return "any"
	default:
		return "any"
	}
}

func Exec(inputDir string, outputFile string, flags ...string) {
	structsMap := make(map[string]*goStruct)

	err := filepath.Walk(inputDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(info.Name(), ".go") {
			return nil
		}
		if strings.HasSuffix(info.Name(), "_test.go") {
			return nil
		}
		parseFile(path, structsMap)
		return nil
	})
	if err != nil {
		os.Exit(1)
	}
	processRawFields(structsMap)
	fmt.Println(outputFile)
	for _, goStruct := range structsMap {
		fmt.Println(goStruct.ident, goStruct.comment)
	}
	tsTypes := buildTSTypes(structsMap)
	generateTSFile(tsTypes, outputFile)
}
func Run() {
	if len(os.Args) < 3 {
		os.Exit(1)
	}

	inputDir := os.Args[1]
	outputFile := os.Args[2]
	Exec(inputDir, outputFile)
}
