package solana

import (
	"encoding/binary"
	"errors"
	"fmt"
)

// SPLTokenInstruction is the SPL Token program instruction discriminator.
type SPLTokenInstruction uint8

const (
	SPLTokenInstructionInitializeMint    SPLTokenInstruction = 0
	SPLTokenInstructionInitializeAccount SPLTokenInstruction = 1
	SPLTokenInstructionTransfer          SPLTokenInstruction = 3
	SPLTokenInstructionApprove           SPLTokenInstruction = 5
	SPLTokenInstructionMintTo            SPLTokenInstruction = 8
	SPLTokenInstructionBurn              SPLTokenInstruction = 9
)

// SPLToken provides helpers for building SPL Token and Token-2022 instructions.
type SPLToken struct{}

// NewSPLToken creates an SPL token helper.
func NewSPLToken() *SPLToken {
	return &SPLToken{}
}

// BuildTransfer builds an SPL Token transfer instruction.
//
// Instruction accounts (in order):
//
//	[0] source token account (writable)
//	[1] destination token account (writable)
//	[2] authority (signer)
//	[3+] multisig signers (optional)
//
// Data layout:
//
//	[0]    instruction discriminator (3)
//	[1:9]  amount (uint64 LE)
func (s *SPLToken) BuildTransfer(source, destination, authority PublicKey, amount uint64, signers ...PublicKey) (Instruction, error) {
	if source.Equal(destination) {
		return Instruction{}, errors.New("source and destination must differ")
	}

	accountMetas := []AccountMeta{
		{PublicKey: source, IsWritable: true, IsSigner: false},
		{PublicKey: destination, IsWritable: true, IsSigner: false},
		{PublicKey: authority, IsWritable: false, IsSigner: true},
	}
	for _, signer := range signers {
		accountMetas = append(accountMetas, AccountMeta{PublicKey: signer, IsWritable: false, IsSigner: true})
	}

	data := make([]byte, 9)
	data[0] = byte(SPLTokenInstructionTransfer)
	binary.LittleEndian.PutUint64(data[1:9], amount)

	return Instruction{
		ProgramID: mustPublicKey(TokenProgram),
		Accounts:  accountMetas,
		Data:      data,
	}, nil
}

// BuildMintTo builds an SPL Token mint-to instruction.
//
// Instruction accounts (in order):
//
//	[0] mint (writable)
//	[1] destination token account (writable)
//	[2] authority (signer)
//	[3+] multisig signers (optional)
//
// Data layout:
//
//	[0]    instruction discriminator (8)
//	[1:9]  amount (uint64 LE)
func (s *SPLToken) BuildMintTo(mint, destination, authority PublicKey, amount uint64, signers ...PublicKey) (Instruction, error) {
	accountMetas := []AccountMeta{
		{PublicKey: mint, IsWritable: true, IsSigner: false},
		{PublicKey: destination, IsWritable: true, IsSigner: false},
		{PublicKey: authority, IsWritable: false, IsSigner: true},
	}
	for _, signer := range signers {
		accountMetas = append(accountMetas, AccountMeta{PublicKey: signer, IsWritable: false, IsSigner: true})
	}

	data := make([]byte, 9)
	data[0] = byte(SPLTokenInstructionMintTo)
	binary.LittleEndian.PutUint64(data[1:9], amount)

	return Instruction{
		ProgramID: mustPublicKey(TokenProgram),
		Accounts:  accountMetas,
		Data:      data,
	}, nil
}

// BuildBurn builds an SPL Token burn instruction.
//
// Instruction accounts (in order):
//
//	[0] token account (writable)
//	[1] mint (writable)
//	[2] authority (signer)
//	[3+] multisig signers (optional)
//
// Data layout:
//
//	[0]    instruction discriminator (9)
//	[1:9]  amount (uint64 LE)
func (s *SPLToken) BuildBurn(tokenAccount, mint, authority PublicKey, amount uint64, signers ...PublicKey) (Instruction, error) {
	accountMetas := []AccountMeta{
		{PublicKey: tokenAccount, IsWritable: true, IsSigner: false},
		{PublicKey: mint, IsWritable: true, IsSigner: false},
		{PublicKey: authority, IsWritable: false, IsSigner: true},
	}
	for _, signer := range signers {
		accountMetas = append(accountMetas, AccountMeta{PublicKey: signer, IsWritable: false, IsSigner: true})
	}

	data := make([]byte, 9)
	data[0] = byte(SPLTokenInstructionBurn)
	binary.LittleEndian.PutUint64(data[1:9], amount)

	return Instruction{
		ProgramID: mustPublicKey(TokenProgram),
		Accounts:  accountMetas,
		Data:      data,
	}, nil
}

// DeriveAssociatedTokenAccount derives the Associated Token Account (ATA) PDA
// for a given wallet owner and token mint.
//
// The ATA is derived as:
//
//	FindProgramAddress(seeds=[owner, token_program, mint], program_id=ASSOCIATED_TOKEN_PROGRAM)
func (s *SPLToken) DeriveAssociatedTokenAccount(owner PublicKey, mint PublicKey) (PublicKey, uint8, error) {
	seeds := [][]byte{owner.Bytes(), mustPublicKey(TokenProgram).Bytes(), mint.Bytes()}
	return DeriveProgramAddress(seeds, mustPublicKey(AssociatedTokenProgram))
}

// Token2022 is the Token-2022 program helper (extensions such as transfer
// hooks, interest-bearing tokens, confidential transfers).
type Token2022 struct{}

// NewToken2022 creates a Token-2022 helper.
func NewToken2022() *Token2022 {
	return &Token2022{}
}

// BuildTransfer builds a Token-2022 transfer instruction using the Token-2022
// program ID instead of the original SPL Token program.
func (t *Token2022) BuildTransfer(source, destination, authority PublicKey, amount uint64, signers ...PublicKey) (Instruction, error) {
	instr, err := NewSPLToken().BuildTransfer(source, destination, authority, amount, signers...)
	if err != nil {
		return Instruction{}, err
	}
	instr.ProgramID = mustPublicKey(Token2022Program)
	return instr, nil
}

// ExtensionType is a Token-2022 extension type.
type ExtensionType uint16

const (
	ExtensionTypeTransferFee          = 0
	ExtensionTypeMemoTransfer         = 1
	ExtensionTypeConfidentialTransfer = 2
	ExtensionTypeInterestBearing      = 3
	ExtensionTypeTransferHook         = 4
)

// String returns the extension name.
func (e ExtensionType) String() string {
	names := map[ExtensionType]string{
		ExtensionTypeTransferFee:          "transfer_fee",
		ExtensionTypeMemoTransfer:         "memo_transfer",
		ExtensionTypeConfidentialTransfer: "confidential_transfer",
		ExtensionTypeInterestBearing:      "interest_bearing",
		ExtensionTypeTransferHook:         "transfer_hook",
	}
	if n, ok := names[e]; ok {
		return n
	}
	return fmt.Sprintf("extension-%d", uint16(e))
}

var _ = errors.New
