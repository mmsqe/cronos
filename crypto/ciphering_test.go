package crypto

import (
	"encoding/hex"
	"fmt"
	"testing"

	"github.com/btcsuite/btcd/btcec/v2"
	"github.com/cometbft/cometbft/crypto/secp256k1"
	"github.com/stretchr/testify/assert"
)

func TestEncryptAndDecrypt(t *testing.T) {
	secret := []byte("secret")
	privateKey := secp256k1.GenPrivKeySecp256k1(secret)
	priv, _ := btcec.PrivKeyFromBytes(privateKey)
	publicKey := priv.PubKey()
	message := []byte("Hello, this is a test message.")
	encryptedMessage, err := Encrypt(publicKey, message)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Encrypted Message:", hex.EncodeToString(encryptedMessage))
	decryptedMessage, err := Decrypt(priv, encryptedMessage)
	if !assert.NoError(t, err) {
		return
	}
	fmt.Println("Decrypted Message:", string(decryptedMessage))
	assert.Equal(t, string(message), string(decryptedMessage))
}
