package types

import (
	"fmt"
	"os"

	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/tendermint/tendermint/libs/log"

	"github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/forbole/juno/v3/node/remote"

	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	mintkeeper "github.com/elesto-dao/elesto/v4/x/mint/keeper"
	minttypes "github.com/elesto-dao/elesto/v4/x/mint/types"
	"github.com/forbole/juno/v3/node/local"

	nodeconfig "github.com/forbole/juno/v3/node/config"

	banksource "github.com/elesto-dao/bdjuno/modules/bank/source"
	localbanksource "github.com/elesto-dao/bdjuno/modules/bank/source/local"
	remotebanksource "github.com/elesto-dao/bdjuno/modules/bank/source/remote"
	distrsource "github.com/elesto-dao/bdjuno/modules/distribution/source"
	localdistrsource "github.com/elesto-dao/bdjuno/modules/distribution/source/local"
	remotedistrsource "github.com/elesto-dao/bdjuno/modules/distribution/source/remote"
	govsource "github.com/elesto-dao/bdjuno/modules/gov/source"
	localgovsource "github.com/elesto-dao/bdjuno/modules/gov/source/local"
	remotegovsource "github.com/elesto-dao/bdjuno/modules/gov/source/remote"
	mintsource "github.com/elesto-dao/bdjuno/modules/mint/source"
	localmintsource "github.com/elesto-dao/bdjuno/modules/mint/source/local"
	remotemintsource "github.com/elesto-dao/bdjuno/modules/mint/source/remote"
	slashingsource "github.com/elesto-dao/bdjuno/modules/slashing/source"
	localslashingsource "github.com/elesto-dao/bdjuno/modules/slashing/source/local"
	remoteslashingsource "github.com/elesto-dao/bdjuno/modules/slashing/source/remote"
	stakingsource "github.com/elesto-dao/bdjuno/modules/staking/source"
	localstakingsource "github.com/elesto-dao/bdjuno/modules/staking/source/local"
	remotestakingsource "github.com/elesto-dao/bdjuno/modules/staking/source/remote"
)

type Sources struct {
	BankSource     banksource.Source
	DistrSource    distrsource.Source
	GovSource      govsource.Source
	MintSource     mintsource.Source
	SlashingSource slashingsource.Source
	StakingSource  stakingsource.Source
}

func BuildSources(nodeCfg nodeconfig.Config, encodingConfig *params.EncodingConfig) (*Sources, error) {
	switch cfg := nodeCfg.Details.(type) {
	case *remote.Details:
		return buildRemoteSources(cfg)
	case *local.Details:
		return buildLocalSources(cfg, encodingConfig)

	default:
		return nil, fmt.Errorf("invalid configuration type: %T", cfg)
	}
}

func buildLocalSources(cfg *local.Details, encodingConfig *params.EncodingConfig) (*Sources, error) {
	source, err := local.NewSource(cfg.Home, encodingConfig)
	if err != nil {
		return nil, err
	}

	app := simapp.NewSimApp(
		log.NewTMLogger(log.NewSyncWriter(os.Stdout)), source.StoreDB, nil, true, map[int64]bool{},
		cfg.Home, 0, simapp.MakeTestEncodingConfig(), simapp.EmptyAppOptions{},
	)
	type SimApp2 struct {
		*simapp.SimApp
		MintKeeper mintkeeper.Keeper
	}

	app2 := SimApp2{
		SimApp: app,
		MintKeeper: mintkeeper.NewKeeper(
			app.AppCodec(),
			app.GetKey(minttypes.ModuleName),
			app.GetSubspace(minttypes.ModuleName),
			app.AccountKeeper,
			app.BankKeeper,
			app.DistrKeeper,
			authtypes.FeeCollectorName,
		),
	}

	sources := &Sources{
		BankSource:     localbanksource.NewSource(source, banktypes.QueryServer(app2.BankKeeper)),
		DistrSource:    localdistrsource.NewSource(source, distrtypes.QueryServer(app2.DistrKeeper)),
		GovSource:      localgovsource.NewSource(source, govtypes.QueryServer(app2.GovKeeper)),
		MintSource:     localmintsource.NewSource(source, minttypes.QueryServer(app2.MintKeeper)),
		SlashingSource: localslashingsource.NewSource(source, slashingtypes.QueryServer(app2.SlashingKeeper)),
		StakingSource:  localstakingsource.NewSource(source, stakingkeeper.Querier{Keeper: app2.StakingKeeper}),
	}

	// Mount and initialize the stores
	err = source.MountKVStores(app, "keys")
	if err != nil {
		return nil, err
	}

	err = source.MountTransientStores(app, "tkeys")
	if err != nil {
		return nil, err
	}

	err = source.MountMemoryStores(app, "memKeys")
	if err != nil {
		return nil, err
	}

	err = source.InitStores()
	if err != nil {
		return nil, err
	}

	return sources, nil
}

func buildRemoteSources(cfg *remote.Details) (*Sources, error) {
	source, err := remote.NewSource(cfg.GRPC)
	if err != nil {
		return nil, fmt.Errorf("error while creating remote source: %s", err)
	}

	return &Sources{
		BankSource:     remotebanksource.NewSource(source, banktypes.NewQueryClient(source.GrpcConn)),
		DistrSource:    remotedistrsource.NewSource(source, distrtypes.NewQueryClient(source.GrpcConn)),
		GovSource:      remotegovsource.NewSource(source, govtypes.NewQueryClient(source.GrpcConn)),
		MintSource:     remotemintsource.NewSource(source, minttypes.NewQueryClient(source.GrpcConn)),
		SlashingSource: remoteslashingsource.NewSource(source, slashingtypes.NewQueryClient(source.GrpcConn)),
		StakingSource:  remotestakingsource.NewSource(source, stakingtypes.NewQueryClient(source.GrpcConn)),
	}, nil
}
