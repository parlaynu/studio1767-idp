package auth

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"time"
)

type stateData struct {
	Stamp       time.Time `json:"stamp"`
	Nonce       string    `json:"nonce"`
	RedirectURL string    `json:"redirect_url"`
	Encoded     string    `json:"encoded,omitempty"`
}

func newStateEncoder(encKey []byte) *stateEncoder {
	s := stateEncoder{
		key: encKey,
	}

	return &s
}

type stateEncoder struct {
	key []byte
}

// ============================================================================

func (se *stateEncoder) encode(redirectURL string) (*stateData, error) {

	// setup the encryption cipher
	block, err := aes.NewCipher(se.key)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to create aes cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to create gcm cipher: %w", err)
	}

	// generate a nonce
	nonce := make([]byte, aesgcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("auth/state: failed to read random bits: %w", err)
	}
	b64nonce := base64.RawStdEncoding.EncodeToString(nonce)

	// construct the state data struct
	sd := stateData{}
	sd.Stamp = time.Now()
	sd.Nonce = b64nonce
	sd.RedirectURL = redirectURL

	// encode as a json string
	jstate, err := json.Marshal(sd)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to marshal state data: %w", err)
	}

	// encrypt
	estate := aesgcm.Seal(nil, nonce, jstate, nil)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to seal the state: %w", err)
	}

	// encode as a URL friendly string
	sd.Encoded = base64.RawURLEncoding.EncodeToString(estate)

	return &sd, nil
}

// ============================================================================

func (se *stateEncoder) decode(bstate, b64nonce string) (*stateData, error) {

	// setup the encryption cipher
	block, err := aes.NewCipher(se.key)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to create aes cipher: %w", err)
	}

	aesgcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to create gcm cipher: %w", err)
	}

	// decode the nonce
	nonce, err := base64.RawStdEncoding.DecodeString(b64nonce)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to decode nonce: %w", err)
	}

	// decode the state from base64
	estate, err := base64.RawURLEncoding.DecodeString(bstate)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to decode state: %w", err)
	}

	// decrypt to get the json back
	jstate, err := aesgcm.Open(nil, nonce, estate, nil)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to decrypt state: %w", err)
	}

	// unmarshal the json back to the struct
	sd := stateData{}
	err = json.Unmarshal(jstate, &sd)
	if err != nil {
		return nil, fmt.Errorf("auth/state: failed to unmarshal state: %w", err)
	}

	// put the encoded data into the structure
	sd.Encoded = bstate

	// make sure the nonces match
	if b64nonce != sd.Nonce {
		// this should never happen... just here for a sanity check during
		//   development
		return nil, fmt.Errorf("auth/state: token nonce is invalid")
	}

	// return the struct
	return &sd, nil
}
