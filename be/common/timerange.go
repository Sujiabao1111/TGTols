package common

import "time"

const (
	YYYYMMDDhhmmss = "2006-01-02 15:04:05"
	YYYYMMDD       = "2006-01-02"
)

type TimeRangeIter struct {
	start    time.Time
	end      time.Time
	interval time.Duration
	current  time.Time
	index    int
}

func NewTimerange(start time.Time, end time.Time, interval time.Duration) *TimeRangeIter {
	return &TimeRangeIter{
		start:    start,
		end:      end,
		interval: interval,
		current:  start,
		index:    0,
	}
}

// Next scans for the next time.
func (i *TimeRangeIter) Next() bool {
	var next time.Time
	if i.index == 0 {
		next = i.current
	} else {
		next = i.current.Add(i.interval)
	}

	if i.end.Equal(next) || i.end.After(next) {
		i.current = next
		i.index++
		return true
	}
	return false
}

// Current returns the latest unscanned time.
func (i *TimeRangeIter) Current() time.Time {
	return i.current
}

func GenerateDateRange(startDate, endDate time.Time) []string {
	result := make([]string, 0)
	iter := NewTimerange(startDate, endDate, time.Hour*24)
	for iter.Next() {
		t := iter.Current()
		result = append(result, t.Format(YYYYMMDD))
	}
	return result
}

func ParseDate(someDate string, useLocal bool) (time.Time, error) {
	if useLocal {
		return time.ParseInLocation(YYYYMMDD, someDate, time.Local)
	}
	return time.Parse(YYYYMMDD, someDate)
}

func ParseDateTime(someDate, timeformate string, useLocal bool) (time.Time, error) {
	if useLocal {
		return time.ParseInLocation(timeformate, someDate, time.Local)
	}
	return time.Parse(timeformate, someDate)
}

// begin of day
func Bod(t time.Time) time.Time {
	year, month, day := t.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, t.Location())
}

func Truncate(t time.Time) time.Time {
	return t.Truncate(24 * time.Hour)
}

// begin of month
func Bom(now time.Time) time.Time {
	currentYear, currentMonth, _ := now.Date()
	currentLocation := now.Location()
	firstOfMonth := time.Date(currentYear, currentMonth, 1, 0, 0, 0, 0, currentLocation)
	return firstOfMonth
}
