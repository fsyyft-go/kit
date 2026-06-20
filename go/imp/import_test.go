package imp

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCheckAliasAllowsBlankImport 验证副作用导入使用的空白标识符别名不会被误判为非法别名。
func TestCheckAliasAllowsBlankImport(t *testing.T) {
	issue := checkAlias(ImportInfo{
		Path:  "github.com/go-kratos/kratos/v2/encoding/json",
		Alias: "_",
		Group: GroupThirdParty,
	})
	assert.Empty(t, issue)
}

// TestImport 测试导入语句检查功能。
// 该测试函数遍历项目中的所有 Go 文件，检查每个文件的导入语句是否符合规范，
// 包括分组、排序和别名规则。如果发现问题，会记录并输出所有问题。
func TestImport(t *testing.T) {
	oldWd, err := os.Getwd()
	require.NoError(t, err)
	t.Cleanup(func() {
		require.NoError(t, os.Chdir(oldWd))
	})

	root := "../.."
	require.NoError(t, os.Chdir(root))
	pwd, err := os.Getwd()
	require.NoError(t, err)
	t.Logf("根目录 %s", pwd)

	allIssues, err := Check(".")
	require.NoError(t, err)
	assert.Empty(t, allIssues)
}
