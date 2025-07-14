package bittorrent

import "testing"


func TestByteToString(t *testing.T) {
	oneKByte := 1024;

	oneKByteAndBytes := oneKByte + 102;

	oneMByte := oneKByte * 1024;

	oneMByteAndBytes := oneMByte + 102;

	oneGByte := oneMByte * 1024;

	res := bytesToString(oneKByte)

	if res != "1.0 KiB" {
		t.Errorf("expected 1.0 KiB. got: %s", res)
	}

	res = bytesToString(oneKByteAndBytes)

	if res != "1.1 KiB" {
		t.Errorf("expected 1.1 KiB. got: %s", res)
	}


	res = bytesToString(oneMByte)

	if res != "1.0 MiB" {
		t.Errorf("expected 1.0 MiB. got: %s", res)
	}

	res = bytesToString(oneMByteAndBytes)

	if res != "1.0 MiB" {
		t.Errorf("expected 1.0 MiB. got: %s", res)
	}

	res = bytesToString(oneGByte)

	if res != "1.0 GiB" {
		t.Errorf("expected 1.0 GiB. got: %s", res)
	}

}