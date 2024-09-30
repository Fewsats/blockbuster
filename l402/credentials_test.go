package l402_test

import (
	"bytes"
	"encoding/hex"
	"testing"

	"github.com/fewsats/blockbuster/l402"
	"github.com/stretchr/testify/require"
)

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
		// Credentials from SULU.
		{
			name:        "valid credentials",
			macBase64:   "AgEEbHNhdAJCAADH5A9NwOPeOFA3YjkJJ4Ntb4J4XEJZBlZ6fgHPUFJEJBtl5cJmilPoo7LPe75LuoeUd0L7jIeZQAF+RIbZKfnzAAIVc2VydmljZXM9d2VhdGhlcm1hbjowAAIYd2VhdGhlcm1hbl9jYXBhYmlsaXRpZXM9AAIhd2VhdGhlcm1hbl92YWxpZF91bnRpbD0xNzE2NDI0NDk1AAAGIJJEOu4KqC0axUALWUTyHc/32mcNGzbfW1F7c7Rjft3D",
			preimageHex: "022b8a0996fa9a6b78c7615cc17de0456a0baf777e1d95fffd7db4fc460af85d",

			expectedVersion:        0,
			expectedPaymentHashHex: "c7e40f4dc0e3de38503762390927836d6f82785c425906567a7e01cf50524424",
			expectedIdentifierHex:  "1b65e5c2668a53e8a3b2cf7bbe4bba87947742fb8c879940017e4486d929f9f3",
		},
		// Credentials from Fewsats
		{
			name:        "valid credentials 2",
			macBase64:   "AgELZmV3c2F0cy5jb20CQgAANc89pN/e+gGjhZZZ1EfrLusHDJxmEPT6pSsVEKTF9ZdxM1SLOcCUuDEgBSsQaF18xcaZeUg1ucnoWv1TN3CoXwACLGZpbGVfaWQ9M2QwNDY3NWMtY2FhNi00N2FkLWJiOGItODY2YTVlZjNkZmI2AAIfZXhwaXJlc19hdD0yMDI0LTA2LTE0VDAxOjM4OjQ2WgAABiBpAqEE5i0Q4WoF/dYKryocjIhVdb3NeNrikU8xliXU6g==",
			preimageHex: "5bdf3bac241faf6eacb035e0b9aa911a615e62b80bef3f91d415e561b2a4da7a",

			expectedVersion:        0,
			expectedPaymentHashHex: "35cf3da4dfdefa01a3859659d447eb2eeb070c9c6610f4faa52b1510a4c5f597",
			expectedIdentifierHex:  "7133548b39c094b83120052b10685d7cc5c699794835b9c9e85afd533770a85f",
		},
		// Credentials from LSAT Playground
		{
			name:        "valid credentials 3",
			macBase64:   "MDAzNmxvY2F0aW9uIGh0dHBzOi8vbHNhdC1wbGF5Z3JvdW5kLmJ1Y2tvLnZlcmNlbC5hcHAKMDA5NGlkZW50aWZpZXIgMDAwMGJlMmJmMmJiZTcxNjRiMmVjODdkNWQ4NWMzNzY4MTcxYmMwMDExNzFhYzVkNWQxZGQ1ZmU0NjVhNTgwMTQ3YTU1YTRmZTFkYWMwMmQzM2Y5OGVkY2M2MTQzMTA2YmFiNGQ4MGM4MTVjMDM1ZWJjYzk2YTNmZjhjOGJkZDE1N2I0CjAwMmZzaWduYXR1cmUgE5Lj9rZ8vgiPfs456rukyEHCkcj6WMWcGVBe3bzRtIcK",
			preimageHex: "022b8a0996fa9a6b78c7615cc17de0456a0baf777e1d95fffd7db4fc460af85d",

			// lsat-js library is encoding the macaroon as ASCII as raw bytes wwhich means that unmarshalling the
			// macaroon does not work properly. The mac.id starts with "0000..." but in bytes begins with "48 48"
			// (0 char is 48 in ASCII)
			expectErr: "unkown version: 12336",
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
