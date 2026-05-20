#!/usr/bin/env sh
set -eu

HOOK_DIR="$(cd "$(dirname "$0")" && pwd)"
. "$HOOK_DIR/lib_json.sh"

payload="$(cat | tr '\n' ' ')"
command="$(extract_json_field "$payload" "command")"

emit_decision() {
  decision="$1"
  reason="$2"
  escaped="$(printf '%s' "$reason" | sed 's/"/\\\"/g')"
  printf '{"hookSpecificOutput":{"hookEventName":"PreToolUse","permissionDecision":"%s","permissionDecisionReason":"%s"}}\n' "$decision" "$escaped"
  exit 0
}

case "$command" in
  *"sudo "*)
    emit_decision "deny" "ハーネス内で sudo は禁止です。プロジェクトローカルのコマンドを使用してください。"
    ;;
  *"git push --force"*|*"git push -f"*)
    emit_decision "deny" "Force push はスキャフォールドにより禁止されています。"
    ;;
  *"git reset --hard"*)
    emit_decision "deny" "Hard reset はスキャフォールドにより禁止されています。"
    ;;
  *".git/"*">"*|*"> .git"*|*"tee .git"*)
    emit_decision "ask" ".git への直接書き込みには明示的な確認が必要です。"
    ;;
  *".env"*">"*|*"> .env"*|*"tee .env"*)
    emit_decision "ask" "シークレットファイルへの書き込みには明示的な確認が必要です。"
    ;;
  *"rm -rf "*)
    emit_decision "ask" "再帰削除には明示的な確認が必要です。"
    ;;
  *"gh pr create"*)
    emit_decision "ask" "gh pr create を検出。/pr スキル経由で実行していますか？ /pr スキルは日本語テンプレート・事前チェック・プランアーカイブを強制します。"
    ;;
esac

case "$command" in
  *"git commit"*"-m "*)
    msg_part="${command#*-m }"
    case "$msg_part" in
      '"'*'`'*|'"'*'$('*)
        emit_decision "deny" "コミットメッセージのダブルクォート内にバッククォートまたは \$() を検出しました。シークレット漏洩の恐れがあります。シングルクォートまたは HEREDOC を使用してください。"
        ;;
    esac
    ;;
esac

exit 0
