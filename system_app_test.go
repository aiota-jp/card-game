package main

import "testing"

func TestPreferencesBackupValidation(t *testing.T) {
	for _, raw := range []string{
		`{"format":"CardGameLoungeBackup/1","settings":{"cardlounge.lounge.voice":"true","cardlounge.lounge.rate":"1.794","cardlounge.music.volume":"25","poker.tutorial.completed.v1.holdem":"1"}}`,
		`{"format":"CardGameLoungeBackup/1","settings":{}}`,
	} {
		if err := validatePreferences(raw); err != nil {
			t.Fatal(err)
		}
	}
	for _, raw := range []string{
		`{"format":"ShogiBackup/1","settings":{}}`,
		`{"format":"CardGameLoungeBackup/1","settings":{"cardlounge.music.volume":"101"}}`,
		`{"format":"CardGameLoungeBackup/1","settings":{"cardlounge.lounge.rate":"NaN"}}`,
		`{"format":"CardGameLoungeBackup/1","settings":{"cardlounge.lounge.voice":"yes"}}`,
		`{"format":"CardGameLoungeBackup/1","settings":{"unrelated":"1"}}`,
		`{"format":"CardGameLoungeBackup/1","settings":null}`,
		`not json`,
	} {
		if validatePreferences(raw) == nil {
			t.Fatalf("accepted invalid backup %s", raw)
		}
	}
}
