#!/bin/sh
# Suggestion quality benchmark.
# Feeds real command sequences to tip observe and scores how well
# the primary suggestion matches what the user actually did next.
#
# Usage:
#   sh quality-bench.sh              # static workflows only
#   sh quality-bench.sh --llm        # + LLM fallback
#
# Scoring:
#   1.0 = exact match
#   0.5 = same tool + subcommand
#   0.25 = same tool
#   0.0 = miss
set -e

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
FIXTURES="$SCRIPT_DIR/../fixtures/real-sessions.yaml"

LLM_MODE=""
if [ "$1" = "--llm" ]; then
  LLM_MODE="yes"
  if [ -z "$TIP_LLM_PROVIDER" ]; then
    echo "ERROR: --llm requires TIP_LLM_PROVIDER env var"
    exit 1
  fi
fi

# Enable tip
tip on >/dev/null 2>&1

# Enable LLM if requested
if [ -n "$LLM_MODE" ]; then
  export TIP_LLM_FALLBACK=true
fi

# Setup minimal git repo
REPO=$(mktemp -d)
cd "$REPO"
git init -q
git config user.email "bench@test" && git config user.name "bench"
echo "x" > f.txt && git add . && git commit -q -m "init"

total=0
hits_exact=0
hits_subcmd=0
hits_tool=0
misses=0

strip_ansi() {
  sed 's/\x1b\[[0-9;]*m//g'
}

score_suggestion() {
  local suggestion="$1"
  local actual="$2"

  local s_tool=$(echo "$suggestion" | awk '{print $1}')
  local s_sub=$(echo "$suggestion" | awk '{print $2}')
  local a_tool=$(echo "$actual" | awk '{print $1}')
  local a_sub=$(echo "$actual" | awk '{print $2}')

  if [ "$suggestion" = "$actual" ]; then
    echo "exact"
  elif [ "$s_tool" = "$a_tool" ] && [ -n "$s_sub" ] && \
       [ "$s_sub" = "$a_sub" ]; then
    echo "subcmd"
  elif [ "$s_tool" = "$a_tool" ]; then
    echo "tool"
  else
    echo "miss"
  fi
}

MODE_LABEL="static workflows"
if [ -n "$LLM_MODE" ]; then
  MODE_LABEL="static + LLM ($TIP_LLM_PROVIDER)"
fi

echo "╔═══════════════════════════════════════════════════════╗"
echo "║  tip Suggestion Quality Benchmark                    ║"
echo "║  Mode: $MODE_LABEL"
echo "╚═══════════════════════════════════════════════════════╝"
echo ""
printf "%-36s │ %-24s │ %-24s │ %s\n" \
  "Command" "tip suggested" "user did next" "Score"
echo "─────────────────────────────────────┼──────────────────────────┼──────────────────────────┼───────"

session_name=""
cmd=""
next=""

while IFS= read -r line; do
  case "$line" in
    *"- name:"*)
      session_name=$(echo "$line" | sed 's/.*- name: *//')
      ;;
    *"commands:"*)
      echo ""
      echo "── $session_name ──"
      ;;
    *"- cmd:"*)
      cmd=$(echo "$line" | sed 's/.*- cmd: *"//' | sed 's/"$//')
      ;;
    *"next:"*)
      next=$(echo "$line" | sed 's/.*next: *"//' | sed 's/"$//')

      if [ -n "$cmd" ] && [ -n "$next" ]; then
        tip_out=$(tip observe --cmd-line "$cmd" --exit 0 2>&1 \
          | strip_ansi | head -1 | sed 's/^[^a-zA-Z]*//')

        if [ -z "$tip_out" ]; then
          tip_out="(none)"
        fi

        result=$(score_suggestion "$tip_out" "$next")
        total=$((total + 1))

        case "$result" in
          exact)   hits_exact=$((hits_exact + 1)); mark="1.0" ;;
          subcmd)  hits_subcmd=$((hits_subcmd + 1)); mark="0.5" ;;
          tool)    hits_tool=$((hits_tool + 1)); mark="0.25" ;;
          miss)    misses=$((misses + 1)); mark="0.0" ;;
        esac

        cmd_short=$(echo "$cmd" | cut -c1-34)
        tip_short=$(echo "$tip_out" | cut -c1-22)
        next_short=$(echo "$next" | cut -c1-22)

        printf "%-36s │ %-24s │ %-24s │ %s\n" \
          "$cmd_short" "$tip_short" "$next_short" "$mark"
      fi

      cmd=""
      next=""
      ;;
  esac
done < "$FIXTURES"

echo "─────────────────────────────────────┼──────────────────────────┼──────────────────────────┼───────"
echo ""

# Summary (pure shell math, no bc)
total_score=$((hits_exact * 100 + hits_subcmd * 50 + hits_tool * 25))
if [ "$total" -gt 0 ]; then
  weighted=$((total_score / total))
  w_int=$((weighted / 100))
  w_frac=$((weighted % 100))
else
  w_int=0
  w_frac=0
fi

echo "Summary ($total commands, mode: $MODE_LABEL):"
echo "  Exact match (1.0):     $hits_exact"
echo "  Same subcommand (0.5): $hits_subcmd"
echo "  Same tool (0.25):      $hits_tool"
echo "  Miss (0.0):            $misses"
echo ""
printf "  Weighted score: %d.%02d / 1.0\n" "$w_int" "$w_frac"
echo ""

# Cleanup
rm -rf "$REPO"
