// Copyright 2025 fsyyft-go
//
// Licensed under the MIT License. See LICENSE file in the project root for full license information.

package imp

import (
	"errors"
	"fmt"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ImportGroup 表示导入分组的枚举类型。
type ImportGroup int

const (
	// GroupBuiltin 表示内置包分组。
	GroupBuiltin ImportGroup = iota
	// GroupThirdParty 表示第三方包分组。
	GroupThirdParty
	// GroupKit 表示 fsyyft-go 相关包分组。
	GroupKit
	// GroupApp 表示项目内包分组。
	GroupApp
)

// ImportInfo 存储单个导入的信息，包括路径、别名和分组。
type ImportInfo struct {
	// Path 导入包的路径。
	Path string
	// Alias 导入包的别名，如果没有别名则为空。
	Alias string
	// Group 导入包所属的分组。
	Group ImportGroup
}

// checkImports 检查单个文件的导入语句是否符合规范。
// 参数：
//   - filePath：要检查的 Go 文件路径。
//
// 返回值：
//   - []string：发现的问题列表，如果没有问题则为空切片。
//   - error：检查过程中发生的错误。
func checkImports(filePath string) ([]string, error) {
	fset := token.NewFileSet()
	src, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	file, err := parser.ParseFile(fset, filePath, src, parser.ParseComments)
	if err != nil {
		return nil, err
	}

	var issues []string

	// 检查是否有 import 声明
	if len(file.Imports) == 0 {
		return issues, nil
	}

	// 解析导入信息
	var imports []ImportInfo
	for _, imp := range file.Imports {
		path := strings.Trim(imp.Path.Value, `"`)
		alias := ""
		if imp.Name != nil {
			alias = imp.Name.Name
		}

		group := determineGroup(path, alias)
		imports = append(imports, ImportInfo{
			Path:  path,
			Alias: alias,
			Group: group,
		})
	}

	// 检查分组和排序
	issues = append(issues, checkGroupingAndSorting(filePath, imports)...)

	return issues, nil
}

// determineGroup 根据导入路径和别名确定所属分组。
// 参数：
//   - path：导入包的路径。
//   - alias：导入包的别名。
//
// 返回值：
//   - ImportGroup：确定的分组类型。
func determineGroup(path, alias string) ImportGroup {
	if strings.HasPrefix(path, "github.com/fsyyft-go/kit") {
		return GroupKit
	}
	if strings.HasPrefix(alias, "kit") {
		return GroupKit
	}
	if strings.HasPrefix(alias, "app") {
		return GroupApp
	}
	if isBuiltinPackage(path) {
		return GroupBuiltin
	}
	return GroupThirdParty
}

// isBuiltinPackage 检查给定的路径是否为 Go 内置包。
// 参数：
//   - path：要检查的包路径。
//
// 返回值：
//   - bool：如果是内置包则返回 true，否则返回 false。
func isBuiltinPackage(path string) bool {
	builtinPackages := []string{
		"archive", "bufio", "builtin", "bytes", "compress", "container", "context",
		"crypto", "database", "debug", "encoding", "errors", "expvar", "flag", "fmt",
		"go", "hash", "html", "image", "internal", "io", "log", "math", "mime",
		"net", "os", "path", "plugin", "reflect", "regexp", "runtime", "slices", "sort",
		"strconv", "strings", "sync", "syscall", "testing", "text", "time",
		"unicode", "unsafe",
	}
	for _, pkg := range builtinPackages {
		if strings.HasPrefix(path, pkg+"/") || path == pkg {
			return true
		}
	}
	return false
}

// checkGroupingAndSorting 检查导入的分组和排序是否符合规范。
// 参数：
//   - filePath：文件路径，用于错误信息。
//   - imports：导入信息列表。
//
// 返回值：
//   - []string：发现的问题列表。
func checkGroupingAndSorting(filePath string, imports []ImportInfo) []string {
	var issues []string

	// 按组分组
	groups := make(map[ImportGroup][]ImportInfo)
	for _, imp := range imports {
		groups[imp.Group] = append(groups[imp.Group], imp)
	}

	// 检查每个组的排序
	for group, groupImports := range groups {
		if !isSorted(groupImports) {
			groupName := getGroupName(group)
			issues = append(issues, fmt.Sprintf("%s: %s 组未按字母顺序排序", filePath, groupName))
		}
	}

	// 检查别名规范
	for _, imp := range imports {
		if issue := checkAlias(imp); issue != "" {
			issues = append(issues, fmt.Sprintf("%s: %s", filePath, issue))
		}
	}

	return issues
}

// isSorted 检查导入列表是否按路径字母顺序排序。
// 参数：
//   - imports：导入信息列表。
//
// 返回值：
//   - bool：如果已排序则返回 true，否则返回 false。
func isSorted(imports []ImportInfo) bool {
	paths := make([]string, len(imports))
	for i, imp := range imports {
		paths[i] = imp.Path
	}
	return sort.StringsAreSorted(paths)
}

// getGroupName 获取分组的中文名称。
// 参数：
//   - group：分组类型。
//
// 返回值：
//   - string：分组的中文名称。
func getGroupName(group ImportGroup) string {
	switch group {
	case GroupBuiltin:
		return "内置包"
	case GroupThirdParty:
		return "第三方包"
	case GroupKit:
		return "fsyyft-go 包"
	case GroupApp:
		return "项目内包"
	default:
		return "未知"
	}
}

// checkAlias 检查导入的别名是否符合规范。
// 参数：
//   - imp：导入信息。
//
// 返回值：
//   - string：如果不符合规范则返回错误信息，否则返回空字符串。
func checkAlias(imp ImportInfo) string {
	if imp.Group == GroupKit {
		if !strings.HasPrefix(imp.Alias, "kit") {
			return fmt.Sprintf("fsyyft-go 包 %s 应使用 kit 前缀别名", imp.Path)
		}
	}
	if imp.Group == GroupApp {
		if !strings.HasPrefix(imp.Alias, "app") {
			return fmt.Sprintf("项目内包 %s 应使用 app 前缀别名", imp.Path)
		}
	}
	// 检查别名只能包含小写字母和数字
	if imp.Alias != "" {
		for _, r := range imp.Alias {
			if !((r >= 'a' && r <= 'z') || (r >= '0' && r <= '9')) {
				return fmt.Sprintf("别名 %s 只能包含小写字母和数字", imp.Alias)
			}
		}
	}
	return ""
}

// walkGoFiles 遍历指定目录下的所有 Go 文件，返回文件路径列表。
// 参数：
//   - root：根目录路径。
//
// 返回值：
//   - []string：Go 文件路径列表。
//   - error：遍历过程中的错误，如果有的话。
func walkGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// Check 检查指定目录下的所有 Go 文件的导入语句是否符合规范。
// 参数：
//   - dir：要检查的目录路径。
//
// 返回值：
//   - []string：所有发现的问题列表，如果没有问题则为空切片。
//   - error：检查过程中发生的错误，如果成功则返回 nil。
func Check(dir string) ([]string, error) {
	// 遍历指定目录下的所有 Go 文件。
	files, err := walkGoFiles(dir)
	if err != nil {
		return nil, fmt.Errorf("遍历文件失败: %w", err)
	}

	allIssues := make([]string, 0, len(files)) // 预分配容量以提高性能。
	errs := make([]error, 0, len(files))       // 预分配容量。

	// 逐个检查每个 Go 文件的导入语句。
	for _, file := range files {
		issues, err := checkImports(file)
		if err != nil {
			errs = append(errs, fmt.Errorf("检查文件 %s 失败: %w", file, err))
		} else {
			allIssues = append(allIssues, issues...)
		}
	}

	// 如果有检查错误，则返回合并的错误信息。
	if len(errs) > 0 {
		return nil, fmt.Errorf("检查文件过程出现问题：%w", errors.Join(errs...))
	}

	return allIssues, nil
}
