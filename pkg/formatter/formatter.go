package formatter

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strings"
	"time"

	"github.com/beevik/etree"
	"github.com/fatih/color"
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
