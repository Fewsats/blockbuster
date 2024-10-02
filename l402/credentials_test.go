package l402_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/fewsats/blockbuster/l402"
	"github.com/stretchr/testify/require"
)

const macBase64 = "AgELZmV3c2F0cy5jb20CQgAAPm7liCfo4ClO2QCUZGJZ6P2fzmrjz9mvreU5cKQs30M8EVtMs-PbDVuhiaoTBNNg8ULIvf-89xHY-MPnE2RxZwACH2V4cGlyZXNfYXQ9MjAyNS0wOS0yN1QxNToxMzo1N1oAAixleHRlcm5hbF9pZD1mNDMzZTM1YmEzMzk0NDQxYmIxNzQ5YWFiMjFiMTdlOQAABiBXk7cYhCcslZf5ssgEym6wWNa10aUIS1R5z6H31QMXog"

func TestDecodeMacIdentifier(t *testing.T) {
	testCases := []struct {
		name          string
		id            []byte
		expectedVer   uint16
		expectedHash  [32]byte
		expectedToken [32]byte
		expectErr     string
	}{
		{
			name:          "valid version 0",
			id:            append(append([]byte{0, 0}, bytes.Repeat([]byte{0x01}, 32)...), bytes.Repeat([]byte{0x02}, 32)...),
			expectedVer:   0,
			expectedHash:  [32]byte{0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01, 0x01},
			expectedToken: [32]byte{0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02, 0x02},
			expectErr:     "",
		},
		{
			name:      "invalid version",
			id:        []byte{0x01, 0x00},
			expectErr: "unkown version",
		},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ver, hash, token, err := l402.DecodeMacIdentifier(tc.id)
			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
				return
			}

			require.NoError(t, err)

			require.Equal(t, tc.expectedVer, ver)
			require.Equal(t, tc.expectedHash, hash)
			require.Equal(t, tc.expectedToken, token)
		})
	}
}

func TestDecodeL402Credentials(t *testing.T) {
	testCase := []struct {
		name        string
		macBase64   string
		preimageHex string

		expectedVersion        uint16
		expectedPaymentHashHex string
		expectedIdentifierHex  string

		expectErr string
	}{
		// Credentials from Fewsats
		{
			name:        "valid credentials",
			macBase64:   macBase64,
			preimageHex: "5bdf3bac241faf6eacb035e0b9aa911a615e62b80bef3f91d415e561b2a4da7a",

			expectedVersion:        0,
			expectedPaymentHashHex: "3e6ee58827e8e0294ed90094646259e8fd9fce6ae3cfd9afade53970a42cdf43",
			expectedIdentifierHex:  "00003e6ee58827e8e0294ed90094646259e8fd9fce6ae3cfd9afade53970a42cdf433c115b4cb3e3db0d5ba189aa1304d360f142c8bdffbcf711d8f8c3e713647167",
		},
	}

	for _, tc := range testCase {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			creds, err := l402.DecodeL402Credentials(
				tc.macBase64, tc.preimageHex,
			)

			if tc.expectErr != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), tc.expectErr)
				return
			}

			require.NoError(t, err)

			require.NoError(t, err)
			require.Equal(t, tc.expectedVersion, creds.Version)
			require.Equal(t, tc.expectedPaymentHashHex,
				hex.EncodeToString(creds.PaymentHash[:]))
			require.Equal(t, tc.expectedIdentifierHex, creds.Identifier)
		})
	}

}

func TestValidatePreimage(t *testing.T) {
	randomCreds := &l402.Credentials{
		Preimage:    [32]byte{},
		PaymentHash: [32]byte{0x01},
	}

	err := randomCreds.ValidatePreimage()
	require.Error(t, err)
	require.Contains(t, err.Error(), "does not match payment hash")

	pre, _ := hex.DecodeString("5bdf3bac241faf6eacb035e0b9aa911a615e62b80bef3f91d415e561b2a4da7a")
	hash, _ := hex.DecodeString("35cf3da4dfdefa01a3859659d447eb2eeb070c9c6610f4faa52b1510a4c5f597")

	var preimage, paymentHash [32]byte
	copy(preimage[:], pre)
	copy(paymentHash[:], hash)

	validCreds := &l402.Credentials{
		Preimage:    preimage,
		PaymentHash: paymentHash,
	}

	err = validCreds.ValidatePreimage()
	require.NoError(t, err)
}

func TestVerifyMacaroon(t *testing.T) {
	// TODO(positiveblue): Verify macaroon including caveats.
}
