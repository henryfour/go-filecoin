package account

import (
	"github.com/ipfs/go-cid"
	"reflect"

	xerrors "github.com/pkg/errors"

	"github.com/filecoin-project/go-filecoin/internal/pkg/encoding"
	"github.com/filecoin-project/go-filecoin/internal/pkg/types"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/abi"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/actor"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/address"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/errors"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/internal/dispatch"
	"github.com/filecoin-project/go-filecoin/internal/pkg/vm/internal/runtime"
)

// Actor is the builtin actor responsible for individual accounts.
// More details on future responsibilities can be found at https://github.com/filecoin-project/specs/blob/master/spec.md#account-actor.
//
// Actor __is__ shared between multiple accounts, as it is the
// underlying code.
// TODO make singleton vs not more clear
type Actor struct{}

// NewActor creates a new account actor.
func NewActor(balance types.AttoFIL) (*actor.Actor, error) {
	return actor.NewActor(types.AccountActorCodeCid, balance), nil
}

// State is the account actors storage.
type State struct {
	// Address is a public key based address that can be used to verify signatures
	Address address.Address
}

// NewState creates a new actor state.
func NewState(addr address.Address) *State {
	return &State{Address: addr}
}

// Actor methods
const (
	Constructor types.MethodID = types.ConstructorMethodID
)

//
// ExecutableActor impl for Actor
//

// Ensure AccountActor is an ExecutableActor at compile time.
var _ dispatch.ExecutableActor = (*Actor)(nil)

// signatures are the publicly (externally callable) methods of the AccountActor.
var signatures = dispatch.Exports{
	Constructor: &dispatch.FunctionSignature{
		Params: []abi.Type{abi.Address},
		Return: []abi.Type{},
	},
}

// Method returns method definition for a given method id.
func (a *Actor) Method(id types.MethodID) (dispatch.Method, *dispatch.FunctionSignature, bool) {
	switch id {
	case Constructor:
		return reflect.ValueOf((*Impl)(a).Constructor), signatures[Constructor], true
	}
	return nil, nil, false
}

// InitializeState for account actors does nothing.
func (*Actor) InitializeState(storage runtime.LegacyStorage, initializerData interface{}) error {
	state, ok := initializerData.(*State)
	if !ok {
		return errors.NewFaultError("Initial state to account actor is not a account.State struct")
	}

	if state.Address.Protocol() != address.SECP256K1 && state.Address.Protocol() != address.BLS {
		return errors.NewRevertError("Attempt to create account actor with wrong type of address")
	}

	stateBytes, err := encoding.Encode(state)
	if err != nil {
		return xerrors.Wrap(err, "failed to cbor marshal objecinitalizerDatat")
	}

	id, err := storage.Put(stateBytes)
	if err != nil {
		return err
	}

	return storage.LegacyCommit(id, cid.Undef)
}

//
// vm methods for actor
//

// Impl is the VM implementation of the actor.
type Impl Actor

// Constructor initializes the actor's state
func (impl *Impl) Constructor(ctx runtime.InvocationContext, addr address.Address) (uint8, error) {
	err := (*Actor)(impl).InitializeState(ctx.Runtime().LegacyStorage(), NewState(addr))
	if err != nil {
		return errors.CodeError(err), errors.RevertErrorWrap(err, "Could not initialize account state")
	}
	return 0, nil
}
