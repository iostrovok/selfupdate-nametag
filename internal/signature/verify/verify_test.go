package verify_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"nametag/internal/signature/verify"
)

func TestCommonSyntax(t *testing.T) {
	assert.Nil(t, nil, "Common syntax error")
}

func Test_Verify_1(t *testing.T) {
	sign := "YHetKjxJG06JrgYLHLdYCLr3PzSZmRnD6Jos4H_XH5o="
	fileSign := "ckhg1Wfbhck2hFxU7hmxN0x8GD1xP2QqGp_C0tWvYa22T81_97jAuycwZDwixhFGTqCG4vSb9f02hW2aa1VDBNhxJds6Vp_sUii0M-hZrEpUfklJ6jBIZtnD6r0agngsmOfZwaCeWmmneydcNdKMNh-KOBxwegbBy6JrVF0wkqiU8sVac7fpuhJN0G8dPvECTKC9hVqse6coaWlg2ia4O77Jf-UJgi8q3Qc_lyQBsPsdo7teQj77PIr2wTd889SmJCsDaweWAKBl7oPXGgo55Czk_UdHE-hQj98u3xxdcnpc970NR73HWGdoJfXnVNbYGgR6qnnAyFiNBrMsexw4KQ=="

	v, err := verify.New()
	assert.Nil(t, err, "new verify")

	err = v.Verify(sign, fileSign)
	assert.Nil(t, err, "Verify")
}

func Test_Verify_2(t *testing.T) {
	sign := "TJka38g4fFiHYalqPdjaBXp5ChPji7-GVFrqr4UxB_4="
	fileSign := "OIit71Cv213APiiN5aBpK8eTJ1jqSgARbpkzT_DqfAKiSS-tT1ChcFjJ4oD1eb4-PC6uchnPYCiIogS0xv-UQPNiKImNfGf1gsJkRJ2srXznmxeHAtCePBjPWgXkKVWD39rl7yjOzQKarHTBugZLK539XpcIXgYM7s74RxdAyPcInJG0sxIDWdOH-KitFMzhI7kihXgD7Hxtct15cbseZrcEyGKYfieIh8nMuzosjAHXDpdEqGMG6aJbszuBaW4YE_gH3FQdPCv3zcfpQc7febwXk6oYCPycah7XrpD7tDiTwy8HhZObG8MXmq7Zd3tes12dlTWfUA07Fv9foWQAUA=="

	v, err := verify.New()
	assert.Nil(t, err, "new verify")

	err = v.Verify(sign, fileSign)
	assert.Nil(t, err, "Verify")
}
