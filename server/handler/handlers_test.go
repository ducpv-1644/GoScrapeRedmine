package handler

import (
	"testing"
)

type UserHandlerTest struct{}

func TestQuarterRange(t *testing.T) {
	dayStart, dayend := quarterRange(2021, 3)
	if dayStart != "07/01/2021" && dayend != "09/30/2021" {
		t.Errorf("%s %s", dayStart, dayend)
	}

}

func TestWeekRange(t *testing.T) {
	dayStart, dayend := weekRange(2021, 34)
	if dayStart != "08/23/2021" && dayend != "08/29/2021" {
		t.Errorf("%s %s", dayStart, dayend)
	}

}
