package l402

import (
	"bytes"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"encoding/hex"
	"fmt"

	"gopkg.in/macaroon.v2"
)

// Credentials represents the credentials for an L402 challenge in the
// Authorization header.
type Credentials struct {
	// Macaroon is the credentials for the L402 challenge in V0.
	Macaroon *macaroon.Macaroon

	// Preimage is the preimage for the payment request.
	Preimage [32]byte

	// Version is the version of the macaroon.
	Version uint16

	// PaymentHash is the payment hash of the macaroon.
	PaymentHash [32]byte

	// tokenID is the tokneID of the macaroon.
	TokenID [32]byte
}

// DecodeMacIdentifier decodes the macaroon identifier into its version,
// payment hash and user ID.
func DecodeMacIdentifier(id []byte) (uint16, [32]byte, [32]byte, error) {
	r := bytes.NewReader(id)

	var version uint16
	if err := binary.Read(r, byteOrder, &version); err != nil {
		return 0, [32]byte{}, [32]byte{}, err
	}

	switch version {
	// A version 0 identifier consists of its linked payment hash, followed
	// by the user ID.
	case 0:
		var paymentHash [32]byte
		if _, err := r.Read(paymentHash[:]); err != nil {
			return 0, [32]byte{}, [32]byte{}, err
		}
		var tokenID [32]byte
		if _, err := r.Read(tokenID[:]); err != nil {
			return 0, [32]byte{}, [32]byte{}, err
		}

		return version, paymentHash, tokenID, nil
	}

	return 0, [32]byte{}, [32]byte{}, fmt.Errorf("unkown version: %d", version)
}

// DecodeL402Credentials decodes the L402 credentials from the given encoded
// credentials from the Authorization header.
func DecodeL402Credentials(macBase64, preimageHex string) (*Credentials,
	error) {

	macBytes, err := base64.StdEncoding.DecodeString(macBase64)
	if err != nil {
		return nil, fmt.Errorf("unable to marshal macaroon: %v", err)
	}

	mac := &macaroon.Macaroon{}
	err = mac.UnmarshalBinary(macBytes)
	if err != nil {
		return nil, fmt.Errorf("invalid macaroon: %s", macBase64)
	}

	version, paymentHash, tokenID, err := DecodeMacIdentifier(mac.Id())
	if err != nil {
		return nil, fmt.Errorf("unable to decode macaroon identifier: %v", err)
	}

	// Decode te preimage.
	if len(preimageHex) != 64 {
		return nil, fmt.Errorf("invalid preimage: %s", preimageHex)
	}

	var preimage [32]byte
	pre, err := hex.DecodeString(preimageHex)
	if err != nil {
		return nil, fmt.Errorf("unable to decode preimage: %s", preimageHex)
	}
	copy(preimage[:], pre)

	return &Credentials{
		Macaroon:    mac,
		Preimage:    preimage,
		Version:     version,
		PaymentHash: paymentHash,
		TokenID:     tokenID,
	}, nil
}

// VerifyPreimage checks that the preimage matches the payment hash of the
// macaroon.
func (c *Credentials) ValidatePreimage() error {
	preimageHash := sha256.Sum256(c.Preimage[:])
	if !bytes.Equal(preimageHash[:], c.PaymentHash[:]) {
		return fmt.Errorf("preimage(%x) does not match payment hash(%x)",
			c.Preimage, c.PaymentHash)
	}

	return nil
}

// VerifyMacaroon verifies the macaroon with the given root key and checks
// that all the caveats are valid.
func (c *Credentials) VerifyMacaroon(rootKey [32]byte) error {
	_, err := c.Macaroon.VerifySignature(rootKey[:], nil)
	if err != nil {
		return fmt.Errorf("unable to verify macaroon: %v", err)
	}

	// TODO(positiveblue): Add caveat logic.

	return nil
}
