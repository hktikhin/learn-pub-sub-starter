package main

import (
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	type testCase struct {
		gamelog    GameLog
		encodedHex string
	}

	runCases := []testCase{
		{
			gamelog: GameLog{
				CurrentTime: time.Date(1776, 1, 1, 0, 0, 0, 0, time.UTC),
				Message:     "Crossed the Delaware",
				Username:    "Washington",
			},
			encodedHex: "3e7f0301010747616d654c6f6701ff80000103010b43757272656e7454696d6501ff820001074d657373616765010c000108557365726e616d65010c00000010ff810501010454696d6501ff8200000036ff80010f010000000d0aaaf98000000000ffff011443726f73736564207468652044656c6177617265010a57617368696e67746f6e00",
		},
	}

	submitCases := append(runCases, []testCase{
		{
			gamelog: GameLog{
				CurrentTime: time.Date(1804, 12, 2, 0, 0, 0, 0, time.UTC),
				Message:     "Crowned Emperor",
				Username:    "Napoleon",
			},
			encodedHex: "3e7f0301010747616d654c6f6701ff80000103010b43757272656e7454696d6501ff820001074d657373616765010c000108557365726e616d65010c00000010ff810501010454696d6501ff820000002fff80010f010000000d410f7c8000000000ffff010f43726f776e656420456d7065726f7201084e61706f6c656f6e00",
		},
	}...)

	testCases := runCases
	if withSubmit {
		testCases = submitCases
	}

	skipped := len(submitCases) - len(testCases)

	var passed, failed int
	for _, test := range testCases {
		encoded, err := encode(test.gamelog)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}
		encodedHex := fmt.Sprintf("%x", encoded)
		decoded, err := decode(encoded)
		if err != nil {
			t.Fatalf("decode failed: %v", err)
		}
		if decoded != test.gamelog {
			failed++
			t.Errorf(`---------------------------------
Test Decode Failed:
  gamelog: %v
=>
  expected encoded hex: %v
  actual encoded hex: %v
  expected decoded gamelog: %v
  actual decoded gamelog: %v
`,
				test.gamelog,
				test.encodedHex,
				encodedHex,
				test.gamelog,
				decoded,
			)
		} else if encodedHex != test.encodedHex {
			failed++
			t.Errorf(`---------------------------------
Test Encode Failed:
  gamelog: %v
=>
  expected encoded hex: %v
  actual encoded hex: %v
  expected decoded gamelog: %v
  actual decoded gamelog: %v
`,
				test.gamelog,
				test.encodedHex,
				encodedHex,
				test.gamelog,
				decoded,
			)
		} else {
			passed++
			fmt.Printf(`---------------------------------
Test Passed:
  gamelog: %v
=>
  expected encoded hex: %v
  actual encoded hex: %v
  expected decoded gamelog: %v
  actual decoded gamelog: %v
`,
				test.gamelog,
				test.encodedHex,
				encodedHex,
				test.gamelog,
				decoded,
			)
		}
	}

	fmt.Println("---------------------------------")
	if skipped > 0 {
		fmt.Printf("%d passed, %d failed, %d skipped\n", passed, failed, skipped)
	} else {
		fmt.Printf("%d passed, %d failed\n", passed, failed)
	}
}

// withSubmit is set at compile time depending
// on which button is used to run the tests
var withSubmit = true
