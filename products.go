package tesla

import (
	"encoding/json"
)

type Product struct {
	EnergySiteId      int64      `json:"energy_site_id,omitempty"`
	ResourceType      string     `json:"resource_type"`
	ID                Identifier `json:"id"`
	AssetSiteId       string     `json:"asset_site_id,omitempty"`
	GatewayId         string     `json:"gateway_id,omitempty"`
	WarpSiteNumber    string     `json:"warp_site_number,omitempty"`
	PercentageCharged float64    `json:"percentage_charged,omitempty"`
	BatteryType       string     `json:"battery_type,omitempty"`
	BackupCapable     bool       `json:"backup_capable,omitempty"`
	BatteryPower      int64      `json:"battery_power,omitempty"`

	c *Client

	// These are no longer returned.
	// https://github.com/teslamotors/vehicle-command/issues/215
	// EnergyLeft        float64    `json:"energy_left,omitempty"`
	// TotalPackEnergy   uint64     `json:"total_pack_energy,omitempty"`
}

type Identifier string

func (i *Identifier) UnmarshalJSON(data []byte) error {
	if n := len(data); n > 1 && data[0] == '"' && data[n-1] == '"' {
		return json.Unmarshal(data, (*string)(i))
	}

	*i = Identifier(data)

	return nil
}

// ProductResponse contains the product details from the Tesla API.
type ProductsResponse struct {
	Response []*Product `json:"response"`
	Count    int        `json:"count"`
}

// Products fetches the products associated to a Tesla account via the API.
func (c *Client) Products() ([]*Product, error) {
	productsResponse := &ProductsResponse{}
	if err := c.getJSON(c.baseURL+"/products", productsResponse); err != nil {
		return nil, err
	}
	for _, v := range productsResponse.Response {
		v.c = c
	}
	return productsResponse.Response, nil
}
