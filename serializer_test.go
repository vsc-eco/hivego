package hivego

import (
	"bytes"
	"testing"
)

func TestOpIdB(t *testing.T) {
	got := opIdB("custom_json")
	expected := byte(18)

	if got != expected {
		t.Error("Expected", expected, "got")
	}
}

func TestRefBlockNumB(t *testing.T) {
	got := refBlockNumB(36029)
	expected := []byte{189, 140}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestRefBlockPrefixB(t *testing.T) {
	got := refBlockPrefixB(1164960351)
	expected := []byte{95, 226, 111, 69}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestExpTimeB(t *testing.T) {
	got, _ := expTimeB("2016-08-08T12:24:17")
	expected := []byte{241, 121, 168, 87}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestCountOpsB(t *testing.T) {
	got := countOpsB(getTwoTestOps())
	expected := []byte{2}

	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

//func TestExtensionsB

func TestAppendVString(t *testing.T) {
	var buf bytes.Buffer
	got := appendVString("xeroc", &buf)
	expected := []byte{5, 120, 101, 114, 111, 99}
	if !bytes.Equal(got.Bytes(), expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestAppendVStringArray(t *testing.T) {
	var buf bytes.Buffer
	got := appendVStringArray([]string{"xeroc", "piston"}, &buf).Bytes()
	expected := []byte{2, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeTx(t *testing.T) {
	got, _ := SerializeTx(getTestVoteTx())
	expected := []byte{189, 140, 95, 226, 111, 69, 241, 121, 168, 87, 1, 0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39, 0}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOps(t *testing.T) {
	got, _ := serializeOps(getTwoTestOps())
	expected := []byte{2, 0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39, 18, 0, 1, 5, 120, 101, 114, 111, 99, 7, 116, 101, 115, 116, 45, 105, 100, 17, 123, 34, 116, 101, 115, 116, 107, 34, 58, 34, 116, 101, 115, 116, 118, 34, 125}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpVoteOperation(t *testing.T) {
	got, _ := getTestVoteOp().SerializeOp()
	expected := []byte{0, 5, 120, 101, 114, 111, 99, 5, 120, 101, 114, 111, 99, 6, 112, 105, 115, 116, 111, 110, 16, 39}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpCustomJsonOperation(t *testing.T) {
	got, _ := getTestCustomJsonOp().SerializeOp()
	expected := []byte{18, 0, 1, 5, 120, 101, 114, 111, 99, 7, 116, 101, 115, 116, 45, 105, 100, 17, 123, 34, 116, 101, 115, 116, 107, 34, 58, 34, 116, 101, 115, 116, 118, 34, 125}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpAccCreate(t *testing.T) {
	got, _ := getTestAccCreateOp().SerializeOp()
	// A7 fix: appended extensions byte (0x00) at the end
	expected := []byte{9, 0, 0, 0, 0, 0, 0, 0, 0, 35, 32, 188, 190, 8, 109, 105, 108, 111, 45, 104, 112, 114, 5, 115, 97, 103, 97, 114, 1, 0, 0, 0, 0, 1, 2, 10, 101, 192, 10, 6, 132, 7, 65, 238, 81, 177, 178, 164, 187, 202, 162, 70, 26, 198, 248, 227, 102, 116, 96, 8, 245, 232, 159, 143, 49, 25, 233, 1, 0, 1, 0, 0, 0, 0, 1, 2, 230, 133, 92, 84, 69, 147, 205, 33, 156, 229, 25, 24, 141, 64, 161, 175, 30, 167, 60, 77, 126, 254, 76, 187, 23, 100, 49, 125, 227, 69, 253, 62, 1, 0, 1, 0, 0, 0, 0, 1, 3, 218, 87, 20, 133, 59, 128, 38, 218, 28, 118, 84, 21, 96, 239, 240, 79, 44, 186, 162, 251, 54, 44, 172, 207, 105, 111, 223, 252, 39, 33, 155, 17, 1, 0, 2, 162, 67, 136, 113, 93, 177, 204, 115, 66, 216, 45, 167, 57, 203, 32, 27, 157, 79, 56, 22, 31, 105, 89, 95, 114, 222, 224, 88, 67, 229, 237, 104, 0, 0}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpAccountUpdateOperation(t *testing.T) {
	got, _ := getTestAccountUpdateOp().SerializeOp()
	expected := []byte{10, 12, 115, 110, 105, 112, 101, 114, 100, 117, 101, 108, 49, 55, 0, 0, 0, 2, 248, 203, 193, 109, 141, 110, 237, 126, 105, 254, 86, 201, 65, 157, 81, 189, 244, 224, 193, 227, 202, 141, 140, 24, 154, 173, 150, 112, 27, 195, 12, 77, 13, 123, 34, 102, 111, 111, 34, 58, 34, 98, 97, 114, 34, 125}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpTransfer(t *testing.T) {
	got, _ := getTestTransferOp().SerializeOp()
	expected := []byte{2, 10, 116, 105, 98, 102, 111, 120, 46, 118,
		115, 99, 11, 118, 115, 99, 46, 103, 97, 116,
		101, 119, 97, 121, 232, 3, 0, 0, 0, 0,
		0, 0, 35, 32, 188, 190, 9, 116, 111, 61,
		116, 105, 98, 102, 111, 120}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpClaimAccount(t *testing.T) {
	got, _ := getTestClaimAcc().SerializeOp()
	expected := []byte{
		22, 10, 116, 101, 99, 104, 99, 111,
		100, 101, 114, 120, 0, 0, 0, 0,
		0, 0, 0, 0, 35, 32, 188, 190,
		0,
	}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}

func TestSerializeOpTransferToVesting(t *testing.T) {
	got, _ := getTestPowerUp().SerializeOp()
	expected := []byte{3, 9, 105, 110, 105, 116, 109, 105, 110, 101, 114, 10, 109, 97, 103, 105, 46, 116, 101, 115, 116, 49, 160, 134, 1, 0, 0, 0, 0, 0, 35, 32, 188, 190}
	if !bytes.Equal(got, expected) {
		t.Error("Expected", expected, "got", got)
	}
}
