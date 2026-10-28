package main

import (
	"strings"
	"testing"
)

func TestAdditionalHostsAndBackup(t *testing.T) {
	for _, host := range []string{"shizue", "kathy"} {
		if !validPreference("cardlounge.lounge.character", host) {
			t.Fatal("backup rejects host", host)
		}
		app := NewBaccaratApp()
		for _, variant := range []string{"punto", "chemin", "banque"} {
			view, err := app.OpenGame(variant, host)
			if err != nil {
				t.Fatal(err)
			}
			if len(view.Seats) != 4 || view.Seats[1].ID != host || view.Seats[1].Name != strings.ToUpper(host) {
				t.Fatalf("wrong four-seat host: %+v", view.Seats)
			}
			if variant == "banque" {
				if _, err = app.SetRole(true); err != nil {
					t.Fatal(err)
				}
			}
			if _, err = app.Deal(20, "player"); err != nil {
				t.Fatal(variant, err)
			}
			for i := 0; i < 10; i++ {
				view = app.Snapshot()
				if view.Done {
					break
				}
				action := "stand"
				if view.Phase == "dealt" {
					action = "continue"
				}
				if _, err = app.Act(action); err != nil {
					t.Fatal(variant, view.Phase, err)
				}
			}
			if !app.Snapshot().Done {
				t.Fatal("hand did not settle", variant)
			}
			saved := app.Snapshot()
			app.OpenGame(variant, "misaki")
			resumed, err := app.OpenGame(variant, host)
			if err != nil || resumed.Chips != saved.Chips || resumed.Hand != saved.Hand {
				t.Fatal("host session lost", host, variant)
			}
			reset := app.Reset()
			if reset.Seats[1].ID != host {
				t.Fatal("reset lost host", host)
			}
		}
	}
}
