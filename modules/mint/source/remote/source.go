package remote

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/elesto-dao/elesto/v4/x/mint"
	minttypes "github.com/elesto-dao/elesto/v4/x/mint/types"
	"github.com/forbole/juno/v3/node/remote"

	mintsource "github.com/elesto-dao/bdjuno/modules/mint/source"
)

var (
	_ mintsource.Source = &Source{}
)

// Source implements mintsource.Source using a remote node
type Source struct {
	*remote.Source
	querier minttypes.QueryClient
}

// NewSource returns a new Source instance
func NewSource(source *remote.Source, querier minttypes.QueryClient) *Source {
	return &Source{
		Source:  source,
		querier: querier,
	}
}

// GetInflation implements mintsource.Source
func (s Source) GetInflation(height int64) (sdk.Dec, error) {

	// TODO replace with the native function in the mint module
	currentEpoch := height / mint.BlocksPerEpoch
	startSupply := sdk.NewDec(200_000_000_000_000)
	for i := int64(0); i < currentEpoch; i++ {
		startSupply = startSupply.Add(sdk.NewDec(mint.BlocksPerEpoch * minttypes.BlockInflationDistribution[i].BlockInflation))
	}
	endSupply := startSupply.Add(sdk.NewDec(mint.BlocksPerEpoch * minttypes.BlockInflationDistribution[currentEpoch].BlockInflation))
	// this is inflation between 0-1
	inflation := endSupply.Sub(startSupply).Quo(startSupply)
	// adjust
	inflation = inflation.Mul(sdk.NewDec(10_000)).TruncateDec().Quo(sdk.NewDec(100))
	return inflation, nil

}

// Params implements mintsource.Source
func (s Source) Params(height int64) (minttypes.Params, error) {
	res, err := s.querier.Params(remote.GetHeightRequestContext(s.Ctx, height), &minttypes.QueryParamsRequest{})
	if err != nil {
		return minttypes.Params{}, nil
	}

	return res.Params, nil
}
