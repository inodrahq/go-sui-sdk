package tx

import "github.com/inodrahq/go-sui-sdk/txn"

// ImmOrOwned creates an ObjectArg for an immutable or owned object.
func ImmOrOwned(ref txn.ObjectRef) txn.CallArg {
	return txn.ObjectCallArg(txn.ImmOrOwnedObject(ref))
}

// Shared creates an ObjectArg for a shared object.
func Shared(ref txn.SharedObjectRef) txn.CallArg {
	return txn.ObjectCallArg(txn.SharedObject(ref))
}

// Receiving creates an ObjectArg for a receiving object.
func Receiving(ref txn.ObjectRef) txn.CallArg {
	return txn.ObjectCallArg(txn.Receiving(ref))
}
