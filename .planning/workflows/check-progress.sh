#!/bin/bash
set -e

# Check Progress Workflow
# Verifies planning structure, loads context, and routes to next action

cd "$(dirname "$0")/../.."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Step 1: Verify planning structure
verify_structure() {
    echo -e "${BLUE}🔍 Verifying planning structure...${NC}\n"
    
    if ! test -d ".planning"; then
        echo -e "${RED}No planning structure found.${NC}"
        echo ""
        echo "Run /gsd-new-project to start a new project."
        exit 1
    fi
    
    if [ ! -f ".planning/STATE.md" ]; then
        echo -e "${YELLOW}STATE.md missing. This may be a new project.${NC}"
        echo ""
        echo "Run /gsd-new-project to initialize planning."
        exit 1
    fi
    
    if [ ! -f ".planning/ROADMAP.md" ] && [ ! -f ".planning/PROJECT.md" ]; then
        echo -e "${YELLOW}Neither ROADMAP.md nor PROJECT.md found.${NC}"
        echo ""
        echo "Run /gsd-new-project to initialize planning."
        exit 1
    fi
    
    if [ ! -f ".planning/ROADMAP.md" ] && [ -f ".planning/PROJECT.md" ]; then
        echo -e "${GREEN}✓ Milestone archived - ROUTE F (between milestones)${NC}"
        route_between_milestones
        exit 0
    fi
    
    echo -e "${GREEN}✓ Planning structure verified${NC}"
}

# Step 2: Load full project context
load_context() {
    echo ""
    echo -e "${BLUE}📚 Loading project context...${NC}\n"
    
    STATE_PATH=".planning/STATE.md"
    ROADMAP_PATH=".planning/ROADMAP.md"
    PROJECT_PATH=".planning/PROJECT.md"
    CONFIG_PATH=".planning/config.json"
    
    # Read files and extract key sections
    if [ -f "$STATE_PATH" ]; then
        echo ".STATE.md loaded."
        STATE_LOADED=true
    fi
    
    if [ -f "$ROADMAP_PATH" ]; then
        echo "ROADMAP.md loaded."
        ROADMAP_LOADED=true
    fi
    
    if [ -f "$PROJECT_PATH" ]; then
        echo "PROJECT.md loaded."
        PROJECT_LOADED=true
    fi
    
    if [ -f "$CONFIG_PATH" ]; then
        echo "config.json loaded."
        CONFIG_LOADED=true
    fi
}

# Step 3: Gather recent work context
gather_recent_work() {
    echo ""
    echo -e "${BLUE}📅 Gathering recent work context...${NC}\n"
    
    RECENT_SUMMARIES=$(ls -lt .planning/phases/*/*-SUMMARY.md 2>/dev/null | head -3 | awk '{print $NF}')
    
    if [ -z "$RECENT_SUMMARIES" ]; then
        echo "No recent summaries found."
        return
    fi
    
    echo "Recent completions:"
    echo "$RECENT_SUMMARIES" | while read -r summary; do
        if [ -f "$summary" ]; then
            plan_file="${summary/-SUMMARY.md/-PLAN.md}"
            if [ -f "$plan_file" ]; then
                plan_name=$(basename "$plan_file")
                # Extract objective from SUMMARY or PLAN
                objective=$(grep -A2 "## Objective" "$summary" 2>/dev/null | tail -1 || grep -A2 "## Objective" "$plan_file" 2>/dev/null | tail -1)
                echo "  • $plan_name: ${objective:-Work completed}"
            fi
        fi
    done
}

# Step 4: Parse current position
parse_current_position() {
    echo ""
    echo -e "${BLUE}📍 Parsing current position...${NC}\n"
    
    # Extract current phase from STATE.md
    if [ -f ".planning/STATE.md" ]; then
        CURRENT_PHASE_LINE=$(grep "Phase:" .planning/STATE.md | head -1)
        PHASE_NUMBER=$(echo "$CURRENT_PHASE_LINE" | grep -oE "Phase: [0-9]+" | grep -oE "[0-9]+" | head -1)
        PHASE_TOTAL=$(echo "$CURRENT_PHASE_LINE" | grep -oE "of [0-9]+" | grep -oE "[0-9]+" | head -1)
        PHASE_STATUS=$(echo "$CURRENT_PHASE_LINE" | sed 's/.*Status: //')
        
        echo "Phase: $PHASE_NUMBER of $PHASE_TOTAL"
    fi
    
    # Get current phase directory from STATE.md
    if [ -f ".planning/STATE.md" ]; then
        CURRENT_PHASE_DIR=$(grep "Phase in progress" .planning/STATE.md 2>/dev/null | head -1 | sed 's/.*: //' | sed 's/\(.*\)/\1/')
        if [ -n "$CURRENT_PHASE_DIR" ]; then
            PHASE_DIR_NAME=$(basename "$CURRENT_PHASE_DIR" | sed 's/ \(.*\)/\1/')
        else
            PHASE_DIR_NAME="0$current_phase_num"
        fi
    fi
    
    # Fallback to first existing phase directory
    if [ -z "$PHASE_DIR_NAME" ]; then
        PHASE_DIR_NAME=$(ls -d .planning/phases/* 2>/dev/null | head -1 | xargs basename)
    fi
    
    # Check for CONTEXT.md
    if [ -f ".planning/phases/$PHASE_DIR_NAME/CONTEXT.md" ]; then
        echo "Context: ✓ CONTEXT.md exists"
        CONTEXT_EXISTS=true
    else
        echo "Context: - No CONTEXT.md"
        CONTEXT_EXISTS=false
    fi
    
    # Count pending todos
    if [ -d ".planning/todos/pending/" ]; then
        TODOS_COUNT=$(ls .planning/todos/pending/*.md 2>/dev/null | wc -l | tr -d ' ')
    else
        TODOS_COUNT=0
    fi
    echo "Pending todos: $TODOS_COUNT"
    
    # Check for active debug sessions
    if [ -d ".planning/debug/" ]; then
        DEBUG_COUNT=$(ls .planning/debug/*.md 2>/dev/null | grep -v resolved | wc -l | tr -d ' ')
    else
        DEBUG_COUNT=0
    fi
    if [ "$DEBUG_COUNT" -gt 0 ]; then
        echo "Active debug sessions: $DEBUG_COUNT"
    fi
}

calculate_progress() {
    # Calculate progress bar
    local completed=$1
    local total=$2
    
    [ "$total" -eq 0 ] && total=1
    [ "$completed" -gt "$total" ] && completed=$total
    
    local percentage=$((completed * 100 / total))
    local filled=$((percentage / 10))
    local empty=$((10 - filled))
    
    printf "["
    for ((i=0; i<filled; i++)); do printf "█"; done
    for ((i=0; i<empty; i++)); do printf "░"; done
    printf "] %d%%" "$percentage"
}

# Step 5: Present rich status report
present_status() {
    echo ""
    echo -e "${BLUE}📊 Status Report${NC}"
    echo "════════════════════════════════════════════════════════════"
    echo ""
    
    # Extract project name
    PROJECT_NAME="Unknown"
    if [ -f ".planning/PROJECT.md" ]; then
        PROJECT_NAME=$(grep "What This Is" -A1 .planning/PROJECT.md | tail -1 | sed 's/^[[:space:]]*//' | cut -c1-40)
    fi
    echo -e "${BLUE}# $PROJECT_NAME${NC}"
    echo ""
    
    # Calculate progress
    if [ -f ".planning/STATE.md" ]; then
        COMPLETED=$(grep "Total plans completed:" .planning/STATE.md | head -1 | grep -oE "[0-9]+")
        TOTAL_PLANS=$(grep "Average duration:" .planning/STATE.md 2>/dev/null | wc -l)
        if [ -z "$COMPLETED" ]; then
            COMPLETED=$(ls .planning/phases/*/*-SUMMARY.md 2>/dev/null | wc -l | tr -d ' ')
            TOTAL_PLANS=$(find .planning/phases -name "*-PLAN.md" 2>/dev/null | wc -l | tr -d ' ')
        fi
    else
        COMPLETED=$(ls .planning/phases/*/*-SUMMARY.md 2>/dev/null | wc -l | tr -d ' ')
        TOTAL_PLANS=$(find .planning/phases -name "*-PLAN.md" 2>/dev/null | wc -l | tr -d ' ')
    fi
    
    DEFAULT_TOTAL=10
    [ "$TOTAL_PLANS" -eq 0 ] && TOTAL_PLANS=$DEFAULT_TOTAL
    
    echo -e "Progress: $(calculate_progress $COMPLETED $TOTAL_PLANS) $COMPLETED of $TOTAL_PLANS plans complete"
    
    # Extract profile from config if available
    if [ -f ".planning/config.json" ]; then
        PROFILE=$(cat .planning/config.json | grep -oE '"profile": "[^"]*"' | head -1 | cut -d'"' -f4)
        [ -z "$PROFILE" ] && PROFILE="balanced"
    else
        PROFILE="balanced"
    fi
    echo -e "Profile: $PROFILE"
    echo ""
    
    # Recent work
    echo "Recent work:"
    RECENT=$(ls -lt .planning/phases/*/*-SUMMARY.md 2>/dev/null | head -3 | awk '{print $NF}')
    if [ -n "$RECENT" ]; then
        echo "$RECENT" | while read -r summary; do
            if [ -f "$summary" ]; then
                plan_file="${summary/-SUMMARY.md/-PLAN.md}"
                if [ -f "$plan_file" ]; then
                    plan_name=$(basename "$plan_file")
                    objective=$(grep -A2 "## Objective" "$summary" 2>/dev/null | tail -1 || grep -A2 "## Objective" "$plan_file" 2>/dev/null | tail -1)
                    echo "  • $plan_name: ${objective:-Work completed}"
                fi
            fi
        done
    fi
    echo ""
    
    # Current position
    if [ -n "$PHASE_NUMBER" ] && [ -n "$PHASE_TOTAL" ]; then
        echo "Current position:"
        echo "  Phase: $PHASE_NUMBER of $PHASE_TOTAL"
        
        if [ -n "$PHASE_DIR_NAME" ]; then
            PHASE_DIR_PATH=".planning/phases/$PHASE_DIR_NAME"
            if [ -f "$PHASE_DIR_PATH/CONTEXT.md" ]; then
                echo "  Context: ✓"
            else
                echo "  Context: -"
            fi
        fi
        echo ""
    fi
    
    # Key decisions
    if [ -f ".planning/STATE.md" ]; then
        echo "Key decisions:"
        grep -A50 "## Decisions" .planning/STATE.md | grep "^\- \[" | head -3 | while read -r line; do
            echo "  $line"
        done
        echo ""
    fi
    
    # Blockers
    if [ -f ".planning/STATE.md" ]; then
        BLOCKERS=$(grep -A20 "## Blockers" .planning/STATE.md | grep "^\- " | head -3)
        if [ -n "$BLOCKERS" ]; then
            echo "Blockers/Concerns:"
            echo "$BLOCKERS" | while read -r line; do
                [ -n "$line" ] && echo "  • $line"
            done
            echo ""
        fi
    fi
    
    # Pending todos
    echo "Pending todos: $TODOS_COUNT"
    if [ "$TODOS_COUNT" -gt 0 ]; then
        echo "  - /gsd-check-todos to review"
    fi
    echo ""
    
    # Active debug sessions
    if [ "$DEBUG_COUNT" -gt 0 ]; then
        echo "Active debug sessions: $DEBUG_COUNT"
        echo "  - /gsd-debug to continue"
        echo ""
    fi
    
    # What's next
    echo "What's next:"
    if [ -f ".planning/ROADMAP.md" ]; then
        if [ -n "$PHASE_NUMBER" ] && [ "$PHASE_NUMBER" -lt "$PHASE_TOTAL" ]; then
            NEXT_PHASE=$((PHASE_NUMBER + 1))
            NEXT_PHASE_INFO=$(grep "### Phase $NEXT_PHASE" -A5 .planning/ROADMAP.md | head -6)
            echo "  Phase $NEXT_PHASE approaching"
        elif [ "$PHASE_NUMBER" -eq "$PHASE_TOTAL" ]; then
            echo "  Milestone approaching completion"
        else
            echo "  Continue current phase"
        fi
    else
        echo "  Current phase in progress"
    fi
    echo ""
}

# Step 6: Determine next action
determine_next_action() {
    echo "════════════════════════════════════════════════════════════"
    echo -e "${BLUE}▶ Determining next action...${NC}\n"
    
    local current_phase_dir=""
    local current_phase_num=""
    local current_phase_name=""
    
    # Find current phase
    if [ -n "$PHASE_DIR_NAME" ] && [ -d ".planning/phases/$PHASE_DIR_NAME" ]; then
        current_phase_dir=".planning/phases/$PHASE_DIR_NAME"
        current_phase_num=$PHASE_NUMBER
        current_phase_name=$PHASE_DIR_NAME
    else
        # Get first existing phase directory
        current_phase_dir=$(ls -d .planning/phases/* 2>/dev/null | head -1)
        if [ -n "$current_phase_dir" ]; then
            current_phase_name=$(basename "$current_phase_dir")
            # Try to extract phase number from directory name
            current_phase_num=$(echo "$current_phase_name" | grep -oE "^[0-9]+" | head -1)
        fi
    fi
    
    echo "Current phase: $current_phase_dir"
    
    # Count plans and summaries
    if [ -n "$current_phase_dir" ]; then
        PHASE_PLANS=$(ls -1 "$current_phase_dir"/*-PLAN.md 2>/dev/null | wc -l | tr -d ' ')
        PHASE_SUMMARIES=$(ls -1 "$current_phase_dir"/*-SUMMARY.md 2>/dev/null | wc -l | tr -d ' ')
        PHASE_UAT=$(ls -1 "$current_phase_dir"/*-UAT.md 2>/dev/null | wc -l | tr -d ' ')
        
        echo "Phase has $PHASE_PLANS plans, $PHASE_SUMMARIES summaries, $PHASE_UAT UAT files"
    else
        PHASE_PLANS=0
        PHASE_SUMMARIES=0
        PHASE_UAT=0
        echo "No phase directory found"
    fi
    
    # Check for UAT gaps
    UAT_WITH_GAPS=0
    if [ -n "$current_phase_dir" ]; then
        UAT_WITH_GAPS=$(grep -l "status: diagnosed" "$current_phase_dir"/*-UAT.md 2>/dev/null | wc -l | tr -d ' ')
    fi
    
    echo "UAT gaps: $UAT_WITH_GAPS"
    
    # Route based on conditions
    if [ "$UAT_WITH_GAPS" -gt 0 ]; then
        echo ""
        echo -e "${YELLOW}⚠️  UAT gaps found - Route E${NC}"
        echo ""
        echo "## ⚠ UAT Gaps Found"
        echo ""
        echo "Phase UAT has gaps requiring fixes."
        echo ""
        echo "Command: \`/gsd-plan-phase $current_phase_num --gaps\`"
        echo ""
        echo "**Also available:**"
        echo "- \`/gsd-execute-phase $current_phase_num\` — execute phase plans"
        echo "- \`/gsd-verify-work $current_phase_num\` — run more UAT testing"
        echo ""
        exit 0
    fi
    
    if [ "$PHASE_PLANS" -gt 0 ] && [ "$PHASE_SUMMARIES" -lt "$PHASE_PLANS" ]; then
        # Route A: Unexecuted plan exists
        echo ""
        echo -e "${GREEN}✓ Unexecuted plan - Route A${NC}"
        echo ""
        
        # Find first incomplete plan
        INCOMPLETE_PLAN=""
        for plan in "$current_phase_dir"/*-PLAN.md; do
            if [ -f "$plan" ]; then
                summary="${plan/PLAN.md/SUMMARY.md}"
                if [ ! -f "$summary" ]; then
                    INCOMPLETE_PLAN="$plan"
                    break
                fi
            fi
        done
        
        if [ -n "$INCOMPLETE_PLAN" ]; then
            PLAN_NAME=$(basename "$INCOMPLETE_PLAN")
            OBJECTIVE=$(grep "## Objective" -A2 "$INCOMPLETE_PLAN" 2>/dev/null | tail -1)
            
            echo "## ▶ Next Up"
            echo ""
            echo "**$PLAN_NAME** — ${OBJECTIVE:-Work to be done}"
            echo ""
            echo "Command: \`/gsd-execute-phase $current_phase_num\`"
            echo ""
            echo "\`/new\` first → fresh context window"
            echo ""
            exit 0
        fi
    fi
    
    if [ "$PHASE_SUMMARIES" -eq "$PHASE_PLANS" ] && [ "$PHASE_PLANS" -gt 0 ]; then
        # Phase complete, check milestone
        echo ""
        echo -e "${GREEN}✓ Phase complete - checking milestone${NC}"
        
        # Count total phases in roadmap
        TOTAL_PHASES=$(grep -c "^### Phase" .planning/ROADMAP.md 2>/dev/null || echo "1")
        
        if [ -n "$current_phase_num" ] && [ "$current_phase_num" -lt "$TOTAL_PHASES" ]; then
            # Route C: More phases remain
            NEXT_PHASE=$((current_phase_num + 1))
            NEXT_PHASE_INFO=$(grep "### Phase $NEXT_PHASE" -A2 .planning/ROADMAP.md | head -3)
            NEXT_PHASE_NAME=$(echo "$NEXT_PHASE_INFO" | tail -1 | sed 's/.*- //')
            
            echo "## ✓ Current Phase Complete"
            echo ""
            echo "## ▶ Next Up"
            echo ""
            echo "**Phase $NEXT_PHASE: $NEXT_PHASE_NAME**"
            echo ""
            echo "Command: \`/gsd-discuss-phase $NEXT_PHASE\` — gather context"
            echo ""
            echo "\`/new\` first → fresh context window"
            echo ""
            echo "**Also available:**"
            echo "- \`/gsd-plan-phase $NEXT_PHASE\` — skip discussion"
            echo "- \`/gsd-verify-work $current_phase_num\` — user acceptance test"
            echo ""
            exit 0
        else
            # Route D: Milestone complete
            echo "## 🎉 Milestone Complete"
            echo ""
            echo "All phases finished!"
            echo ""
            echo "Command: \`/gsd-complete-milestone\`"
            echo ""
            echo "\`/new\` first → fresh context window"
            echo ""
            echo "**Also available:**"
            echo "- \`/gsd-verify-work\` — user acceptance test"
            echo ""
            exit 0
        fi
    fi
    
    # Route B: Phase not yet planned
    echo ""
    echo -e "${GREEN}✓ Phase needs planning - Route B${NC}"
    echo ""
    
    if [ -f ".planning/phases/$current_phase_name/CONTEXT.md" ]; then
        echo "## ▶ Next Up"
        echo ""
        echo "**Phase $current_phase_num: $current_phase_name**"
        echo ""
        echo "Command: \`/gsd-plan-phase $current_phase_num\`"
        echo ""
        echo "\`/new\` first → fresh context window"
        echo ""
    else
        echo "## ▶ Next Up"
        echo ""
        echo "**Phase $current_phase_num: $current_phase_name**"
        echo ""
        echo "Command: \`/gsd-discuss-phase $current_phase_num\` — gather context"
        echo ""
        echo "\`/new\` first → fresh context window"
        echo ""
        echo "**Also available:**"
        echo "- \`/gsd-plan-phase $current_phase_num\` — skip discussion, plan directly"
        echo "- \`/gsd-list-phase-assumptions $current_phase_num\` — see OpenCode's assumptions"
        echo ""
    fi
    
    exit 0
}

# Route F: Between milestones
route_between_milestones() {
    echo -e "${YELLOW}Milestone archived - ROUTE F${NC}"
    echo ""
    echo "Ready to start the next milestone cycle."
    echo ""
    echo "Command: \`/gsd-new-milestone\`"
    echo ""
    echo "\`/new\` first → fresh context window"
}

# Main workflow
main() {
    verify_structure
    load_context
    gather_recent_work
    parse_current_position
    present_status
    determine_next_action
}

main "$@"
