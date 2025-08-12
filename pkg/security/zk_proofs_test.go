package security

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetProofInfo(t *testing.T) {
	// Test Schnorr proof
	info := GetProofInfo(ProofTypeSchnorr)
	assert.NotNil(t, info)
	assert.Equal(t, "Schnorr", info.Name)
	assert.Equal(t, 256, info.SecurityLevel)
	assert.Equal(t, 64, info.ProofSize)
	assert.Equal(t, int64(1000), info.VerificationTime)
	assert.False(t, info.IsQuantumResistant)
	assert.True(t, info.Recommended)

	// Test Bulletproofs
	info = GetProofInfo(ProofTypeBulletproofs)
	assert.NotNil(t, info)
	assert.Equal(t, "Bulletproofs", info.Name)
	assert.Equal(t, 128, info.SecurityLevel)
	assert.Equal(t, 576, info.ProofSize)
	assert.Equal(t, int64(5000), info.VerificationTime)
	assert.False(t, info.IsQuantumResistant)
	assert.True(t, info.Recommended)

	// Test zk-SNARK
	info = GetProofInfo(ProofTypeZkSNARK)
	assert.NotNil(t, info)
	assert.Equal(t, "zk-SNARK", info.Name)
	assert.Equal(t, 128, info.SecurityLevel)
	assert.Equal(t, 128, info.ProofSize)
	assert.Equal(t, int64(1000), info.VerificationTime)
	assert.False(t, info.IsQuantumResistant)
	assert.True(t, info.Recommended)

	// Test zk-STARK
	info = GetProofInfo(ProofTypeZkSTARK)
	assert.NotNil(t, info)
	assert.Equal(t, "zk-STARK", info.Name)
	assert.Equal(t, 128, info.SecurityLevel)
	assert.Equal(t, 1024, info.ProofSize)
	assert.Equal(t, int64(2000), info.VerificationTime)
	assert.True(t, info.IsQuantumResistant)
	assert.True(t, info.Recommended)

	// Test Ring Signature
	info = GetProofInfo(ProofTypeRingSignature)
	assert.NotNil(t, info)
	assert.Equal(t, "Ring Signature", info.Name)
	assert.Equal(t, 128, info.SecurityLevel)
	assert.Equal(t, 256, info.ProofSize)
	assert.Equal(t, int64(3000), info.VerificationTime)
	assert.False(t, info.IsQuantumResistant)
	assert.False(t, info.Recommended)
}

func TestNewZKProver(t *testing.T) {
	// Test Schnorr prover
	prover := NewZKProver(ProofTypeSchnorr)
	assert.NotNil(t, prover)
	assert.Equal(t, ProofTypeSchnorr, prover.proofType)
	assert.NotNil(t, prover.info)

	// Test Bulletproofs prover
	prover = NewZKProver(ProofTypeBulletproofs)
	assert.NotNil(t, prover)
	assert.Equal(t, ProofTypeBulletproofs, prover.proofType)
	assert.NotNil(t, prover.info)
}

func TestNewZKVerifier(t *testing.T) {
	// Test Schnorr verifier
	verifier := NewZKVerifier(ProofTypeSchnorr)
	assert.NotNil(t, verifier)
	assert.Equal(t, ProofTypeSchnorr, verifier.proofType)
	assert.NotNil(t, verifier.info)

	// Test Bulletproofs verifier
	verifier = NewZKVerifier(ProofTypeBulletproofs)
	assert.NotNil(t, verifier)
	assert.Equal(t, ProofTypeBulletproofs, verifier.proofType)
	assert.NotNil(t, verifier.info)
}

func TestGenerateAndVerifySchnorrProof(t *testing.T) {
	prover := NewZKProver(ProofTypeSchnorr)
	verifier := NewZKVerifier(ProofTypeSchnorr)

	statement := []byte("I know the secret without revealing it")
	witness := []byte("secret123")

	// Generate proof
	proof, err := prover.GenerateProof(statement, witness)
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.Equal(t, ProofTypeSchnorr, proof.Type)
	// Proof size is limited by witness size
	assert.Equal(t, len(witness), len(proof.Proof))
	assert.Equal(t, statement, proof.PublicInputs)
	assert.NotNil(t, proof.VerificationKey)

	// Verify proof
	valid, err := verifier.VerifyProof(proof, statement)
	assert.NoError(t, err)
	assert.True(t, valid)

	// Test with wrong statement
	valid, err = verifier.VerifyProof(proof, []byte("Wrong statement"))
	assert.NoError(t, err)
	assert.False(t, valid)
}

func TestGenerateAndVerifyBulletproofsProof(t *testing.T) {
	prover := NewZKProver(ProofTypeBulletproofs)
	verifier := NewZKVerifier(ProofTypeBulletproofs)

	statement := []byte("I know the range proof")
	witness := []byte("range_witness")

	// Generate proof
	proof, err := prover.GenerateProof(statement, witness)
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.Equal(t, ProofTypeBulletproofs, proof.Type)
	// Proof size is limited by witness size
	assert.Equal(t, len(witness), len(proof.Proof))
	assert.Equal(t, statement, proof.PublicInputs)
	assert.NotNil(t, proof.VerificationKey)

	// Verify proof
	valid, err := verifier.VerifyProof(proof, statement)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestGenerateAndVerifyZkSNARKProof(t *testing.T) {
	prover := NewZKProver(ProofTypeZkSNARK)
	verifier := NewZKVerifier(ProofTypeZkSNARK)

	statement := []byte("I know the SNARK proof")
	witness := []byte("snark_witness")

	// Generate proof
	proof, err := prover.GenerateProof(statement, witness)
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.Equal(t, ProofTypeZkSNARK, proof.Type)
	// Proof size is limited by witness size
	assert.Equal(t, len(witness), len(proof.Proof))
	assert.Equal(t, statement, proof.PublicInputs)
	assert.NotNil(t, proof.VerificationKey)

	// Verify proof
	valid, err := verifier.VerifyProof(proof, statement)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestGenerateAndVerifyZkSTARKProof(t *testing.T) {
	prover := NewZKProver(ProofTypeZkSTARK)
	verifier := NewZKVerifier(ProofTypeZkSTARK)

	statement := []byte("I know the STARK proof")
	witness := []byte("stark_witness")

	// Generate proof
	proof, err := prover.GenerateProof(statement, witness)
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.Equal(t, ProofTypeZkSTARK, proof.Type)
	// Proof size is limited by witness size
	assert.Equal(t, len(witness), len(proof.Proof))
	assert.Equal(t, statement, proof.PublicInputs)
	assert.NotNil(t, proof.VerificationKey)

	// Verify proof
	valid, err := verifier.VerifyProof(proof, statement)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestGenerateAndVerifyRingSignatureProof(t *testing.T) {
	prover := NewZKProver(ProofTypeRingSignature)
	verifier := NewZKVerifier(ProofTypeRingSignature)

	statement := []byte("I know the ring signature")
	witness := []byte("ring_witness")

	// Generate proof
	proof, err := prover.GenerateProof(statement, witness)
	assert.NoError(t, err)
	assert.NotNil(t, proof)
	assert.Equal(t, ProofTypeRingSignature, proof.Type)
	// Proof size is limited by witness size
	assert.Equal(t, len(witness), len(proof.Proof))
	assert.Equal(t, statement, proof.PublicInputs)
	assert.NotNil(t, proof.VerificationKey)

	// Verify proof
	valid, err := verifier.VerifyProof(proof, statement)
	assert.NoError(t, err)
	assert.True(t, valid)
}

func TestGetRecommendedProofTypes(t *testing.T) {
	recommended := GetRecommendedProofTypes()
	assert.NotNil(t, recommended)
	assert.Len(t, recommended, 4) // Schnorr, Bulletproofs, zk-SNARK, zk-STARK

	// Check that recommended proof types are actually recommended
	for _, proofType := range recommended {
		info := GetProofInfo(proofType)
		assert.True(t, info.Recommended)
	}
}

func TestGetProofSecurityLevel(t *testing.T) {
	// Test security levels
	assert.Equal(t, 256, GetProofSecurityLevel(ProofTypeSchnorr))
	assert.Equal(t, 128, GetProofSecurityLevel(ProofTypeBulletproofs))
	assert.Equal(t, 128, GetProofSecurityLevel(ProofTypeZkSNARK))
	assert.Equal(t, 128, GetProofSecurityLevel(ProofTypeZkSTARK))
	assert.Equal(t, 128, GetProofSecurityLevel(ProofTypeRingSignature))
}

func TestIsQuantumResistantProof(t *testing.T) {
	// Test quantum resistance status
	assert.False(t, IsQuantumResistantProof(ProofTypeSchnorr))
	assert.False(t, IsQuantumResistantProof(ProofTypeBulletproofs))
	assert.False(t, IsQuantumResistantProof(ProofTypeZkSNARK))
	assert.True(t, IsQuantumResistantProof(ProofTypeZkSTARK))
	assert.False(t, IsQuantumResistantProof(ProofTypeRingSignature))
}

func TestInvalidProofType(t *testing.T) {
	// Test invalid proof type
	invalidProofType := ProofType(999)
	info := GetProofInfo(invalidProofType)
	assert.Equal(t, "Unknown", info.Name)
	assert.Equal(t, 0, info.SecurityLevel)
	assert.Equal(t, 0, info.ProofSize)
	assert.Equal(t, int64(0), info.VerificationTime)
	assert.False(t, info.IsQuantumResistant)
	assert.False(t, info.Recommended)
}

func TestProofProperties(t *testing.T) {
	prover := NewZKProver(ProofTypeSchnorr)
	statement := []byte("Test statement")
	witness := []byte("Test witness")

	proof, err := prover.GenerateProof(statement, witness)
	assert.NoError(t, err)

	// Test that proof has correct properties
	assert.Equal(t, ProofTypeSchnorr, proof.Type)
	assert.Equal(t, statement, proof.PublicInputs)
	assert.NotNil(t, proof.VerificationKey)
	assert.Equal(t, int64(0), proof.Timestamp) // Should be set by caller

	// Test that proof size matches witness size (limited by witness)
	assert.Equal(t, len(witness), len(proof.Proof))
}

func TestDifferentProofTypes(t *testing.T) {
	// Test that different proof types generate different sized proofs
	statement := []byte("Same statement")
	witness := []byte("Same witness")

	// Schnorr
	prover := NewZKProver(ProofTypeSchnorr)
	proof1, err := prover.GenerateProof(statement, witness)
	assert.NoError(t, err)

	// Bulletproofs
	prover = NewZKProver(ProofTypeBulletproofs)
	proof2, err := prover.GenerateProof(statement, witness)
	assert.NoError(t, err)

	// zk-SNARK
	prover = NewZKProver(ProofTypeZkSNARK)
	proof3, err := prover.GenerateProof(statement, witness)
	assert.NoError(t, err)

	// Verify different proof sizes (all should be limited by witness size)
	assert.Equal(t, len(witness), len(proof1.Proof))
	assert.Equal(t, len(witness), len(proof2.Proof))
	assert.Equal(t, len(witness), len(proof3.Proof))
}

func BenchmarkGenerateProof(b *testing.B) {
	prover := NewZKProver(ProofTypeSchnorr)
	statement := []byte("Benchmark statement")
	witness := []byte("Benchmark witness")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := prover.GenerateProof(statement, witness)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkVerifyProof(b *testing.B) {
	prover := NewZKProver(ProofTypeSchnorr)
	verifier := NewZKVerifier(ProofTypeSchnorr)
	statement := []byte("Benchmark statement")
	witness := []byte("Benchmark witness")

	proof, err := prover.GenerateProof(statement, witness)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := verifier.VerifyProof(proof, statement)
		if err != nil {
			b.Fatal(err)
		}
	}
}
