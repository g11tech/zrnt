package beacon

import (
	"bytes"
	"github.com/protolambda/zssz"
	"github.com/protolambda/ztyp/codec"
	. "github.com/protolambda/ztyp/view"
)

type BeaconState struct {
	// Versioning
	GenesisTime           Timestamp
	GenesisValidatorsRoot Root
	Slot                  Slot
	Fork                  Fork
	// History
	LatestBlockHeader BeaconBlockHeader
	// BlockRoots is a SLOTS_PER_HISTORICAL_ROOT vector
	BlockRoots        []Root
	// StateRoots is a SLOTS_PER_HISTORICAL_ROOT vector
	StateRoots        []Root
	HistoricalRoots   HistoricalRoots
	// Eth1
	Eth1Data      Eth1Data
	Eth1DataVotes Eth1DataVotes
	DepositIndex  DepositIndex
	// Registry
	Validators ValidatorRegistry
	Balances   Balances
	// RandaoMixes is a EPOCHS_PER_HISTORICAL_VECTOR vector
	RandaoMixes []Root
	// Slashings is a EPOCHS_PER_SLASHINGS_VECTOR vector
	Slashings []Gwei
	// Attestations
	PreviousEpochAttestations PendingAttestations
	CurrentEpochAttestations  PendingAttestations
	// Finality
	JustificationBits           JustificationBits
	PreviousJustifiedCheckpoint Checkpoint
	CurrentJustifiedCheckpoint  Checkpoint
	FinalizedCheckpoint         Checkpoint
}

// Hack to make state fields consistent and verifiable without using many hardcoded indices
// A trade-off to interpret the state as tree, without generics, and access fields by index very fast.
const (
	_stateGenesisTime = iota
	_stateGenesisValidatorsRoot
	_stateSlot
	_stateFork
	_stateLatestBlockHeader
	_stateBlockRoots
	_stateStateRoots
	_stateHistoricalRoots
	_stateEth1Data
	_stateEth1DataVotes
	_stateDepositIndex
	_stateValidators
	_stateBalances
	_stateRandaoMixes
	_stateSlashings
	_statePreviousEpochAttestations
	_stateCurrentEpochAttestations
	_stateJustificationBits
	_statePreviousJustifiedCheckpoint
	_stateCurrentJustifiedCheckpoint
	_stateFinalizedCheckpoint
)

func (c *Phase0Config) BeaconState() *ContainerTypeDef {
	return ContainerType("BeaconState", []FieldDef{
		// Versioning
		{"genesis_time", Uint64Type},
		{"genesis_validators_root", RootType},
		{"slot", SlotType},
		{"fork", ForkType},
		// History
		{"latest_block_header", BeaconBlockHeaderType},
		{"block_roots", c.BatchRoots()},
		{"state_roots", c.BatchRoots()},
		{"historical_roots", c.HistoricalRoots()},
		// Eth1
		{"eth1_data", Eth1DataType},
		{"eth1_data_votes", c.Eth1DataVotes()},
		{"eth1_deposit_index", Uint64Type},
		// Registry
		{"validators", c.ValidatorsRegistry()},
		{"balances", c.RegistryBalances()},
		// Randomness
		{"randao_mixes", c.RandaoMixes()},
		// Slashings
		{"slashings", c.Slashings()}, // Per-epoch sums of slashed effective balances
		// Attestations
		{"previous_epoch_attestations", c.PendingAttestations()},
		{"current_epoch_attestations", c.PendingAttestations()},
		// Finality
		{"justification_bits", JustificationBitsType},     // Bit set for every recent justified epoch
		{"previous_justified_checkpoint", CheckpointType}, // Previous epoch snapshot
		{"current_justified_checkpoint", CheckpointType},
		{"finalized_checkpoint", CheckpointType},
	})
}

// To load a state:
//
//     state, err := beacon.AsBeaconStateView(beacon.BeaconStateType.Deserialize(reader, size))
func AsBeaconStateView(v View, err error) (*BeaconStateView, error) {
	c, err := AsContainer(v, err)
	return &BeaconStateView{c}, err
}

type BeaconStateView struct {
	*ContainerView
}

func (c *Phase0Config) NewBeaconStateView() *BeaconStateView {
	return &BeaconStateView{ContainerView: c.BeaconState().New()}
}

func (state *BeaconStateView) GenesisTime() (Timestamp, error) {
	return AsTimestamp(state.Get(_stateGenesisTime))
}

func (state *BeaconStateView) SetGenesisTime(t Timestamp) error {
	return state.Set(_stateGenesisTime, Uint64View(t))
}

func (state *BeaconStateView) GenesisValidatorsRoot() (Root, error) {
	return AsRoot(state.Get(_stateGenesisValidatorsRoot))
}

func (state *BeaconStateView) SetGenesisValidatorsRoot(r Root) error {
	rv := RootView(r)
	return state.Set(_stateGenesisValidatorsRoot, &rv)
}

func (state *BeaconStateView) Slot() (Slot, error) {
	return AsSlot(state.Get(_stateSlot))
}

func (state *BeaconStateView) SetSlot(slot Slot) error {
	return state.Set(_stateSlot, Uint64View(slot))
}

func (state *BeaconStateView) Fork() (*ForkView, error) {
	return AsFork(state.Get(_stateFork))
}

func (state *BeaconStateView) SetFork(f Fork) error {
	return state.Set(_stateFork, f.View())
}

func (state *BeaconStateView) LatestBlockHeader() (*BeaconBlockHeaderView, error) {
	return AsBeaconBlockHeader(state.Get(_stateLatestBlockHeader))
}

func (state *BeaconStateView) SetLatestBlockHeader(v *BeaconBlockHeaderView) error {
	return state.Set(_stateLatestBlockHeader, v)
}

func (state *BeaconStateView) BlockRoots() (*BatchRootsView, error) {
	return AsBatchRoots(state.Get(_stateBlockRoots))
}

func (state *BeaconStateView) StateRoots() (*BatchRootsView, error) {
	return AsBatchRoots(state.Get(_stateStateRoots))
}

func (state *BeaconStateView) HistoricalRoots() (*HistoricalRootsView, error) {
	return AsHistoricalRoots(state.Get(_stateHistoricalRoots))
}

func (state *BeaconStateView) Eth1Data() (*Eth1DataView, error) {
	return AsEth1Data(state.Get(_stateEth1Data))
}
func (state *BeaconStateView) SetEth1Data(v *Eth1DataView) error {
	return state.Set(_stateEth1Data, v)
}

func (state *BeaconStateView) Eth1DataVotes() (*Eth1DataVotesView, error) {
	return AsEth1DataVotes(state.Get(_stateEth1DataVotes))
}

func (state *BeaconStateView) DepositIndex() (DepositIndex, error) {
	return AsDepositIndex(state.Get(_stateDepositIndex))
}

func (state *BeaconStateView) IncrementDepositIndex() error {
	depIndex, err := state.DepositIndex()
	if err != nil {
		return err
	}
	return state.Set(_stateDepositIndex, Uint64View(depIndex+1))
}

func (state *BeaconStateView) Validators() (*ValidatorsRegistryView, error) {
	return AsValidatorsRegistry(state.Get(_stateValidators))
}

func (state *BeaconStateView) Balances() (*RegistryBalancesView, error) {
	return AsRegistryBalances(state.Get(_stateBalances))
}

func (state *BeaconStateView) RandaoMixes() (*RandaoMixesView, error) {
	return AsRandaoMixes(state.Get(_stateRandaoMixes))
}

func (state *BeaconStateView) SetRandaoMixes(v *RandaoMixesView) error {
	return state.Set(_stateRandaoMixes, v)
}

func (state *BeaconStateView) Slashings() (*SlashingsView, error) {
	return AsSlashings(state.Get(_stateSlashings))
}

func (state *BeaconStateView) PreviousEpochAttestations() (*PendingAttestationsView, error) {
	return AsPendingAttestations(state.Get(_statePreviousEpochAttestations))
}

func (state *BeaconStateView) CurrentEpochAttestations() (*PendingAttestationsView, error) {
	return AsPendingAttestations(state.Get(_stateCurrentEpochAttestations))
}

func (state *BeaconStateView) JustificationBits() (*JustificationBitsView, error) {
	return AsJustificationBits(state.Get(_stateJustificationBits))
}

func (state *BeaconStateView) PreviousJustifiedCheckpoint() (*CheckpointView, error) {
	return AsCheckPoint(state.Get(_statePreviousJustifiedCheckpoint))
}

func (state *BeaconStateView) CurrentJustifiedCheckpoint() (*CheckpointView, error) {
	return AsCheckPoint(state.Get(_stateCurrentJustifiedCheckpoint))
}

func (state *BeaconStateView) FinalizedCheckpoint() (*CheckpointView, error) {
	return AsCheckPoint(state.Get(_stateFinalizedCheckpoint))
}

func (state *BeaconStateView) IsValidIndex(index ValidatorIndex) (bool, error) {
	vals, err := state.Validators()
	if err != nil {
		return false, err
	}
	count, err := vals.Length()
	if err != nil {
		return false, err
	}
	return uint64(index) < count, nil
}

// Raw converts the tree-structured state into a flattened native Go structure.
func (state *BeaconStateView) Raw() (*BeaconState, error) {
	var buf bytes.Buffer
	if err := state.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
		return nil, err
	}
	var raw BeaconState
	err := zssz.Decode(bytes.NewReader(buf.Bytes()), uint64(len(buf.Bytes())), &raw, BeaconStateSSZ)
	if err != nil {
		return nil, err
	}
	return &raw, nil
}
