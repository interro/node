/*
 * Copyright (C) 2017 The "MysteriumNetwork/node" Authors.
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <http://www.gnu.org/licenses/>.
 */

package endpoints

import (
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	"github.com/mysteriumnetwork/node/server"
	"github.com/mysteriumnetwork/node/server/metrics"
	dto_discovery "github.com/mysteriumnetwork/node/service_discovery/dto"
	"github.com/mysteriumnetwork/node/tequilapi/utils"
)

// swagger:model ProposalsList
type proposalsRes struct {
	Proposals []proposalRes `json:"proposals"`
}

// swagger:model ServiceLocationDTO
type locationRes struct {
	// Autonomous System Number
	// example: AS00001
	ASN string `json:"asn"`

	// example: NL
	Country string `json:"country,omitempty"`

	// example: Amsterdam
	City string `json:"city,omitempty"`
}

// swagger:model ServiceDefinitionDTO
type serviceDefinitionRes struct {
	LocationOriginate locationRes `json:"locationOriginate"`
}

// swagger:model ProposalDTO
type proposalRes struct {
	// per provider unique serial number of service description provided
	// example: 5
	ID int `json:"id"`

	// provider who offers service
	// example: 0x0000000000000000000000000000000000000001
	ProviderID string `json:"providerId"`

	// type of service provider offers
	// example: openvpn
	ServiceType string `json:"serviceType"`

	// qualitative service definition
	ServiceDefinition serviceDefinitionRes `json:"serviceDefinition"`

	// Metrics of the service
	Metrics json.RawMessage `json:"metrics,omitempty"`
}

func proposalToRes(p dto_discovery.ServiceProposal) proposalRes {
	return proposalRes{
		ID:          p.ID,
		ProviderID:  p.ProviderID,
		ServiceType: p.ServiceType,
		ServiceDefinition: serviceDefinitionRes{
			LocationOriginate: locationRes{
				ASN:     p.ServiceDefinition.GetLocation().ASN,
				Country: p.ServiceDefinition.GetLocation().Country,
				City:    p.ServiceDefinition.GetLocation().City,
			},
		},
	}
}

func mapProposalsToRes(
	proposalArry []dto_discovery.ServiceProposal,
	f func(dto_discovery.ServiceProposal) proposalRes,
	metrics func(proposalRes) proposalRes,
) []proposalRes {
	proposalsResArry := make([]proposalRes, len(proposalArry))
	for i, proposal := range proposalArry {
		proposalsResArry[i] = metrics(f(proposal))
	}
	return proposalsResArry
}

type proposalsEndpoint struct {
	mysteriumClient      server.Client
	mysteriumMorqaClient metrics.QualityOracle
}

// NewProposalsEndpoint creates and returns proposal creation endpoint
func NewProposalsEndpoint(mc server.Client, morqaClient metrics.QualityOracle) *proposalsEndpoint {
	return &proposalsEndpoint{mc, morqaClient}
}

// swagger:operation GET /proposals Proposal listProposals
// ---
// summary: Returns proposals
// description: Returns list of proposals filtered by provider id
// parameters:
//   - in: query
//     name: providerId
//     description: id of provider proposals
//     example: "0x0000000000000000000000000000000000000001"
//     type: string
// responses:
//   200:
//     description: List of proposals
//     schema:
//       "$ref": "#/definitions/ProposalsList"
//   500:
//     description: Internal server error
//     schema:
//       "$ref": "#/definitions/ErrorMessageDTO"
func (pe *proposalsEndpoint) List(resp http.ResponseWriter, req *http.Request, params httprouter.Params) {
	providerID := req.URL.Query().Get("providerId")
	fetchConnectCounts := req.URL.Query().Get("fetchConnectCounts")
	proposals, err := pe.mysteriumClient.FindProposals(providerID)
	if err != nil {
		utils.SendError(resp, err, http.StatusInternalServerError)
		return
	}

	addMetricsToRes := noMetrics
	if fetchConnectCounts == "true" {
		addMetricsToRes = addMetrics(pe.mysteriumMorqaClient)
	}

	proposalsRes := proposalsRes{mapProposalsToRes(proposals, proposalToRes, addMetricsToRes)}
	utils.WriteAsJSON(proposalsRes, resp)
}

// AddRoutesForProposals attaches proposals endpoints to router
func AddRoutesForProposals(router *httprouter.Router, mc server.Client, morqaClient metrics.QualityOracle) {
	pe := NewProposalsEndpoint(mc, morqaClient)
	router.GET("/proposals", pe.List)
}

func noMetrics(p proposalRes) proposalRes { return p }

func addMetrics(mc metrics.QualityOracle) func(p proposalRes) proposalRes {
	receivedMetrics := mc.ProposalsMetrics()
	proposalsMetrics := make(map[string]json.RawMessage, len(receivedMetrics))
	var proposal struct{ ProposalID proposalRes }

	for _, m := range receivedMetrics {
		json, err := metrics.Parse(m, &proposal)
		if err != nil {
			return noMetrics
		}
		p := proposal.ProposalID
		proposalsMetrics[p.ProviderID+"-"+p.ServiceType] = json
	}

	return func(p proposalRes) proposalRes {
		if metrics, ok := proposalsMetrics[p.ProviderID+"-"+p.ServiceType]; ok {
			p.Metrics = metrics
			return p
		}
		p.Metrics = []byte("{}")
		return p
	}
}
