package main

import "testing"

func TestAuditRejectNonNumericPreferenceLiterals(t *testing.T) {
	for _, key := range []string{"cardlounge.music.volume", "cardlounge.sound.volume", "cardlounge.lounge.volume", "cardlounge.lounge.rate"} {
		for _, value := range []string{"null", "true", "false", "\"25\"", "0x10", "+25", "025", "NaN", "Infinity"} {
			if validPreference(key, value) {
				t.Errorf("accepted non-number %s=%q", key, value)
			}
		}
	}
	for _, value := range []string{"25", "0", "100", "2.5e1", " 25 "} {
		if !validPreference("cardlounge.music.volume", value) {
			t.Errorf("rejected legitimate number %q", value)
		}
	}
}

func TestAuditBaccaratRolePersistsBeforeFirstDeal(t *testing.T) {
	a := NewBaccaratApp()
	for _, host := range []string{"misaki", "ayaka", "rina", "shizue", "kathy"} {
		if _, err := a.OpenGame("banque", host); err != nil {
			t.Fatal(err)
		}
		if _, err := a.SetRole(true); err != nil {
			t.Fatal(err)
		}
		if _, err := a.OpenGame("punto", host); err != nil {
			t.Fatal(err)
		}
		view, err := a.OpenGame("banque", host)
		if err != nil {
			t.Fatal(err)
		}
		if view.Role != "banker" {
			t.Errorf("%s: reopening undelivered session lost banker role", host)
		}
		if reset := a.Reset(); reset.Role != "banker" {
			t.Errorf("%s: reset lost banker role", host)
		}
	}
}
