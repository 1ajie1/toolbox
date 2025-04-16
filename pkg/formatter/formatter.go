package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/fatih/color"
	"github.com/tidwall/gjson"
	"github.com/tidwall/pretty"
	"gopkg.in/yaml.v3"
)

// FormatType 表示支持的格式化类型
type FormatType string

// 支持的格式化类型
const (
	FormatJSON FormatType = "json"
	FormatXML  FormatType = "xml"
	FormatYAML FormatType = "yaml"
)

// Options 格式化选项
type Options struct {
	Format  FormatType // 格式类型
	Pretty  bool       // 是否美化输出
	Indent  int        // 缩进数量
	Compact bool       // 是否压缩输出
	Color   bool       // 是否彩色输出
}

// 默认缩进值
const (
	DefaultJSONIndent = 4
	DefaultXMLIndent  = 4
	DefaultYAMLIndent = 2
)

// 获取格式对应的默认缩进值
func (o Options) GetIndent() int {
	// 如果已设置缩进值，则使用该值
	if o.Indent > 0 {
		return o.Indent
	}

	// 根据格式返回默认缩进值
	switch o.Format {
	case FormatJSON:
		return DefaultJSONIndent
	case FormatXML:
		return DefaultXMLIndent
	case FormatYAML:
		return DefaultYAMLIndent
	default:
		return 2 // 通用默认值
	}
}

// Result 格式化结果
type Result struct {
	Output      string        // 格式化后的输出
	InputSize   int64         // 输入大小
	OutputSize  int64         // 输出大小
	Duration    time.Duration // 处理耗时
	ContentType string        // 内容类型
}

// Format 根据选项格式化数据
func Format(input io.Reader, opts Options) (*Result, error) {
	startTime := time.Now()

	// 读取输入数据
	data, err := ioutil.ReadAll(input)
	if err != nil {
		return nil, fmt.Errorf("读取输入失败: %v", err)
	}

	inputSize := int64(len(data))

	// 如果是JSON格式，尝试处理和修复
	if opts.Format == FormatJSON {
		// 尝试修复JSON
		content := string(data)
		fixedJSON, isFixed := TryFixJSON(content)
		if isFixed {
			// 使用修复后的JSON
			data = []byte(fixedJSON)
		}
	}

	// 根据格式进行处理
	var output []byte
	var contentType string

	switch opts.Format {
	case FormatJSON:
		contentType = "application/json"

		// 确保输入是有效的JSON
		var jsonObj interface{}
		if err := json.Unmarshal(data, &jsonObj); err != nil {
			return nil, fmt.Errorf("解析JSON失败: %v", err)
		}

		if opts.Pretty {
			// 美化JSON
			jsonData, err := json.MarshalIndent(jsonObj, "", strings.Repeat(" ", opts.GetIndent()))
			if err != nil {
				return nil, fmt.Errorf("生成美化JSON失败: %v", err)
			}

			if opts.Color {
				output = pretty.Color(jsonData, nil)
			} else {
				output = jsonData
			}
		} else if opts.Compact {
			// 压缩JSON
			var buf bytes.Buffer
			if err := json.Compact(&buf, data); err != nil {
				return nil, fmt.Errorf("压缩JSON失败: %v", err)
			}
			output = buf.Bytes()
		} else {
			// 使用原始数据
			output = data
		}

	case FormatXML:
		contentType = "application/xml"

		// 使用etree库解析和格式化XML
		doc := etree.NewDocument()
		doc.ReadSettings.CharsetReader = nil
		err := doc.ReadFromBytes(data)
		if err != nil {
			return nil, fmt.Errorf("解析XML失败: %v", err)
		}

		if opts.Pretty {
			// 美化XML，设置缩进
			settings := etree.NewIndentSettings()
			settings.Spaces = opts.GetIndent()
			doc.IndentWithSettings(settings)
			xmlBytes, err := doc.WriteToBytes()
			if err != nil {
				return nil, fmt.Errorf("美化XML失败: %v", err)
			}

			if opts.Color {
				// 为XML添加颜色
				coloredXML := colorizeXML(string(xmlBytes))
				output = []byte(coloredXML)
			} else {
				output = xmlBytes
			}
		} else if opts.Compact {
			// 压缩XML - 不使用缩进
			doc.Indent(0) // 不缩进
			xmlBytes, err := doc.WriteToBytes()
			if err != nil {
				return nil, fmt.Errorf("压缩XML失败: %v", err)
			}
			// 去除额外的空白
			xmlStr := string(xmlBytes)
			xmlStr = strings.ReplaceAll(xmlStr, "\n", "")
			xmlStr = strings.ReplaceAll(xmlStr, "\r", "")
			output = []byte(xmlStr)
		} else {
			// 默认格式化
			indentValue := opts.GetIndent()
			doc.Indent(indentValue) // 使用格式对应的默认缩进
			xmlBytes, err := doc.WriteToBytes()
			if err != nil {
				return nil, fmt.Errorf("格式化XML失败: %v", err)
			}

			if opts.Color {
				// 为XML添加颜色
				coloredXML := colorizeXML(string(xmlBytes))
				output = []byte(coloredXML)
			} else {
				output = xmlBytes
			}
		}

	case FormatYAML:
		contentType = "application/yaml"

		// 检查YAML是否有效
		var yamlObj interface{}
		if err := yaml.Unmarshal(data, &yamlObj); err != nil {
			return nil, fmt.Errorf("解析YAML失败: %v", err)
		}

		// 创建编码器，设置缩进
		var buf bytes.Buffer
		encoder := yaml.NewEncoder(&buf)
		encoder.SetIndent(opts.GetIndent()) // 使用格式对应的默认缩进

		// 将数据编码为YAML
		if err := encoder.Encode(yamlObj); err != nil {
			return nil, fmt.Errorf("生成YAML失败: %v", err)
		}
		encoder.Close()

		// 获取格式化后的YAML
		yamlData := buf.Bytes()

		if opts.Color && opts.Pretty {
			// 为YAML添加颜色
			coloredYAML := colorizeYAML(string(yamlData))
			output = []byte(coloredYAML)
		} else {
			output = yamlData
		}

	default:
		return nil, fmt.Errorf("不支持的格式: %s", opts.Format)
	}

	duration := time.Since(startTime)

	// 生成结果
	result := &Result{
		Output:      string(output),
		InputSize:   inputSize,
		OutputSize:  int64(len(output)),
		Duration:    duration,
		ContentType: contentType,
	}

	return result, nil
}

// colorizeXML 为XML添加ANSI颜色
func colorizeXML(xml string) string {
	// 创建彩色对象
	tagColor := color.New(color.FgCyan).SprintFunc()
	attrNameColor := color.New(color.FgYellow).SprintFunc()
	attrValueColor := color.New(color.FgGreen).SprintFunc()

	// 正则表达式匹配XML的不同部分
	tagRegex := regexp.MustCompile(`</?[^>\s]+`)
	attrNameRegex := regexp.MustCompile(`\s([a-zA-Z0-9_:-]+)=`)
	attrValueRegex := regexp.MustCompile(`="([^"]*)"`)

	// 添加颜色
	coloredXML := xml

	// 为标签添加颜色
	coloredXML = tagRegex.ReplaceAllStringFunc(coloredXML, func(tag string) string {
		return tagColor(tag)
	})

	// 为属性名添加颜色
	coloredXML = attrNameRegex.ReplaceAllStringFunc(coloredXML, func(attr string) string {
		// 仅对属性名着色
		name := strings.TrimSpace(strings.Split(attr, "=")[0])
		return " " + attrNameColor(name) + "="
	})

	// 为属性值添加颜色
	coloredXML = attrValueRegex.ReplaceAllStringFunc(coloredXML, func(value string) string {
		parts := strings.SplitN(value, "\"", 3)
		if len(parts) >= 3 {
			return "=\"" + attrValueColor(parts[1]) + "\""
		}
		return value
	})

	return coloredXML
}

// colorizeYAML 为YAML添加ANSI颜色
func colorizeYAML(yamlStr string) string {
	// 创建彩色对象
	keyColor := color.New(color.FgCyan).SprintFunc()
	valueColor := color.New(color.FgGreen).SprintFunc()
	dashColor := color.New(color.FgYellow).SprintFunc()

	// 正则表达式匹配YAML的不同部分
	keyRegex := regexp.MustCompile(`^(\s*)([^:\n-]+):`)
	valueRegex := regexp.MustCompile(`: (.+)$`)
	dashRegex := regexp.MustCompile(`^(\s*)(- )`)

	// 分割行并处理
	lines := strings.Split(yamlStr, "\n")
	for i, line := range lines {
		// 为键添加颜色
		line = keyRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := keyRegex.FindStringSubmatch(match)
			if len(parts) >= 3 {
				return parts[1] + keyColor(parts[2]) + ":"
			}
			return match
		})

		// 为值添加颜色
		line = valueRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := valueRegex.FindStringSubmatch(match)
			if len(parts) >= 2 {
				return ": " + valueColor(parts[1])
			}
			return match
		})

		// 为破折号添加颜色
		line = dashRegex.ReplaceAllStringFunc(line, func(match string) string {
			parts := dashRegex.FindStringSubmatch(match)
			if len(parts) >= 3 {
				return parts[1] + dashColor(parts[2])
			}
			return match
		})

		lines[i] = line
	}

	return strings.Join(lines, "\n")
}

// ToFile 将结果保存到文件
func (r *Result) ToFile(path string) error {
	return ioutil.WriteFile(path, []byte(r.Output), 0644)
}

// FormatFile 格式化文件内容
func FormatFile(path string, opts Options) (*Result, error) {
	file, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取文件失败: %v", err)
	}

	return Format(bytes.NewReader(file), opts)
}

// HandlePowerShellEscaping 处理 PowerShell 的转义字符问题
func HandlePowerShellEscaping(content string) string {
	// 检测常见的 PowerShell 转义模式

	// 处理 PS 的反引号转义 (例如: `")
	content = strings.ReplaceAll(content, "`\"", "\"")
	content = strings.ReplaceAll(content, "`n", "\n")
	content = strings.ReplaceAll(content, "`r", "\r")
	content = strings.ReplaceAll(content, "`t", "\t")

	// 处理双引号内部的转义引号 (PowerShell 中嵌套引号的方式)
	if strings.HasPrefix(content, "\"") && strings.HasSuffix(content, "\"") {
		// 去除首尾引号
		unquoted := content[1 : len(content)-1]
		// 如果是 Windows PowerShell 通常的转义模式 (`")，已经在上面处理过了
		// 如果存在 "" 连续引号（另一种转义方式），转换为单个引号
		unquoted = strings.ReplaceAll(unquoted, "\"\"", "\"")
		content = unquoted
	}

	// 如果内容是一个格式不完整的 JSON 字符串，可能是未正确引用的字符串
	// 例如 {name:"value"} 这种格式，尝试修复成标准 JSON
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") &&
		!strings.Contains(content, "\"") && strings.Contains(content, ":") {
		// 这可能是一个没有引号的键值对
		content = strings.ReplaceAll(content, "'", "\"")
	}

	return content
}

// IsNumeric 检查字符串是否为数字
func IsNumeric(s string) bool {
	// 检查是否为整数
	if _, err := strconv.ParseInt(s, 10, 64); err == nil {
		return true
	}
	// 检查是否为浮点数
	if _, err := strconv.ParseFloat(s, 64); err == nil {
		return true
	}
	return false
}

// TryFixJSON 尝试修复无效的 JSON 字符串
func TryFixJSON(content string) (string, bool) {
	// 如果内容已经是有效的 JSON，直接返回
	if gjson.Valid(content) {
		return content, true
	}

	// 如果内容被单引号包裹，则去除这些引号
	if strings.HasPrefix(content, "'") && strings.HasSuffix(content, "'") {
		content = content[1 : len(content)-1]
		// 再次检查是否有效
		if gjson.Valid(content) {
			return content, true
		}
	}

	// 检查引号内容
	if strings.Count(content, "\"") == 0 && strings.Count(content, "'") > 0 {
		// 可能使用了单引号而不是双引号，尝试替换
		content = strings.ReplaceAll(content, "'", "\"")
		if gjson.Valid(content) {
			return content, true
		}
	}

	// 解析并重新格式化JSON
	if strings.HasPrefix(content, "{") && strings.HasSuffix(content, "}") {
		// 对象处理
		parsed := parseJSONObject(content)
		if gjson.Valid(parsed) {
			return parsed, true
		}
	} else if strings.HasPrefix(content, "[") && strings.HasSuffix(content, "]") {
		// 数组处理
		parsed := parseJSONArray(content)
		if gjson.Valid(parsed) {
			return parsed, true
		}
	}

	return content, false
}

// parseJSONObject 解析 JSON 对象字符串
func parseJSONObject(content string) string {
	// 移除空格并创建一个空的JSON对象
	result := make(map[string]interface{})

	// 去除首尾的花括号
	trimContent := strings.TrimSpace(content)
	if len(trimContent) < 2 {
		return "{}"
	}
	trimContent = trimContent[1 : len(trimContent)-1]

	// 拆分为键值对
	pairs := splitJSONPairs(trimContent)

	// 处理每一个键值对
	for _, pair := range pairs {
		// 分割键和值
		parts := strings.SplitN(pair, ":", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 处理键
		if strings.HasPrefix(key, "\"") && strings.HasSuffix(key, "\"") {
			// 已有引号
			key = key[1 : len(key)-1]
		} else if strings.HasPrefix(key, "'") && strings.HasSuffix(key, "'") {
			// 单引号
			key = key[1 : len(key)-1]
		}

		// 处理值
		var parsedValue interface{}

		if value == "null" {
			parsedValue = nil
		} else if value == "true" {
			parsedValue = true
		} else if value == "false" {
			parsedValue = false
		} else if IsNumeric(value) {
			// 数字
			if strings.Contains(value, ".") {
				// 浮点数
				if floatVal, err := strconv.ParseFloat(value, 64); err == nil {
					parsedValue = floatVal
				} else {
					parsedValue = value
				}
			} else {
				// 整数
				if intVal, err := strconv.ParseInt(value, 10, 64); err == nil {
					parsedValue = intVal
				} else {
					parsedValue = value
				}
			}
		} else if strings.HasPrefix(value, "\"") && strings.HasSuffix(value, "\"") {
			// 双引号字符串
			parsedValue = value[1 : len(value)-1]
		} else if strings.HasPrefix(value, "'") && strings.HasSuffix(value, "'") {
			// 单引号字符串
			parsedValue = value[1 : len(value)-1]
		} else if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
			// 数组 - 解析为JSON并反序列化为实际的数组
			parsedArrayJSON := parseJSONArray(value)
			var array []interface{}
			if err := json.Unmarshal([]byte(parsedArrayJSON), &array); err == nil {
				parsedValue = array
			} else {
				// 如果反序列化失败，使用原始内容
				parsedValue = value
			}
		} else if strings.HasPrefix(value, "{") && strings.HasSuffix(value, "}") {
			// 对象 - 解析为JSON并反序列化为实际的对象
			parsedObjJSON := parseJSONObject(value)
			var obj map[string]interface{}
			if err := json.Unmarshal([]byte(parsedObjJSON), &obj); err == nil {
				parsedValue = obj
			} else {
				// 如果反序列化失败，使用原始内容
				parsedValue = value
			}
		} else {
			// 其他情况，假设为字符串
			parsedValue = value
		}

		// 添加到结果对象
		result[key] = parsedValue
	}

	// 将结果转为JSON字符串
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "{}"
	}

	return string(jsonBytes)
}

// parseJSONArray 解析 JSON 数组字符串
func parseJSONArray(content string) string {
	// 去除首尾的方括号
	trimContent := strings.TrimSpace(content)
	if len(trimContent) < 2 {
		return "[]"
	}
	trimContent = trimContent[1 : len(trimContent)-1]

	// 如果数组为空，直接返回
	if trimContent == "" {
		return "[]"
	}

	// 分割数组元素
	elements := splitJSONArray(trimContent)

	// 创建结果数组
	result := make([]interface{}, 0, len(elements))

	// 处理每个元素
	for _, elem := range elements {
		elem = strings.TrimSpace(elem)

		if elem == "" {
			continue
		} else if elem == "null" {
			result = append(result, nil)
		} else if elem == "true" {
			result = append(result, true)
		} else if elem == "false" {
			result = append(result, false)
		} else if IsNumeric(elem) {
			// 数字
			if strings.Contains(elem, ".") {
				if floatVal, err := strconv.ParseFloat(elem, 64); err == nil {
					result = append(result, floatVal)
				} else {
					result = append(result, elem)
				}
			} else {
				if intVal, err := strconv.ParseInt(elem, 10, 64); err == nil {
					result = append(result, intVal)
				} else {
					result = append(result, elem)
				}
			}
		} else if strings.HasPrefix(elem, "\"") && strings.HasSuffix(elem, "\"") {
			// 双引号字符串
			result = append(result, elem[1:len(elem)-1])
		} else if strings.HasPrefix(elem, "'") && strings.HasSuffix(elem, "'") {
			// 单引号字符串
			result = append(result, elem[1:len(elem)-1])
		} else if strings.HasPrefix(elem, "[") && strings.HasSuffix(elem, "]") {
			// 嵌套数组 - 解析为JSON并反序列化为实际的数组
			parsedArrayJSON := parseJSONArray(elem)
			var nestedArray []interface{}
			if err := json.Unmarshal([]byte(parsedArrayJSON), &nestedArray); err == nil {
				result = append(result, nestedArray)
			} else {
				// 如果反序列化失败，使用原始字符串
				result = append(result, elem)
			}
		} else if strings.HasPrefix(elem, "{") && strings.HasSuffix(elem, "}") {
			// 嵌套对象 - 解析为JSON并反序列化为实际的对象
			parsedObjJSON := parseJSONObject(elem)
			var nestedObj map[string]interface{}
			if err := json.Unmarshal([]byte(parsedObjJSON), &nestedObj); err == nil {
				result = append(result, nestedObj)
			} else {
				// 如果反序列化失败，使用原始字符串
				result = append(result, elem)
			}
		} else {
			// 默认为未引用的字符串
			result = append(result, elem)
		}
	}

	// 将结果转为JSON字符串
	jsonBytes, err := json.Marshal(result)
	if err != nil {
		return "[]"
	}

	return string(jsonBytes)
}

// splitJSONPairs 拆分 JSON 对象的键值对
func splitJSONPairs(content string) []string {
	var result []string
	var buffer strings.Builder
	var inQuotes bool
	var quoteChar rune
	var bracketCount, braceCount int

	for _, char := range content {
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			}
			buffer.WriteRune(char)
		case '{':
			braceCount++
			buffer.WriteRune(char)
		case '}':
			braceCount--
			buffer.WriteRune(char)
		case '[':
			bracketCount++
			buffer.WriteRune(char)
		case ']':
			bracketCount--
			buffer.WriteRune(char)
		case ',':
			if !inQuotes && bracketCount == 0 && braceCount == 0 {
				result = append(result, buffer.String())
				buffer.Reset()
			} else {
				buffer.WriteRune(char)
			}
		default:
			buffer.WriteRune(char)
		}
	}

	// 添加最后一个部分
	if buffer.Len() > 0 {
		result = append(result, buffer.String())
	}

	return result
}

// splitJSONArray 拆分 JSON 数组的元素
func splitJSONArray(content string) []string {
	var result []string
	var buffer strings.Builder
	var inQuotes bool
	var quoteChar rune
	var bracketCount, braceCount int

	for _, char := range content {
		switch char {
		case '"', '\'':
			if !inQuotes {
				inQuotes = true
				quoteChar = char
			} else if char == quoteChar {
				inQuotes = false
			}
			buffer.WriteRune(char)
		case '{':
			braceCount++
			buffer.WriteRune(char)
		case '}':
			braceCount--
			buffer.WriteRune(char)
		case '[':
			bracketCount++
			buffer.WriteRune(char)
		case ']':
			bracketCount--
			buffer.WriteRune(char)
		case ',':
			if !inQuotes && bracketCount == 0 && braceCount == 0 {
				result = append(result, buffer.String())
				buffer.Reset()
			} else {
				buffer.WriteRune(char)
			}
		default:
			buffer.WriteRune(char)
		}
	}

	// 添加最后一个部分
	if buffer.Len() > 0 {
		result = append(result, buffer.String())
	}

	return result
}

// ExtractContentWithDelimiter 从文本中提取被特定分隔符包围的内容
// 例如：从 #{"name":"value"}# 中提取出 {"name":"value"}
func ExtractContentWithDelimiter(content string, delimiter string) (string, bool) {
	if delimiter == "" {
		return content, false
	}

	// 检查内容是否被分隔符包围
	if strings.HasPrefix(content, delimiter) && strings.HasSuffix(content, delimiter) {
		// 去除首尾的分隔符
		extracted := content[len(delimiter) : len(content)-len(delimiter)]
		return extracted, true
	}

	return content, false
}
