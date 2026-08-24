package services

import (
	"testing"
	"time"

	"gogogo/models/dtos"
)

func TestBuildDailyWeeklyChallengeReferenceIDIncludesUserID(t *testing.T) {
	loc := activityTimeLocation()
	periodStart := time.Date(2026, 6, 4, 0, 0, 0, 0, loc)

	first := buildDailyWeeklyChallengeReferenceID(900, "daily", 1, periodStart)
	second := buildDailyWeeklyChallengeReferenceID(901, "daily", 1, periodStart)

	if first == second {
		t.Fatalf("expected unique reference ids per user, got %q", first)
	}

	want := "daily-weekly-challenge-900-daily-1-20260604000000"
	if first != want {
		t.Fatalf("expected %q, got %q", want, first)
	}
}

func TestPlanWeeklySpinWheelTodayTicketReconciliationClearsUnqualifiedTickets(t *testing.T) {
	now := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	rows := []weeklySpinTicketRecord{
		{
			Progress: dtos.UserActivityProgress{ID: 11, DayNumber: 1, CreatedAt: now},
			Data:     dtos.WeeklySpinWheelTicketProgressData{DailyTicketIndex: 1, IsBaseTicket: true},
		},
	}

	keepRows, deleteRows := planWeeklySpinWheelTodayTicketReconciliation(rows, 0)

	if len(keepRows) != 0 {
		t.Fatalf("expected no kept rows, got %d", len(keepRows))
	}
	if len(deleteRows) != 1 || deleteRows[0].Progress.ID != 11 {
		t.Fatalf("expected row 11 to be deleted, got %#v", deleteRows)
	}
}

func TestPlanWeeklySpinWheelTodayTicketReconciliationKeepsBaseTicketAndRenumbers(t *testing.T) {
	now := time.Date(2026, 6, 5, 12, 0, 0, 0, time.UTC)
	rows := []weeklySpinTicketRecord{
		{
			Progress: dtos.UserActivityProgress{ID: 21, DayNumber: 2, CreatedAt: now.Add(2 * time.Minute)},
			Data:     dtos.WeeklySpinWheelTicketProgressData{DailyTicketIndex: 2, IsBaseTicket: false},
		},
		{
			Progress: dtos.UserActivityProgress{ID: 22, DayNumber: 1, CreatedAt: now},
			Data:     dtos.WeeklySpinWheelTicketProgressData{DailyTicketIndex: 1, IsBaseTicket: true},
		},
		{
			Progress: dtos.UserActivityProgress{ID: 23, DayNumber: 3, CreatedAt: now.Add(3 * time.Minute)},
			Data:     dtos.WeeklySpinWheelTicketProgressData{DailyTicketIndex: 3, IsBaseTicket: false},
		},
	}

	keepRows, deleteRows := planWeeklySpinWheelTodayTicketReconciliation(rows, 1)

	if len(keepRows) != 1 {
		t.Fatalf("expected 1 kept row, got %d", len(keepRows))
	}
	if keepRows[0].Progress.ID != 22 {
		t.Fatalf("expected row 22 to remain, got %d", keepRows[0].Progress.ID)
	}
	if keepRows[0].Progress.DayNumber != 1 || keepRows[0].Data.DailyTicketIndex != 1 {
		t.Fatalf("expected remaining row to be renumbered to 1, got day=%d index=%d", keepRows[0].Progress.DayNumber, keepRows[0].Data.DailyTicketIndex)
	}
	if !keepRows[0].Data.IsBaseTicket {
		t.Fatalf("expected remaining row to stay the base ticket")
	}
	if len(deleteRows) != 2 {
		t.Fatalf("expected 2 deleted rows, got %d", len(deleteRows))
	}
	if deleteRows[0].Progress.ID != 21 || deleteRows[1].Progress.ID != 23 {
		t.Fatalf("expected rows 21 and 23 to be deleted, got %#v", deleteRows)
	}
}
