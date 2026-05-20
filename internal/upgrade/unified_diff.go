package upgrade

import (
	"fmt"
	"strings"
)

const diffSeparator = " │ "

func UnifiedDiff(oldText, newText []byte, oldLabel, newLabel string) string {
	oldLines, oldNL := splitLines(oldText)
	newLines, newNL := splitLines(newText)

	if equalSlices(oldLines, newLines) && oldNL == newNL {
		return ""
	}

	ops := lcsDiff(oldLines, newLines)
	blocks := groupDiffBlocks(ops, 3)

	width := gutterWidth(blocks)
	blank := strings.Repeat(" ", width)

	var b strings.Builder
	fmt.Fprintf(&b, "--- %s\n", oldLabel)
	fmt.Fprintf(&b, "+++ %s\n", newLabel)

	for _, block := range blocks {
		fmt.Fprintf(&b, "@@ 旧 %s  →  新 %s @@\n",
			formatRange(block.oldStart, block.oldCount),
			formatRange(block.newStart, block.newCount))

		oldNo := block.oldStart
		newNo := block.newStart
		for _, op := range block.ops {
			oldCol := blank
			newCol := blank
			var prefix byte
			switch op.kind {
			case opEqual:
				oldNo++
				newNo++
				oldCol = fmt.Sprintf("%*d", width, oldNo)
				newCol = fmt.Sprintf("%*d", width, newNo)
				prefix = ' '
			case opDel:
				oldNo++
				oldCol = fmt.Sprintf("%*d", width, oldNo)
				prefix = '-'
			case opAdd:
				newNo++
				newCol = fmt.Sprintf("%*d", width, newNo)
				prefix = '+'
			}
			fmt.Fprintf(&b, "%s %s%s%c%s\n", oldCol, newCol, diffSeparator, prefix, op.line)
		}
	}

	if !oldNL || !newNL {
		b.WriteString("\\ No newline at end of file\n")
	}

	return b.String()
}

func formatRange(start, count int) string {
	if count == 0 {
		return "(空)"
	}
	if count == 1 {
		return fmt.Sprintf("L%d", start+1)
	}
	return fmt.Sprintf("L%d–%d", start+1, start+count)
}

func gutterWidth(blocks []diffBlock) int {
	maxLine := 0
	for _, block := range blocks {
		if v := block.oldStart + block.oldCount; v > maxLine {
			maxLine = v
		}
		if v := block.newStart + block.newCount; v > maxLine {
			maxLine = v
		}
	}
	w := 1
	for n := maxLine; n >= 10; n /= 10 {
		w++
	}
	if w < 2 {
		w = 2
	}
	return w
}

func splitLines(text []byte) (lines []string, hasTrailingNewline bool) {
	if len(text) == 0 {
		return nil, false
	}
	s := string(text)
	hasTrailingNewline = strings.HasSuffix(s, "\n")
	if hasTrailingNewline {
		s = s[:len(s)-1]
	}
	return strings.Split(s, "\n"), hasTrailingNewline
}

func equalSlices(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

type opKind int

const (
	opEqual opKind = iota
	opDel
	opAdd
)

type diffOp struct {
	kind opKind
	line string
}

func lcsDiff(oldLines, newLines []string) []diffOp {
	m, n := len(oldLines), len(newLines)
	dp := make([][]int, m+1)
	for i := range dp {
		dp[i] = make([]int, n+1)
	}
	for i := m - 1; i >= 0; i-- {
		for j := n - 1; j >= 0; j-- {
			if oldLines[i] == newLines[j] {
				dp[i][j] = dp[i+1][j+1] + 1
			} else if dp[i+1][j] >= dp[i][j+1] {
				dp[i][j] = dp[i+1][j]
			} else {
				dp[i][j] = dp[i][j+1]
			}
		}
	}

	ops := make([]diffOp, 0, m+n)
	i, j := 0, 0
	for i < m && j < n {
		if oldLines[i] == newLines[j] {
			ops = append(ops, diffOp{kind: opEqual, line: oldLines[i]})
			i++
			j++
		} else if dp[i+1][j] >= dp[i][j+1] {
			ops = append(ops, diffOp{kind: opDel, line: oldLines[i]})
			i++
		} else {
			ops = append(ops, diffOp{kind: opAdd, line: newLines[j]})
			j++
		}
	}
	for ; i < m; i++ {
		ops = append(ops, diffOp{kind: opDel, line: oldLines[i]})
	}
	for ; j < n; j++ {
		ops = append(ops, diffOp{kind: opAdd, line: newLines[j]})
	}
	return ops
}

type diffBlock struct {
	oldStart, newStart int
	oldCount, newCount int
	ops                []diffOp
}

func groupDiffBlocks(ops []diffOp, context int) []diffBlock {
	type indexed struct {
		op     diffOp
		oldIdx int
		newIdx int
	}
	idx := make([]indexed, len(ops))
	oi, ni := 0, 0
	for k, op := range ops {
		idx[k] = indexed{op: op, oldIdx: oi, newIdx: ni}
		switch op.kind {
		case opEqual:
			oi++
			ni++
		case opDel:
			oi++
		case opAdd:
			ni++
		}
	}

	var blocks []diffBlock
	i := 0
	for i < len(idx) {
		if idx[i].op.kind == opEqual {
			i++
			continue
		}
		start := i
		for start > 0 && idx[start-1].op.kind == opEqual && i-start < context {
			start--
		}

		end := i
		for end < len(idx) {
			if idx[end].op.kind != opEqual {
				end++
				continue
			}
			look := end
			for look < len(idx) && idx[look].op.kind == opEqual && look-end < 2*context {
				look++
			}
			if look < len(idx) && idx[look].op.kind != opEqual {
				end = look
				continue
			}
			trailing := context
			if end+trailing > len(idx) {
				trailing = len(idx) - end
			}
			end += trailing
			break
		}

		hOps := make([]diffOp, 0, end-start)
		for k := start; k < end; k++ {
			hOps = append(hOps, idx[k].op)
		}
		var oldCount, newCount int
		for _, op := range hOps {
			switch op.kind {
			case opEqual:
				oldCount++
				newCount++
			case opDel:
				oldCount++
			case opAdd:
				newCount++
			}
		}
		blocks = append(blocks, diffBlock{
			oldStart: idx[start].oldIdx,
			newStart: idx[start].newIdx,
			oldCount: oldCount,
			newCount: newCount,
			ops:      hOps,
		})
		i = end
	}
	return blocks
}
