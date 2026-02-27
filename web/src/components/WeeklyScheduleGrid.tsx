import { useState, useEffect, useCallback, Fragment } from 'react';
import clsx from 'clsx';

// Type Definitions
export type ScheduleGrid = boolean[][];

export interface WeeklyScheduleGridProps {
  grid: ScheduleGrid;
  onChange: (grid: ScheduleGrid) => void;
  disabled?: boolean;
}

// Constants
const DAYS = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
const HOURS = Array.from({ length: 24 }, (_, i) => i);

/**
 * WeeklyScheduleGrid Component
 * 
 * A 7×24 visual grid for selecting sleep/wake schedules with click-drag painting.
 * - Sleep cells are displayed as indigo
 * - Wake cells are displayed as dark slate
 * - Click-drag paints cells with consistent state
 * 
 * @param grid - Current 7×24 schedule grid (true=sleep, false=wake)
 * @param onChange - Callback when grid changes
 * @param disabled - Whether the grid is interactive
 */
export function WeeklyScheduleGrid({ 
  grid, 
  onChange, 
  disabled = false 
}: WeeklyScheduleGridProps) {
  const [isDragging, setIsDragging] = useState(false);
  const [paintMode, setPaintMode] = useState<boolean | null>(null);

  /**
   * Updates a single cell in the grid (immutable)
   */
  const updateCell = useCallback((day: number, hour: number, value: boolean) => {
    const newGrid = grid.map((dayRow, dayIdx) => 
      dayIdx === day 
        ? dayRow.map((hourVal, hourIdx) => 
            hourIdx === hour ? value : hourVal
          )
        : dayRow
    );
    onChange(newGrid);
  }, [grid, onChange]);

  /**
   * Handles mousedown on a grid cell
   */
  const handleCellMouseDown = useCallback((day: number, hour: number) => {
    if (disabled) return;

    setIsDragging(true);
    const newPaintMode = !grid[day][hour]; // Toggle opposite of current
    setPaintMode(newPaintMode);
    updateCell(day, hour, newPaintMode);
  }, [disabled, grid, updateCell]);

  /**
   * Handles mouseenter on a grid cell (for drag painting)
   */
  const handleCellMouseEnter = useCallback((day: number, hour: number) => {
    if (disabled || !isDragging || paintMode === null) return;

    updateCell(day, hour, paintMode);
  }, [disabled, isDragging, paintMode, updateCell]);

  /**
   * Handles mouseup - stops dragging
   */
  const handleMouseUp = useCallback(() => {
    setIsDragging(false);
    setPaintMode(null);
  }, []);

  // Attach document-level mouseup listener to handle drag release outside grid
  useEffect(() => {
    document.addEventListener('mouseup', handleMouseUp);
    return () => {
      document.removeEventListener('mouseup', handleMouseUp);
    };
  }, [handleMouseUp]);

  return (
    <div 
      className="bg-slate-700 rounded-lg p-2 overflow-x-auto"
      role="grid"
      aria-label="Weekly schedule grid"
    >
      {/* Header row with hour labels */}
      <div className="grid grid-cols-[auto_repeat(24,1fr)] gap-px mb-1">
        <div className="bg-slate-800 px-2 py-1 text-xs text-slate-400" />
        {HOURS.map((hour) => (
          <div 
            key={hour} 
            className="bg-slate-800 px-1 py-1 text-xs text-center text-slate-400 min-w-[2rem]"
          >
            {hour}
          </div>
        ))}
      </div>

      {/* Grid rows - one per day */}
      <div className="grid grid-cols-[auto_repeat(24,1fr)] gap-px">
        {DAYS.map((day, dayIdx) => (
            <Fragment key={day}>
            {/* Day label column */}
            <div className="bg-slate-800 px-2 py-1.5 text-sm font-medium text-slate-200 sticky left-0 z-10">
              {day}
            </div>

            {/* Hour cells for this day */}
            {HOURS.map((hour) => {
              const isSleep = grid[dayIdx][hour];
              
              // Determine cell styling based on state
              const cellClasses = clsx(
                'h-7 cursor-pointer transition-colors duration-150 ease-in-out',
                'border border-slate-600/50',
                'hover:brightness-110',
                isSleep 
                  ? 'bg-indigo-600 hover:bg-indigo-500' 
                  : 'bg-slate-900 hover:bg-slate-800',
                isDragging 
                  ? (paintMode === isSleep ? 'brightness-125' : 'brightness-75')
                  : ''
              );

              return (
                <div
                  key={`${dayIdx}-${hour}`}
                  className={cellClasses}
                  role="gridcell"
                  aria-label={`${day} at ${hour}:00 - ${isSleep ? 'Sleep' : 'Wake'}`}
                  onMouseDown={() => handleCellMouseDown(dayIdx, hour)}
                  onMouseEnter={() => handleCellMouseEnter(dayIdx, hour)}
                  onMouseLeave={() => {
                    if (isDragging) {
                      // Reset to current state when leaving cell during drag
                      updateCell(dayIdx, hour, grid[dayIdx][hour]);
                    }
                  }}
                  onMouseUp={() => {
                    handleMouseUp();
                    // Ensure final state is correct
                    updateCell(dayIdx, hour, grid[dayIdx][hour]);
                  }}
                />
              );
            })}
            </Fragment>
        ))}
      </div>
      
      <div className="mt-3 text-xs text-slate-500 flex gap-4 ml-1">
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded bg-slate-900" />
          <span>Wake</span>
        </div>
        <div className="flex items-center gap-1.5">
          <div className="w-3 h-3 rounded bg-indigo-600" />
          <span>Sleep</span>
        </div>
      </div>
    </div>
  );
}

export default WeeklyScheduleGrid;
