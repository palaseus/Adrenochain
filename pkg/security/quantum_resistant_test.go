package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetAlgorithmInfo(t *testing.T) {
	// Test lattice-based algorithm
	info := GetAlgorithmInfo(AlgorithmLatticeBased)
	assert.NotNil(t, info)
	assert.Equal(t, "Lattice-Based (CRYSTALS-Kyber)", info.Name)
	assert.Equal(t, 256, info.SecurityLevel)
	assert.Equal(t, 1184, info.KeySize)
	assert.Equal(t, 2400, info.SignatureSize)
	assert.True(t, info.IsPostQuantum)
	assert.True(t, info.Recommended)

	// Test hash-based algorithm
	info = GetAlgorithmInfo(AlgorithmHashBased)
	assert.NotNil(t, info)
	assert.Equal(t, "Hash-Based (SPHINCS+)", info.Name)
	assert.Equal(t, 256, info.SecurityLevel)
	assert.Equal(t, 64, info.KeySize)
	assert.Equal(t, 8080, info.SignatureSize)
	assert.True(t, info.IsPostQuantum)
	assert.True(t, info.Recommended)

	// Test code-based algorithm
	info = GetAlgorithmInfo(AlgorithmCodeBased)
	assert.NotNil(t, info)
	assert.Equal(t, "Code-Based (Classic McEliece)", info.Name)
	assert.Equal(t, 256, info.SecurityLevel)
	assert.Equal(t, 1357824, info.KeySize)
	assert.Equal(t, 240, info.SignatureSize)
	assert.True(t, info.IsPostQuantum)
	assert.False(t, info.Recommended) // Large key size

	// Test multivariate algorithm
	info = GetAlgorithmInfo(AlgorithmMultivariate)
	assert.NotNil(t, info)
	assert.Equal(t, "Multivariate (Rainbow)", info.Name)
	assert.Equal(t, 128, info.SecurityLevel)
	assert.Equal(t, 103648, info.KeySize)
	assert.Equal(t, 66, info.SignatureSize)
	assert.True(t, info.IsPostQuantum)
	assert.False(t, info.Recommended) // Lower security level

	// Test isogeny-based algorithm
	info = GetAlgorithmInfo(AlgorithmIsogenyBased)
	assert.NotNil(t, info)
	assert.Equal(t, "Isogeny-Based (SIKE)", info.Name)
	assert.Equal(t, 128, info.SecurityLevel)
	assert.Equal(t, 751, info.KeySize)
	assert.Equal(t, 751, info.SignatureSize)
	assert.True(t, info.IsPostQuantum)
	assert.False(t, info.Recommended) // Lower security level
}

func TestNewQuantumResistantCrypto(t *testing.T) {
	// Test lattice-based crypto
	crypto := NewQuantumResistantCrypto(AlgorithmLatticeBased)
	assert.NotNil(t, crypto)
	assert.Equal(t, AlgorithmLatticeBased, crypto.algorithm)
	assert.NotNil(t, crypto.info)

	// Test hash-based crypto
	crypto = NewQuantumResistantCrypto(AlgorithmHashBased)
	assert.NotNil(t, crypto)
	assert.Equal(t, AlgorithmHashBased, crypto.algorithm)
	assert.NotNil(t, crypto.info)
}

func TestGenerateKeyPair(t *testing.T) {
	// Test lattice-based key generation
	crypto := NewQuantumResistantCrypto(AlgorithmLatticeBased)
	keyPair, err := crypto.GenerateKeyPair()
	assert.NoError(t, err)
	assert.NotNil(t, keyPair)
	assert.Equal(t, AlgorithmLatticeBased, keyPair.Algorithm)
	assert.Equal(t, 1184, len(keyPair.PrivateKey))
	assert.Equal(t, 1184, len(keyPair.PublicKey))
	assert.NotNil(t, keyPair.AlgorithmInfo)

	// Test hash-based key generation
	crypto = NewQuantumResistantCrypto(AlgorithmHashBased)
	keyPair, err = crypto.GenerateKeyPair()
	assert.NoError(t, err)
	assert.NotNil(t, keyPair)
	assert.Equal(t, AlgorithmHashBased, keyPair.Algorithm)
	assert.Equal(t, 64, len(keyPair.PrivateKey))
	assert.Equal(t, 64, len(keyPair.PublicKey))
}

func TestSignAndVerify(t *testing.T) {
	// Test lattice-based signing and verification
	crypto := NewQuantumResistantCrypto(AlgorithmLatticeBased)
	keyPair, err := crypto.GenerateKeyPair()
	assert.NoError(t, err)

	message := []byte("Hello, quantum-resistant world!")
	signature, err := crypto.Sign(keyPair.PrivateKey, message)
	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, AlgorithmLatticeBased, signature.Algorithm)
	// Signature size is limited by private key size
	assert.Equal(t, len(keyPair.PrivateKey), len(signature.Signature))

	// Verify the signature
	valid, err := crypto.Verify(signature, message)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test with wrong message
	valid, err = crypto.Verify(signature, []byte("Wrong message"))
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestHashBasedSignAndVerify(t *testing.T) {
	// Test hash-based signing and verification
	crypto := NewQuantumResistantCrypto(AlgorithmHashBased)
	keyPair, err := crypto.GenerateKeyPair()
	assert.NoError(t, err)

	message := []byte("Hash-based quantum-resistant message")
	signature, err := crypto.Sign(keyPair.PrivateKey, message)
	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, AlgorithmHashBased, signature.Algorithm)
	// Signature size is limited by private key size
	assert.Equal(t, len(keyPair.PrivateKey), len(signature.Signature))

	// Verify the signature
	valid, err := crypto.Verify(signature, message)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestCodeBasedSignAndVerify(t *testing.T) {
	// Test code-based signing and verification
	crypto := NewQuantumResistantCrypto(AlgorithmCodeBased)
	keyPair, err := crypto.GenerateKeyPair()
	assert.NoError(t, err)

	message := []byte("Code-based quantum-resistant message")
	signature, err := crypto.Sign(keyPair.PrivateKey, message)
	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, AlgorithmCodeBased, signature.Algorithm)
	assert.Equal(t, 240, len(signature.Signature))

	// Verify the signature
	valid, err := crypto.Verify(signature, message)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestMultivariateSignAndVerify(t *testing.T) {
	// Test multivariate signing and verification
	crypto := NewQuantumResistantCrypto(AlgorithmMultivariate)
	keyPair, err := crypto.GenerateKeyPair()
	assert.NoError(t, err)

	message := []byte("Multivariate quantum-resistant message")
	signature, err := crypto.Sign(keyPair.PrivateKey, message)
	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, AlgorithmMultivariate, signature.Algorithm)
	assert.Equal(t, 66, len(signature.Signature))

	// Verify the signature
	valid, err := crypto.Verify(signature, message)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestIsogenyBasedSignAndVerify(t *testing.T) {
	// Test isogeny-based signing and verification
	crypto := NewQuantumResistantCrypto(AlgorithmIsogenyBased)
	keyPair, err := crypto.GenerateKeyPair()
	assert.NoError(t, err)

	message := []byte("Isogeny-based quantum-resistant message")
	signature, err := crypto.Sign(keyPair.PrivateKey, message)
	assert.NoError(t, err)
	assert.NotNil(t, signature)
	assert.Equal(t, AlgorithmIsogenyBased, signature.Algorithm)
	assert.Equal(t, 751, len(signature.Signature))

	// Verify the signature
	valid, err := crypto.Verify(signature, message)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestGetRecommendedAlgorithms(t *testing.T) {
	recommended := GetRecommendedAlgorithms()
	assert.NotNil(t, recommended)
	assert.Len(t, recommended, 2) // Only lattice-based and hash-based are recommended

	// Check that recommended algorithms are actually recommended
	for _, algorithm := range recommended {
		info := GetAlgorithmInfo(algorithm)
		assert.True(t, info.Recommended)
	}
}

func TestGetAlgorithmSecurityLevel(t *testing.T) {
	// Test security levels
	assert.Equal(t, 256, GetAlgorithmSecurityLevel(AlgorithmLatticeBased))
	assert.Equal(t, 256, GetAlgorithmSecurityLevel(AlgorithmHashBased))
	assert.Equal(t, 256, GetAlgorithmSecurityLevel(AlgorithmCodeBased))
	assert.Equal(t, 128, GetAlgorithmSecurityLevel(AlgorithmMultivariate))
	assert.Equal(t, 128, GetAlgorithmSecurityLevel(AlgorithmIsogenyBased))
}

func TestIsPostQuantum(t *testing.T) {
	// Test post-quantum status
	assert.True(t, IsPostQuantum(AlgorithmLatticeBased))
	assert.True(t, IsPostQuantum(AlgorithmHashBased))
	assert.True(t, IsPostQuantum(AlgorithmCodeBased))
	assert.True(t, IsPostQuantum(AlgorithmMultivariate))
	assert.True(t, IsPostQuantum(AlgorithmIsogenyBased))
}

func TestInvalidAlgorithm(t *testing.T) {
	// Test invalid algorithm
	invalidAlgorithm := QuantumResistantAlgorithm(999)
	info := GetAlgorithmInfo(invalidAlgorithm)
	assert.Equal(t, "Unknown", info.Name)
	assert.Equal(t, 0, info.SecurityLevel)
	assert.Equal(t, 0, info.KeySize)
	assert.Equal(t, 0, info.SignatureSize)
	assert.False(t, info.IsPostQuantum)
	assert.False(t, info.Recommended)
}

func TestKeyPairProperties(t *testing.T) {
	crypto := NewQuantumResistantCrypto(AlgorithmLatticeBased)
	keyPair, err := crypto.GenerateKeyPair()
	assert.NoError(t, err)

	// Test that private and public keys are different
	assert.NotEqual(t, keyPair.PrivateKey, keyPair.PublicKey)

	// Test that keys have correct sizes
	assert.Equal(t, keyPair.AlgorithmInfo.KeySize, len(keyPair.PrivateKey))
	assert.Equal(t, keyPair.AlgorithmInfo.KeySize, len(keyPair.PublicKey))

	// Test that algorithm info is correct
	assert.Equal(t, AlgorithmLatticeBased, keyPair.Algorithm)
	assert.Equal(t, 256, keyPair.AlgorithmInfo.SecurityLevel)
}

func BenchmarkGenerateKeyPair(b *testing.B) {
	crypto := NewQuantumResistantCrypto(AlgorithmLatticeBased)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.GenerateKeyPair()
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkSign(b *testing.B) {
	crypto := NewQuantumResistantCrypto(AlgorithmLatticeBased)
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		b.Fatal(err)
	}

	message := []byte("Benchmark message")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.Sign(keyPair.PrivateKey, message)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerify(b *testing.B) {
	crypto := NewQuantumResistantCrypto(AlgorithmLatticeBased)
	keyPair, err := crypto.GenerateKeyPair()
	if err != nil {
		b.Fatal(err)
	}

	message := []byte("Benchmark message")
	signature, err := crypto.Sign(keyPair.PrivateKey, message)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := crypto.Verify(signature, message)
		if err != nil {
			b.Fatal(err)
		}
	}
}
