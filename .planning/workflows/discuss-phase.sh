#!/bin/bash
set -e

# Discuss Phase Workflow
# Extract implementation decisions for a phase

cd "$(dirname "$0")/../.."

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

PHASE="${1:-}"

# Validate phase
validate_phase() {
    if [ -z "$PHASE" ]; then
        echo -e "${RED}Phase number required${NC}"
        echo "Usage: ./discuss-phase.sh <phase-number>"
        exit 1
    fi
    
    if ! grep -q "^### Phase $PHASE:" .planning/ROADMAP.md 2>/dev/null; then
        echo -e "${RED}Phase $PHASE not found in roadmap${NC}"
        exit 1
    fi
    
    echo -e "${GREEN}✓ Phase $PHASE validated${NC}"
}

# Check existing context
check_existing() {
    local PADDED=$(printf "%02d" $PHASE)
    
    # Check for both padded and unpadded directory patterns
    local existing=$(find .planning/phases -name "${PADDED}-CONTEXT.md" -o -name "*-CONTEXT.md" 2>/dev/null | grep -E "(^|/)${PADDED}(-|$)" | head -1)
    
    if [ -n "$existing" ] && [ -f "$existing" ]; then
        echo -e "${YELLOW}Found existing context: $existing${NC}"
        echo ""
        echo "Options:"
        echo "1. Update it"
        echo "2. View it"
        echo "3. Skip"
        echo ""
        exit 0
    fi
    
    echo -e "${GREEN}No existing context - creating new${NC}"
}

# Analyze phase for gray areas
analyze_phase() {
    if [ -f ".planning/ROADMAP.md" ]; then
        echo -e "${BLUE}Analyzing Phase $PHASE...${NC}\n"
        
        # Extract phase goal
        local goal=$(grep "Phase $PHASE:" -A5 .planning/ROADMAP.md | grep "Goal" | sed 's/.*Goal: //')
        local name=$(grep "### Phase $PHASE:" .planning/ROADMAP.md | sed 's/.*: //')
        
        echo "Phase $PHASE: $name"
        echo "Goal: $goal"
        echo ""
        
        # Analyze gray areas based on phase domain
        case $PHASE in
            1)
                # Multi-cloud discovery - already complete, no new discussion needed
                echo -e "${YELLOW}Phase 1 already complete. No discussion needed.${NC}"
                ;;
            2)
                # Manual Control & Audit - UI for bulk operations and audit log
                echo "Gray areas for Phase 2 (Manual Control & Audit):"
                echo ""
                echo "1. Audit log UI layout - Card view vs table view vs timeline?"
                echo "   - What info to show per event? Timestamp, user, action, details?"
                echo "   - Filter capabilities needed? Date range, action type, instance?"
                echo "   - Pagination vs infinite scroll?"
                echo ""
                echo "2. Confirmation dialog behavior - When to show, what to include?"
                echo "   - Confirmation text style and detail level?"
                echo "   - Auto-cancel after timeout?"
                echo "   - 'Do not ask again' option?"
                echo ""
                echo "3. Bulk operations feedback - How to show progress and results?"
                echo "   - Inline progress indicators vs toast notifications?"
                echo "   - Success/failure reporting per instance?"
                echo "   - Summary view after bulk operation completes?"
                echo ""
                ;;
            3)
                # Basic Scheduling - UI for schedule creation
                echo "Gray areas for Phase 3 (Basic Scheduling):"
                echo ""
                echo "1. Schedule creation flow - Wizard vs form vs quick action?"
                echo "   - How many steps to create a schedule?"
                echo "   - Pre-filled values from selection context?"
                echo "   - Preview of affected instances before creating?"
                echo ""
                echo "2. Time picker UX - Custom widget or native datetime?" 
                echo "   - Time zone handling strategy?"
                echo "   - Recurrence patterns (daily, weekly, custom)?"
                echo "   - Preset options (e.g., 'Weekdays 9am-9pm')?"
                echo ""
                echo "3. Schedule list layout - Cards vs table vs collapsible list?"
                echo "   - What metadata displays per schedule?"
                echo "   - Sorting and filtering options?"
                echo "   - Color coding by type or status?"
                echo ""
                ;;
            4)
                # Advanced Schedule Filtering - Regex-based filters
                echo "Gray areas for Phase 4 (Advanced Schedule Filtering):"
                echo ""
                echo "1. Filter builder UI - Form vs code editor vs drag-drop?"
                echo "   - Visual builder with dropdowns or regex text input?"
                echo "   - Operator selection (AND/OR) placement?"
                echo "   - Live preview of matched instances?"
                echo ""
                echo "2. Regex pattern documentation - Help text, examples, validation?"
                echo "   - Inline documentation for regex syntax?"
                echo "   - Pre-built common patterns button?"
                echo "   - Validation feedback and error messages?"
                echo ""
                echo "3. Preview behavior - When and how to show preview?"
                echo "   - Auto-preview as user types or explicit 'Preview' button?"
                echo "   - How many results to show in preview?"
                echo "   - Preview pagination?"
                echo ""
                ;;
            5)
                # Activity Analysis - Metrics and pattern detection
                echo "Gray areas for Phase 5 (Activity Analysis):"
                echo ""
                echo "1. Metrics visualization - Charts, tables, or cards?"
                echo "   - Which metrics to show per instance?"
                echo "   - Time range selector and presets?"
                echo "   - Aggregation options (daily, hourly, weekly)?"
                echo ""
                echo "2. Pattern detection display - How to show patterns?"
                echo "   - Highlighted ranges on charts?"
                echo "   - List of detected patterns?"
                echo "   - Confidence scores for detected patterns?"
                echo ""
                echo "3. Inactivity alerting - When and how to show inactivity?"
                echo "   - Threshold configuration (hours, days)?"
                echo "   - Visual indicators on instance cards?"
                echo "   - Notification system integration?"
                echo ""
                ;;
            6)
                # Intelligent Recommendations - AI-based suggestions
                echo "Gray areas for Phase 6 (Intelligent Recommendations):"
                echo ""
                echo "1. Recommendation card design - What info per card?"
                echo "   - Sleep/wake times displayed prominently?"
                echo "   - Expected savings calculation and display?"
                echo "   - Confidence score and explanation?"
                echo ""
                echo "2. Recommendation workflow - How to review and apply?"
                echo "   - Single card view vs batch review?"
                echo "   - Apply to schedule vs create one-off?"
                echo "   - 'Apply all' vs each card action?"
                echo ""
                echo "3. Scheduling flexibility - User control over AI recommendations?"
                echo "   - Edit times before applying?"
                echo "   - Temporary override options?"
                echo "   - Adjust confidence threshold for showing?"
                echo ""
                ;;
            *)
                echo -e "${YELLOW}Phase $PHASE - needs analysis${NC}"
                ;;
        esac
    fi
}

# Present gray areas for selection
present_gray_areas() {
    echo -e "${BLUE}Which areas do you want to discuss?${NC}\n"
    echo "1. [Specific area]" - description
    echo "2. [Specific area]" - description
    echo "3. [Specific area]" - description
    echo "4. [Specific area]" - description
    echo ""
    echo "Enter numbers (comma-separated): "
}

# Main workflow
main() {
    echo ""
    echo "╔════════════════════════════════════════════════════════════╗"
    echo "║  DISCUSS PHASE WORKFLOW                                     ║"
    echo "╚════════════════════════════════════════════════════════════╝"
    echo ""
    
    validate_phase
    check_existing
    analyze_phase
    present_gray_areas
    
    echo "Selection: "
    read -r SELECTION
    
    echo ""
    echo "Processing $SELECTION..."
    echo ""
}

main "$@"
