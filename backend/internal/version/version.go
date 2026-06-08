package version

import (
	"strconv"
	"strings"
)

// Compare 比较两个版本号
// 返回: -1 (a < b), 0 (a == b), 1 (a > b)
// 支持多种版本命名规则:
//   - SemVer: 1.2.3, 1.2.3-beta.1
//   - 日期版本: 2024.01.15, 2024-01-15
//   - 前缀数字: v1.2.3, V2.0
//   - 混合: 1.0-alpha, 2.1-build123
func Compare(a, b string) int {
	a = normalize(a)
	b = normalize(b)

	if a == b {
		return 0
	}

	aParts := splitVersion(a)
	bParts := splitVersion(b)

	maxLen := len(aParts)
	if len(bParts) > maxLen {
		maxLen = len(bParts)
	}

	for i := 0; i < maxLen; i++ {
		var aTok, bTok token
		if i < len(aParts) {
			aTok = aParts[i]
		}
		if i < len(bParts) {
			bTok = bParts[i]
		}

		if aTok.isNum && bTok.isNum {
			// 数字比较
			if aTok.num != bTok.num {
				if aTok.num < bTok.num {
					return -1
				}
				return 1
			}
		} else if aTok.isNum {
			// 数字 > 非数字 (例如: 1.0 > 1.0-alpha)
			return 1
		} else if bTok.isNum {
			return -1
		} else {
			// 都是字符串: 字典序比较，但特殊字符串有优先级
			r := comparePreRelease(aTok.str, bTok.str)
			if r != 0 {
				return r
			}
		}
	}

	return 0
}

// Latest 返回版本列表中最新的版本号
func Latest(versions []string) string {
	if len(versions) == 0 {
		return ""
	}
	latest := versions[0]
	for _, v := range versions[1:] {
		if Compare(v, latest) > 0 {
			latest = v
		}
	}
	return latest
}

// Sort 对版本号列表排序（升序）
func Sort(versions []string) {
	// 简单的插入排序，对小列表足够
	for i := 1; i < len(versions); i++ {
		for j := i; j > 0 && Compare(versions[j], versions[j-1]) < 0; j-- {
			versions[j], versions[j-1] = versions[j-1], versions[j]
		}
	}
}

// --- 内部实现 ---

type token struct {
	isNum bool
	num   int64
	str   string
}

// normalize 标准化版本号字符串
func normalize(v string) string {
	v = strings.TrimSpace(v)
	// 去除前缀 v/V
	v = strings.TrimPrefix(v, "v")
	v = strings.TrimPrefix(v, "V")
	// 将 - 替换为 . 统一分隔符（保留后面的 pre-release 标识）
	return v
}

// splitVersion 将版本号拆分为 token 列表
func splitVersion(v string) []token {
	var tokens []token

	var buf strings.Builder
	var hasDigit bool
	var hasAlpha bool

	flush := func() {
		if buf.Len() == 0 {
			return
		}
		s := buf.String()
		if hasDigit && !hasAlpha {
			if n, err := strconv.ParseInt(s, 10, 64); err == nil {
				tokens = append(tokens, token{isNum: true, num: n})
			} else {
				tokens = append(tokens, token{str: s})
			}
		} else {
			tokens = append(tokens, token{str: s})
		}
		buf.Reset()
		hasDigit = false
		hasAlpha = false
	}

	for i := 0; i < len(v); i++ {
		c := v[i]
		switch {
		case c == '.' || c == '-' || c == '_' || c == '/':
			flush()
		case c >= '0' && c <= '9':
			if hasAlpha && buf.Len() > 0 {
				flush()
			}
			buf.WriteByte(c)
			hasDigit = true
		case (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z'):
			if hasDigit && buf.Len() > 0 {
				flush()
			}
			buf.WriteByte(c)
			hasAlpha = true
		default:
			flush()
		}
	}
	flush()

	return tokens
}

// comparePreRelease 比较预发布标识的优先级
// 优先级: alpha < beta < pre < rc < (无标识/正式)
func comparePreRelease(a, b string) int {
	rank := func(s string) int {
		s = strings.ToLower(s)
		switch {
		case strings.Contains(s, "alpha"):
			return 1
		case strings.Contains(s, "beta"):
			return 2
		case strings.Contains(s, "pre"):
			return 3
		case strings.Contains(s, "rc"):
			return 4
		case strings.Contains(s, "snapshot"):
			return 0
		default:
			return 5
		}
	}

	ra := rank(a)
	rb := rank(b)
	if ra != rb {
		if ra < rb {
			return -1
		}
		return 1
	}

	// 同级别，按字典序
	if a < b {
		return -1
	}
	if a > b {
		return 1
	}
	return 0
}
