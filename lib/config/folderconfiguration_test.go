// Copyright (C) 2026 The Syncthing Authors.
//
// This Source Code Form is subject to the terms of the Mozilla Public
// License, v. 2.0. If a copy of the MPL was not distributed with this file,
// You can obtain one at https://mozilla.org/MPL/2.0/.

package config

import (
	"slices"
	"testing"
	"time"
)

func TestSanitizeScheduledScanTimes(t *testing.T) {
	cases := []struct {
		in  []string
		out []string
	}{
		{nil, nil},
		{[]string{}, nil},
		{[]string{"not-a-time"}, nil},
		{[]string{"04:00"}, []string{"04:00"}},
		{[]string{"16:00", "04:00", "12:00"}, []string{"04:00", "12:00", "16:00"}},
		{[]string{"04:00", "04:00", " 04:00 "}, []string{"04:00"}},
		{[]string{"4:0", "04:00", "bogus", "25:00", "12:61"}, []string{"04:00"}},
	}
	for _, tc := range cases {
		got := sanitizeScheduledScanTimes(tc.in)
		if !slices.Equal(got, tc.out) {
			t.Errorf("sanitizeScheduledScanTimes(%v) = %v, want %v", tc.in, got, tc.out)
		}
	}
}

func TestFolderConfigurationNextScheduledScan(t *testing.T) {
	loc := time.UTC
	now := time.Date(2026, 9, 23, 10, 30, 0, 0, loc)

	t.Run("no times configured", func(t *testing.T) {
		f := FolderConfiguration{}
		if _, ok := f.NextScheduledScan(now); ok {
			t.Error("expected no schedule when ScheduledScanTimes is empty")
		}
	})

	t.Run("next time later today", func(t *testing.T) {
		f := FolderConfiguration{ScheduledScanTimes: []string{"04:00", "12:00", "16:00"}}
		d, ok := f.NextScheduledScan(now)
		if !ok {
			t.Fatal("expected a schedule")
		}
		want := 90 * time.Minute // 10:30 -> 12:00
		if d != want {
			t.Errorf("got %v, want %v", d, want)
		}
	})

	t.Run("wraps to tomorrow when all times passed", func(t *testing.T) {
		f := FolderConfiguration{ScheduledScanTimes: []string{"04:00", "09:00"}}
		d, ok := f.NextScheduledScan(now)
		if !ok {
			t.Fatal("expected a schedule")
		}
		want := 17*time.Hour + 30*time.Minute // 10:30 -> 04:00 next day
		if d != want {
			t.Errorf("got %v, want %v", d, want)
		}
	})

	t.Run("unsorted input still finds the nearest time", func(t *testing.T) {
		f := FolderConfiguration{ScheduledScanTimes: []string{"23:00", "10:31", "02:00"}}
		d, ok := f.NextScheduledScan(now)
		if !ok {
			t.Fatal("expected a schedule")
		}
		want := time.Minute // 10:30 -> 10:31
		if d != want {
			t.Errorf("got %v, want %v", d, want)
		}
	})
}
