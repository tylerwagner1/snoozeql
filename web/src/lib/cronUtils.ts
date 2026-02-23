/**
 * CRON Conversion Utilities
 * 
 * Utilities to convert between the visual 7×24 grid representation
 * and CRON expressions for schedule storage and execution.
 */

// Type Definitions

export type ScheduleGrid = boolean[][];

/**
 * Creates an empty 7×24 grid where all cells are false (wake state)
 */
export function createEmptyGrid(): ScheduleGrid {
  return Array(7).fill(null).map(() => Array(24).fill(false));
}

/**
 * Formats a hour as a time string
 * @param hour - Hour value (0-23)
 * @param use24h - Whether to use 24-hour format (default: false)
 * @returns Formatted time string (e.g., "10pm" or "22:00")
 */
export function formatHour(hour: number, use24h: boolean = false): string {
  if (use24h) {
    return `${hour.toString().padStart(2, '0')}:00`;
  }
  
  const suffix = hour >= 12 ? 'pm' : 'am';
  const displayHour = hour === 0 ? 12 : hour > 12 ? hour - 12 : hour;
  return `${displayHour}${suffix}`;
}

/**
 * Gets the abbreviated day name for a grid index
 * @param dayIndex - Grid index (0=Monday, 6=Sunday)
 * @returns Abbreviated day name
 */
export function getDayName(dayIndex: number): string {
  const days = ['Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat', 'Sun'];
  return days[dayIndex] || '';
}

/**
 * Converts a grid to CRON expressions
 * 
 * For simplicity in Phase 3, this assumes a single contiguous sleep window per day
 * and generates a single sleep/wake CRON pair based on the majority pattern across days.
 * 
 * @param grid - 7×24 schedule grid (true = sleep, false = wake)
 * @returns { sleepCron: string, wakeCron: string } or null if no sleep hours
 */
export function gridToCron(grid: ScheduleGrid): { sleepCron: string; wakeCron: string } | null {
  // Find all sleep hours per day
  const daySleepPatterns = grid.map((dayGrid) => {
    const sleepHours: number[] = [];
    dayGrid.forEach((isSleep, hour) => {
      if (isSleep) sleepHours.push(hour);
    });
    return sleepHours;
  });

  // Check if there are any sleep hours at all
  const hasSleep = daySleepPatterns.some((hours) => hours.length > 0);
  if (!hasSleep) {
    return null;
  }

  // Find the most common sleep pattern across days
  // For simplicity: find first contiguous sleep block in the most common pattern
  const patternCounts = new Map<string, number>();
  
  daySleepPatterns.forEach((hours) => {
    if (hours.length === 0) return;
    
    // Create a simple string representation of the pattern
    // Format: "start-end" for contiguous blocks, or "multiple" for non-contiguous
    const hasMultipleBlocks = !isContiguousHours(hours);
    const key = hasMultipleBlocks 
      ? 'multiple' 
      : `${hours[0]}-${hours[hours.length - 1]}`;
    
    patternCounts.set(key, (patternCounts.get(key) || 0) + 1);
  });

  // Get the most common pattern
  let sleepStart = 22; // Default: 10pm
  let sleepEnd = 7;    // Default: 7am

  if (patternCounts.size > 0) {
    const mostCommon = Array.from(patternCounts.entries()).sort(
      ([, a], [, b]) => b - a
    )[0][0];

    if (mostCommon !== 'multiple') {
      const [startStr, endStr] = mostCommon.split('-');
      sleepStart = parseInt(startStr, 10);
      sleepEnd = parseInt(endStr, 10);
    }
  }

  // Convert to CRON format
  // Days mapping: grid[0]=Monday → CRON day 1, grid[6]=Sunday → CRON day 0
  // For weekdays (Mon-Fri), we use days 1-5
  const sleepDays = daySleepPatterns.map((hours, dayIdx) => 
    hours.length > 0 ? convertGridDayToCronDay(dayIdx) : null
  ).filter((day): day is number => day !== null);

  const uniqueDays = Array.from(new Set(sleepDays)).sort((a, b) => a - b);
  const daysCron = uniqueDays.join(',');

  // CRON format: minute hour * * day-of-week
  // We use minute 0 for simplicity
  const sleepCron = `0 ${sleepStart} * ${daysCron}`;
  const wakeCron = `0 ${sleepEnd} * ${daysCron}`;

  return { sleepCron, wakeCron };
}

/**
 * Checks if hours form a contiguous block
 */
function isContiguousHours(hours: number[]): boolean {
  if (hours.length <= 1) return true;
  
  for (let i = 1; i < hours.length; i++) {
    if (hours[i] !== hours[i - 1] + 1) {
      return false;
    }
  }
  return true;
}

/**
 * Converts grid day index to CRON day number
 * @param gridDay - Grid index (0=Monday, 6=Sunday)
 * @returns CRON day (0=Sunday, 1=Monday, ..., 6=Saturday)
 */
function convertGridDayToCronDay(gridDay: number): number {
  // grid: [Mon=0, Tue=1, Wed=2, Thu=3, Fri=4, Sat=5, Sun=6]
  // cron: [Sun=0, Mon=1, Tue=2, Wed=3, Thu=4, Fri=5, Sat=6]
  // Mapping: gridDay → (gridDay + 1) % 7
  return (gridDay + 1) % 7;
}

/**
 * Parses a CRON expression and populates a grid
 * 
 * @param sleepCron - Sleep CRON expression (e.g., "0 22 * * 1-5")
 * @param wakeCron - Wake CRON expression (e.g., "0 7 * * 1-5")
 * @returns Populated 7×24 schedule grid
 */
export function cronToGrid(sleepCron: string, wakeCron: string): ScheduleGrid {
  const grid = createEmptyGrid();

  try {
    const sleepParts = parseCronPart(sleepCron);
    const wakeParts = parseCronPart(wakeCron);

    if (!sleepParts || !wakeParts) {
      return grid;
    }

    const { hour: sleepHour, days: sleepDays } = sleepParts;
    const { hour: wakeHour, days: wakeDays } = wakeParts;

    // Get all days to mark (union of sleep and wake days)
    const allDays = new Set([...sleepDays, ...wakeDays]);

    allDays.forEach((cronDay) => {
      const gridDay = convertCronDayToGridDay(cronDay);
      
      // Handle overnight schedules (sleep hour > wake hour)
      if (sleepHour > wakeHour) {
        // Sleep from sleepHour to 23:59, then 00:00 to wakeHour-1
        for (let h = sleepHour; h < 24; h++) {
          grid[gridDay][h] = true;
        }
        for (let h = 0; h < wakeHour; h++) {
          grid[gridDay][h] = true;
        }
      } else {
        // Regular schedule: sleep from sleepHour to wakeHour-1
        for (let h = sleepHour; h < wakeHour; h++) {
          grid[gridDay][h] = true;
        }
      }
    });
  } catch (error) {
    // On parse error, return empty grid
    console.warn('Failed to parse CRON expressions:', error);
  }

  return grid;
}

/**
 * Parses a single CRON expression part
 */
function parseCronPart(cron: string): { hour: number; days: number[] } | null {
  const parts = cron.trim().split(/\s+/);
  
  if (parts.length < 3) {
    return null;
  }

  // parts[0] = minute, parts[1] = hour, parts[4] = day-of-week
  const hour = parseInt(parts[1], 10);
  const days = parseDaysCron(parts[4] || '*');

  return { hour, days };
}

/**
 * Parses the day-of-week field from a CRON expression
 */
function parseDaysCron(daysCron: string): number[] {
  if (daysCron === '*') {
    // Every day: 0-6
    return [0, 1, 2, 3, 4, 5, 6];
  }

  const days: number[] = [];
  
  daysCron.split(',').forEach((dayStr) => {
    const day = parseInt(dayStr, 10);
    if (!isNaN(day) && day >= 0 && day <= 6) {
      days.push(day);
    }
  });

  return days;
}

/**
 * Converts CRON day number to grid day index
 * @param cronDay - CRON day (0=Sunday, 6=Saturday)
 * @returns Grid index (0=Monday, 6=Sunday)
 */
function convertCronDayToGridDay(cronDay: number): number {
  // cron: [Sun=0, Mon=1, Tue=2, Wed=3, Thu=4, Fri=5, Sat=6]
  // grid: [Mon=0, Tue=1, Wed=2, Thu=3, Fri=4, Sat=5, Sun=6]
  // Mapping: cronDay → (cronDay + 6) % 7
  return (cronDay + 6) % 7;
}

/**
 * Creates a human-readable summary of a grid
 * 
 * @param grid - 7×24 schedule grid
 * @returns { activeDays: string, sleepHours: string }
 */
export function formatGridSummary(grid: ScheduleGrid): { activeDays: string; sleepHours: string } {
  // Find active days (days with any sleep hours)
  const activeDaysIndices: number[] = [];
  
  grid.forEach((dayGrid, dayIdx) => {
    if (dayGrid.some((hour) => hour)) {
      activeDaysIndices.push(dayIdx);
    }
  });

  if (activeDaysIndices.length === 0) {
    return {
      activeDays: 'No active days',
      sleepHours: 'No sleep hours'
    };
  }

  // Format active days
  let activeDays = formatActiveDays(activeDaysIndices);

  // Find common sleep hours across all active days
  const sleepHoursString = formatSleepHours(grid, activeDaysIndices);

  return {
    activeDays,
    sleepHours: sleepHoursString || 'No sleep hours'
  };
}

/**
 * Formats the active days as a human-readable string
 */
function formatActiveDays(days: number[]): string {
  if (days.length === 0) return 'No active days';
  if (days.length === 1) return getDayName(days[0]);

  // Check if weekdays (Mon-Fri)
  const isWeekdays = days.length === 5 && 
    days.includes(0) && days.includes(1) && days.includes(2) && 
    days.includes(3) && days.includes(4);

  if (isWeekdays) {
    return 'Weekdays';
  }

  // Check if weekends (Sat-Sun)
  const isWeekends = days.length === 2 && 
    days.includes(5) && days.includes(6);

  if (isWeekends) {
    return 'Weekends';
  }

  // Check if every day
  if (days.length === 7) {
    return 'Every day';
  }

  // Fall back to listing days
  return days.map((dayIdx) => getDayName(dayIdx)).join(', ');
}

/**
 * Formats the sleep hours as a human-readable string
 */
function formatSleepHours(grid: ScheduleGrid, activeDaysIndices: number[]): string {
  // Find the most common sleep pattern across active days
  const hourSleepCounts = Array(24).fill(0);
  
  activeDaysIndices.forEach((dayIdx) => {
    grid[dayIdx].forEach((isSleep, hour) => {
      if (isSleep) hourSleepCounts[hour]++;
    });
  });

  const maxCount = Math.max(...hourSleepCounts);
  if (maxCount === 0) {
    return '';
  }

  // Find contiguous sleep blocks
  const activeHours: number[] = [];
  const isSleepThreshold = Math.ceil(activeDaysIndices.length / 2); // Majority

  for (let h = 0; h < 24; h++) {
    if (hourSleepCounts[h] >= isSleepThreshold) {
      activeHours.push(h);
    }
  }

  if (activeHours.length === 0) {
    return '';
  }

  // Find contiguous blocks
  let sleepStart = activeHours[0];
  let sleepEnd = activeHours[activeHours.length - 1];

  // Handle overnight schedules
  const isOvernight = sleepStart > sleepEnd || (sleepEnd - sleepStart) > 12;

  // If we have a simple contiguous range
  if (!isOvernight && sleepEnd - sleepStart < 12) {
    return `${formatHour(sleepStart)}-${formatHour(sleepEnd)}`;
  }

  // Forovernight or complex patterns
  return `${formatHour(sleepStart)}-${formatHour(sleepEnd)}`;
}
