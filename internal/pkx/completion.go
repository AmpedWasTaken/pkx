package pkx

import (
	"fmt"
	"io"
)

// Completion scripts are generated locally. Nothing is installed or executed.
func writeCompletion(shell string, out, stderr io.Writer) int {
	var script string
	switch shell {
	case "bash":
		script = bashCompletion
	case "zsh":
		script = zshCompletion
	case "fish":
		script = fishCompletion
	case "powershell", "pwsh":
		script = powershellCompletion
	default:
		fmt.Fprintf(stderr, "pkx: unsupported shell %q (choose bash, zsh, fish or powershell)\n", shell)
		return 2
	}
	if _, err := io.WriteString(out, script); err != nil {
		fmt.Fprintf(stderr, "pkx: couldn't write completion: %v\n", err)
		return 1
	}
	return 0
}

const bashCompletion = `# pkx bash completion; source this file in your interactive shell.
_pkx_complete() {
  local cur prev subcommand
  COMPREPLY=()
  cur="${COMP_WORDS[COMP_CWORD]}"
  prev="${COMP_WORDS[COMP_CWORD-1]}"
  subcommand="${COMP_WORDS[1]}"
  case "$prev" in
    --dir) COMPREPLY=( $(compgen -d -- "$cur") ); return ;;
    --depth) return ;;
    --only) COMPREPLY=( $(compgen -W 'node npm pnpm yarn bun php composer python pip uv poetry rust cargo go' -- "$cur") ); return ;;
    completion) COMPREPLY=( $(compgen -W 'bash zsh fish powershell' -- "$cur") ); return ;;
  esac
  if [[ "$COMP_CWORD" -eq 1 ]]; then
    COMPREPLY=( $(compgen -W 'status plan up scan install doctor bundle init version completion help ls i' -- "$cur") )
    return
  fi
  if [[ "$subcommand" == bundle && "$COMP_CWORD" -eq 2 ]]; then
    COMPREPLY=( $(compgen -W 'init install' -- "$cur") )
    return
  fi
  if [[ "$cur" == -* ]]; then
    COMPREPLY=( $(compgen -W '--dir --depth --only --dry-run --json --help' -- "$cur") )
  fi
}
complete -F _pkx_complete pkx
`

const zshCompletion = `#compdef pkx
_pkx() {
  local -a commands managers
  commands=(status plan up scan install doctor bundle init version completion help ls i)
  managers=(node npm pnpm yarn bun php composer python pip uv poetry rust cargo go)
  if (( CURRENT == 2 )); then
    _describe 'pkx command' commands
    return
  fi
  case "$words[CURRENT-1]" in
    --dir) _files -/; return ;;
    --only) _describe 'package manager' managers; return ;;
    completion) compadd bash zsh fish powershell; return ;;
  esac
  if [[ "$words[2]" == bundle && CURRENT == 3 ]]; then
    compadd init install
    return
  fi
  _arguments '*: :((--dir --depth --only --dry-run --json --help))'
}
compdef _pkx pkx
`

const fishCompletion = `# pkx fish completion
complete -c pkx -f
complete -c pkx -n '__fish_use_subcommand' -a 'status plan up scan install doctor bundle init version completion help ls i'
complete -c pkx -n '__fish_seen_subcommand_from bundle' -a 'init install'
complete -c pkx -n '__fish_seen_subcommand_from completion' -a 'bash zsh fish powershell'
complete -c pkx -l dir -r -F
complete -c pkx -l depth -r
complete -c pkx -l only -r -a 'node npm pnpm yarn bun php composer python pip uv poetry rust cargo go'
complete -c pkx -l dry-run
complete -c pkx -l json
complete -c pkx -l help
`

const powershellCompletion = `# pkx PowerShell completion
Register-ArgumentCompleter -Native -CommandName pkx -ScriptBlock {
    param($wordToComplete, $commandAst, $cursorPosition)
    $words = @($commandAst.CommandElements | ForEach-Object { $_.Extent.Text })
    $previous = if ($words.Count -ge 2) { $words[$words.Count - 2] } else { '' }
    $options = switch ($previous) {
        '--only' { @('node','npm','pnpm','yarn','bun','php','composer','python','pip','uv','poetry','rust','cargo','go'); break }
        'completion' { @('bash','zsh','fish','powershell'); break }
        'bundle' { @('init','install'); break }
        default {
            if ($words.Count -le 2) { @('status','plan','up','scan','install','doctor','bundle','init','version','completion','help','ls','i') }
            else { @('--dir','--depth','--only','--dry-run','--json','--help') }
        }
    }
    foreach ($option in $options) {
        if ($option -like "$wordToComplete*") {
            [System.Management.Automation.CompletionResult]::new($option, $option, 'ParameterValue', $option)
        }
    }
}
`
