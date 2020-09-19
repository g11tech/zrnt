package beacon

import (
	"context"
	"errors"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"

	"github.com/protolambda/zrnt/eth2/util/bls"
	. "github.com/protolambda/ztyp/view"
)

type ProposerSlashing struct {
	SignedHeader1 SignedBeaconBlockHeader
	SignedHeader2 SignedBeaconBlockHeader
}

func (a *ProposerSlashing) Deserialize(dr *codec.DecodingReader) error {
	return dr.Container(&a.SignedHeader1, &a.SignedHeader2)
}

func (p *ProposerSlashing) HashTreeRoot(hFn tree.HashFn) Root {
	return hFn.HashTreeRoot(&p.SignedHeader1, &p.SignedHeader2)
}

var ProposerSlashingType = ContainerType("ProposerSlashing", []FieldDef{
	{"header_1", SignedBeaconBlockHeaderType},
	{"header_2", SignedBeaconBlockHeaderType},
})

func (c *Phase0Config) BlockProposerSlashings() ListTypeDef {
	return ListType(ProposerSlashingType, c.MAX_PROPOSER_SLASHINGS)
}

type ProposerSlashings struct {
	Items []ProposerSlashing
	Limit uint64
}

func (a *ProposerSlashings) Deserialize(dr *codec.DecodingReader) error {
	return dr.List(func() codec.Deserializable {
		i := len(a.Items)
		a.Items = append(a.Items, ProposerSlashing{})
		return &a.Items[i]
	}, ProposerSlashingType.TypeByteLength(), a.Limit)
}

func (li *ProposerSlashings) HashTreeRoot(hFn tree.HashFn) Root {
	length := uint64(len(li.Items))
	return hFn.ComplexListHTR(func(i uint64) tree.HTR {
		if i < length {
			return &li.Items[i]
		}
		return nil
	}, length, li.Limit)
}

func (spec *Spec) ProcessProposerSlashings(ctx context.Context, epc *EpochsContext, state *BeaconStateView, ops []ProposerSlashing) error {
	for i := range ops {
		select {
		case <-ctx.Done():
			return TransitionCancelErr
		default: // Don't block.
			break
		}
		if err := spec.ProcessProposerSlashing(epc, state, &ops[i]); err != nil {
			return err
		}
	}
	return nil
}

func (spec *Spec) ProcessProposerSlashing(epc *EpochsContext, state *BeaconStateView, ps *ProposerSlashing) error {
	// Verify header slots match
	if ps.SignedHeader1.Message.Slot != ps.SignedHeader2.Message.Slot {
		return errors.New("proposer slashing requires slashing headers to have the same slot")
	}
	// Verify header proposer indices match
	if ps.SignedHeader1.Message.ProposerIndex != ps.SignedHeader2.Message.ProposerIndex {
		return errors.New("proposer slashing headers proposer-indices do not match")
	}
	proposerIndex := ps.SignedHeader1.Message.ProposerIndex
	// Verify header proposer index is valid
	if valid, err := state.IsValidIndex(proposerIndex); err != nil {
		return err
	} else if !valid {
		return errors.New("invalid proposer index")
	}
	// Verify the headers are different
	if ps.SignedHeader1.Message == ps.SignedHeader2.Message {
		return errors.New("proposer slashing requires two different headers")
	}
	currentEpoch := epc.CurrentEpoch.Epoch
	// Verify the proposer is slashable
	validators, err := state.Validators()
	if err != nil {
		return err
	}
	validator, err := validators.Validator(proposerIndex)
	if err != nil {
		return err
	}
	if slashable, err := spec.IsSlashable(validator, currentEpoch); err != nil {
		return err
	} else if !slashable {
		return errors.New("proposer slashing requires proposer to be slashable")
	}
	domain, err := state.GetDomain(spec.DOMAIN_BEACON_PROPOSER, spec.SlotToEpoch(ps.SignedHeader1.Message.Slot))
	if err != nil {
		return err
	}
	pubkey, ok := epc.PubkeyCache.Pubkey(proposerIndex)
	if !ok {
		return errors.New("could not find pubkey of proposer")
	}
	// Verify signatures
	if !bls.Verify(
		pubkey,
		ComputeSigningRoot(ps.SignedHeader1.Message.HashTreeRoot(tree.GetHashFn()), domain),
		ps.SignedHeader1.Signature) {
		return errors.New("proposer slashing header 1 has invalid BLS signature")
	}
	if !bls.Verify(
		pubkey,
		ComputeSigningRoot(ps.SignedHeader2.Message.HashTreeRoot(tree.GetHashFn()), domain),
		ps.SignedHeader2.Signature) {
		return errors.New("proposer slashing header 2 has invalid BLS signature")
	}
	return spec.SlashValidator(epc, state, proposerIndex, nil)
}
