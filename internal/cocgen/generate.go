// Package cocgen 提供 Clash of Clans API 客户端代码生成能力。
//
// 从官方 Swagger 2.0 规范生成 spec.yaml / types.go / api.go。
// 权威来源:pkg/cocapi/official-swagger.yaml(从 https://developer.clashofclans.com 下载)
//
// 通过 cobra 子命令调用:warspark cocapi generate
// 或独立运行:go run scripts/generate-coc-api.go
package cocgen

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ============================================================
// Swagger 2.0 规范数据结构(只解析用到的字段)
// ============================================================

type Swagger struct {
	Swagger     string              `yaml:"swagger"`
	Paths       map[string]PathItem `yaml:"paths"`
	Definitions map[string]Schema   `yaml:"definitions"`
}

type PathItem struct {
	Get  *Operation `yaml:"get"`
	Post *Operation `yaml:"post"`
}

type Operation struct {
	Summary     string              `yaml:"summary"`
	Description string              `yaml:"description"`
	OperationID string              `yaml:"operationId"`
	Tags        []string            `yaml:"tags"`
	Parameters  []Parameter         `yaml:"parameters"`
	Responses   map[string]Response `yaml:"responses"`
}

type Parameter struct {
	Name     string  `yaml:"name"`
	In       string  `yaml:"in"` // path / query / body
	Required bool    `yaml:"required"`
	Type     string  `yaml:"type"`   // string / integer / boolean
	Schema   *Schema `yaml:"schema"` // for in: body
}

type Response struct {
	Description string  `yaml:"description"`
	Schema      *Schema `yaml:"schema"`
}

type Schema struct {
	Ref        string            `yaml:"$ref"`
	Type       string            `yaml:"type"`       // object / array / string / integer / boolean
	Items      *Schema           `yaml:"items"`      // for type: array
	Properties map[string]Schema `yaml:"properties"` // for type: object
	Enum       []string          `yaml:"enum"`
}

// ============================================================
// 生成用的中间结构
// ============================================================

type GenEndpoint struct {
	GoName      string // SearchClans
	OperationID string // searchClans
	Method      string // GET / POST
	Path        string // /clans/{clanTag}
	PathParams  []GenParam
	QueryParams []GenParam
	BodyType    string // 请求 body 的 Go 类型(verifytoken)
	ReturnType  string // 响应 Go 类型
	Summary     string
	Description string
	Tag         string
}

type GenParam struct {
	Name   string // clanTag
	GoType string // string / int
	IsTag  bool   // 是否是 *Tag(path 参数名含 tag)
}

// ============================================================
// 主流程
// ============================================================

// Generate 从 swaggerPath 读取官方规范,在 dir 目录下生成 spec.yaml/types.go/api.go。
func Generate(swaggerPath, dir string) (endpoints, defs int, err error) {
	data, err := os.ReadFile(swaggerPath)
	if err != nil {
		return 0, 0, fmt.Errorf("read swagger: %w", err)
	}

	var sw Swagger
	if err := yaml.Unmarshal(data, &sw); err != nil {
		return 0, 0, fmt.Errorf("parse swagger: %w", err)
	}

	eps := collectEndpoints(&sw)
	defNames := orderDefinitions(&sw)

	if err := genSpecYAML(filepath.Join(dir, "spec.yaml"), eps); err != nil {
		return 0, 0, err
	}
	if err := genTypesGo(filepath.Join(dir, "types.go"), &sw, defNames); err != nil {
		return 0, 0, err
	}
	if err := genAPIGo(filepath.Join(dir, "api.go"), eps); err != nil {
		return 0, 0, err
	}
	return len(eps), len(defNames), nil
}

// ============================================================
// 端点收集
// ============================================================

func collectEndpoints(sw *Swagger) []GenEndpoint {
	var eps []GenEndpoint
	for path, item := range sw.Paths {
		if item.Get != nil {
			eps = append(eps, buildEndpoint(path, "GET", item.Get, sw))
		}
		if item.Post != nil {
			eps = append(eps, buildEndpoint(path, "POST", item.Post, sw))
		}
	}
	// 按 path 排序,输出稳定
	sort.Slice(eps, func(i, j int) bool {
		if eps[i].Path != eps[j].Path {
			return eps[i].Path < eps[j].Path
		}
		return eps[i].Method < eps[j].Method
	})
	// 生成稳定的 Go 方法名(用 operationId 转 PascalCase)
	for i := range eps {
		eps[i].GoName = pascalCase(eps[i].OperationID)
	}
	return eps
}

func buildEndpoint(path, method string, op *Operation, sw *Swagger) GenEndpoint {
	ep := GenEndpoint{
		Method:      method,
		Path:        path,
		OperationID: op.OperationID,
		Summary:     op.Summary,
		Description: cleanDesc(op.Description),
		Tag:         firstOr(op.Tags, ""),
	}
	for _, p := range op.Parameters {
		switch p.In {
		case "path":
			ep.PathParams = append(ep.PathParams, GenParam{
				Name:   p.Name,
				GoType: swaggerToGoType(p.Type),
				IsTag:  strings.Contains(strings.ToLower(p.Name), "tag"),
			})
		case "query":
			ep.QueryParams = append(ep.QueryParams, GenParam{
				Name:   p.Name,
				GoType: swaggerToGoType(p.Type),
			})
		case "body":
			if p.Schema != nil && p.Schema.Ref != "" {
				ep.BodyType = refToGoName(p.Schema.Ref)
			}
		}
	}
	// 响应类型:取 200 的 schema
	if r, ok := op.Responses["200"]; ok && r.Schema != nil {
		// 如果响应是 $ref 指向一个 array 定义,实际 API 返回 {items, paging} 包装对象。
		// 用 XResponse 包装类型,避免与被引用为字段的 array alias 冲突。
		if r.Schema.Ref != "" {
			refName := refToGoName(r.Schema.Ref)
			if def, ok := sw.Definitions[refName]; ok && def.Type == "array" {
				ep.ReturnType = refName + "Response"
			} else {
				ep.ReturnType = refName
			}
		} else {
			ep.ReturnType = schemaToGoType(r.Schema, sw)
		}
	}
	return ep
}

// schemaToGoType 把 swagger schema 转成 Go 类型名。
// 对于 $ref 返回引用的类型名;对于 array 返回 "[]" + 元素类型。
func schemaToGoType(s *Schema, sw *Swagger) string {
	if s == nil {
		return ""
	}
	if s.Ref != "" {
		return refToGoName(s.Ref)
	}
	switch s.Type {
	case "array":
		if s.Items != nil {
			return "[]" + schemaToGoType(s.Items, sw)
		}
		return "[]any"
	case "string":
		return "string"
	case "integer":
		return "int"
	case "boolean":
		return "bool"
	default:
		return "any"
	}
}

// swaggerToGoType 把 swagger 基础类型转 Go 类型。
func swaggerToGoType(t string) string {
	switch t {
	case "string":
		return "string"
	case "integer":
		return "int"
	case "boolean":
		return "bool"
	default:
		return "string"
	}
}

// refToGoName 把 "#/definitions/ClanWar" 转成 "ClanWar"。
func refToGoName(ref string) string {
	const prefix = "#/definitions/"
	if strings.HasPrefix(ref, prefix) {
		return ref[len(prefix):]
	}
	return ref
}

// pascalCase 把 camelCase 转成 PascalCase,并对常见缩写做全大写处理。
// 例: "searchClans" -> "SearchClans", "clanId" -> "ClanID", "badgeUrls" -> "BadgeUrls"。
// 注意:"urls" 不在缩写表里(因为是复数),只有 "url" 单数才大写。
func pascalCase(s string) string {
	if s == "" {
		return s
	}
	// 整词匹配缩写
	if upper, ok := pascalAcronyms[strings.ToLower(s)]; ok {
		return upper
	}
	// 按驼峰边界拆分:大写字母前断开,数字前断开
	var parts []string
	start := 0
	for i := 1; i < len(s); i++ {
		if s[i] >= 'A' && s[i] <= 'Z' {
			parts = append(parts, s[start:i])
			start = i
		}
	}
	parts = append(parts, s[start:])
	var b strings.Builder
	for _, p := range parts {
		if upper, ok := pascalAcronyms[strings.ToLower(p)]; ok {
			b.WriteString(upper)
		} else if p != "" {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	result := b.String()
	if result == "" {
		result = strings.ToUpper(s[:1]) + s[1:]
	}
	return result
}

var pascalAcronyms = map[string]string{
	"id":    "ID",
	"url":   "URL",
	"urls":  "URLs",
	"api":   "API",
	"ip":    "IP",
	"http":  "HTTP",
	"https": "HTTPS",
	"ttl":   "TTL",
	"jwt":   "JWT",
}

func firstOr(s []string, def string) string {
	if len(s) > 0 {
		return s[0]
	}
	return def
}

func cleanDesc(s string) string {
	// 去掉多行描述里的换行和多余空格,单行化
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "  ", " ")
	return s
}

// ============================================================
// definition 排序(按名字字母序,稳定输出)
// ============================================================

func orderDefinitions(sw *Swagger) []string {
	var names []string
	for name := range sw.Definitions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// ============================================================
// 生成 spec.yaml
// ============================================================

func genSpecYAML(path string, eps []GenEndpoint) error {
	var b strings.Builder
	b.WriteString("# Clash of Clans API 端点清单(由 generate-coc-api.go 从 official-swagger.yaml 生成)\n")
	b.WriteString("# 权威来源: pkg/cocapi/official-swagger.yaml — 请勿手动编辑本文件\n\n")
	b.WriteString("endpoints:\n")
	for _, e := range eps {
		fmt.Fprintf(&b, "  - name: %s\n", e.GoName)
		fmt.Fprintf(&b, "    method: %s\n", e.Method)
		fmt.Fprintf(&b, "    path: %s\n", e.Path)
		if len(e.PathParams) > 0 || len(e.QueryParams) > 0 || e.BodyType != "" {
			b.WriteString("    params:\n")
			for _, p := range e.PathParams {
				fmt.Fprintf(&b, "      - {name: %s, in: path, type: %s}\n", p.Name, goTypeToYaml(p.GoType))
			}
			for _, p := range e.QueryParams {
				fmt.Fprintf(&b, "      - {name: %s, in: query, type: %s}\n", p.Name, goTypeToYaml(p.GoType))
			}
			if e.BodyType != "" {
				fmt.Fprintf(&b, "      - {name: body, in: body, type: %s}\n", e.BodyType)
			}
		}
		fmt.Fprintf(&b, "    response_type: %s\n", e.ReturnType)
		fmt.Fprintf(&b, "    description: %q\n", e.Summary)
	}
	return os.WriteFile(path, []byte(b.String()), 0o644)
}

func goTypeToYaml(t string) string {
	switch t {
	case "int":
		return "int"
	case "bool":
		return "bool"
	default:
		return "string"
	}
}

// ============================================================
// 生成 types.go
// ============================================================

func genTypesGo(path string, sw *Swagger, defs []string) error {
	var b strings.Builder
	b.WriteString("// Code generated by scripts/generate-coc-api.go from official-swagger.yaml; DO NOT EDIT.\n")
	b.WriteString("// 权威来源: pkg/cocapi/official-swagger.yaml (官方 Swagger 2.0 规范)\n")
	b.WriteString("// 重新生成: go run scripts/generate-coc-api.go\n\n")
	b.WriteString("package cocapi\n\n")
	// 导入 encoding/json 和 fmt:覆盖类型(JsonLocalizedName、Long)的 UnmarshalJSON 需要。
	b.WriteString("import (\n\t\"encoding/json\"\n\t\"fmt\"\n)\n\n")
	// Paging 是列表响应的通用分页结构,官方规范未单独定义,这里手写注入。
	b.WriteString("// Paging 列表响应的分页信息。官方规范里 list 端点响应被简化为 array,\n")
	b.WriteString("// 但实际响应是 {items, paging},故在生成的 List struct 里统一引用本类型。\n")
	b.WriteString("type Paging struct {\n")
	b.WriteString("\tCursors Cursors `json:\"cursors,omitempty\"`\n")
	b.WriteString("}\n\n")
	b.WriteString("// Cursors 分页游标。\n")
	b.WriteString("type Cursors struct {\n")
	b.WriteString("\tAfter  string `json:\"after,omitempty\"`\n")
	b.WriteString("\tBefore string `json:\"before,omitempty\"`\n")
	b.WriteString("}\n\n")

	// 收集端点返回类型,判断哪些 definition 被端点直接作为响应
	eps := collectEndpoints(sw)
	endpointReturnTypes := map[string]bool{}
	// arrayResponseWrappers 记录需要生成 XResponse 包装类型的 array definition 名
	arrayResponseWrappers := map[string]string{} // refName -> item Go type
	for _, e := range eps {
		endpointReturnTypes[e.ReturnType] = true
		// 如果 ReturnType 是 XResponse 形式,说明对应的 array definition 需要包装
		if strings.HasSuffix(e.ReturnType, "Response") {
			baseName := strings.TrimSuffix(e.ReturnType, "Response")
			if def, ok := sw.Definitions[baseName]; ok && def.Type == "array" && def.Items != nil {
				itemType := schemaToGoType(def.Items, sw)
				arrayResponseWrappers[baseName] = itemType
			}
		}
	}

	for _, name := range defs {
		schema := sw.Definitions[name]
		genDefinition(&b, name, schema, sw, endpointReturnTypes[name])
	}

	// 生成 array 响应的 {Items, Paging} 包装类型
	for baseName, itemType := range arrayResponseWrappers {
		fmt.Fprintf(&b, "// %sResponse 是 %s 端点响应的包装(带分页)。\n", baseName, baseName)
		fmt.Fprintf(&b, "// 官方规范把 %s 定义为 array,但实际响应是 {items, paging} 对象。\n", baseName)
		fmt.Fprintf(&b, "type %sResponse struct {\n\tItems []%s `json:\"items,omitempty\"`\n\tPaging Paging `json:\"paging,omitempty\"`\n}\n\n",
			baseName, itemType)
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// definitionOverrides 对官方规范中定义不准确的类型做特殊处理。
// key 是 definition 名,value 是要生成的 Go 代码(不含前导注释)。
// 这些覆盖基于实际 API 调用验证,弥补官方 swagger 与运行时的偏差。
var definitionOverrides = map[string]string{
	// JsonLocalizedName 官方定义为空 object,但实际 API 返回纯字符串(英文名)。
	// 用自定义类型兼容 string 和 object 两种情况。
	"JsonLocalizedName": `type JsonLocalizedName string

// UnmarshalJSON 兼容官方规范(object)与实际响应(string)。
func (j *JsonLocalizedName) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		*j = JsonLocalizedName(s)
		return nil
	}
	// object 形式:取 en 字段或首个字符串值
	var m map[string]any
	if err := json.Unmarshal(data, &m); err == nil {
		if v, ok := m["en"].(string); ok {
			*j = JsonLocalizedName(v)
			return nil
		}
		for _, v := range m {
			if s, ok := v.(string); ok {
				*j = JsonLocalizedName(s)
				return nil
			}
		}
	}
	*j = ""
	return nil
}`,
	// Long 官方定义为空 object,但实际 API 返回 64 位整数(赛季 ID 等)。
	"Long": `type Long int64

// UnmarshalJSON 兼容数字和字符串形式的数字。
func (l *Long) UnmarshalJSON(data []byte) error {
	var n int64
	if err := json.Unmarshal(data, &n); err == nil {
		*l = Long(n)
		return nil
	}
	var s string
	if err := json.Unmarshal(data, &s); err == nil {
		var n int64
		if _, err := fmt.Sscanf(s, "%d", &n); err == nil {
			*l = Long(n)
			return nil
		}
	}
	*l = 0
	return nil
}`,
	// Float 官方定义为空 object,但实际 API 返回裸数字(如摧毁百分比)。
	// 用自定义类型兼容 number 和 object 两种情况。
	"Float": `type Float float64

// UnmarshalJSON 兼容裸数字和 object 形式。
func (f *Float) UnmarshalJSON(data []byte) error {
	var n float64
	if err := json.Unmarshal(data, &n); err == nil {
		*f = Float(n)
		return nil
	}
	// object 形式:取首个数值
	var m map[string]any
	if err := json.Unmarshal(data, &m); err == nil {
		for _, v := range m {
			if n, ok := v.(float64); ok {
				*f = Float(n)
				return nil
			}
		}
	}
	*f = 0
	return nil
}`,
}

// genDefinition 生成单个 definition 的 Go 类型。
func genDefinition(b *strings.Builder, name string, s Schema, sw *Swagger, isEndpointReturn bool) {
	// 优先使用覆盖定义
	if code, ok := definitionOverrides[name]; ok {
		fmt.Fprintf(b, "// %s 官方规范定义不完整,基于实际 API 调用验证做特殊处理。\n%s\n\n", name, code)
		return
	}
	switch {
	case s.Type == "array":
		// List 类型:始终生成 type alias(如 type ClanMemberList []ClanMember)。
		// 端点响应的 {Items, Paging} 包装由 endpointListWrapper 单独处理,避免与字段引用冲突。
		itemType := "any"
		if s.Items != nil {
			itemType = schemaToGoType(s.Items, sw)
		}
		fmt.Fprintf(b, "type %s []%s\n\n", name, itemType)
	case s.Type == "object" && len(s.Properties) == 0:
		// 空 object(如 JsonNode、Long、badgeUrls)按动态 map 处理
		fmt.Fprintf(b, "// %s 官方规范未定义具体字段,按动态对象处理。\ntype %s map[string]any\n\n", name, name)
	case s.Type == "object":
		genStruct(b, name, s, sw)
	case s.Type == "string":
		fmt.Fprintf(b, "type %s string\n\n", name)
	case s.Type == "integer":
		fmt.Fprintf(b, "type %s int\n\n", name)
	case s.Type == "boolean":
		fmt.Fprintf(b, "type %s bool\n\n", name)
	default:
		// 无 type 字段,只有 $ref 之类的——这种情况 coc 规范里没有,兜底
		fmt.Fprintf(b, "type %s any\n\n", name)
	}
}

// genStruct 生成 object 类型为 Go struct。
func genStruct(b *strings.Builder, name string, s Schema, sw *Swagger) {
	fmt.Fprintf(b, "// %s 对应官方 definition %s。\ntype %s struct {\n", name, name, name)
	// 按 field 名字母序,输出稳定
	var fields []string
	for f := range s.Properties {
		fields = append(fields, f)
	}
	sort.Strings(fields)
	for _, f := range fields {
		prop := s.Properties[f]
		goType := schemaToGoType(&prop, sw)
		// 字段名 PascalCase
		fieldName := pascalCase(f)
		fmt.Fprintf(b, "\t%s %s `json:\"%s,omitempty\"`\n", fieldName, goType, f)
	}
	b.WriteString("}\n\n")
}

// ============================================================
// 生成 api.go
// ============================================================

// chineseComments 以 operationId 为 key,提供每个端点的中文注释。
// 注释格式:"简短说明。参数/特性补充。"
// 当 API 变动新增端点时,在此表添加条目;未收录的端点会回退到官方英文 summary。
var chineseComments = map[string]string{
	"searchClans":                 "搜索部族。按名称和/或多种条件过滤;若用 name 至少 3 字符,至少需指定一个过滤条件。",
	"getClan":                     "获取部族信息。clanTag 须含 # 前缀,如 #2PP。",
	"getClanMembers":              "列出部族成员。支持分页(limit/after/before)。",
	"getClanWarLog":               "获取部族战争日志。支持分页。",
	"getCurrentWar":               "获取部族当前战争信息。含双方部族、成员、攻击详情。",
	"getClanWarLeagueGroup":       "获取部族当前部落战联赛(CWL)分组信息。含参赛部族和各轮次 warTag。",
	"getClanWarLeagueWar":         "获取单个 CWL 战争详情。warTag 从 leaguegroup 的 rounds 中获取。",
	"getCapitalRaidSeasons":       "获取部族都城突袭赛季记录。支持分页。",
	"getBattleLog":                "获取玩家对战日志(1v1 攻防记录)。",
	"getPlayer":                   "获取玩家信息。playerTag 须含 # 前缀,如 #2ABC。",
	"verifyToken":                 "验证玩家 API token。用于校验玩家身份,token 可在游戏设置中获取,一次性使用。",
	"getLeagueHistory":            "获取玩家联赛历史记录。",
	"getLeagues":                  "列出所有联赛。支持分页。",
	"getLeague":                   "获取单个联赛信息。leagueId 为字符串形式的数字 ID。",
	"getLeagueSeasons":            "获取联赛赛季列表。仅传奇联赛有赛季数据。",
	"getLeagueSeasonRankings":     "获取联赛赛季排名。返回该赛季玩家排名列表。",
	"getLeagueTiers":              "列出联赛段位。",
	"getLeagueTier":               "获取单个联赛段位信息。",
	"getLeagueGroup":              "获取联赛分组信息。按联赛组 tag 和赛季 ID 查询。",
	"getWarLeagues":               "列出所有部落战联赛。支持分页。",
	"getWarLeague":                "获取单个部落战联赛信息。",
	"getBuilderBaseLeagues":       "列出所有建筑大师联赛。支持分页。",
	"getBuilderBaseLeague":        "获取单个建筑大师联赛信息。",
	"getCapitalLeagues":           "列出所有都城联赛。支持分页。",
	"getCapitalLeague":            "获取单个都城联赛信息。",
	"getLocations":                "列出所有地区(国家/全球)。支持分页。",
	"getLocation":                 "获取单个地区信息。",
	"getClanRanking":              "获取指定地区的部族排名。locationId 为数字,全球用 32000006。",
	"getPlayerRanking":            "获取指定地区的玩家排名。",
	"getClanCapitalRanking":       "获取指定地区的都城排名。",
	"getClanBuilderBaseRanking":   "获取指定地区的建筑大师部族排名。",
	"getPlayerBuilderBaseRanking": "获取指定地区的建筑大师玩家排名。",
	"getClanLabels":               "列出所有部族标签。支持分页。",
	"getPlayerLabels":             "列出所有玩家标签。支持分页。",
	"getCurrentGoldPassSeason":    "获取当前金币通行证赛季信息。含起止时间。",
}

// chineseParamComments 参数名 → 中文说明,用于注释里描述路径参数。
var chineseParamComments = map[string]string{
	"clanTag":        "部族 tag(含 #)",
	"playerTag":      "玩家 tag(含 #)",
	"warTag":         "CWL 战争 tag(含 #)",
	"leagueId":       "联赛 ID(字符串形式)",
	"locationId":     "地区 ID(数字,全球为 32000006)",
	"leagueTierId":   "联赛段位 ID",
	"leagueGroupTag": "联赛组 tag",
	"leagueSeasonId": "联赛赛季 ID",
}

func genAPIGo(path string, eps []GenEndpoint) error {
	var b strings.Builder
	b.WriteString("// Code generated by scripts/generate-coc-api.go from official-swagger.yaml; DO NOT EDIT.\n")
	b.WriteString("// 权威来源: pkg/cocapi/official-swagger.yaml\n")
	b.WriteString("// 重新生成: go run scripts/generate-coc-api.go\n\n")
	b.WriteString("package cocapi\n\n")

	needStrconv := needsStrconv(eps)
	b.WriteString("import (\n\t\"context\"\n")
	if needStrconv {
		b.WriteString("\t\"strconv\"\n")
	}
	b.WriteString(")\n\n")

	for _, e := range eps {
		// 查询参数 struct
		if len(e.QueryParams) > 0 {
			fmt.Fprintf(&b, "// Query%s 是 %s 的查询参数。零值不传。\ntype Query%s struct {\n",
				e.GoName, e.GoName, e.GoName)
			for _, q := range e.QueryParams {
				fmt.Fprintf(&b, "\t%s %s `json:\"%s,omitempty\"`\n", pascalCase(q.Name), q.GoType, q.Name)
			}
			b.WriteString("}\n\n")
		}

		// 方法注释(中文优先,回退官方英文)
		comment, hasZh := chineseComments[e.OperationID]
		if !hasZh {
			comment = e.Summary
		}
		fmt.Fprintf(&b, "// %s %s\n//\n// 官方端点: %s %s\n", e.GoName, comment, e.Method, e.Path)

		// 路径参数中文说明
		if len(e.PathParams) > 0 {
			var paramDocs []string
			for _, p := range e.PathParams {
				if zh, ok := chineseParamComments[p.Name]; ok {
					paramDocs = append(paramDocs, p.Name+": "+zh)
				}
			}
			if len(paramDocs) > 0 {
				fmt.Fprintf(&b, "// 参数: %s\n", strings.Join(paramDocs, ", "))
			}
		}
		// body 参数说明
		if e.BodyType != "" {
			fmt.Fprintf(&b, "// 请求体: %s\n", e.BodyType)
		}
		// 返回类型说明
		fmt.Fprintf(&b, "// 返回: %s\n", e.ReturnType)

		// 方法签名
		fmt.Fprintf(&b, "func (c *Client) %s(%s) (%s, error) {\n",
			e.GoName, methodArgs(e), e.ReturnType)
		fmt.Fprintf(&b, "\tvar result %s\n", e.ReturnType)

		// 构造路径参数
		callArgsStr := ""
		if len(e.PathParams) > 0 {
			var args []string
			for _, p := range e.PathParams {
				switch p.GoType {
				case "int":
					args = append(args, "strconv.Itoa("+p.Name+")")
				case "string":
					if p.IsTag {
						args = append(args, "NormalizeTag("+p.Name+")")
					} else {
						args = append(args, p.Name)
					}
				default:
					args = append(args, p.Name)
				}
			}
			callArgsStr = ", " + strings.Join(args, ", ")
		}

		pathFmt := pathToFormat(e.Path)
		switch {
		case len(e.QueryParams) > 0:
			// 构造 queryMap
			b.WriteString("\tqueryMap := make(map[string]string)\n")
			for _, q := range e.QueryParams {
				field := pascalCase(q.Name)
				switch q.GoType {
				case "int":
					fmt.Fprintf(&b, "\tif query.%s != 0 { queryMap[%q] = strconv.Itoa(query.%s) }\n", field, q.Name, field)
				case "bool":
					fmt.Fprintf(&b, "\tif query.%s { queryMap[%q] = \"true\" }\n", field, q.Name)
				default:
					fmt.Fprintf(&b, "\tif query.%s != \"\" { queryMap[%q] = query.%s }\n", field, q.Name, field)
				}
			}
			fmt.Fprintf(&b, "\tif err := c.GetWithQuery(ctx, %q, queryMap, &result%s); err != nil {\n", pathFmt, callArgsStr)
		case e.Method == "GET":
			fmt.Fprintf(&b, "\tif err := c.Get(ctx, %q, &result%s); err != nil {\n", pathFmt, callArgsStr)
		case e.Method == "POST":
			bodyArg := "nil"
			if e.BodyType != "" {
				bodyArg = "body"
			}
			fmt.Fprintf(&b, "\tif err := c.Post(ctx, %q, %s, &result%s); err != nil {\n", pathFmt, bodyArg, callArgsStr)
		}
		b.WriteString("\t\treturn result, err\n\t}\n")
		b.WriteString("\treturn result, nil\n}\n\n")
	}

	return os.WriteFile(path, []byte(b.String()), 0o644)
}

// methodArgs 生成方法签名参数。
func methodArgs(e GenEndpoint) string {
	args := []string{"ctx context.Context"}
	for _, p := range e.PathParams {
		args = append(args, fmt.Sprintf("%s %s", p.Name, p.GoType))
	}
	if e.BodyType != "" {
		args = append(args, "body "+e.BodyType)
	}
	if len(e.QueryParams) > 0 {
		args = append(args, "query Query"+e.GoName)
	}
	return strings.Join(args, ", ")
}

// pathToFormat 把 /clans/{clanTag} 转成 /clans/{}(client.go 的 resolvePath 用 {})。
func pathToFormat(p string) string {
	var b strings.Builder
	i := 0
	for i < len(p) {
		c := p[i]
		if c == '{' {
			// 找到对应的 },替换成 {}
			j := strings.Index(p[i:], "}")
			if j < 0 {
				b.WriteString(p[i:])
				break
			}
			b.WriteString("{}")
			i += j + 1
		} else {
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

func needsStrconv(eps []GenEndpoint) bool {
	for _, e := range eps {
		for _, p := range e.PathParams {
			if p.GoType == "int" {
				return true
			}
		}
		for _, q := range e.QueryParams {
			if q.GoType == "int" {
				return true
			}
		}
	}
	return false
}

// ============================================================
// 工具
// ============================================================
