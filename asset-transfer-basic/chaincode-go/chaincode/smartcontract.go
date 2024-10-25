package chaincode

import (
	"encoding/json"
	"fmt"
	"strconv"
	"github.com/hyperledger/fabric-contract-api-go/v2/contractapi"
)

// SmartContract provMpines functions for managing an Asset
type SmartContract struct {
	contractapi.Contract
}

// Asset describes basic details of what makes up a simple asset
// Insert struct field in alphabetic order => to achieve determinism across languages
// golang keeps the order when marshal to json but doesn't order automatically
type Asset struct {
	DealerID       string`json:"DealerID"`
	Msisdn          string `json:"Msisdn"`
	Mpin             string `json:"Mpin"`
	Balance          int `json:"Balance"`
	Status           string    `json:"Status"`
	TransAmount     int `json:"TransAmount"`
	TransType       string `json:"TransType"`
	Remarks       string`json:"Remarks"`
}

// InitLedger adds a base set of assets to the ledger
func (s *SmartContract) InitLedger(ctx contractapi.TransactionContextInterface) error {
	assets := []Asset{
		{DealerID: "DEALER001", Msisdn: "1234567890", Mpin: "1234", Balance: 10000, Status: "Active", TransAmount: 5000, TransType: "Credit", Remarks: "Initial deposit"},
		{DealerID: "DEALER002", Msisdn: "0987654321", Mpin: "5678", Balance: 15000, Status: "Active", TransAmount: 2000, TransType: "Debit", Remarks: "Payment for stock"},
		{DealerID: "DEALER003", Msisdn: "1122334455", Mpin: "9101", Balance: 25000, Status: "Inactive", TransAmount: 3000, TransType: "Credit", Remarks: "Refund from supplier"},
		{DealerID: "DEALER004", Msisdn: "2233445566", Mpin: "1213", Balance: 5000, Status: "Active", TransAmount: 1000, TransType: "Debit", Remarks: "Payment for delivery"},
		{DealerID: "DEALER005", Msisdn: "3344556677", Mpin: "1415", Balance: 20000, Status: "Active", TransAmount: 7000, TransType: "Credit", Remarks: "Monthly sales revenue"},
	}

	for _, asset := range assets {
		assetJSON, err := json.Marshal(asset)
		if err != nil {
			return err
		}

		err = ctx.GetStub().PutState(asset.DealerID, assetJSON)
		if err != nil {
			return fmt.Errorf("failed to put to world state. %v", err)
		}
	}

	return nil
}

// CreateAsset issues a new asset to the world state with given details.
func (s *SmartContract) CreateAsset(ctx contractapi.TransactionContextInterface, DealerID string , Mpin string, Msisdn string, Status string, Balance int,  TransAmount int, TransType string, Remarks string) error {
	exists, err := s.AssetExists(ctx, DealerID)
	if err != nil {
		return err
	}
	if exists {
		return fmt.Errorf("the asset %s already exists", DealerID)
	}

	asset := Asset{
		DealerID:   DealerID,
		Msisdn:     Msisdn,
		Mpin:       Mpin,
		Balance:    Balance,
		Status:     Status,
		TransAmount: TransAmount,
		TransType:  TransType,
		Remarks:    Remarks,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(DealerID, assetJSON)
}

// ReadAsset returns the asset stored in the world state with given Mpin.
func (s *SmartContract) ReadAsset(ctx contractapi.TransactionContextInterface, DealerID string) (*Asset, error) {
	assetJSON, err := ctx.GetStub().GetState(DealerID)
	if err != nil {
		return nil, fmt.Errorf("failed to read from world state: %v", err)
	}
	if assetJSON == nil {
		return nil, fmt.Errorf("the asset %s does not exist", DealerID)
	}

	var asset Asset
	err = json.Unmarshal(assetJSON, &asset)
	if err != nil {
		return nil, err
	}

	return &asset, nil
}

// UpdateAsset updates an existing asset in the world state with provMpined parameters.
func (s *SmartContract) UpdateAsset(ctx contractapi.TransactionContextInterface, DealerID string, Msisdn string, Status string, Balance int, Mpin string, TransAmount int, TransType string, Remarks string) error {
	exists, err := s.AssetExists(ctx, DealerID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", DealerID)
	}

	// overwriting original asset with new asset
	asset := Asset{
		DealerID:   DealerID,
		Msisdn:     Msisdn,
		Mpin:       Mpin,
		Balance:    Balance,
		Status:     Status,
		TransAmount: TransAmount,
		TransType:  TransType,
		Remarks:    Remarks,
	}
	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return err
	}

	return ctx.GetStub().PutState(DealerID, assetJSON)
}

// DeleteAsset deletes an given asset from the world state.
func (s *SmartContract) DeleteAsset(ctx contractapi.TransactionContextInterface, DealerID string) error {
	exists, err := s.AssetExists(ctx, DealerID)
	if err != nil {
		return err
	}
	if !exists {
		return fmt.Errorf("the asset %s does not exist", DealerID)
	}

	return ctx.GetStub().DelState(DealerID)
}

// AssetExists returns true when asset with given Mpin exists in world state
func (s *SmartContract) AssetExists(ctx contractapi.TransactionContextInterface, DealerID string) (bool, error) {
	assetJSON, err := ctx.GetStub().GetState(DealerID)
	if err != nil {
		return false, fmt.Errorf("failed to read from world state: %v", err)
	}

	return assetJSON != nil, nil
}

// TransferAsset updates the Balance field of asset with given Mpin in world state, and returns the old Balance.
func (s *SmartContract) TransferAsset(ctx contractapi.TransactionContextInterface, DealerID string, newBalance int) (string, error) {
	asset, err := s.ReadAsset(ctx, DealerID)
	if err != nil {
		return "", err
	}

	oldBalance := asset.Balance
	asset.Balance = newBalance

	assetJSON, err := json.Marshal(asset)
	if err != nil {
		return "", err
	}

	err = ctx.GetStub().PutState(DealerID, assetJSON)
	if err != nil {
		return "", err
	}

	return strconv.Itoa(oldBalance), nil
}

// GetAllAssets returns all assets found in world state
func (s *SmartContract) GetAllAssets(ctx contractapi.TransactionContextInterface) ([]*Asset, error) {
	// range query with empty string for startKey and endKey does an
	// open-ended query of all assets in the chaincode namespace.
	resultsIterator, err := ctx.GetStub().GetStateByRange("", "")
	if err != nil {
		return nil, err
	}
	defer resultsIterator.Close()

	var assets []*Asset
	for resultsIterator.HasNext() {
		queryResponse, err := resultsIterator.Next()
		if err != nil {
			return nil, err
		}

		var asset Asset
		err = json.Unmarshal(queryResponse.Value, &asset)
		if err != nil {
			return nil, err
		}
		assets = append(assets, &asset)
	}

	return assets, nil
}
