// Code generated by mockery. DO NOT EDIT.

package mock_txcompressor

import (
	claimcompressor "github.com/fiwallets/zkevm-bridge-service/etherman/smartcontracts/claimcompressor"
	bind "github.com/fiwallets/go-ethereum/accounts/abi/bind"

	common "github.com/fiwallets/go-ethereum/common"

	mock "github.com/stretchr/testify/mock"

	types "github.com/fiwallets/go-ethereum/core/types"
)

// EthermanI is an autogenerated mock type for the EthermanI type
type EthermanI struct {
	mock.Mock
}

type EthermanI_Expecter struct {
	mock *mock.Mock
}

func (_m *EthermanI) EXPECT() *EthermanI_Expecter {
	return &EthermanI_Expecter{mock: &_m.Mock}
}

// CompressClaimCall provides a mock function with given fields: mainnetExitRoot, rollupExitRoot, claimData
func (_m *EthermanI) CompressClaimCall(mainnetExitRoot common.Hash, rollupExitRoot common.Hash, claimData []claimcompressor.ClaimCompressorCompressClaimCallData) ([]byte, error) {
	ret := _m.Called(mainnetExitRoot, rollupExitRoot, claimData)

	if len(ret) == 0 {
		panic("no return value specified for CompressClaimCall")
	}

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(common.Hash, common.Hash, []claimcompressor.ClaimCompressorCompressClaimCallData) ([]byte, error)); ok {
		return rf(mainnetExitRoot, rollupExitRoot, claimData)
	}
	if rf, ok := ret.Get(0).(func(common.Hash, common.Hash, []claimcompressor.ClaimCompressorCompressClaimCallData) []byte); ok {
		r0 = rf(mainnetExitRoot, rollupExitRoot, claimData)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(common.Hash, common.Hash, []claimcompressor.ClaimCompressorCompressClaimCallData) error); ok {
		r1 = rf(mainnetExitRoot, rollupExitRoot, claimData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanI_CompressClaimCall_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'CompressClaimCall'
type EthermanI_CompressClaimCall_Call struct {
	*mock.Call
}

// CompressClaimCall is a helper method to define mock.On call
//   - mainnetExitRoot common.Hash
//   - rollupExitRoot common.Hash
//   - claimData []claimcompressor.ClaimCompressorCompressClaimCallData
func (_e *EthermanI_Expecter) CompressClaimCall(mainnetExitRoot interface{}, rollupExitRoot interface{}, claimData interface{}) *EthermanI_CompressClaimCall_Call {
	return &EthermanI_CompressClaimCall_Call{Call: _e.mock.On("CompressClaimCall", mainnetExitRoot, rollupExitRoot, claimData)}
}

func (_c *EthermanI_CompressClaimCall_Call) Run(run func(mainnetExitRoot common.Hash, rollupExitRoot common.Hash, claimData []claimcompressor.ClaimCompressorCompressClaimCallData)) *EthermanI_CompressClaimCall_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(common.Hash), args[1].(common.Hash), args[2].([]claimcompressor.ClaimCompressorCompressClaimCallData))
	})
	return _c
}

func (_c *EthermanI_CompressClaimCall_Call) Return(_a0 []byte, _a1 error) *EthermanI_CompressClaimCall_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanI_CompressClaimCall_Call) RunAndReturn(run func(common.Hash, common.Hash, []claimcompressor.ClaimCompressorCompressClaimCallData) ([]byte, error)) *EthermanI_CompressClaimCall_Call {
	_c.Call.Return(run)
	return _c
}

// SendCompressedClaims provides a mock function with given fields: auth, compressedTxData
func (_m *EthermanI) SendCompressedClaims(auth *bind.TransactOpts, compressedTxData []byte) (*types.Transaction, error) {
	ret := _m.Called(auth, compressedTxData)

	if len(ret) == 0 {
		panic("no return value specified for SendCompressedClaims")
	}

	var r0 *types.Transaction
	var r1 error
	if rf, ok := ret.Get(0).(func(*bind.TransactOpts, []byte) (*types.Transaction, error)); ok {
		return rf(auth, compressedTxData)
	}
	if rf, ok := ret.Get(0).(func(*bind.TransactOpts, []byte) *types.Transaction); ok {
		r0 = rf(auth, compressedTxData)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*types.Transaction)
		}
	}

	if rf, ok := ret.Get(1).(func(*bind.TransactOpts, []byte) error); ok {
		r1 = rf(auth, compressedTxData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EthermanI_SendCompressedClaims_Call is a *mock.Call that shadows Run/Return methods with type explicit version for method 'SendCompressedClaims'
type EthermanI_SendCompressedClaims_Call struct {
	*mock.Call
}

// SendCompressedClaims is a helper method to define mock.On call
//   - auth *bind.TransactOpts
//   - compressedTxData []byte
func (_e *EthermanI_Expecter) SendCompressedClaims(auth interface{}, compressedTxData interface{}) *EthermanI_SendCompressedClaims_Call {
	return &EthermanI_SendCompressedClaims_Call{Call: _e.mock.On("SendCompressedClaims", auth, compressedTxData)}
}

func (_c *EthermanI_SendCompressedClaims_Call) Run(run func(auth *bind.TransactOpts, compressedTxData []byte)) *EthermanI_SendCompressedClaims_Call {
	_c.Call.Run(func(args mock.Arguments) {
		run(args[0].(*bind.TransactOpts), args[1].([]byte))
	})
	return _c
}

func (_c *EthermanI_SendCompressedClaims_Call) Return(_a0 *types.Transaction, _a1 error) *EthermanI_SendCompressedClaims_Call {
	_c.Call.Return(_a0, _a1)
	return _c
}

func (_c *EthermanI_SendCompressedClaims_Call) RunAndReturn(run func(*bind.TransactOpts, []byte) (*types.Transaction, error)) *EthermanI_SendCompressedClaims_Call {
	_c.Call.Return(run)
	return _c
}

// NewEthermanI creates a new instance of EthermanI. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEthermanI(t interface {
	mock.TestingT
	Cleanup(func())
}) *EthermanI {
	mock := &EthermanI{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
