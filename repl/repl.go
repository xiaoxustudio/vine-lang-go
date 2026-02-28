package repl

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"vine-lang/env"
	"vine-lang/ipt"
	"vine-lang/lexer"
	"vine-lang/parser"
	"vine-lang/utils"
	"vine-lang/verror"
)

const (
	PROMPT       = "vine> "
	MULTI_PROMPT = "...   "
)

func init() {
	env.SetExecuteCode(executeCodeForModule)
}

func executeCodeForModule(filename string, code string, wk env.Workspace) (any, error) {
	lex := lexer.New(filename, code)
	lex.Parse()

	p := parser.CreateParser(lex)

	e := env.New(wk)
	e.FileName = filename

	i := ipt.New(p, e)

	return i.EvalSafeWithDefer()
}

// REPL 交互式环境结构
type REPL struct {
	env       *env.Environment
	scanner   *bufio.Scanner
	multiLine bool
	buffer    strings.Builder
}

// New 创建新的 REPL 实例
func New() *REPL {
	return &REPL{
		env:       env.New(env.Workspace{FileName: "<repl>"}),
		scanner:   bufio.NewScanner(os.Stdin),
		multiLine: false,
	}
}

// Start 启动 REPL
func Start() {
	repl := New()
	repl.printWelcome()
	repl.run()
}

// printWelcome 打印欢迎信息
func (r *REPL) printWelcome() {
	fmt.Println("Vine Language REPL")
	fmt.Println("input .help to see helpful info. .exit to quit")
	fmt.Println()
}

// run 运行 REPL 主循环
func (r *REPL) run() {
	for {
		// 显示提示符
		if r.multiLine {
			fmt.Print(MULTI_PROMPT)
		} else {
			fmt.Print(PROMPT)
		}

		// 读取输入
		if !r.scanner.Scan() {
			break
		}

		line := r.scanner.Text()

		// 处理特殊命令
		if !r.multiLine && strings.HasPrefix(line, ".") {
			if r.handleCommand(line) {
				return
			}
			continue
		}

		// 处理多行输入
		if r.handleMultiLine(line) {
			continue
		}

		// 执行代码
		r.execute(line)
	}
}

// handleCommand 处理 REPL 命令
func (r *REPL) handleCommand(cmd string) bool {
	cmd = strings.TrimSpace(cmd)

	switch cmd {
	case ".exit", ".quit":
		fmt.Println("quit!")
		return true
	case ".help":
		r.printHelp()
	case ".clear":
		r.clearEnv()
	case ".env":
		r.printEnv()
	case ".reset":
		r.reset()
	default:
		fmt.Printf("unknown command: %s\n", cmd)
		fmt.Println("input .help to see helpful info.")
	}

	return false
}

// handleMultiLine 处理多行输入
func (r *REPL) handleMultiLine(line string) bool {
	trimmed := strings.TrimSpace(line)

	// 检查是否开始多行输入
	if !r.multiLine {
		// 检查是否需要多行输入（以 : 结尾或有未闭合的括号）
		if r.needsMoreInput(line) {
			r.multiLine = true
			r.buffer.WriteString(line)
			r.buffer.WriteString("\n")
			return true
		}
		return false
	}

	// 多行模式中
	// 空行结束多行输入
	if trimmed == "" {
		code := r.buffer.String()
		r.buffer.Reset()
		r.multiLine = false
		r.execute(code)
		return true
	}

	// 显式的 end 命令也可以结束
	if trimmed == "end" {
		code := r.buffer.String()
		r.buffer.Reset()
		r.multiLine = false
		r.execute(code)
		return true
	}

	// 继续累积多行输入
	r.buffer.WriteString(line)
	r.buffer.WriteString("\n")

	// 检查是否可以结束（括号已匹配且不以 : 结尾）
	currentCode := r.buffer.String()
	if !r.needsMoreInput(strings.TrimSpace(currentCode)) && !strings.HasSuffix(trimmed, ":") {
		// 自动结束多行输入
		r.buffer.Reset()
		r.multiLine = false
		r.execute(currentCode)
		return true
	}

	return true
}

// needsMoreInput 检查是否需要更多输入
func (r *REPL) needsMoreInput(code string) bool {
	trimmed := strings.TrimSpace(code)

	// 以 : 结尾需要多行
	if strings.HasSuffix(trimmed, ":") {
		return true
	}

	// 检查括号是否匹配
	return !r.isBalanced(code)
}

// isBalanced 检查括号是否平衡
func (r *REPL) isBalanced(code string) bool {
	stack := []rune{}
	pairs := map[rune]rune{
		')': '(',
		'}': '{',
		']': '[',
	}

	inString := false
	var stringChar rune
	escaped := false

	for _, ch := range code {
		// 处理转义字符
		if escaped {
			escaped = false
			continue
		}

		if ch == '\\' {
			escaped = true
			continue
		}

		// 处理字符串
		if ch == '"' || ch == '\'' {
			if !inString {
				inString = true
				stringChar = ch
			} else if ch == stringChar {
				inString = false
			}
			continue
		}

		// 字符串内部不检查括号
		if inString {
			continue
		}

		// 检查括号
		switch ch {
		case '(', '{', '[':
			stack = append(stack, ch)
		case ')', '}', ']':
			if len(stack) == 0 {
				return false
			}
			if stack[len(stack)-1] != pairs[ch] {
				return false
			}
			stack = stack[:len(stack)-1]
		}
	}

	return len(stack) == 0
}

// execute 执行代码
func (r *REPL) execute(code string) {
	defer func() {
		if rec := recover(); rec != nil {
			r.handleError(rec)
		}
	}()

	code = strings.TrimSpace(code)
	if code == "" {
		return
	}

	// 词法分析
	lex := lexer.New("<repl>", code)
	lex.Parse()

	// 语法分析
	p := parser.CreateParser(lex)

	// 解释执行
	i := ipt.New(p, r.env)
	result, err := i.EvalSafeWithDefer()

	if err != nil {
		fmt.Printf("repl errors: %v\n", err)
		return
	}

	if result != nil {
		fmt.Printf("%s", utils.TrasformPrintString(result))
		fmt.Println()
	}
}

// handleError 处理错误
func (r *REPL) handleError(rec interface{}) {
	switch err := rec.(type) {
	case verror.VError:
		fmt.Println(err.Error())
	case verror.ParseVError:
		fmt.Println(err.Error())
	case verror.InterpreterVError:
		fmt.Println(err.Error())
	case verror.LexerVError:
		fmt.Println(err.Error())
	default:
		fmt.Printf("repl errors: %v\n", rec)
	}
}

// printHelp 打印帮助信息
func (r *REPL) printHelp() {
	fmt.Println("REPL command list:")
	fmt.Println("  .help          show helpful info")
	fmt.Println("  .exit, .quit   exit REPL")
	fmt.Println("  .clear         clear all output")
	fmt.Println("  .env           show environment variables")
	fmt.Println("  .reset         reset environment")
	fmt.Println()
	fmt.Println("Multi line input:")
	fmt.Println("  Enable multiline mode for lines ending in:")
	fmt.Println("  Enter blank line or end to end multi line input")
	fmt.Println("  Unfosed parentheses will automatically enter multiline mode")
	fmt.Println()
	fmt.Println("example:")
	fmt.Println("  vine> let a = 10")
	fmt.Println("  vine> for let i = 0; i < 5; i++ :")
	fmt.Println("  ...       glb.print(i)")
	fmt.Println("  ...   ")
	fmt.Println("  vine> let obj = {")
	fmt.Println("  ...       name: 'test',")
	fmt.Println("  ...       value: 42")
	fmt.Println("  ...   }")
}

// clearEnv 清空屏幕
func (r *REPL) clearEnv() {
	// Windows 和 Unix 系统的清屏命令不同
	fmt.Print("\033[H\033[2J")
}

// printEnv 打印环境变量
func (r *REPL) printEnv() {
	fmt.Println("current environment variables:")
	r.env.Print()
}

// reset 重置环境
func (r *REPL) reset() {
	r.env = env.New(r.env.WorkSpace)
	r.buffer.Reset()
	r.multiLine = false
	fmt.Println("environment is reset!")
}
